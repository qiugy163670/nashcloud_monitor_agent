package tail

import (
	"encoding/json"
)

type Log struct {
	Log    string `json:"log"`
	Stream string `json:"stream"`
	Time   string `json:"time"`
}

func Json2Struct(jsonStr string) Log {
	//	jsonStr := `
	//{
	//	"log": "[2021-08-04 05:53:50] \u001b[32minfo\u001b[39m: ⛓  Got new block 2715180(0xad391f2c1d47a5f68637de50a9be989fde0a6c6052bef48218f0b361e6aa2fcd)\n",
	//	"stream": "stdout",
	//	"time": "2021-08-04T05:53:50.751753157Z"
	//}`
	var log Log
	json.Unmarshal([]byte(jsonStr), &log)
	//fmt.Println(log.Time)
	return log
}
