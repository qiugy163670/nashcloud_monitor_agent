package agent

import (
	"crypto/rand"
	"fmt"
	"github.com/bitly/go-simplejson"
	log "github.com/cihub/seelog"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kirinlabs/HttpRequest"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"math"
	"math/big"
	ca "nashcloud_monitor_agent/src/cmd"
	"nashcloud_monitor_agent/src/config"
	"nashcloud_monitor_agent/src/constants"
	ci "nashcloud_monitor_agent/src/crust_info"
	er "nashcloud_monitor_agent/src/error"
	"nashcloud_monitor_agent/src/local"
	"nashcloud_monitor_agent/src/utils"
	"os"
	"strconv"
	"strings"
	"time"
)

func Init() {
	config.SetupLogger()
}

func getDiskSN(path string) string {
	// 捕获异常
	defer func() {
		if r := recover(); r != nil {

		}

	}()
	pac := ca.ProcessAgentCheck{
		BinPath: "/bin/sh",
	}
	disks := GetDisks()
	if disks.Len() > 2 && strings.Contains(path, "sd") && !strings.Contains(path, "loop") {
		cmd := "hdparm  -I /dev/" + path + " |grep 'Serial Number'"
		//fmt.Println(cmd)
		err, list := pac.ExecCmd(cmd)
		if err != nil {
			fmt.Println(err)
		}
		return strings.ReplaceAll(list.Front().Value.(string), "Serial Number:", "")
	}
	return "is raid"
}

