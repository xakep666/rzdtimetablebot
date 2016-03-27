package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
)

func recvNodeNameSearchSave(bot *tgbotapi.BotAPI, update tgbotapi.Update, data []interface{}) (newfn sorterFunc, newdata []interface{}) {
	myLogf(LogDebug, "Handling node selection request to save code for %s [%d]", update.Message.Chat.FirstName, update.Message.Chat.ID)
	nname := update.Message.Text
	node, found := nodeByName(nname)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardHide{HideKeyboard: true}
	if !found {
		msg.Text = "Такой узел не зарегистрирован"
		bot.Send(msg)
		newfn, _ = askCommand(bot, update, nil)
		return
	}
	newdata = append(newdata, node)
	msg.Text = "Введите код станции для сохранения (найти можно через /searchcode или /showline)"
	bot.Send(msg)
	newfn = recvSaveCode
	return
}

func recvSaveCode(bot *tgbotapi.BotAPI, update tgbotapi.Update, data []interface{}) (newfn sorterFunc, newdata []interface{}) {
	myLogf(LogDebug, "Handling search substring to save code for %s [%d]", update.Message.Chat.FirstName, update.Message.Chat.ID)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	if len(data) <= 0 {
		myLogf(LogError, "recvSaveCode must have node argument %d [%d]", update.Message.Chat.FirstName, update.Message.Chat.ID)
		msg.Text = "Ошибка функции"
		bot.Send(msg)
		newfn, _ = askCommand(bot, update, nil)
		return
	}
	node, isnode := data[0].(Node)
	if !isnode {
		myLogf(LogError, "recvSaveCode must have node argument %d [%d]", update.Message.Chat.FirstName, update.Message.Chat.ID)
		msg.Text = "Ошибка функции"
		bot.Send(msg)
		newfn, _ = askCommand(bot, update, nil)
		return
	}
	code, err := strconv.Atoi(update.Message.Text)
	if err != nil {
		msg.Text = "Введено не число"
		bot.Send(msg)
		newfn, _ = askCommand(bot, update, nil)
		return
	}
	stationName, isexist := node.stations[code]
	if !isexist {
		msg.Text = "Введенный код станции отсутствует в этом узле"
		bot.Send(msg)
		newfn, _ = askCommand(bot, update, nil)
		return
	}
	err = settings.dbinterface.AddStation(code, node.NodeName, update.Message.Chat.ID)
	if err != nil {
		myLogf(LogError, "Error when adding to %s (%d) from %s to database for %d (%s)", stationName, code, node.NodeName, update.Message.Chat.ID, err.Error())
		msg.Text = "Ошибка добавления в базу"
		bot.Send(msg)
		newfn, _ = askCommand(bot, update, nil)
		return
	}
	msg.Text = fmt.Sprintf("Станция %s (%d) [узел %s] добавлена в сохраненные", stationName, code, node.NodeName)
	bot.Send(msg)
	newfn, _ = askCommand(bot, update, nil)
	return
}
