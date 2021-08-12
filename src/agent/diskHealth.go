package agent

import (
	"container/list"
	"fmt"
	log "github.com/cihub/seelog"
	ca "nashcloud_monitor_agent/src/cmd"
	"nashcloud_monitor_agent/src/config"
	"nashcloud_monitor_agent/src/utils"
	"strings"
)

func GetAndPushDiskInfo() {
	disks := getDisks()

	db, err := config.GetDBConnection()
	if err != nil {
		err := log.Errorf("get db connection failed: %s from %s", err.Error())
		if err != nil {
			return
		}
		return
	}

	if disks.Len() > 2 {
		for disk := disks.Front(); disk != nil; disk = disk.Next() {
			strs := strings.Split(disk.Value.(string), " ")
			sn := getDiskSN(strs[1])
			stmt, err := db.Prepare("insert into monitor_disk_check (diskMountName,diskSN,diskVal,host_ip) values (?,?,?,?)")
			if err != nil {
				log.Errorf("prepare add disk partition detail failed: %s from %s", err.Error())
				return
			}
			_, err = stmt.Exec(strs[1], sn, strs[0], utils.GetHostIp())
			if err != nil {
			}
		}
	} else {
		fmt.Println(disks)
	}
}

func getDisks() list.List {
	pac := ca.ProcessAgentCheck{
		BinPath: "/bin/sh",
	}
	err, list := pac.ExecCmd("cat /proc/partitions |grep 'sd[a-z]'|awk '{print $3,$4}'")
	if err != nil {
		fmt.Println(err)
	}
	return list
}

//func getDiskSN(path string) string {
//	pac := ca.ProcessAgentCheck{
//		BinPath: "/bin/sh",
//	}
//	cmd := "hdparm  -I /dev/" + path + " |grep 'Serial Number'"
//	//fmt.Println(cmd)
//	err, list := pac.ExecCmd(cmd)
//	if err != nil {
//		fmt.Println(err)
//	}
//	return strings.ReplaceAll(list.Front().Value.(string), "Serial Number:", "")
//}
