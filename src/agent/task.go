package agent

import (
	"fmt"
	log "github.com/cihub/seelog"
	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"nashcloud_monitor_agent/src/config"
	"nashcloud_monitor_agent/src/utils"
	"os"
)

func Init() {
	config.SetupLogger()
}
func CollectJob() {
	collectJob()
}
func collectJob() {
	Init()
	defer log.Flush()
	fmt.Println("collectJob")
	tmpName, _ := os.Hostname()

	//获取机器信息
	hostInfos, err := host.Info()
	if err != nil {
		log.Error("get host info failed: %s from %s", err.Error(), tmpName)
		return
	}
	//查询该机器是否存在，不存在注册
	db, err := config.GetDBConnection()
	if err != nil {
		log.Error("get db connection failed: %s from %s", err.Error(), tmpName)
		return
	}
	var id, dateTime, hostName, hostIp, os, platform, company, platformVersion, kernelVersion string
	err = db.QueryRow("select * from nash_servers where `host_name` = ?", hostInfos.Hostname).Scan(&id, &hostName, &hostIp, &os, &platform, &platformVersion, &kernelVersion, &company, &dateTime)
	if err != nil {
		log.Error("query host info failed: %s from %s", err.Error(), tmpName)
		return
	}
	tmpIp := utils.GetHostIp()
	if hostIp != tmpIp || os != hostInfos.OS || platform != hostInfos.Platform || platformVersion != hostInfos.PlatformVersion || kernelVersion != hostInfos.KernelVersion {
		stmt, err := db.Prepare("update nash_servers set host_ip = ?, os = ?, platform = ?, platform_version = ?, kernel_version = ? where `host_name` = ?")
		if err != nil {
			log.Error("prepare update host info failed: %s from %s", err.Error(), tmpName)
		} else {
			_, err := stmt.Exec(tmpIp, hostInfos.OS, hostInfos.Platform, hostInfos.PlatformVersion, hostInfos.KernelVersion, hostInfos.Hostname)
			if err != nil {
				log.Error("formal update host info failed: %s from %s", err.Error(), tmpName)
			}
		}
	}

	//获取cpu信息
	cpuInfos, err := cpu.Times(false)
	if err != nil {
		log.Error("get cpu info failed: %s from %s", err.Error(), tmpName)
		return
	}
	//获取负载信息
	loadInfo, err := load.Avg()
	if err != nil {
		log.Error("get load avg info failed: %s from %s", err.Error(), tmpName)
		return
	}
	loadMisInfo, err := load.Misc()
	if err != nil {
		log.Error("get load misc info failed: %s from %s", err.Error(), tmpName)
		return
	}
	//获取内存信息
	swapMemInfo, err := mem.SwapMemory()
	if err != nil {
		log.Error("get swap mem info failed: %s from %s", err.Error(), tmpName)
		return
	}
	virtualMemInfo, err := mem.VirtualMemory()
	if err != nil {
		log.Error("get virtual mem info failed: %s from %s", err.Error(), tmpName)
		return
	}
	//获取磁盘信息
	disk.IOCounters()
	disk.Partitions(false)

	//获取网络信息
	netInfo, err := net.IOCounters(false)
	if err != nil {
		log.Error("get net info failed: %s from %s", err.Error(), tmpName)
		return
	}
	//先查询上次累加值
	var netBytesRev, netBytesSend, netPackageRev, netPackageSend, netDropRev, netDropSend, netErrorRev, netErrorSend uint64
	db.QueryRow("select * from net_record limit 1").Scan(netBytesRev, netBytesSend, netPackageRev, netPackageSend, netDropRev, netDropSend, netErrorRev, netErrorSend)
	//网络指标是累加值，所以需要记录每次的累加值
	//stmt, err := db.Prepare("update net_record set net_bytes_rev = ?, net_bytes_send = ?, net_package_rev = ?, net_package_send = ?, net_drop_rev = ?, net_drop_send = ?, net_error_rev, net_error_send")
	//if err != nil {
	//	log.Error("prepare add host indicator failed: %s from %s", err.Error(), tmpName)
	//	return
	//}
	//fmt.Println(netInfo[0].Errout)
	//_, err = stmt.Exec(netInfo[0].BytesRecv, netInfo[0].BytesSent, netInfo[0].PacketsRecv, netInfo[0].PacketsSent, netInfo[0].Dropin, netInfo[0].Dropout, netInfo[0].Errin, netInfo[0].Errout)
	//if err != nil {
	//	log.Error("prepare add host indicator failed: %s from %s", err.Error(), tmpName)
	//	return
	//}
	//记录本次机器指标信息
	stmt, err := db.Prepare("insert into monitor_host_indicator (host_name, host_ip, procs, cpu_user, cpu_sys, cpu_idle, cpu_iowait, cpu_irq, cpu_sofirg, load1, load5, load15, load_process_total, load_process_run, mem_swap_total, mem_swap_used, mem_swap_free, mem_swap_percent, mem_vtotal, mem_vused, mem_vfree, mem_vpercent, net_traffic_rev, net_traffic_sent, net_drop_rev, net_drop_sent, net_error_rev, net_error_sent) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		log.Error("prepare add host indicator failed: %s from %s", err.Error(), tmpName)
		return
	}
	_, err = stmt.Exec(hostName, tmpIp, hostInfos.Procs,
		cpuInfos[0].User, cpuInfos[0].System, cpuInfos[0].Idle, cpuInfos[0].Iowait, cpuInfos[0].Irq, cpuInfos[0].Softirq,
		loadInfo.Load1, loadInfo.Load5, loadInfo.Load15, loadMisInfo.ProcsTotal, loadMisInfo.ProcsRunning,
		swapMemInfo.Total, swapMemInfo.Used, swapMemInfo.Free, swapMemInfo.UsedPercent, virtualMemInfo.Total, virtualMemInfo.Used, virtualMemInfo.Free, virtualMemInfo.UsedPercent,
		(netInfo[0].BytesRecv-netBytesRev)/300, (netInfo[0].BytesSent-netBytesSend)/300, 100*(netInfo[0].Dropin-netDropRev)/(netInfo[0].PacketsRecv-netPackageRev), 100*(netInfo[0].Dropout-netDropSend)/(netInfo[0].PacketsSent-netPackageSend), 100*(netInfo[0].Errin-netErrorRev)/(netInfo[0].PacketsRecv-netPackageRev), 100*(netInfo[0].Errout-netErrorSend)/(netInfo[0].PacketsSent-netPackageSend))
	if err != nil {
		log.Error("formal add host indicator failed: %s from %s", err.Error(), tmpName)
	}
}

func ExecuteTask() {
	fmt.Println("task")
	c := cron.New()

	//AddFunc
	spec := "0 */5 * * * ?"

	//AddJob方法
	c.AddFunc(spec, collectJob)

	//启动计划任务
	c.Start()

	//关闭着计划任务, 但是不能关闭已经在执行中的任务.

	//select {}
}
