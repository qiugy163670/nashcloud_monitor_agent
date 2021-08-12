package cmd

import (
	"container/list"
	log "github.com/cihub/seelog"
)

func RemoteExec(cmdStr string) string {
	list, err := remoteExec(cmdStr)
	if err != nil {
		return err.Error()
	}
	jsonStr := "{\"strs\":["
	for str := list.Front(); str != nil; str = str.Next() {
		jsonStr = jsonStr + "\"" + str.Value.(string) + "\","
	}
	jsonStr = jsonStr + "\"end\"]}"
	return jsonStr
}

func remoteExec(cmdStr string) (list.List, error) {
	pac := ProcessAgentCheck{
		BinPath: "/bin/sh",
	}
	err, res := pac.ExecCmd(cmdStr)
	if err != nil {
		log.Info("error: ", err)
	}
	return res, err
}
