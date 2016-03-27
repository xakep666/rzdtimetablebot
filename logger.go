package main

import (
	"github.com/fiam/gounidecode/unidecode"
	"log"
)

const (
	LogError = iota
	LogInfo
	LogDebug
)

func myLog(level int, data ...interface{}) {
	if level > settings.LogLevel {
		return
	}
	var newdata []interface{}
	switch level {
	case LogError:
		newdata = append(newdata, "ERROR:")
	case LogInfo:
		newdata = append(newdata, "INFO:")
	case LogDebug:
		newdata = append(newdata, "DEBUG:")
	default:
		newdata = append(newdata, "<INVALID LEVEL>:")
	}
	for _, v := range data {
		if str, isstr := v.(string); isstr && settings.DoTranslit {
			newdata = append(newdata, unidecode.Unidecode(str))
		} else {
			newdata = append(newdata, v)
		}
	}
	log.Println(newdata...)
}

func myLogf(level int, format string, data ...interface{}) {
	if level > settings.LogLevel {
		return
	}
	var newdata []interface{}
	for _, v := range data {
		if str, isstr := v.(string); isstr && settings.DoTranslit {
			newdata = append(newdata, unidecode.Unidecode(str))
		} else {
			newdata = append(newdata, v)
		}
	}
	switch level {
	case LogError:
		log.Printf("ERROR: "+format, newdata...)
	case LogInfo:
		log.Printf("INFO: "+format, newdata...)
	case LogDebug:
		log.Printf("DEBUG: "+format, newdata...)
	default:
		log.Printf("<INVALID LEVEL>: "+format, newdata...)
	}
}
