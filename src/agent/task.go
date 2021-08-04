package agent

import (
	"config"
	log "github.com/cihub/seelog"
	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"os"
	"utils"
)

func Init() {
	config.SetupLogger()
}

func collectJob() {
	Init()
	defer log.Flush()

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
	var hostName, hostIp, os, platform, platformVersion, kernelVersion string
	err = db.QueryRow("select * from nash_servers where `name` = ?", hostInfos.Hostname).Scan(&hostName, &hostIp, &os, &platform, &platformVersion, &kernelVersion)
	if err != nil {
		log.Error("query host info failed: %s from %s", err.Error(), tmpName)
		return
	}
	tmpIp := utils.GetHostIp()
	if hostIp != tmpIp || os != hostInfos.OS || platform != hostInfos.Platform || platformVersion != hostInfos.PlatformVersion || kernelVersion != hostInfos.KernelVersion {
		stmt, err := db.Prepare("update nash_servers set host_ip = ?, os = ?, platform = ?, platform_version = ?, kernel_version = ? where `name` = ?")
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

	//获取网络信息
	//net.Addr{}

	//获取进程信息
	//process

	//记录本次机器指标信息
	stmt, err := db.Prepare("insert into monitor_host_indicator (host_name, host_ip, cpu_user, cpu_sys, cpu_idle, cpu_iowait, cpu_irq, cpu_sofirg, load1, load5, load15, load_process_total, load_process_run, mem_swap_total, mem_swap_used, mem_swap_free, mem_swap_percent, mem_vtotal, _mem_vused, mem_vfree, mem_vpercent) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		log.Error("prepare add host indicator failed: %s from %s", err.Error(), tmpName)
		return
	}
	_, err = stmt.Exec(hostName, tmpIp,
		cpuInfos[0].User, cpuInfos[0].System, cpuInfos[0].Idle, cpuInfos[0].Iowait, cpuInfos[0].Irq, cpuInfos[0].Softirq,
		loadInfo.Load1, loadInfo.Load5, loadInfo.Load15, loadMisInfo.ProcsTotal, loadMisInfo.ProcsRunning,
		swapMemInfo.Total, swapMemInfo.Used, swapMemInfo.Free, swapMemInfo.UsedPercent, virtualMemInfo.Total, virtualMemInfo.Used, virtualMemInfo.Free, virtualMemInfo.UsedPercent,
	)
	if err != nil {
		log.Error("formal add host indicator failed: %s from %s", err.Error(), tmpName)
	}
}

func ExecuteTask() {
	c := cron.New()

	//AddFunc
	spec := "0 */5 * * * ?"

	//AddJob方法
	c.AddFunc(spec, collectJob)

	//启动计划任务
	c.Start()

	//关闭着计划任务, 但是不能关闭已经在执行中的任务.
	defer c.Stop()

	select {}
}
