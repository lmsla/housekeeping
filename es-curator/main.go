package main

import (
	"es-curator/job"
	"es-curator/utils"
	"es-curator/global"
	"sync"
)

func main1() {
	utils.LoadEnvironment()
	job.SetElkClient()

	if global.EnvConfig.INFORMATION.Execute_cron == true {
		utils.LoadCrontab()
		wg := new(sync.WaitGroup)
		num := 1
		wg.Add(num)
		wg.Wait()
	}else if global.EnvConfig.INFORMATION.Execute_cron == false {
		job.Action_controll()
	}



}


func main() {
	job.Test111()
}