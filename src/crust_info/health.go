package crust_info

import (
	ca "nashcloud_monitor_agent/src/cmd"
)

func CheckHealth() string {
	pac := ca.ProcessAgentCheck{
		BinPath: "/bin/sh",
	}
	_, list := pac.ExecCmd("crust status |grep running |wc -l")

	//fmt.Println(list.Front().Value)
	return list.Front().Value.(string)
}
