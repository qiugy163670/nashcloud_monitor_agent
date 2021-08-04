package agent

import (
	"github.com/cihub/seelog"
	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"nashcloud_monitor_agent_sync/src/config"
)

func Init() {
	config.SetupLogger()
}

func collectJob() {
	Init()
	defer seelog.Flush()

	//获取机器信息
	hostInfos, err := host.Info()
	if err != nil {

	}
	//查询该机器是否存在，不存在注册
	db, err := config.GetDBConnection()
	if err != nil {

	}
	var hostName, os, platform, platformVersion, kernelVersion string
	err = db.QueryRow("select * from nash_servers where `name` = ?", hostInfos.Hostname).Scan(&hostName, &os, &platform, &platformVersion, &kernelVersion)
	if err != nil {

	}
	if hostName != hostInfos.Hostname || os != hostInfos.OS || platform != hostInfos.Platform || platformVersion != hostInfos.PlatformVersion || kernelVersion != hostInfos.KernelVersion {
		stmt, err := db.Prepare("update nash_servers set os = ?, platform = ?, platform_version = ?, kernel_version = ? where `name` = ?")
		if err != nil {

		}
		stmt.Exec(hostInfos.OS, hostInfos.Platform, hostInfos.PlatformVersion, hostInfos.KernelVersion, hostInfos.Hostname)
	}

	//获取cpu信息
	cpuInfos, err := cpu.Times(false)
	if err != nil {

	}
	print(cpuInfos[0])

	//获取负载信息
	load.Avg()
	load.Misc()

	//获取内存信息
	mem.SwapMemory()

	//获取磁盘信息
	//disk.Usage()

	//获取网络信息
	//net.Addr{}

	//获取进程信息
	//process
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
