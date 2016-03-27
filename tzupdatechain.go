package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
)

func recvTZ(bot *tgbotapi.BotAPI, update tgbotapi.Update, data []interface{}) (newfn sorterFunc, newdata []interface{}) {
	myLogf(LogDebug, "Handling timezone setting for %s [%d]", update.Message.Chat.FirstName, update.Message.Chat.ID)
	offset, err := strconv.ParseFloat(update.Message.Text,64)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
    if err!=nil || !isValidTZOffset(offset) {
        newfn=recvTZ
        msg.Text="Введено неверное количество часов для временной зоны (должно быть целым или полуцелым)! "+
                 "Введите верное число, иначе не сможете пользоваться ботом!"
        bot.Send(msg)
        return
    }
    err=settings.dbinterface.SetTimeZone(update.Message.Chat.ID,offset)
    if err!=nil {
        myLogf(LogError,"Update timezone for %s [%d] failed (%s)",update.Message.Chat.ID,update.Message.Chat.FirstName,err.Error())
        newfn=recvTZ
        msg.Text="Ошибка базы данных, попробуйте ввести ещё раз"
        bot.Send(msg)
        return
    }
    msg.Text=fmt.Sprintf("Временная зона %+f установлена",offset)
    bot.Send(msg)
    newfn,_=askCommand(bot,update,nil)
    return
}