package error

import (
	log "github.com/cihub/seelog"
	"nashcloud_monitor_agent/src/config"
	"nashcloud_monitor_agent/src/local"
	"time"
)

func ErrorHandler(error string) {
	ip := local.GetLocal().Ip
	db, err := config.GetDBConnection()
	if err != nil {
		err := log.Error("get db connection failed: %s from %s", err.Error())
		if err != nil {
			return
		}
		return
	}
	stmt, err := db.Prepare("INSERT INTO monitor_agent_error_handler ( `date_time`, `error`, `host_ip`) VALUES ( ?,?,?);")
	if err != nil {
		log.Errorf("insert crust info  failed: %s from %s", err.Error())
		return
	}
	_, err = stmt.Exec(time.Now().Unix(), error, ip)
	if err != nil {
		log.Errorf("insert error")
	}
}
