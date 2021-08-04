package main

import (
	"nashcloud_monitor_agent_sync/src/agent"
	_ "nashcloud_monitor_agent_sync/src/init"
)

func main() {
	agent.ExecuteTask()
}
