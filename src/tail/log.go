package tail

import (
	"encoding/json"
	"fmt"
	log "github.com/cihub/seelog"
	"nashcloud_monitor_agent/src/config"
	"nashcloud_monitor_agent/src/local"
	"nashcloud_monitor_agent/src/utils"
	"strings"
)

type Log struct {
	Log    string `json:"log"`
	Stream string `json:"stream"`
	Time   string `json:"time"`
}

type MainLog struct {
	time           string
	bigTaskCount   string
	smallTaskCount string
	sealQueCount   string
	pullQueCount   string
	newBlock       string
	localIp        string
	error          string
	hostName       string
}

func logPush(mainLog MainLog) {
	db, err := config.GetDBConnection()
	if err != nil {
		log.Error("get db connection failed: %s from %s", err.Error())
		return
	}
	stmt, err := db.Prepare("INSERT INTO monitor_crust_que ( Ip, hostName, pullQue, smallTaskQue, bigTaskQue, sealQue, block, error) VALUES (?,?,?,?,?,?,?,?);")
	if err != nil {
		log.Error("prepare add host indicator failed: %s from %s", err.Error())
		return
	}
	_, err = stmt.Exec(mainLog.localIp, mainLog.hostName, mainLog.pullQueCount, mainLog.smallTaskCount, mainLog.bigTaskCount, mainLog.sealQueCount, "nil", "nil")
	if err != nil {
		log.Error("prepare add host indicator failed: %s from %s", err.Error())
		return
	}
}

var mainLog MainLog
var count = -1

func MainLogSync(log string, time string) {
	//Checking pulling queue start
	if strings.Index(log, "Sealing queue length") != -1 {
		mainLog.time = utils.UTCTransLocal(time)
		mainLog.hostName = local.GetLocal().HostName
		mainLog.localIp = local.GetLocal().Ip
		index := strings.Index(log, "/5000")
		if index == -1 {
			mainLog.pullQueCount = "0/5000"
		} else {
			mainLog.sealQueCount = strings.ReplaceAll(log[index-3:index+5], ":", "")
		}
		count = 0
	}
	if count == 0 && strings.Index(log, "Pulling queue length") != -1 {
		index := strings.Index(log, "/5000")
		if index == -1 {
			mainLog.pullQueCount = "0"
		} else {
			mainLog.pullQueCount = strings.ReplaceAll(log[index-3:index+5], ":", "")
		}

		count++
	}
	if count == 1 && strings.Index(log, "Ipfs small task") != -1 {
		index := strings.Index(log, "/250")
		if index == -1 {
			mainLog.pullQueCount = "0"
		} else {
			mainLog.smallTaskCount = strings.ReplaceAll(log[index-3:index+4], ":", "")
		}
		count++
	}
	if count == 2 && strings.Index(log, "Ipfs big task count") != -1 {
		index := strings.Index(log, "/500")
		if index == -1 {
			mainLog.pullQueCount = "0"
		} else {
			mainLog.bigTaskCount = strings.ReplaceAll(log[index-3:index+4], ":", "")
		}
		count++
	}
	if strings.Index(log, "Checking pulling queue end") != -1 {
		fmt.Println(mainLog)
		logPush(mainLog)
		//mainLog = MainLog{}
		count = -1
	}
}

func Json2Struct(jsonStr string) Log {
	var log Log
	json.Unmarshal([]byte(jsonStr), &log)
	return log
}
