package init

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"nashcloud_monitor_agent/src/cmd"
	"os"
	"sync"
)

var gLocker sync.Mutex
var gCondition *sync.Cond

var origin = "http://wwww.nashcloud.cn/"
var url = "ws://nashcloud.cn:8080/agentServer/172.28.101.12"

type CmdJson struct {
	MsgType    string `json:"MsgType"`
	MsgContent string `json:"MsgContent"`
}

func Json2Struct(jsonStr string) CmdJson {
	var cmdJson CmdJson
	json.Unmarshal([]byte(jsonStr), &cmdJson)
	return cmdJson
}
func checkErr(err error, extra string) bool {
	if err != nil {
		formatStr := " Err : %s\n"
		if extra != "" {
			formatStr = extra + formatStr
		}
		fmt.Fprintf(os.Stderr, formatStr, err.Error())
		return true
	}
	return false
}

func Close(conn *websocket.Conn) {
	conn.Close()
}
func clientConnHandler(conn *websocket.Conn) {
	gLocker.Lock()
	defer gLocker.Unlock()
	defer conn.Close()
	request := make([]byte, 128)
	for {
		readLen, err := conn.Read(request)
		if checkErr(err, "Read") {
			gCondition.Signal()
			break
		}
		if readLen == 0 {
			fmt.Println("Server connection close!")
			gCondition.Signal()
			break
		} else {
			s := string(request[:readLen])
			fmt.Println(s)
			cmdJson := Json2Struct(s)
			if cmdJson.MsgType == "cmd" {
				fmt.Println(cmdJson.MsgContent)
				s2 := cmd.RemoteExec(cmdJson.MsgContent)
				conn.Write([]byte("{\"MsgType\":\"res\",\"MsgContent\":\"" + s2 + "\"}"))
			}
		}
		request = make([]byte, 128)
	}
}

func Conn() {
	conn, err := websocket.Dial(url, "", origin)
	if checkErr(err, "Dial") {
		return
	}

	gLocker.Lock()
	gCondition = sync.NewCond(&gLocker)
	_, err = conn.Write([]byte("{\"msg\":\"hello\"}"))
	go clientConnHandler(conn)

	for {
		gCondition.Wait()
		break
	}

	gLocker.Unlock()
	fmt.Println("Client finish.")
}
