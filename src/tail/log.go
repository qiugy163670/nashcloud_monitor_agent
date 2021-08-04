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
	var log Log
	json.Unmarshal([]byte(jsonStr), &log)
	//fmt.Println(log.Time)
	return log
}