func collectDiskIndicator(name, mount string, diskIoInfo disk.IOCountersStat, stat *disk.UsageStat) {
	tmpName, _ := os.Hostname()
	tmpIp := utils.GetHostIp()
	dateTime := time.Now().Unix()
	db, err := config.GetDBConnection()
	if err != nil {
		log.Errorf("get db connection failed: %s from %s, %s", err.Error(), tmpName, tmpIp)
		return
	}
	//查询上次累加值
	var readCountAcc, writeCountAcc, readBytesAcc, writeBytesAcc, readTimeAcc, writeTimeAcc, ioTimeAcc, weightedIoAcc uint64 = 0, 0, 0, 0, 0, 0, 0, 0
	err = db.QueryRow("select net_bytes_rev,net_bytes_send,net_package_rev,net_package_send,net_drop_rev,net_drop_send,net_error_rev,net_error_send from net_record where host_ip = ? and `name` = ?", tmpIp, name).Scan(&readCountAcc, &writeCountAcc, &readBytesAcc, &writeBytesAcc, &readTimeAcc, &writeTimeAcc, &ioTimeAcc, &weightedIoAcc)
	if err != nil {
		if strings.Contains(err.Error(), constants.NO_ROWS_IN_DB) {
			stmt, err := db.Prepare("insert into net_record (name,host_ip,net_bytes_rev,net_bytes_send,net_package_rev,net_package_send,net_drop_rev,net_drop_send,net_error_rev,net_error_send) values (?,?,?,?,?,?,?,?,?,?)")
			defer stmt.Close()
			if err != nil {
				log.Errorf("prepare insert net_record of self_disk failed: %s from %s", err.Error(), tmpName)
				return
			} else {
				_, err := stmt.Exec(name, tmpIp, diskIoInfo.ReadCount, diskIoInfo.WriteCount, diskIoInfo.ReadBytes, diskIoInfo.WriteBytes, diskIoInfo.ReadTime, diskIoInfo.WriteTime, diskIoInfo.IoTime, diskIoInfo.WeightedIO)
				if err != nil {
					log.Errorf("formal net_record of self_disk failed: %s from %s", err.Error(), tmpName)
				}
			}
		} else {
			log.Errorf("get self_disk total last record failed: %s from %s", err.Error(), tmpName)
			return
		}
	}
	//更新累加值
	stmt, err := db.Prepare("update net_record set net_bytes_rev = ?, net_bytes_send = ?, net_package_rev = ?, net_package_send = ?, net_drop_rev = ?, net_drop_send = ?, net_error_rev = ?, net_error_send = ? where host_ip = ? and `name` = ?")
	defer stmt.Close()
	if err != nil {
		log.Errorf("prepare update net monitor_disk_history self_disk io total failed: %s from %s", err.Error(), tmpName)
		return
	}
	_, err = stmt.Exec(readCountAcc, writeCountAcc, readBytesAcc, writeBytesAcc, readTimeAcc, writeTimeAcc, ioTimeAcc, weightedIoAcc, tmpIp, name)
	if err != nil {
		log.Errorf("update monitor_disk_history self_disk io total failed: %s from %s", err.Error(), tmpName)
		return
	}
	stmt, err = db.Prepare("insert into monitor_disk_indicator (name,host_ip,host_name,device,mount,serial_num,disk_total,disk_used,disk_free,inode_total,inode_used,inode_free,read_count,write_count,read_bytes,write_bytes,read_time,write_time,io_time,weight_io,date_time) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
	defer stmt.Close()
	if err != nil {
		log.Errorf("prepare add disk partition detail failed: %s from %s, %s", err.Error(), tmpName, tmpIp)
		return
	}
	_, err = stmt.Exec(name, tmpIp, tmpName, "/dev/"+name, mount, getDiskSN(name), stat.Total, stat.Used, stat.Free, stat.InodesTotal, stat.InodesUsed, stat.InodesFree, diskIoInfo.ReadCount-readCountAcc, diskIoInfo.WriteCount-writeCountAcc, diskIoInfo.ReadBytes-readBytesAcc, diskIoInfo.WriteBytes-writeBytesAcc, diskIoInfo.ReadTime-readTimeAcc, diskIoInfo.WriteTime-writeTimeAcc, diskIoInfo.IoTime-ioTimeAcc, diskIoInfo.WeightedIO-weightedIoAcc, dateTime-dateTime%300)
	if err != nil {
		log.Errorf("add disk partition detail failed: %s from %s, %s", err.Error(), tmpName, tmpIp)
		return
	}
	defer stmt.Close()

}

func CollectJob(backupJson utils.BackupJson, c utils.Conf) {
	MainLogSync(backupJson, c)
	collectJob()

}

func collectJob() {
	defer func() {
		if r := recover(); r != nil {

		}

	}()
	Init()
	defer log.Flush()

	tmpName, _ := os.Hostname()

	//获取机器信息
	hostInfos, err := host.Info()
	if err != nil {
		log.Errorf("get host info failed: %s from %s", err.Error(), tmpName)
		return
	}
	//查询该机器是否存在，不存在注册
	db, err := config.GetDBConnection()
	if err != nil {
		log.Errorf("get db connection failed: %s from %s", err.Error(), tmpName)
		return
	}
	tmpIp := utils.GetHostIp()

	//获取cpu信息
	cpuInfos, err := cpu.Times(false)
	if err != nil {
		log.Errorf("get cpu info failed: %s from %s", err.Error(), tmpName)
		return
	}
	//cpu是累加值，计算本次cpu值
	var cpuUser, cpuSys, cpuIdle, cpuIOwait, cpuIrq, cpuSofirq float64 = 0, 0, 0, 0, 0, 0
	err = db.QueryRow("select net_bytes_rev,net_bytes_send,net_package_rev,net_package_send,net_drop_rev,net_drop_send from net_record where host_ip = ? and `name` = ?", tmpIp, constants.CPU).Scan(&cpuUser, &cpuSys, &cpuIdle, &cpuIOwait, &cpuIrq, &cpuSofirq)
	if err != nil {
		if strings.Contains(err.Error(), constants.NO_ROWS_IN_DB) {
			stmt, err := db.Prepare("insert into net_record (net_bytes_rev,net_bytes_send,net_package_rev,net_package_send,net_drop_rev,net_drop_send,host_ip,name) values (?,?,?,?,?,?,?,?)")
			defer stmt.Close()
			if err != nil {
				log.Errorf("prepare insert net_record of cpu failed: %s from %s", err.Error(), tmpIp)
				return
			} else {
				_, err := stmt.Exec(cpuInfos[0].User, cpuInfos[0].System, cpuInfos[0].Idle, cpuInfos[0].Iowait, cpuInfos[0].Irq, cpuInfos[0].Softirq, tmpIp, constants.CPU)
				if err != nil {
					log.Errorf("formal insert net_record of cpu failed: %s from %s", err.Error(), tmpName)
				}
			}
		} else {
			log.Errorf("get last cpu info failed: %s from %s", err.Error(), tmpName)
			return
		}
	}
	stmt, err := db.Prepare("update net_record set net_bytes_rev = ?, net_bytes_send = ?, net_package_rev = ?, net_package_send = ?, net_drop_rev = ?, net_drop_send = ? where host_ip = ? and `name` = ?")
	defer stmt.Close()
	if err != nil {
		log.Errorf("prepare update current cpu info failed: %s from %s", err.Error(), tmpName)
	} else {
		_, err := stmt.Exec(cpuInfos[0].User, cpuInfos[0].System, cpuInfos[0].Idle, cpuInfos[0].Iowait, cpuInfos[0].Irq, cpuInfos[0].Softirq, tmpIp, constants.CPU)
		if err != nil {
			log.Errorf("formal update cpu info failed: %s from %s", err.Error(), tmpName)
		}
	}
	//获取负载信息
	loadInfo, err := load.Avg()
	if err != nil {
		log.Errorf("get load avg info failed: %s from %s", err.Error(), tmpName)
		return
	}
	loadMisInfo, err := load.Misc()
	if err != nil {
		log.Errorf("get load misc info failed: %s from %s", err.Error(), tmpName)
		return
	}
	//获取内存信息
	swapMemInfo, err := mem.SwapMemory()
	if err != nil {
		log.Errorf("get swap mem info failed: %s from %s", err.Error(), tmpName)
		return
	}
	virtualMemInfo, err := mem.VirtualMemory()
	if err != nil {
		log.Errorf("get virtual mem info failed: %s from %s", err.Error(), tmpName)
		return
	}
	//获取磁盘信息
	var readCount, writeCount, readBytes, writeBytes, readTime, writeTime, ioTime, weightedIo uint64 = 0, 0, 0, 0, 0, 0, 0, 0
	var diskTotal, diskUsed, diskFree, inodeTotal, inodeUsed, inodeFree uint64 = 0, 0, 0, 0, 0, 0
	partitionInfo, _ := disk.Partitions(false)
	diskIoInfo, _ := disk.IOCounters()
	for _, v := range partitionInfo {
		if strings.Contains(v.Device, "/dev/sd") {
			name := strings.ReplaceAll(v.Device, "/dev/", "") //v.Device[5 : len(v.Device)-1]
			dio := diskIoInfo[name]
			space, _ := disk.Usage(v.Mountpoint)
			diskTotal = diskTotal + space.Total
			diskUsed = diskUsed + space.Used
			diskFree = diskFree + space.Free
			inodeTotal = inodeTotal + space.InodesTotal
			inodeUsed = inodeUsed + space.InodesUsed
			inodeFree = inodeFree + space.InodesFree
			readCount = readCount + dio.ReadCount
			writeCount = writeCount + dio.WriteCount
			readBytes = readBytes + dio.ReadBytes
			writeBytes = writeBytes + dio.WriteBytes
			readTime = readTime + dio.ReadTime
			writeTime = writeTime + dio.WriteTime
			ioTime = ioTime + dio.IoTime
			weightedIo = weightedIo + dio.WeightedIO
			go collectDiskIndicator(name, v.Mountpoint, dio, space)
		}
	}
	//查询上次累加值
	var readCountAcc, writeCountAcc, readBytesAcc, writeBytesAcc, readTimeAcc, writeTimeAcc, ioTimeAcc, weightedIoAcc uint64 = 0, 0, 0, 0, 0, 0, 0, 0
	err = db.QueryRow("select net_bytes_rev,net_bytes_send,net_package_rev,net_package_send,net_drop_rev,net_drop_send,net_error_rev,net_error_send from net_record where host_ip = ? and `name` = ?", tmpIp, constants.DISK_IO_TOTAL).Scan(&readCountAcc, &writeCountAcc, &readBytesAcc, &writeBytesAcc, &readTimeAcc, &writeTimeAcc, &ioTimeAcc, &weightedIoAcc)
	if err != nil {
		if strings.Contains(err.Error(), constants.NO_ROWS_IN_DB) {
			stmt, err := db.Prepare("insert into net_record (name,host_ip,net_bytes_rev,net_bytes_send,net_package_rev,net_package_send,net_drop_rev,net_drop_send,net_error_rev,net_error_send) values (?,?,?,?,?,?,?,?,?,?)")
			defer stmt.Close()
			if err != nil {
				log.Errorf("prepare insert net_record of disk failed: %s from %s", err.Error(), tmpName)
				return
			} else {
				_, err := stmt.Exec(constants.DISK_IO_TOTAL, tmpIp, readCount, writeCount, readBytes, writeBytes, readTime, writeTime, ioTime, weightedIo)
				if err != nil {
					log.Errorf("formal net_record of disk failed: %s from %s", err.Error(), tmpName)
				}
			}
		} else {
			log.Errorf("get disk total last record failed: %s from %s", err.Error(), tmpName)
			return
		}
	}
	//更新累加值
	stmt, err = db.Prepare("update net_record set net_bytes_rev = ?, net_bytes_send = ?, net_package_rev = ?, net_package_send = ?, net_drop_rev = ?, net_drop_send = ?, net_error_rev = ?, net_error_send = ? where host_ip = ? and `name` = ?")
	defer stmt.Close()
	if err != nil {
		log.Errorf("prepare update net monitor_disk_history disk io total failed: %s from %s", err.Error(), tmpName)
		return
	}
	_, err = stmt.Exec(readCountAcc, writeCountAcc, readBytesAcc, writeBytesAcc, readTimeAcc, writeTimeAcc, ioTimeAcc, weightedIoAcc, tmpIp, constants.DISK_IO_TOTAL)
	if err != nil {
		log.Errorf("update monitor_disk_history disk io total failed: %s from %s", err.Error(), tmpName)
		return
	}

	//获取网络信息
	netInfo, err := net.IOCounters(false)
	if err != nil {
		log.Errorf("get net info failed: %s from %s", err.Error(), tmpName)
		return
	}
	//先查询上次累加值
	var netBytesRev, netBytesSend, netPackageRev, netPackageSend, netDropRev, netDropSend, netErrorRev, netErrorSend uint64 = 0, 0, 0, 0, 0, 0, 0, 0
	err = db.QueryRow("select net_bytes_rev,net_bytes_send,net_package_rev,net_package_send,net_drop_rev,net_drop_send,net_error_rev,net_error_send from net_record where host_ip = ? and `name` = ?", tmpIp, constants.NET).Scan(&netBytesRev, &netBytesSend, &netPackageRev, &netPackageSend, &netDropRev, &netDropSend, &netErrorRev, &netErrorSend)
	if err != nil {
		if strings.Contains(err.Error(), constants.NO_ROWS_IN_DB) {
			stmt, err := db.Prepare("insert into net_record (net_bytes_rev,net_bytes_send,net_package_rev,net_package_send,net_drop_rev,net_drop_send,net_error_rev,net_error_send,host_ip,name) values (?,?,?,?,?,?,?,?,?,?)")
			defer stmt.Close()
			if err != nil {
				log.Errorf("prepare insert net_record of net failed: %s from %s", err.Error(), tmpIp)
				return
			} else {
				_, err := stmt.Exec(netInfo[0].BytesRecv, netInfo[0].BytesSent, netInfo[0].PacketsRecv, netInfo[0].PacketsSent, netInfo[0].Dropin, netInfo[0].Dropout, netInfo[0].Errin, netInfo[0].Errout, tmpIp, constants.NET)
				if err != nil {
					log.Errorf("formal insert net_record of net failed: %s from %s", err.Error(), tmpName)
				}
			}
		} else {
			log.Errorf("get net last record failed: %s from %s", err.Error(), tmpName)
			return
		}
	}
	//网络指标是累加值，所以需要记录每次的累加值
	stmt, err = db.Prepare("update net_record set net_bytes_rev = ?, net_bytes_send = ?, net_package_rev = ?, net_package_send = ?, net_drop_rev = ?, net_drop_send = ?, net_error_rev = ?, net_error_send = ? where host_ip = ? and `name` = ?")
	defer stmt.Close()
	if err != nil {
		log.Errorf("prepare add net record failed: %s from %s", err.Error(), tmpName)
		return
	}
	_, err = stmt.Exec(netInfo[0].BytesRecv, netInfo[0].BytesSent, netInfo[0].PacketsRecv, netInfo[0].PacketsSent, netInfo[0].Dropin, netInfo[0].Dropout, netInfo[0].Errin, netInfo[0].Errout, tmpIp, constants.NET)
	if err != nil {
		log.Errorf("add net record failed: %s from %s", err.Error(), tmpName)
		return
	}
	//记录本次机器指标信息
	param := make(map[string]interface{})
	param["hostName"] = tmpName
	param["hostIp"] = tmpIp
	param["procs"] = hostInfos.Procs
	param["cpuUser"] = cpuInfos[0].User - cpuUser
	param["cpuSys"] = cpuInfos[0].System - cpuSys
	param["cpuIdle"] = cpuInfos[0].Idle - cpuIdle
	param["cpuIowait"] = cpuInfos[0].Iowait - cpuIOwait
	param["cpuIrq"] = cpuInfos[0].Irq - cpuIrq
	param["cpuSofirg"] = cpuInfos[0].Softirq - cpuSofirq
	param["load1"] = loadInfo.Load1
	param["load5"] = loadInfo.Load5
	param["load15"] = loadInfo.Load15
	param["loadProcessTotal"] = loadMisInfo.ProcsTotal
	param["loadProcessRun"] = loadMisInfo.ProcsRunning
	param["memSwapTotal"] = swapMemInfo.Total
	param["memSwapUsed"] = swapMemInfo.Used
	param["memSwapFree"] = swapMemInfo.Free
	param["memSwapPercent"] = swapMemInfo.UsedPercent
	param["memVtotal"] = virtualMemInfo.Total
	param["memVused"] = virtualMemInfo.Used
	param["memVfree"] = virtualMemInfo.Free
	param["memVpercent"] = virtualMemInfo.UsedPercent
	param["netTrafficRev"] = (netInfo[0].BytesRecv - netBytesRev) / 300
	param["netTrafficSent"] = (netInfo[0].BytesSent - netBytesSend) / 300
	param["netPackageRev"] = (netInfo[0].PacketsRecv - netPackageRev) / 300
	param["netPackageSent"] = (netInfo[0].PacketsSent - netPackageSend) / 300
	param["netDropRev"] = (netInfo[0].Dropin - netDropRev) / 300

	param["netDropSent"] = (netInfo[0].Dropout - netDropSend) / 30
	param["netErrorRev"] = (netInfo[0].Errin - netErrorRev) / 300
	param["netErrorSent"] = (netInfo[0].Errout - netErrorSend) / 300

	param["diskReadCount"] = (readCount - readCountAcc) / 300
	param["diskWriteCount"] = (writeCount - writeCountAcc) / 300
	param["diskReadBytes"] = (readBytes - readBytesAcc) / 300

	param["diskWriteBytes"] = (writeBytes - writeBytesAcc) / 300
	param["diskReadTime"] = (readTime - readTimeAcc) / 300
	param["diskWriteTime"] = (writeTime - writeTimeAcc) / 300
	param["ioTime"] = (ioTime - ioTimeAcc) / 300
	param["diskTotal"] = diskTotal

	param["diskUsed"] = diskUsed
	param["diskFree"] = diskFree
	param["inodeTotal"] = inodeTotal
	param["inodeUsed"] = inodeUsed
	param["inodeFree"] = inodeFree

	stp := time.Now().Unix()
	param["dateTime"] = stp

	//b, _ := json.Marshal(param)

	postRequest("http://116.62.222.211:9090/hostIndicator", param)

	defer stmt.Close()

}
func RandInt64(min, max int64) int64 {
	if min > max {
		//panic("the min is greater than max!")
	}

	if min < 0 {
		f64Min := math.Abs(float64(min))
		i64Min := int64(f64Min)
		result, _ := rand.Int(rand.Reader, big.NewInt(max+1+i64Min))

		return result.Int64() - i64Min
	} else {
		result, _ := rand.Int(rand.Reader, big.NewInt(max-min+1))
		return min + result.Int64()
	}
	//maxBigInt := big.NewInt(max)
	//i, _ := rand.Int(rand.Reader, maxBigInt)
	//if i.Int64() < min {
	//	RandInt64(min, max)
	//}
	//return i.Int64()
}
func ExecuteTask(backupJson utils.BackupJson, c utils.Conf) {
	var ch chan int
	//定时任务
	//取4：30 -5：30 内的随机时间
	cTime := RandInt64(270, 330)
	log.Info("ticker is ", cTime)
	ticker := time.NewTicker(time.Second * time.Duration(cTime))

	go func() {
		for range ticker.C {
			collectJob()
			MainLogSync(backupJson, c)
		}
		ch <- 1
	}()
	<-ch
}

type MainLog struct {
	time     string
	newBlock string
	localIp  string
	error    string
	hostName string
	ipfs     string
	smanager string
	addr     string
}

var mainLog MainLog
var count = -1

func crustTask(mainLog MainLog) {

	workLoad := GetWorkLoad()

	crustStatus := getCrustStatus()
	if strings.Contains(workLoad, "files") {
		res, err := simplejson.NewJson([]byte(workLoad))
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}

		files := res.Get("files")
		srd := res.Get("srd")

		//fmt.Println(files)

		param := make(map[string]interface{})
		dateTime := time.Now().Unix()
		param["time"] = dateTime
		param["newBlock"] = getNewBlock()
		param["localIp"] = local.GetLocal().Ip
		param["error"] = mainLog.error
		param["hostName"] = local.GetLocal().HostName
		param["ipfs"] = mainLog.ipfs
		param["smanager"] = mainLog.smanager
		param["addr"] = mainLog.addr
		param["filesLost"] = files.Get("lost").Get("num").MustInt()
		param["filesPeeding"] = files.Get("pending").Get("num").MustInt()
		param["filesVaild"] = files.Get("valid").Get("num").MustInt()
		param["srdComplete"] = srd.Get("srd_complete").MustInt()
		param["srdRemainingTask"] = srd.Get("srd_remaining_task").MustInt()
		param["DiskAVA4Srd"] = srd.Get("disk_available_for_srd").MustInt()
		param["apiStatus"] = crustStatus.Api
		param["chainStatus"] = crustStatus.Chain
		param["sworkerStatus"] = crustStatus.Sworker
		param["smanagerStatus"] = crustStatus.Smanager
		param["ipfsStatus"] = crustStatus.Ipfs
		request := postRequest("http://116.62.222.211:9090/indicator", param)
		if strings.Contains(request, "chain-reload") {
			cmdAction("crust reload chain")
		}

	} else {

		param := make(map[string]interface{})
		dateTime := time.Now().Unix()
		param["time"] = dateTime
		param["newBlock"] = getNewBlock()
		param["localIp"] = local.GetLocal().Ip
		param["error"] = mainLog.error
		param["hostName"] = local.GetLocal().HostName
		if strings.Contains(mainLog.ipfs, "dis") && strings.Contains(mainLog.smanager, "dis") {
			mainLog.smanager = "owner"
		}
		param["smanager"] = mainLog.smanager

		param["ipfs"] = mainLog.ipfs
		param["addr"] = mainLog.addr
		param["filesLost"] = nil        //files.Get("lost").Get("num").MustInt()
		param["filesPeeding"] = nil     //files.Get("pending").Get("num").MustInt()
		param["filesVaild"] = nil       //files.Get("valid").Get("num").MustInt()
		param["srdComplete"] = nil      //srd.Get("srd_complete").MustInt()
		param["srdRemainingTask"] = nil //srd.Get("srd_remaining_task").MustInt()
		param["DiskAVA4Srd"] = nil      //srd.Get("disk_available_for_srd").MustInt()
		param["apiStatus"] = crustStatus.Api
		param["chainStatus"] = crustStatus.Chain
		param["sworkerStatus"] = crustStatus.Sworker
		param["smanagerStatus"] = crustStatus.Smanager
		param["ipfsStatus"] = crustStatus.Ipfs
		postRequest("http://116.62.222.211:9090/indicator", param)

	}

}

