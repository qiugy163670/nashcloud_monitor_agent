package tail

import (
	"encoding/json"
	util "nashcloud_monitor_agent_sync/src/util"
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
	newBlock       string
	localIp        string
}

func MainLogSync(log Log) MainLog {
	var mainLog MainLog
	mainLog.time = util.UTCTransLocal(log.Time)

	return mainLog
}

func Json2Struct(jsonStr string) Log {
	var log Log
	json.Unmarshal([]byte(jsonStr), &log)
	return log
}
