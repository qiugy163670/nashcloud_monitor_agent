package agent

import (
	"github.com/robfig/cron"
	"time"
	"log"
)

func collectCPUJob()  {

}

func ExecuteTask() {
	c := cron.New()

	//AddFunc
	spec := "*/5 * * * * ?"
	c.AddFunc(spec, func() {
		log.Println("cron running:", time.Now())
	})

	//AddJob方法
	c.AddJob(spec, collectCPUJob{})
	c.AddJob(spec, collectMemJob{})

	//启动计划任务
	c.Start()

	//关闭着计划任务, 但是不能关闭已经在执行中的任务.
	defer c.Stop()

	select{}
}