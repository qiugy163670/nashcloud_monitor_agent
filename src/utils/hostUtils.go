package utils

import (
	"config"
	"net"
	"os"

	log "github.com/cihub/seelog"
)

var hostIp string = ""

func Init() {
	config.SetupLogger()
}

func GetHostIp() string {
	if hostIp != "" {
		return hostIp
	}
	Init()
	defer log.Flush()

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		hostName, _ := os.Hostname()
		log.Error("get host ip failed: %s from %s", err.Error(), hostName)
		return ""
	}

	for _, address := range addrs {

		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				hostIp = ipnet.IP.String()
				return hostIp
			}
		}
	}
	return ""
}
