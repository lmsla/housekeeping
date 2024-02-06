package main

import (
	"es-curator/global"
	"es-curator/job"
	"es-curator/utils"
	"fmt"
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
	fmt.Println("L")
	fmt.Println(global.EnvConfig.ES.URL)
	// job.SetElkClient()
	// job.Allocation1([]string{"logstash-bimap-test01"})
	// job.CatNodes()
	// job.CatIndices()
	// Action_test()
	// job.Nodetest()
	// job.Test111()
	// job.Allocation([]string{"logstash-bimap-test01"},"include","_tier_preference","data_hot")
	// job.ListTest()
	// job.Action_controll()
	// job.CatIndices()
	// job.CatNodes()
	// job.CatIndices_withPattern([]string{"logstash-atm-sn-20231223"})
}
