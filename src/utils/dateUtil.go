package utils

import (
	"time"
)

func UTCTransLocal(utcTime string) string {
	t, _ := time.Parse(time.RFC3339, utcTime)
	return t.In(time.Local).Format("2006-01-02 15:04:05")

}
