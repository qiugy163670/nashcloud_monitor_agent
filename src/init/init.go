package init

import (
	"fmt"
	ca "nashcloud_monitor_agent_sync/src/cmd"
	ci "nashcloud_monitor_agent_sync/src/crust_info"
	lt "nashcloud_monitor_agent_sync/src/tail"
)

func init() {
	fmt.Println("===============agent start ===============")
	fmt.Println("checking docker ")

	container := ci.GetContainer()
	for k, v := range container {
		fmt.Println(k, "---", v)
	}

	fmt.Println("===============check docker is fine=============== ")
	fmt.Println("===============search crust logs===============")

	crustApiPath := getCrustLogsPath(container["crust-api"])
	ipfs := getCrustLogsPath(container["ipfs"])
	crustSworker := getCrustLogsPath(container["crust-sworker-a"])
	crustSmanager := getCrustLogsPath(container["crust-smanager"])
	crust := getCrustLogsPath(container["crust"])

	fmt.Println("api log file :", crustApiPath)
	fmt.Println("ipfs log file :", ipfs)
	fmt.Println("crustSworker log file :", crustSworker)
	fmt.Println("crustSmanager log file :", crustSmanager)
	fmt.Println("crust log file :", crust)
	fmt.Println("===============search crust logs is fine===============")
	messages := make(chan string, 1)
	go lt.Stream(crustSmanager, messages)
	//go lt.Stream(crustSworker,messages)
	//go lt.Stream(crustApiPath,messages)
	for message := range messages {
		fmt.Println("received", message)
	}

}

func getCrustLogsPath(id string) string {
	logBasePath := "/var/lib/docker/containers/"
	pac := ca.ProcessAgentCheck{
		BinPath: "/bin/sh",
	}
	cmdStr := "sudo ls -l " + logBasePath + "| grep " + id + " | awk '{print $NF}'"
	err, res := pac.ExecCmd(cmdStr)
	if err != nil {
		fmt.Println("error: ", err)
	}
	midPath := res.Front().Value.(string)
	return logBasePath + midPath + "/" + midPath + "-json.log"

}
