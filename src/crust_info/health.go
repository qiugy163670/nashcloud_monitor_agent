package crust_info

import (
	"fmt"
	log "github.com/cihub/seelog"
	ca "nashcloud_monitor_agent/src/cmd"
)

func CheckHealth() string {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recover...:", r)
		}
	}()
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
