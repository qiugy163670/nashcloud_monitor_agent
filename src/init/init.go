package init

import (
	"fmt"
	log "github.com/cihub/seelog"
	"nashcloud_monitor_agent/src/agent"
	ca "nashcloud_monitor_agent/src/cmd"
	er "nashcloud_monitor_agent/src/error"
	"nashcloud_monitor_agent/src/local"
	"nashcloud_monitor_agent/src/utils"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
)

//func startCrustLog(backupJson utils.BackupJson,c utils.Conf ) {
//	messages := make(chan string, 1)
//	agent.CollectJob()
//	//container := ci.GetContainer()
//	//crustApiPath := getCrustLogsPath(container["crust-api"])
//	//ipfs := getCrustLogsPath(container["ipfs"])
//	//crustSworker := getCrustLogsPath(container["crust-sworker-a"])
//	//crustSmanager := getCrustLogsPath(container["crust-smanager"])
//
//	go agent.ExecuteTask()
//	go Conn()
//
//	for message := range messages {
//		log.Info("received", message)
//	}
//
//}

func IsFileExist(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}
func init() {

	Init()
}
func Init() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recover...:", r)
			er.ErrorHandler(r.(string))
		}
	}()
	runtime.GOMAXPROCS(1)
	hostName, _ := os.Hostname()
	ip := utils.GetHostIp()
	local.GetLocal().Ip = ip
	local.GetLocal().HostName = hostName
	log.Info("=============== NashCloud Agent start ===============")
	log.Info(ip)
	log.Info(hostName)
	var c utils.Conf
	pathStr := "/opt/crust/crust-node/config.yaml"

	if IsFileExist(pathStr) {
		c.GetConf(pathStr)
		var backupJson utils.BackupJson
		backupJson = utils.Json2Struct(c.Identity.Backup)
		log.Info("current crust node address : ", backupJson.Address)
		log.Info("current crust node name : ", c.Chain.Name)
		log.Info("current crust node ipfs status : ", c.Node.Ipfs)
		log.Info("current crust node mode : ", c.Node.Smanager)
		go Conn()
		//首次运行时执行一次
		agent.CollectJob(backupJson, c)
		//启动定时任务
		agent.ExecuteTask(backupJson, c)
		//链接socket

	} else {
		log.Info("crust node not installed")

	}

}

func doExec(writer http.ResponseWriter, request *http.Request) {
	///api/v0/diag/sys
	fmt.Println("sss")
}

//func pushCrustNodeInfo(c utils.Conf, address string, ip string, hostName string) {
//	db, err := config.GetDBConnection()
//	if err != nil {
//		err := log.Error("get db connection failed: %s from %s", err.Error())
//		if err != nil {
//			return
//		}
//		return
//	}
//	var addr string
//
//	err = db.QueryRow("select addr from crust_node_info where host_ip = ?", ip).Scan(&addr)
//	if err != nil {
//		//addr not found & insert
//		log.Info("crust info not reg ,just insert it")
//	}
//	if addr == "" {
//		stmt, err := db.Prepare("insert into crust_node_info (addr, smanager, chain, name, ipfs, host_ip, host_name, curr_staking, max_staking) values (?,?,?,?,?,?,?,?,?)")
//		if err != nil {
//			log.Errorf("insert crust info  failed: %s from %s", err.Error())
//			return
//		}
//		_, err = stmt.Exec(address, c.Node.Smanager, c.Node.Chain, c.Chain.Name, c.Node.Ipfs, ip, hostName, "0", "0")
//		if err != nil {
//			log.Errorf("insert error")
//		}
//	}
//
//}

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
