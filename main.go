package main

import (
	"fmt"

	"github.com/shirou/gopsutil/cpu"
)

func main() {
	fmt.Println("1111111111111")
	cpuInfos, _ := cpu.Times(false)
	i := 0
	for item := range cpuInfos {
		fmt.Println("2222222 %d", cpuInfos[i])
		fmt.Println(item)
		i++
	}
	//agent.ExecuteTask()
}
