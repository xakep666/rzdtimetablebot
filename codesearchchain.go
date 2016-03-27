package main

import (
    "github.com/go-telegram-bot-api/telegram-bot-api"
    "fmt"
    "sort"
)

func recvNodeNameSearchCode(bot *tgbotapi.BotAPI, update tgbotapi.Update, data []interface{}) (newfn sorterFunc,newdata []interface{}) {
    myLogf(LogDebug,"Handling node selection request to search code for %s [%d]",update.Message.Chat.FirstName,update.Message.Chat.ID)
    nname:=update.Message.Text
    node,found:=nodeByName(nname)
    if !found {
        msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"Такой узел не зарегистрирован")
        msg.ReplyMarkup=tgbotapi.ReplyKeyboardHide{HideKeyboard:true}
        bot.Send(msg)
        newfn,_=askCommand(bot,update,nil)
        return
    }
    newdata=append(newdata,node)
    newfn=recvSearchSubstr
    msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"Введите строку для поиска")
    msg.ReplyMarkup=tgbotapi.ReplyKeyboardHide{HideKeyboard:true}
    bot.Send(msg)
    return
}

func recvSearchSubstr(bot *tgbotapi.BotAPI, update tgbotapi.Update, data []interface{}) (newfn sorterFunc,newdata []interface{}) {
    myLogf(LogDebug,"Handling search substring to search code for %s [%d]",update.Message.Chat.FirstName,update.Message.Chat.ID)
    if len(data)<=0 {
        myLogf(LogError,"recvSearchSubstr must have node argument %d [%d]",update.Message.Chat.FirstName,update.Message.Chat.ID)
        msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"Ошибка функции")
        bot.Send(msg)
        newfn,_=askCommand(bot,update,nil)
        return
    }
    node,isnode:=data[0].(Node)
    if !isnode {
        myLogf(LogError,"recvSearchSubstr must have node argument %d [%d]",update.Message.Chat.FirstName,update.Message.Chat.ID)
        msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"Ошибка функции")
        bot.Send(msg)
        newfn,_=askCommand(bot,update,nil)
        return
    }
    results:=fuzzySearchStation(node,update.Message.Text)
    msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"")
    msg.Text+="Формат\nКодСтанции|НазваниеСтанции\n"
    var csps CodeStationPairs
    for _,v:=range results {
        csps=append(csps,CodeStationPair{Code:v,Name:node.stations[v]})
    }
    sort.Sort(csps)
    for _,v:=range csps {
        msg.Text+=fmt.Sprintf("%d|%s\n",v.Code,v.Name)
    }
    bot.Send(msg)
    newfn,_=askCommand(bot,update,nil)
    return
}