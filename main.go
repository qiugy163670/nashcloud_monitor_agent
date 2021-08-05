package main

import (
	"nashcloud_monitor_agent/src/agent"
	_ "nashcloud_monitor_agent/src/init"
)

func main() {
	agent.ExecuteTask()
}
