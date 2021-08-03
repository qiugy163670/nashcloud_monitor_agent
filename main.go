package main

import (
	"config"
	"agent"
)

func main() {
	config.InitDBConnection()
	agent.ExecuteTask()
}
