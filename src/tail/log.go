package tail

//
//import (
//	"fmt"
//	log "github.com/cihub/seelog"
//	"nashcloud_monitor_agent/src/config"
//	ci "nashcloud_monitor_agent/src/crust_info"
//	er "nashcloud_monitor_agent/src/error"
//	"nashcloud_monitor_agent/src/local"
//	"nashcloud_monitor_agent/src/utils"
//	"net"
//	"strconv"
//	"strings"
//)
//
//type MainLog struct {
//	time           string
//	newBlock       string
//	localIp        string
//	error          string
//	hostName       string
//	ipfs		   string
//	smanager	   string
//	addr 	string
//	filesLost string
//	filesPeeding string
//	filesVaild string
//	srdComplete string
//	srdRemainingTask string
//	DiskAVA4Srd string
//}
//
//func logPush(mainLog MainLog) {
//	db, err := config.GetDBConnection()
//	if err != nil {
//		log.Error("get db connection failed: %s from %s", err.Error())
//		return
//	}
//	stmt, err := db.Prepare("INSERT INTO monitor_crust_que ( Ip, hostName, pullQue, smallTaskQue, bigTaskQue, sealQue, block, error) VALUES (?,?,?,?,?,?,?,?);")
//	if err != nil {
//		log.Error("prepare add host indicator failed: %s from %s", err.Error())
//		return
//	}
//	_, err = stmt.Exec(mainLog.localIp, mainLog.hostName, mainLog.pullQueCount, mainLog.smallTaskCount, mainLog.bigTaskCount, mainLog.sealQueCount, mainLog.newBlock, "nil")
//	if err != nil {
//		log.Error("prepare add host indicator failed: %s from %s", err.Error())
//		return
//	}
//}
//
//var mainLog MainLog
//var count = -1
//
//
//func MainLogSync(log string, time string ,backupJson utils.BackupJson,c utils.Conf)  {
//
//	defer func() {
//		if r := recover(); r != nil {
//			fmt.Println("recover...:", r)
//			er.ErrorHandler(r.(string))
//		}
//	}()
//			mainLog.time = time
//			mainLog.hostName = local.GetLocal().HostName
//			mainLog.localIp = local.GetLocal().Ip
//			mainLog.ipfs = c.Node.Ipfs
//			mainLog.smanager = c.Node.Smanager
//			mainLog.addr = backupJson.Address
//
//			health := ci.CheckHealth()
//			healthCount, _ := strconv.Atoi(health)
//			mainLog.error = strconv.Itoa(5 - healthCount)
//
//			if count == -1 && strings.Index(log, "task done") != -1 {
//			count ++
//			}
//			if count == 0 && strings.Index(log, "block") != -1 {
//				logPush(mainLog)
//				count = -1
//			}
//
//}
