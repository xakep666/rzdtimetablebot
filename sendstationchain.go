package main

import (
    "github.com/go-telegram-bot-api/telegram-bot-api"
    "fmt"
    
)

func sendStations(bot *tgbotapi.BotAPI, update tgbotapi.Update, data []interface{}) (newfn sorterFunc,newdata []interface{}) {
    myLogf(LogDebug,"Handling send station request for %s [%d]",update.Message.Chat.FirstName,update.Message.Chat.ID)
    text:="Формат\nКодСтанции|НазваниеСтанции|НазваниеУзла\n"
    chinfo,err:=settings.dbinterface.GetUserStations(update.Message.Chat.ID)
    msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"")
    msg.ReplyMarkup=tgbotapi.ReplyKeyboardHide{HideKeyboard:true}
    if err!=nil {
        myLog(LogError,err)
        msg.Text="Ошибка извлечения из базы"
        bot.Send(msg)
        return
    }
    for _,v:=range chinfo.Stations {
        node,_:=nodeByName(v.Node)
        text+=fmt.Sprintf("%d|%s|%s\n",v.Station,node.stations[v.Station],v.Node)
    }
    msg.Text=text
    bot.Send(msg)
    return
}