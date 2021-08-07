package init

import (
	log "github.com/cihub/seelog"
	ca "nashcloud_monitor_agent/src/cmd"
	ci "nashcloud_monitor_agent/src/crust_info"
	"nashcloud_monitor_agent/src/local"
	lt "nashcloud_monitor_agent/src/tail"
	"nashcloud_monitor_agent/src/utils"
	"os"
)

func startCrustLog() {
	messages := make(chan string, 1)
	container := ci.GetContainer()
	//crustApiPath := getCrustLogsPath(container["crust-api"])
	//ipfs := getCrustLogsPath(container["ipfs"])
	//crustSworker := getCrustLogsPath(container["crust-sworker-a"])
	crustSmanager := getCrustLogsPath(container["crust-smanager"])
	//crust := getCrustLogsPath(container["crust"])

	lt.Stream(crustSmanager, messages)
	//go lt.Stream(crustSworker, messages)
	//go lt.Stream(crustApiPath, messages)
	for message := range messages {
		log.Info("received", message)
	}

}
func init() {
	hostName, _ := os.Hostname()
	ip := utils.GetHostIp()
	local.GetLocal().Ip = ip
	local.GetLocal().HostName = hostName
	log.Info("===============agent start ===============")
	log.Info(ip)
	log.Info(hostName)
	startCrustLog()
}

func getCrustLogsPath(id string) string {
	logBasePath := "/var/lib/docker/containers/"
	pac := ca.ProcessAgentCheck{
		BinPath: "/bin/sh",
	}
	cmdStr := "sudo ls -l " + logBasePath + "| grep " + id + " | awk '{print $NF}'"
	err, res := pac.ExecCmd(cmdStr)
	if err != nil {
		log.Info("error: ", err)
	}
	midPath := res.Front().Value.(string)
	return logBasePath + midPath + "/" + midPath + "-json.log"
}