//获取当前块
func getNewBlock() int {

	s := sChainAction("block/header", "-XGET")
	fmt.Println(s)
	res, err := simplejson.NewJson([]byte(s))
	if err != nil {
	}
	return res.Get("number").MustInt()

}

//获取sworker信息

func GetWorkLoad() string {
	s := sWorkAction("workload", "-XGET")
	index := strings.Index(s, "{")
	if index != -1 {
		s = s[index:]
	}
	//fmt.Println(s)
	return s
}

func sWorkAction(apiPath string, action string) string {
	baseUrl := "http://127.0.0.1:12222/api/v0/"
	cmd := "curl -s " + action + " " + baseUrl + apiPath
	return cmdAction(cmd)
}

func sChainAction(apiPath string, action string) string {
	baseUrl := "http://127.0.0.1:56666/api/v1/"
	cmd := "curl -s " + action + " " + baseUrl + apiPath
	return cmdAction(cmd)
}

type CrustStatus struct {
	Chain    string
	Api      string
	Sworker  string
	Smanager string
	Ipfs     string
}

func getCrustStatus() CrustStatus {
	var c CrustStatus
	//快速检查
	quicklyCheck := "crust status |grep running |wc -l"
	allStatus := cmdAction(quicklyCheck)

	if allStatus == "5" {
		c.Sworker = "running"
		c.Chain = "running"
		c.Smanager = "running"
		c.Ipfs = "running"
		c.Api = "running"
		return c
	}

	chain := "crust status  |grep chain|awk '{print $2}'"
	api := "crust status  |grep api|awk '{print $2}'"
	sworker := "crust status  |grep sworker|awk '{print $2}'"
	smanager := "crust status  |grep smanager|awk '{print $2}'"
	ipfs := "crust status  |grep ipfs|awk '{print $2}'"

	chainStatus := cmdAction(chain)
	apiStatus := cmdAction(api)
	sworkerStatus := cmdAction(sworker)
	smanagerStatus := cmdAction(smanager)
	ipfsStatus := cmdAction(ipfs)
	fmt.Println(chainStatus + "," + apiStatus + "," + sworkerStatus + "," + smanagerStatus + "," + ipfsStatus)

	c.Chain = chainStatus       //strings.ReplaceAll(strings.TrimSpace(chainStatus),"chain","")
	c.Api = apiStatus           //strings.ReplaceAll(strings.TrimSpace(apiStatus),"api","")
	c.Smanager = smanagerStatus //strings.ReplaceAll(strings.TrimSpace(smanagerStatus),"smanager","")
	if strings.Contains(sworkerStatus, "run") {
		c.Sworker = "running"
	} else {
		c.Sworker = "error exit!"
	}
	//c.Sworker = sworkerStatus//strings.ReplaceAll(strings.TrimSpace(sworkerStatus),"sworker","")
	c.Ipfs = ipfsStatus //strings.ReplaceAll(strings.TrimSpace(ipfsStatus),"ipfs","")

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recover...:", r)
			er.ErrorHandler(r.(string))
		}
	}()
	return c
}

