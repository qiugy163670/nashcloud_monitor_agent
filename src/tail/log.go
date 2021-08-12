package tail

import (
	"fmt"
	log "github.com/cihub/seelog"
	"nashcloud_monitor_agent/src/agent"
	"nashcloud_monitor_agent/src/config"
	ci "nashcloud_monitor_agent/src/crust_info"
	"nashcloud_monitor_agent/src/local"
	"strconv"
	"strings"
)

type Log struct {
	Log    string `json:"log"`
	Stream string `json:"stream"`
	Time   string `json:"time"`
}

type MainLog struct {
	time           string
	bigTaskCount   string
	smallTaskCount string
	sealQueCount   string
	pullQueCount   string
	newBlock       string
	localIp        string
	error          string
	hostName       string
}

func logPush(mainLog MainLog) {
	db, err := config.GetDBConnection()
	if err != nil {
		log.Error("get db connection failed: %s from %s", err.Error())
		return
	}
	stmt, err := db.Prepare("INSERT INTO monitor_crust_que ( Ip, hostName, pullQue, smallTaskQue, bigTaskQue, sealQue, block, error) VALUES (?,?,?,?,?,?,?,?);")
	if err != nil {
		log.Error("prepare add host indicator failed: %s from %s", err.Error())
		return
	}
	_, err = stmt.Exec(mainLog.localIp, mainLog.hostName, mainLog.pullQueCount, mainLog.smallTaskCount, mainLog.bigTaskCount, mainLog.sealQueCount, mainLog.newBlock, "nil")
	if err != nil {
		log.Error("prepare add host indicator failed: %s from %s", err.Error())
		return
	}
}

var mainLog MainLog
var count = -1
var timeCount = 0
var ver = 0

func MainLogSync(log string, time string) int {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recover...:", r)
		}
	}()
	if ver == 0 && strings.Index(log, "Got new block") != -1 {
		// old crust
		ver = 1
		fmt.Println("is old crust")
	} else if ver == 0 {
		ver = 2
	}
	tempCount := 0
	if ver == 2 {
		if count == -1 && strings.Index(log, "Checking pulling queue") != -1 {
			mainLog.time = time //utils.UTCTransLocal(time)
			mainLog.hostName = local.GetLocal().HostName
			mainLog.localIp = local.GetLocal().Ip

			health := ci.CheckHealth()
			healthCount, _ := strconv.Atoi(health)
			mainLog.error = strconv.Itoa(5 - healthCount)
			//
			//index := strings.Index(log, "/5000")
			//if index == -1 {
			//	mainLog.pullQueCount = "0/5000"
			//} else {
			//	mainLog.sealQueCount = strings.ReplaceAll(log[index-3:index+5], ":", "")
			//}
			count = 0
		}

		if count == 0 && strings.Index(log, "Pulling queue length") != -1 {
			index := strings.Index(log, "/5000")
			if index == -1 {
				mainLog.pullQueCount = "0"
			} else {
				mainLog.pullQueCount = strings.ReplaceAll(log[index-3:index+5], ":", "")
			}
			count++
		}
		if count == 1 && strings.Index(log, "Ipfs small task") != -1 {
			index := strings.Index(log, "/100")
			if index == -1 {
				mainLog.pullQueCount = "0"
			} else {
				mainLog.smallTaskCount = strings.ReplaceAll(log[index-3:index+4], ":", "")
			}
			count++
		}
		if count == 2 && strings.Index(log, "Ipfs big task count") != -1 {
			index := strings.Index(log, "/200")
			if index == -1 {
				mainLog.pullQueCount = "0"
			} else {
				mainLog.bigTaskCount = strings.ReplaceAll(log[index-3:index+4], ":", "")
			}
			count++
		}

		if count == 3 && strings.Index(log, "Checking pulling queue end") != -1 {
			count++
		}
		if count == 4 && strings.Index(log, "block") != -1 {
			start := strings.Index(log, "block")
			end := strings.Index(log, "(0x")
			start = start + 5
			if start < end && end < len(log) {
				mainLog.newBlock = log[start:end]
			}
			fmt.Println(mainLog)
			logPush(mainLog)
			tempCount = count
			count = -1
			if timeCount < 5 {
				timeCount++
			} else {
				agent.CollectJob()
				timeCount = 0
			}
		}
	}
	//Checking pulling queue start
	if ver == 1 {
		if count == -1 && strings.Index(log, "Sealing queue length") != -1 {
			mainLog.time = time //utils.UTCTransLocal(time)
			mainLog.hostName = local.GetLocal().HostName
			mainLog.localIp = local.GetLocal().Ip

			health := ci.CheckHealth()
			healthCount, _ := strconv.Atoi(health)
			mainLog.error = strconv.Itoa(5 - healthCount)

			index := strings.Index(log, "/5000")
			if index == -1 {
				mainLog.pullQueCount = "0/5000"
			} else {
				mainLog.sealQueCount = strings.ReplaceAll(log[index-3:index+5], ":", "")
			}
			count = 0
		}
		if count == 0 && strings.Index(log, "block") != -1 {
			start := strings.Index(log, "block")
			end := strings.Index(log, "(0x")
			start = start + 5
			if start < end && end < len(log) {
				mainLog.newBlock = log[start:end]
			}
		}
		if count == 0 && strings.Index(log, "Pulling queue length") != -1 {
			index := strings.Index(log, "/5000")
			if index == -1 {
				mainLog.pullQueCount = "0"
			} else {
				mainLog.pullQueCount = strings.ReplaceAll(log[index-3:index+5], ":", "")
			}
			count++
		}
		if count == 1 && strings.Index(log, "Ipfs small task") != -1 {
			index := strings.Index(log, "/250")
			if index == -1 {
				mainLog.pullQueCount = "0"
			} else {
				mainLog.smallTaskCount = strings.ReplaceAll(log[index-3:index+4], ":", "")
			}
			count++
		}
		if count == 2 && strings.Index(log, "Ipfs big task count") != -1 {
			index := strings.Index(log, "/500")
			if index == -1 {
				mainLog.pullQueCount = "0"
			} else {
				mainLog.bigTaskCount = strings.ReplaceAll(log[index-3:index+4], ":", "")
			}
			count++
		}
		//tempCount := 0
		if count == 3 && strings.Index(log, "Checking pulling queue end") != -1 {
			fmt.Println(mainLog)
			logPush(mainLog)
			tempCount = count
			count = -1
			if timeCount < 4 {
				timeCount++
			} else {
				agent.CollectJob()
				timeCount = 0
			}
		}
	}
	return tempCount
}
