package main

import (
    "github.com/yhat/scrape"
    "net/http"
    "time"
    "fmt"
    "golang.org/x/net/html"
    "golang.org/x/net/html/atom"
    "strconv"
)

/* Структура таблицы
Пустое_поле Номер_маршрута* Время_прибытия* Время_отправления* Сообщение(<откуда> - <куда>)* Остановки Периодичность Перевозчик Фактическое_движение*
Используем поля, отмеченные *
*/

type TimeTableRow struct{
    RouteCode int
    ArrivalTime,DepartTime *time.Time
    FromTo string
    FactMovement string
}

type StationTimeTableEntry struct {
    Direction string
    TimeTable []TimeTableRow
}

type StationTimeTable []StationTimeTableEntry

func getDirections(root *html.Node) []string {
    //ищем названия направалений (находится в теге <a> с классом j-station-toggler)
    directionmatcher:=func(n *html.Node) bool {
        if n.DataAtom==atom.A && scrape.ByClass("j-station-toggler")(n) {
            return true
        }
        return false
    }
    dirs:=[]string{}
    for _,v:=range scrape.FindAll(root,directionmatcher) {
        dirs=append(dirs,scrape.Text(v))
    }
    return dirs
}

func getTables(root *html.Node) [][]TimeTableRow {
    ttrs:=[][]TimeTableRow{}
    now:=time.Now()
    //ищем таблицы, находящуюся внутри <div> с классом trlist 
    tablematcher:=func(n *html.Node) bool {
        if n.DataAtom==atom.Table && n.Parent!=nil && scrape.ByClass("trlist")(n.Parent) {
            return true
        }
        return false
    }
    //не берем ячейки из thead
    cellmatcher:=func(n *html.Node) bool {
        if n.DataAtom==atom.Tr && n.Parent!=nil && n.Parent.DataAtom!=atom.Thead {
            return true
        }
        return false
    }
    for _,table:=range scrape.FindAll(root,tablematcher) {
        tabletoappend:=[]TimeTableRow{}
        //строки
        for _,row:=range scrape.FindAll(table,cellmatcher) {
            //берем только нужные ячейки
            cells:=scrape.FindAll(row,scrape.ByTag(atom.Td))
            if len(cells)<9 {
                continue
            }
            
            routecodecell:=cells[1]
            arrtimecell:=cells[2]
            deptimecell:=cells[3]
            fromtocell:=cells[4]
            factmovecell:=cells[8]
            
            routecode,err:=strconv.Atoi(scrape.Text(routecodecell))
            if err!=nil {
                myLogf(LogError,"Error on parsing route code, %s\n",scrape.Text(routecodecell))
                continue
            }
            arrtimetext:=scrape.Text(arrtimecell)
            var truearrtime *time.Time
            if arrtimetext!="-" {
                arrtime,err:=time.Parse("15:04",arrtimetext)
                truearrtime=new(time.Time)
                *truearrtime=time.Date(now.Year(),now.Month(),now.Day(),arrtime.Hour(),arrtime.Minute(),0,0,now.Location())
                if err!=nil {
                    myLogf(LogError,"Error on parsing arrival time, %s\n",scrape.Text(arrtimecell))
                    continue
                }
            }
            var truedeptime *time.Time
            deptimetext:=scrape.Text(deptimecell)
            if deptimetext!="-"{
                deptime,err:=time.Parse("15:04",deptimetext)
                truedeptime=new(time.Time)
                *truedeptime=time.Date(now.Year(),now.Month(),now.Day(),deptime.Hour(),deptime.Minute(),0,0,now.Location())            
                if err!=nil {
                    myLogf(LogError,"Error on parsing departure time, %s\n",scrape.Text(deptimecell))
                    continue
                }
            }
            fromto:=scrape.Text(fromtocell)
            factmove:=scrape.Text(factmovecell)
            tabletoappend=append(tabletoappend,TimeTableRow{RouteCode: routecode,ArrivalTime: truearrtime, DepartTime: truedeptime, FromTo: fromto, FactMovement: factmove})
        }
        ttrs=append(ttrs,tabletoappend)
    }
    return ttrs
}

func postProcessForEndStation(stt *StationTimeTable,name string) {
    for i:=0;i<len(*stt);i++ {
        if (*stt)[i].Direction==name {
            (*stt)[i].Direction="ПРИБЫТИЕ"
            return
        }
    }
    finishing:=[]TimeTableRow{}
    for i:=0;i<len(*stt);i++ {
        for j:=0;j<len((*stt)[i].TimeTable); {
            if (*stt)[i].TimeTable[j].DepartTime==nil {
                finishing=append(finishing,(*stt)[i].TimeTable[j])
                copy((*stt)[i].TimeTable[j:],(*stt)[i].TimeTable[j+1:])
                (*stt)[i].TimeTable[len((*stt)[i].TimeTable)-1]=TimeTableRow{}
                (*stt)[i].TimeTable=(*stt)[i].TimeTable[:len((*stt)[i].TimeTable)-1]
                continue
            }
            j++
        }
    }
    if len(finishing)>0 {
        *stt=append(*stt,StationTimeTableEntry{Direction:"ПРИБЫТИЕ",TimeTable:finishing})
        myLogf(LogDebug,"Found finishing routes at %s, adding \"Arrived\" direction",name)
    }
}

func DownloadStationTimeTable(stationcode int) (StationTimeTable,error) {
    myLogf(LogDebug,"Getting timetable for station %s\n",stationByCode(stationcode))
    stts:=StationTimeTable{}
    today:=time.Now().Format("02.01.2006") //<день>.<месяц>.<год>
    resp,err:=http.Get(fmt.Sprintf(TimeTableQueryUrlFormat,today,stationcode))
    if err!=nil {
        return stts,err
    }
    defer resp.Body.Close()
    myLogf(LogDebug,"Downloaded raw html for station %s\n",stationByCode(stationcode))
    root,err:=html.Parse(resp.Body)
    if err!=nil {
        return stts,err
    }
    myLogf(LogDebug,"Parsed html to tree for station %s\n",stationByCode(stationcode))
    directions:=getDirections(root)
    myLogf(LogDebug,"Found %d directions for station %s\n",len(directions),stationByCode(stationcode))
    tables:=getTables(root)
    myLogf(LogDebug,"Found %d tables for station %s\n",len(tables),stationByCode(stationcode))
    if len(directions)!=len(tables) {
        return stts,fmt.Errorf("Mismatched number of directions (%d) and tables(%d) for station %s",
        len(directions),len(tables),stationByCode(stationcode))
    }
    for i:=0;i<len(directions);i++ {
        stts=append(stts,StationTimeTableEntry{Direction:directions[i],TimeTable:tables[i]})
    }
    postProcessForEndStation(&stts,stationByCode(stationcode))
    return stts,nil
}