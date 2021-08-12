package agent

import (
	"fmt"
	log "github.com/cihub/seelog"
	_ "github.com/go-sql-driver/mysql"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"

	ca "nashcloud_monitor_agent/src/cmd"
	"nashcloud_monitor_agent/src/config"
	"nashcloud_monitor_agent/src/constants"
	"nashcloud_monitor_agent/src/utils"
	"os"
	"strings"
	"time"
)

func Init() {
	config.SetupLogger()
}

func getDiskSN(path string) string {
	pac := ca.ProcessAgentCheck{
		BinPath: "/bin/sh",
	}
	cmd := "hdparm  -I /dev/" + path + " |grep 'Serial Number'"
	//fmt.Println(cmd)
	err, list := pac.ExecCmd(cmd)
	if err != nil {
		fmt.Println(err)
	}
	return strings.ReplaceAll(list.Front().Value.(string), "Serial Number:", "")
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
	if err != nil {
		log.Errorf("prepare add disk partition detail failed: %s from %s, %s", err.Error(), tmpName, tmpIp)
		return
	}
	_, err = stmt.Exec(name, tmpIp, tmpName, "/dev/"+name, mount, getDiskSN(name), stat.Total, stat.Used, stat.Free, stat.InodesTotal, stat.InodesUsed, stat.InodesFree, diskIoInfo.ReadCount-readCountAcc, diskIoInfo.WriteCount-writeCountAcc, diskIoInfo.ReadBytes-readBytesAcc, diskIoInfo.WriteBytes-writeBytesAcc, diskIoInfo.ReadTime-readTimeAcc, diskIoInfo.WriteTime-writeTimeAcc, diskIoInfo.IoTime-ioTimeAcc, diskIoInfo.WeightedIO-weightedIoAcc, dateTime-dateTime%300)
	if err != nil {
		log.Errorf("add disk partition detail failed: %s from %s, %s", err.Error(), tmpName, tmpIp)
		return
	}
}

func CollectJob() {
	collectJob()
}

