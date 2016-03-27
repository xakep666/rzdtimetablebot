package main

import (
	"math"
    "sync"
	"github.com/go-telegram-bot-api/telegram-bot-api"
    "strconv"
)

type sorterFunc func(*tgbotapi.BotAPI, tgbotapi.Update, []interface{}) (sorterFunc, []interface{})

type sorterStruct struct {
	fn   sorterFunc
	data []interface{}
}

type msgSorter map[int]sorterStruct

func botRequestHandler(bot *tgbotapi.BotAPI) {
	myLog(LogInfo, "Initialized request handler")
	sorter:=make(msgSorter)
    u:=tgbotapi.NewUpdate(0)
    u.Timeout=60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		myLog(LogError, err)
		return
	}
    var mu sync.Mutex
    for update := range updates {
		go func(bot *tgbotapi.BotAPI, update tgbotapi.Update, sorter *msgSorter) {
            chatsorter := (*sorter)[update.Message.Chat.ID]
            if chatsorter.fn==nil || update.Message.Text=="/cancel" {
                chatsorter.fn=askCommand
                chatsorter.data=nil
            }
			newfn, newdata := chatsorter.fn(bot, update, chatsorter.data)
			mu.Lock() //для синхронизации записи (map позволяет чтение из множества потоков, но запись только из одного)
            (*sorter)[update.Message.Chat.ID] = sorterStruct{fn: newfn, data: newdata}
            mu.Unlock()
		}(bot, update, &sorter)
	}
}

func askCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update, data []interface{}) (newfn sorterFunc,newdata []interface{}) {
    myLogf(LogDebug,"Asking for command %s [%d]",update.Message.Chat.FirstName,update.Message.Chat.ID)
    msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"Выберете команду из списка")
    msg.ReplyMarkup=tgbotapi.ReplyKeyboardHide{HideKeyboard:true}
    bot.Send(msg)
    newfn=recvCommand
    return
}

func recvCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update, data []interface{}) (newfn sorterFunc,newdata []interface{}) {
    myLogf(LogDebug,"Handling command request for %s [%d]",update.Message.Chat.FirstName,update.Message.Chat.ID)
    switch update.Message.Text {
    case "/start": {
        msg:=tgbotapi.NewMessage(update.Message.Chat.ID,StartMessage)
        bot.Send(msg)
    }
    case "/searchcode":
        msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"Выберете узел")
        msg.ReplyMarkup=kbdMarkupAligner(nodeNames())
        bot.Send(msg)
        //что обрабатываем в следующий раз (принимаем имя узла)
        newfn=recvNodeNameSearchCode
    case "/addstation":
        msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"Выберете узел")
        msg.ReplyMarkup=kbdMarkupAligner(nodeNames())
        bot.Send(msg)
        newfn=recvNodeNameSearchSave
    case "/liststations":
        sendStations(bot,update,nil)
        newfn,_=askCommand(bot,update,nil)
    case "/delstation":
        sendStations(bot,update,nil)
        msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"")
        savedstationscodes,err:=savedStationCodes(update.Message.Chat.ID)
        savedstations:=[]string{}
        for _,v:=range savedstationscodes {savedstations=append(savedstations,strconv.Itoa(v))}
        if err!=nil {
            myLogf(LogError,"Cannot extract data from base for user %s (%s)",update.Message.Chat.FirstName,err.Error())
            msg.Text="Ошибка извлечения сохраненных станций из базы"
        } else {
            msg.Text="Выберете код станции, которую нужно удалить"
            msg.ReplyMarkup=kbdMarkupAligner(savedstations)
            newfn=recvStationToDel
        }
        bot.Send(msg)
    case "/showline":
        msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"Выберете узел")
        msg.ReplyMarkup=kbdMarkupAligner(nodeNames())
        bot.Send(msg)
        newfn=recvNodeNameShowLine
    case "/timetablebycode":
        msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"Введите код станции, по которой нужно узнать расписание\n"+
                                "(найти можно через /searchcode или /showline)")
        bot.Send(msg)
        newfn=recvStationCodeTimeTable
   case "/timetablefromsaved":
        sendStations(bot,update,nil)
        msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"")
        userstations,err:=settings.dbinterface.GetUserStations(update.Message.Chat.ID)
        if err!=nil {
            msg.Text="Ошибка функции"
            myLogf(LogError,"Error getting user stations for %s [%d]: %s",
                update.Message.Chat.FirstName,update.Message.Chat.ID,err.Error())
            bot.Send(msg)
            newfn,_=askCommand(bot,update,nil)
        } else {
            msg.Text="Выберете код нужной станции"
            stringcodes:=[]string{}
            for _,v:=range userstations.Stations {
                stringcodes=append(stringcodes,strconv.Itoa(v.Station))
            }
            msg.ReplyMarkup=kbdMarkupAligner(stringcodes)
            bot.Send(msg)
            newfn=recvStationCodeTimeTable
        }
    default:
        msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"Комманда не распознана")
        bot.Send(msg)
        newfn,_=askCommand(bot,update,nil)                   
    }
    return
}

func kbdMarkupAligner(indata []string) (ret tgbotapi.ReplyKeyboardMarkup) {
	arrsize := int(math.Ceil(math.Sqrt(float64(len(indata)))))
    ret.Keyboard=make([][]string,arrsize)
	for i := 0; i < arrsize; i++ {
        ret.Keyboard[i]=make([]string,arrsize)
		for j := 0; j < arrsize; j++ {
			if i*arrsize+j < len(indata) {
				ret.Keyboard[i][j] = indata[i*arrsize+j]
			}
		}
	}
	ret.OneTimeKeyboard = true
	ret.ResizeKeyboard = true
	return
}