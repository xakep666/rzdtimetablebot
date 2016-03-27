package main

import (
    "errors"
)

type DBUserInfo struct {
    ID int
    Stations []NSPair
}

type NSPair struct {
    Station int
    Node string
}

type DBInterface interface{
    AddStation(code int,node string, user int) error //название узла для удобства вывода
    DelStation(code int,user int) error
    GetUserStations(user int) (DBUserInfo,error)
    Close() error
}

func DBInit() (DBInterface,error) {
    switch settings.DBType {
        case DBEmbedded:
            return BoltInit()
        default:
            return nil,errors.New("Invalid databases type in DBInit")
    }   
}

func savedStationCodes(user int) (ret []int,err error) {
    chinfo,err:=settings.dbinterface.GetUserStations(user)
    if err!=nil {
        return
    }
    for _,v:=range chinfo.Stations {
        ret=append(ret,v.Station)
    }
    return
}