func collectJob() {
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
	var hostName, hostIp, os, platform, platformVersion, kernelVersion string
	err = db.QueryRow("select host_name, host_ip, os, platform, platform_version, kernel_version from nash_servers where `host_ip` = ?", utils.GetHostIp()).Scan(&hostName, &hostIp, &os, &platform, &platformVersion, &kernelVersion)
	if err != nil {
		if strings.Contains(err.Error(), constants.NO_ROWS_IN_DB) {
			stmt, err := db.Prepare("insert into nash_servers (host_name, host_ip, os, platform, platform_version, kernel_version, company) values (?,?,?,?,?,?,?)")
			if err != nil {
				log.Errorf("prepare insert host info failed: %s from %s", err.Error(), tmpName)
				return
			} else {
				_, err := stmt.Exec(hostInfos.Hostname, hostIp, hostInfos.OS, hostInfos.Platform, hostInfos.PlatformVersion, hostInfos.KernelVersion, constants.NASH_CLOUD)
				if err != nil {
					log.Errorf("formal insert host info failed: %s from %s", err.Error(), tmpName)
				}
			}
		} else {
			log.Errorf("query host info failed: %s from %s", err.Error(), tmpName)
			return
		}
	}
	tmpIp := utils.GetHostIp()
	if hostIp != tmpIp || os != hostInfos.OS || platform != hostInfos.Platform || platformVersion != hostInfos.PlatformVersion || kernelVersion != hostInfos.KernelVersion {
		stmt, err := db.Prepare("update nash_servers set host_ip = ?, os = ?, platform = ?, platform_version = ?, kernel_version = ? where `host_name` = ?")
		if err != nil {
			log.Errorf("prepare update host info failed: %s from %s", err.Error(), tmpName)
		} else {
			_, err := stmt.Exec(tmpIp, hostInfos.OS, hostInfos.Platform, hostInfos.PlatformVersion, hostInfos.KernelVersion, hostInfos.Hostname)
			if err != nil {
				log.Errorf("formal update host info failed: %s from %s", err.Error(), tmpName)
			}
		}
	}

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
		name := v.Device[5 : len(v.Device)-1]
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
	//查询上次累加值
	var readCountAcc, writeCountAcc, readBytesAcc, writeBytesAcc, readTimeAcc, writeTimeAcc, ioTimeAcc, weightedIoAcc uint64 = 0, 0, 0, 0, 0, 0, 0, 0
	err = db.QueryRow("select net_bytes_rev,net_bytes_send,net_package_rev,net_package_send,net_drop_rev,net_drop_send,net_error_rev,net_error_send from net_record where host_ip = ? and `name` = ?", tmpIp, constants.DISK_IO_TOTAL).Scan(&readCountAcc, &writeCountAcc, &readBytesAcc, &writeBytesAcc, &readTimeAcc, &writeTimeAcc, &ioTimeAcc, &weightedIoAcc)
	if err != nil {
		if strings.Contains(err.Error(), constants.NO_ROWS_IN_DB) {
			stmt, err := db.Prepare("insert into net_record (name,host_ip,net_bytes_rev,net_bytes_send,net_package_rev,net_package_send,net_drop_rev,net_drop_send,net_error_rev,net_error_send) values (?,?,?,?,?,?,?,?,?,?)")
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

	stp := time.Now().Unix()
	stmt, err = db.Prepare("insert into monitor_host_indicator (host_name, host_ip, procs, cpu_user, cpu_sys, cpu_idle, cpu_iowait, cpu_irq, cpu_sofirg, load1, load5, load15, load_process_total, load_process_run, mem_swap_total, mem_swap_used, mem_swap_free, mem_swap_percent, mem_vtotal, mem_vused, mem_vfree, mem_vpercent, net_traffic_rev, net_traffic_sent, net_package_rev, net_package_sent, net_drop_rev, net_drop_sent, net_error_rev, net_error_sent, disk_read_count, disk_write_count, disk_read_bytes, disk_write_bytes, disk_read_time, disk_write_time, io_time, disk_total, disk_used, disk_free, inode_total, inode_used, inode_free, date_time) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		log.Errorf("prepare add all indicator failed: %s from %s", err.Error(), tmpName)
		return
	}

	_, err = stmt.Exec(tmpName, tmpIp, hostInfos.Procs,
		cpuInfos[0].User-cpuUser, cpuInfos[0].System-cpuSys, cpuInfos[0].Idle-cpuIdle, cpuInfos[0].Iowait-cpuIOwait, cpuInfos[0].Irq-cpuIrq, cpuInfos[0].Softirq-cpuSofirq,
		loadInfo.Load1, loadInfo.Load5, loadInfo.Load15, loadMisInfo.ProcsTotal, loadMisInfo.ProcsRunning,
		swapMemInfo.Total, swapMemInfo.Used, swapMemInfo.Free, swapMemInfo.UsedPercent, virtualMemInfo.Total, virtualMemInfo.Used, virtualMemInfo.Free, virtualMemInfo.UsedPercent,
		(netInfo[0].BytesRecv-netBytesRev)/300, (netInfo[0].BytesSent-netBytesSend)/300, (netInfo[0].PacketsRecv-netPackageRev)/300, (netInfo[0].PacketsSent-netPackageSend)/300, (netInfo[0].Dropin-netDropRev)/300, (netInfo[0].Dropout-netDropSend)/300, (netInfo[0].Errin-netErrorRev)/300, (netInfo[0].Errout-netErrorSend)/300,
		(readCount-readCountAcc)/300, (writeCount-writeCountAcc)/300, (readBytes-readBytesAcc)/300, (writeBytes-writeBytesAcc)/300, (readTime-readTimeAcc)/300, (writeTime-writeTimeAcc)/300, (ioTime-ioTimeAcc)/300,
		diskTotal, diskUsed, diskFree, inodeTotal, inodeUsed, inodeFree, stp-stp%300)
	if err != nil {
		log.Errorf("formal add host indicator failed: %s from %s", err.Error(), tmpName)
	}
}

func ExecuteTask() {
	var ch chan int
	//定时任务
	ticker := time.NewTicker(time.Minute * 5)
	go func() {
		for range ticker.C {
			collectJob()
		}
		ch <- 1
	}()
	<-ch
}
