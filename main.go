package main

import (
	"fmt"
	"github.com/robfig/cron/v3"
)

func main() {
	Task()
	crontab := cron.New(cron.WithSeconds())
	task := func() {
		Task()
	}
	spec := "0 0 1 * * ?"
	_, err := crontab.AddFunc(spec, task)
	if err != nil {
		fmt.Println("定时任务添加失败", err)
	}
	crontab.Start()
	select {}
}