func postRequest(url string, param map[string]interface{}) string {
	resp, err := HttpRequest.SetHeaders(map[string]string{
		"Authorization": "NDM1MjU1ZTRiYjgwZTRiOTg4ZTY5N2I2ZTU4MDk5ZTg4M2JkZTU4OGIwMzUzMDMw",
		"Content-Type":  "application/json",
	}).Post(url, param)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Close()

	body, err := resp.Body()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(body))
	return string(body)
}

func getRequest(url string) string {
	resp, err := HttpRequest.SetHeaders(map[string]string{
		"Authorization": "NDM1MjU1ZTRiYjgwZTRiOTg4ZTY5N2I2ZTU4MDk5ZTg4M2JkZTU4OGIwMzUzMDMw",
	}).Get(url)
	if err != nil {

	}
	defer resp.Close()
	body, err := resp.Body()
	if err != nil {
		return ""
	}
	return string(body)
}

func cmdAction(cmd string) string {
	pac := ca.ProcessAgentCheck{
		BinPath: "/bin/sh",
	}
	err, s := pac.ExecCmd4String(cmd)
	if err != nil {
	}
	return s
}

func MainLogSync(backupJson utils.BackupJson, c utils.Conf) {
	defer func() {
		if r := recover(); r != nil {
			//fmt.Println("recover...:", r)
			er.ErrorHandler(r.(string))
		}
	}()

	mainLog.hostName = local.GetLocal().HostName
	mainLog.localIp = local.GetLocal().Ip
	mainLog.ipfs = c.Node.Ipfs
	mainLog.smanager = c.Node.Smanager
	mainLog.addr = backupJson.Address

	health := ci.CheckHealth()
	healthCount, _ := strconv.Atoi(health)
	mainLog.error = strconv.Itoa(5 - healthCount)

	crustTask(mainLog)

}
