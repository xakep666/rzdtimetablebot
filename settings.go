package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
    "time"
)

const (
	DBEmbedded = iota
)

type Settings struct {
	NodesFolder string
	BotKey      string
	LogLevel    int
	DoTranslit  bool
	DBType      int
	DBPath      string
	DBTable     string
	dbinterface DBInterface
}

var settings Settings

type Line struct {
	LineName string
	Stations []int
}

type StationMap map[int]string

type Node struct {
	NodeName  string
	CodesFile string
	Lines     []Line
    TZLocation  string
    TZ        time.Location
	stations  StationMap
}

var nodes []Node

func parseStations(name string) (sm StationMap, err error) {
	sm = make(StationMap)
	sfile, err := os.Open(filepath.Join(settings.NodesFolder, name))
	if err != nil {
		return sm, err
	}
	csvrd := csv.NewReader(sfile)
	csvrd.Comma = ';'
	csvrd.Comment = '#'
	csvrd.FieldsPerRecord = 2
	records, err := csvrd.ReadAll()
	if err != nil {
		myLog(LogError, err)
		return sm, err
	}
	for _, record := range records {
		myLog(LogDebug, "Parsing record", record[0], ";", record[1])
		code, err := strconv.Atoi(record[0])
		if err != nil {
			err = errors.New("Station code must be int, got " + record[0])
			return nil, err
		}
		name := record[1]
		sm[code] = name
	}
	return sm, err
}

func parseNode(filename string) (n Node, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return n, err
	}
	jsondec := json.NewDecoder(file)
	err = jsondec.Decode(&n)
	if err != nil {
		return n, err
	}
	n.stations, err = parseStations(n.CodesFile)
	myLogf(LogInfo, "Found %d stations and %d lines for node \"%s\"\n", len(n.stations), len(n.Lines), n.NodeName)
    loc,err:=time.LoadLocation(n.TZLocation)
    if err !=nil {
        return n,err
    }
    n.TZ=*loc
    myLogf(LogInfo, "Got TZ: [%s] for node %s\n",n.TZLocation,n.NodeName)
	return n, err
}
