package main

import (
	"es-curator/global"
	"es-curator/job"
	"es-curator/utils"
	"sync"
)

func main() {
	utils.LoadEnvironment()
	job.SetElkClient()

	if global.EnvConfig.INFORMATION.Execute_cron == true {
		utils.LoadCrontab()
		wg := new(sync.WaitGroup)
		num := 1
		wg.Add(num)
		wg.Wait()
	} else if global.EnvConfig.INFORMATION.Execute_cron == false {
		job.Action_controll()
	}

}

func main1() {
	utils.LoadEnvironment()
	// job.SetElkClient()
	// job.Allocation1([]string{"logstash-bimap-test01"})
	// job.CatNodes()
	// job.CatIndices()
	// Action_test()
	// job.Nodetest()
	// job.Test111()
	// job.Allocation([]string{"logstash-bimap-test01"},"include","_tier_preference","data_hot")
	// job.GetdataInAction()
}

