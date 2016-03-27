package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"sort"
	"strings"
)

func recvNodeNameShowLine(bot *tgbotapi.BotAPI, update tgbotapi.Update, data []interface{}) (newfn sorterFunc, newdata []interface{}) {
	myLogf(LogDebug, "Handling node selection request to show line for %s [%d]", update.Message.Chat.FirstName, update.Message.Chat.ID)
	nname := update.Message.Text
	node, found := nodeByName(nname)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	lines := []string{}
	for _, line := range node.Lines {
		lines = append(lines, line.LineName)
	}
	msg.ReplyMarkup = kbdMarkupAligner(lines)
	if !found {
		newfn = recvCommand
		msg.Text = "Такой узел не зарегистрирован"
		bot.Send(msg)
		newfn, _ = askCommand(bot, update, nil)
		return
	}
	newdata = append(newdata, node)
	newfn = recvLine
	msg.Text = "Выберите линию (направление)"
	bot.Send(msg)
	return
}

func recvLine(bot *tgbotapi.BotAPI, update tgbotapi.Update, data []interface{}) (newfn sorterFunc, newdata []interface{}) {
	myLogf(LogDebug, "Handling line select request to show stations for %s [%d]", update.Message.Chat.FirstName, update.Message.Chat.ID)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardHide{HideKeyboard: true}
	if len(data) <= 0 {
		myLogf(LogError, "recvLine must have node argument %d [%d]", update.Message.Chat.FirstName, update.Message.Chat.ID)
		msg.Text = "Ошибка функции"
		bot.Send(msg)
		newfn, _ = askCommand(bot, update, nil)
		return
	}
	node, isnode := data[0].(Node)
	if !isnode {
		myLogf(LogError, "recvLine must have node argument %d [%d]", update.Message.Chat.FirstName, update.Message.Chat.ID)
		msg.Text = "Ошибка функции"
		bot.Send(msg)
		newfn, _ = askCommand(bot, update, nil)
		return
	}
	foundline := false
	var line *Line
	for _, v := range node.Lines {
		if strings.ToUpper(v.LineName) == strings.ToUpper(update.Message.Text) {
			line = &v
			foundline = true
			break
		}
	}
	if !foundline {
		msg.Text = "Линия (направление) \"" + update.Message.Text + "\" не найдена"
		bot.Send(msg)
		newfn, _ = askCommand(bot, update, nil)
		return
	}
	msg.Text = "Формат:\nКодСтанции|НазваниеСтанции\n"
	var csps CodeStationPairs
	for _, v := range line.Stations {
		csps = append(csps, CodeStationPair{Code: v, Name: node.stations[v]})
	}
	sort.Sort(csps)
	for _, v := range csps {
		msg.Text += fmt.Sprintf("%d|%s\n", v.Code, v.Name)
	}
	bot.Send(msg)
	newfn, _ = askCommand(bot, update, nil)
	return
}
