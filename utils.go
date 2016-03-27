package main

import "strings"

func stationByCode(code int) string {
    ret:=""
    for _,node:=range nodes {
        ret=node.stations[code]
    }
    return ret
}

func nodeNames() (ret []string) {
    for _,node:=range nodes {
        ret=append(ret,node.NodeName)
    }
    return
}

func nodeByName(name string) (ret Node,found bool){
    for _,node:=range nodes {
        if strings.ToUpper(node.NodeName)==strings.ToUpper(name) {
            return node,true
        }
    }
    found=false
    return
}

func fuzzySearchStation(node Node, substr string) (ret []int) { //коды станций
	myLogf(LogDebug, "Perform search by example %s in node %s", substr, node.NodeName)
	for code, station := range node.stations {
		if strings.Contains(station, strings.ToUpper(substr)) {
			ret = append(ret, code)
		}
	}
	return
}

type CodeStationPair struct {
    Code int
    Name string
} 

type CodeStationPairs []CodeStationPair

func (csps CodeStationPairs) Len() int {
    return len(csps)
}

func (csps CodeStationPairs) Swap(i,j int) {
    csps[i],csps[j]=csps[j],csps[i]
}

func (csps CodeStationPairs) Less(i,j int) bool {
    return csps[i].Name<csps[j].Name
}