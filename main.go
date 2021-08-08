package main

import (
	log "github.com/cihub/seelog"
	_ "nashcloud_monitor_agent/src/init"
)

func main() {

	defer log.Info("exit")
}
