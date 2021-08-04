package tail

import (
	"encoding/json"
	"fmt"
	util "nashcloud_monitor_agent_sync/src/utils"
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
}

func logPush() {
	//db, err := config.GetDBConnection()
	//if err != nil {
	//	log.Error("get db connection failed: %s from %s", err.Error())
	//	return
	//}
	//stmt, err := db.Prepare("update net_record set net_bytes_int = ?, net_bytes_out = ?, net_package_in = ?, net_package_out = ?, net_drop_in = ?, net_drop_out = ?, net_error_in, net_error_out")
	//if err != nil {
	//	log.Error("prepare add host indicator failed: %s from %s", err.Error(), tmpName)
	//	return
	//}
	//_, err = stmt.Exec(netInfo[0].BytesRecv, netInfo[0].BytesSent, netInfo[0].PacketsRecv, netInfo[0].PacketsSent, netInfo[0].Dropin, netInfo[0].Dropout, netInfo[0].Errin, netInfo[0].Errout)
	//if err != nil {
	//	log.Error("prepare add host indicator failed: %s from %s", err.Error(), tmpName)
	//	return
	//}
}

var mainLog MainLog

func MainLogSync(log Log) MainLog {
	mainLog.time = util.UTCTransLocal(log.Time)

	//index := strings.Index(log.Log, "small task")
	if strings.Index(log.Log, "small task") != -1 {
		split := strings.Split(log.Log, ":")
		mainLog.smallTaskCount = split[len(split)-1]
	} else if strings.Index(log.Log, "Pulling queue length") != -1 {
		split := strings.Split(log.Log, ":")
		mainLog.pullQueCount = split[len(split)-1]
	} else if strings.Index(log.Log, " big task") != -1 {
		split := strings.Split(log.Log, ":")
		mainLog.bigTaskCount = split[len(split)-1]
	} else if strings.Index(log.Log, "Sealing queue length") != -1 {
		split := strings.Split(log.Log, ":")
		mainLog.sealQueCount = split[len(split)-1]
		fmt.Println(mainLog)
		mainLog = MainLog{}
	}

	return mainLog
}

func Json2Struct(jsonStr string) Log {
	var log Log
	json.Unmarshal([]byte(jsonStr), &log)
	return log
}
