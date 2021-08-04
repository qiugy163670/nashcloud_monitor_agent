package tail

import (
	"encoding/json"
	log "github.com/cihub/seelog"
	"nashcloud_monitor_agent_sync/src/config"
	util "nashcloud_monitor_agent_sync/src/utils"
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
	error		   string
}

func logPush()  {
	db, err := config.GetDBConnection()
	if err != nil {
		log.Error("get db connection failed: %s from %s", err.Error())
		return
	}
	//stmt, err := db.Prepare("update net_record set net_bytes_int = ?, net_bytes_out = ?, net_package_in = ?, net_package_out = ?, net_drop_in = ?, net_drop_out = ?, net_error_in, net_error_out")
	//if err != nil {
	//	log.Error("prepare add host indicator failed: %s from %s", err.Error(), tmpName)
	//	return
	//}
	//_, err = stmt.Exec(netInfo[0].BytesRecv, netInfo[0].BytesSent, netInfo[0].PacketsRecv, netInfo[0].PacketsSent, netInfo[0].Dropin, netInfo[0].Dropout, netInfo[0].Errin, netInfo[0].Errout)
	//if err != nil {
	//	log.Error("prepare add host indicator failed: %s from %s", err.Error(), tmpName)
	//	return
	}
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
