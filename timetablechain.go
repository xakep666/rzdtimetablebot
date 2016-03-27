package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"strings"
	"time"
)

func recvStationCodeTimeTable(bot *tgbotapi.BotAPI, update tgbotapi.Update, data []interface{}) (newfn sorterFunc, newdata []interface{}) {
	myLogf(LogDebug, "Handling reciving station code for timetable for %s [%d]", update.Message.Chat.FirstName, update.Message.Chat.ID)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	code, err := strconv.Atoi(update.Message.Text)
	if err != nil {
		msg.Text = "Введено не число"
		bot.Send(msg)
		newfn, _ = askCommand(bot, update, nil)
		return
	}
	timetable, err := DownloadStationTimeTable(code)
	if err != nil {
		myLogf(LogError, "Cannot download timetable for station %d (%s)", code, err.Error())
		msg.Text = "Не удалось загрузить расписание"
		bot.Send(msg)
		newfn, _ = askCommand(bot, update, nil)
		return
	}
	directions := []string{}
	for _, v := range timetable {
		directions = append(directions, v.Direction)
	}
	msg.Text = "В какую сторону показать расписание"
	msg.ReplyMarkup = kbdMarkupAligner(directions)
	newdata = append(newdata, timetable)
	newfn = recvShowTimeTable
	bot.Send(msg)
	return
}

func recvShowTimeTable(bot *tgbotapi.BotAPI, update tgbotapi.Update, data []interface{}) (newfn sorterFunc, newdata []interface{}) {
	myLogf(LogDebug, "Handling reciving direction for timetable for %s [%d]", update.Message.Chat.FirstName, update.Message.Chat.ID)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardHide{HideKeyboard: true}
	if len(data) <= 0 {
		myLogf(LogError, "recvShowTimeTalbe must have timetable argument %d [%d]", update.Message.Chat.FirstName, update.Message.Chat.ID)
		msg.Text = "Ошибка функции"
		bot.Send(msg)
		newfn, _ = askCommand(bot, update, nil)
		return
	}
	timetable, istimetable := data[0].(StationTimeTable)
	if !istimetable {
		myLogf(LogError, "recvShowTimeTalbe must have timetable argument %d [%d]", update.Message.Chat.FirstName, update.Message.Chat.ID)
		msg.Text = "Ошибка функции"
		bot.Send(msg)
		newfn, _ = askCommand(bot, update, nil)
		return
	}
	found := false
	tabletoshow := StationTimeTableEntry{}
	for _, v := range timetable {
		if strings.ToUpper(v.Direction) == strings.ToUpper(update.Message.Text) {
			found = true
			tabletoshow = v
			break
		}
	}
	if !found {
		msg.Text = "Такое направление не найдено"
		bot.Send(msg)
		newfn, _ = askCommand(bot, update, nil)
		return
	}
	msg.ReplyMarkup = kbdMarkupAligner([]string{"Да", "Нет"})
	now := time.Now()
	msg.Text = "Формат\nНомерМаршрута|ВремяПрибытия|ВремяОтправления|Сообщение|ФактическоеДвижение\n"
	procarriving := tabletoshow.Direction == "ПРИБЫТИЕ"
	for _, v := range tabletoshow.TimeTable {
		arrtime := v.ArrivalTime
		arrtimetext := ""
		if arrtime != nil {
			arrtimetext = arrtime.Format("15:04")
		} else {
			arrtimetext = "-"
		}
		if procarriving {
			msg.Text += fmt.Sprintf("-----------\n%d|%s|%s|\n%s\n",
				v.RouteCode, arrtimetext, v.FromTo, v.FactMovement)
		} else {
			if v.DepartTime.After(now) {
				msg.Text += fmt.Sprintf("-----------\n%d|%s|%s|%s|\n%s\n",
					v.RouteCode, arrtimetext,
					v.DepartTime.Format("15:04"), v.FromTo, v.FactMovement)
			}
		}
		if len(msg.Text) >= 4000 {
			bot.Send(msg)
			msg.Text = ""
		}
	}
	msg.Text += "-----------\n"
	msg.Text += "Показать полное расписание?"
	newfn = recvShowFullTable
	newdata = append(newdata, tabletoshow)
	bot.Send(msg)
	newfn = recvShowFullTable
	return
}

func recvShowFullTable(bot *tgbotapi.BotAPI, update tgbotapi.Update, data []interface{}) (newfn sorterFunc, newdata []interface{}) {
	myLogf(LogDebug, "Handling show full timetable for %s [%d]", update.Message.Chat.FirstName, update.Message.Chat.ID)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardHide{HideKeyboard: true}
	if len(data) <= 0 {
		myLogf(LogError, "recvShowFullTable must have table argument %d [%d]", update.Message.Chat.FirstName, update.Message.Chat.ID)
		msg.Text = "Ошибка функции"
		bot.Send(msg)
		newfn, _ = askCommand(bot, update, nil)
		return
	}
	table, istable := data[0].(StationTimeTableEntry)
	if !istable {
		myLogf(LogError, "recvShowFullTable must have table argument %d [%d]", update.Message.Chat.FirstName, update.Message.Chat.ID)
		msg.Text = "Ошибка функции"
		bot.Send(msg)
		newfn, _ = askCommand(bot, update, nil)
		return
	}
	if strings.ToUpper(update.Message.Text) != "ДА" {
		newfn, _ = askCommand(bot, update, nil)
		return
	}
	msg.Text = "Формат\nНомерМаршрута|ВремяПрибытия|ВремяОтправления|Сообщение|ФактическоеДвижение\n"
	procarriving := table.Direction == "ПРИБЫТИЕ"
	for _, v := range table.TimeTable {
		arrtime := v.ArrivalTime
		arrtimetext := ""
		if arrtime != nil {
			arrtimetext = arrtime.Format("15:04")
		} else {
			arrtimetext = "-"
		}
		if procarriving {
			msg.Text += fmt.Sprintf("-----------\n%d|%s|%s|\n%s\n",
				v.RouteCode, arrtimetext, v.FromTo, v.FactMovement)
		} else {
			msg.Text += fmt.Sprintf("-----------\n%d|%s|%s|%s|\n%s\n",
				v.RouteCode, arrtimetext,
				v.DepartTime.Format("15:04"), v.FromTo, v.FactMovement)
		}
		if len(msg.Text) >= 4000 {
			bot.Send(msg)
			msg.Text = ""
		}
	}
	msg.Text += "-----------\n"
	bot.Send(msg)
	newfn, _ = askCommand(bot, update, nil)
	return
}
