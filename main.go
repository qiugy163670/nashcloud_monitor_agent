package main

import (
	"fmt"
	er "nashcloud_monitor_agent/src/error"
	_ "nashcloud_monitor_agent/src/init"
)

func main() {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recover...:", r)
			er.ErrorHandler(r.(string))
		}
	}()

}
