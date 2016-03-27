package main

import (
    "github.com/go-telegram-bot-api/telegram-bot-api"
    "strconv"
    "fmt"
)

func recvStationToDel(bot *tgbotapi.BotAPI, update tgbotapi.Update, data []interface{}) (newfn sorterFunc,newdata []interface{}) {
    myLogf(LogDebug,"Handling removing station for %s [%d]",update.Message.Chat.FirstName,update.Message.Chat.ID)
    code,err:=strconv.Atoi(update.Message.Text)
    msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"")
    msg.ReplyMarkup=tgbotapi.ReplyKeyboardHide{HideKeyboard:true}
    if err!=nil {
        msg.Text="Введено не число"
        bot.Send(msg)
        newfn,_=askCommand(bot,update,nil)
        return
    }
    err=settings.dbinterface.DelStation(code,update.Message.Chat.ID)
    if err!=nil {
        myLogf(LogInfo,"Cannot delete code %d from saved for user %s",code,update.Message.Chat.FirstName)
        msg.Text="Ошибка удаления (код введен верно?)"
        bot.Send(msg)
        newfn,_=askCommand(bot,update,nil)
        return
    }
    msg.Text=fmt.Sprintf("Станция с кодом %d удалена из сохраненных",code)
    bot.Send(msg)
    newfn,_=askCommand(bot,update,nil)
    return
}