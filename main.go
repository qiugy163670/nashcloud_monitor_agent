package main

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	_ "nashcloud_monitor_agent_sync/src/init"
)

func main() {
	cpuInfos, _ := cpu.Times(false)
	i := 0
	for item := range cpuInfos {
		fmt.Printf("%d\n", cpuInfos[i])
		fmt.Println(item)
		i++
	}

}
