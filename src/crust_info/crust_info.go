package crust_info

import (
	"fmt"
	ca "nashcloud_monitor_agent_sync/src/cmd"
	"strings"
)

func GetContainer() map[string]string {
	pac := ca.ProcessAgentCheck{
		BinPath: "/bin/sh",
	}

	err, res := pac.ExecCmd("sudo docker ps | awk '{print $NF,$1}'")
	if err != nil {
		fmt.Println(err)
	}
	docker := make(map[string]string)
	for e := res.Front(); e != nil; e = e.Next() {
		str := e.Value.(string)
		strs := strings.Split(str, " ")
		docker[strs[0]] = strs[1]
	}

	return docker

}
