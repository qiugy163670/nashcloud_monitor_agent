package crust_info

import (
	log "github.com/cihub/seelog"
	ca "nashcloud_monitor_agent/src/cmd"
)

func CheckHealth() string {
	pac := ca.ProcessAgentCheck{
		BinPath: "/bin/sh",
	}
	err, list := pac.ExecCmd("crust status |grep running |wc -l")
	if err != nil {
		log.Errorf("check health fair")
	}
	res := "0"
	if list.Len() > 0 {
		res = list.Front().Value.(string)
	}
	return res
}
