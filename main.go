package main

import (
	"flag"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

const StartMessage = "Стартовое сообщение"
const TimeTableQueryUrlFormat = "http://pass.rzd.ru/tablo/public/ru?STRUCTURE_ID=704&layer_id=5366&refererLayerId=5368&date=%s&id=%d"

func initNodes() {
	nodes = []Node{}
	myLog(LogDebug, "Listing nodes folder", settings.NodesFolder)
	files, err := ioutil.ReadDir(settings.NodesFolder)
	if err != nil {
		log.Fatalln("Cannot access nodes directory: ", err)
	}
	for _, finfo := range files {
		fname := finfo.Name()
		if filepath.Ext(fname) == ".json" {
			myLog(LogDebug, "Try parse", fname)
			if !finfo.Mode().IsRegular() {
				log.Println("Found non-file with .json extension")
				continue
			}
			node, err := parseNode(filepath.Join(settings.NodesFolder, fname))
			if err != nil {
				log.Fatalln("Cannot parse node:", err)
			}
			myLog(LogInfo, "Registered node", node.NodeName)
			nodes = append(nodes, node)
		}
	}
}

func init() {
	flag.StringVar(&settings.NodesFolder, "nodesdir", "nodes", "Directory with nodes description files")
	flag.StringVar(&settings.BotKey, "botkey", "<default>", "API key for telegram bot (if starts with env, takes key from os environment variable)")
	flag.IntVar(&settings.LogLevel, "loglevel", LogError, "Log Level: 0 - errors only, 1 - info and errors, 2 - debug,info,error")
	flag.BoolVar(&settings.DoTranslit, "translit", false, "Translit symbols on logging")
	flag.IntVar(&settings.DBType, "dbtype", DBEmbedded, "Database type - 0 - embedded(boltdb)")
	flag.StringVar(&settings.DBPath, "dbpath", "tgbot.db", "Path where we can find database")
	flag.StringVar(&settings.DBTable, "dbtable", "rzdtimetablebot", "Name of table in database")
	flag.Parse()
	if settings.LogLevel != LogError &&
		settings.LogLevel != LogInfo &&
		settings.LogLevel != LogDebug {
		flag.PrintDefaults()
		log.Fatalf("Log level must be %d,%d or %d\n", LogError, LogInfo, LogDebug)
	}
	if settings.BotKey == "<default>" {
		flag.PrintDefaults()
		log.Fatalln("Bot key must be defined")
	}
	if strings.HasPrefix(settings.BotKey, "env") {
		settings.BotKey = os.Getenv(strings.TrimPrefix(settings.BotKey, "env"))
		if settings.BotKey == "" {
			flag.PrintDefaults()
			log.Fatalln("Bot key must be defined")
		}
	}
	if settings.DBType != DBEmbedded {
		flag.PrintDefaults()
		log.Fatalf("Database type must be %d\n", DBEmbedded)
	}
	//sighup для перечитывания файлов
	sigch := make(chan os.Signal)
	signal.Notify(sigch, syscall.SIGHUP)
	go func() {
		for sig := range sigch {
			switch sig {
			case syscall.SIGHUP:
				myLog(LogInfo, "SIGHUP Catched, reloading nodes")
				initNodes()
			}
		}
	}()
}

func main() {
	//найти все .json файлы в папке из аргументов запуска и обработать
	initNodes()
	//инициализация базы
	var err error
	settings.dbinterface, err = DBInit()
	if err != nil {
		log.Fatalln(err)
	}
	defer settings.dbinterface.Close()
	//инициализация бота
	bot, err := tgbotapi.NewBotAPI(settings.BotKey)
	if err != nil {
		log.Fatalln(err)
	}
	//bot.Debug=settings.LogLevel==LogDebug
	myLogf(LogInfo, "Authorized under %s\n", bot.Self.UserName)
	botRequestHandler(bot)
}
