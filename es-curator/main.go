package main

import (
	"es-curator/global"
	"es-curator/job"
	"es-curator/utils"
	"es-curator/log_record"
	"sync"
	// "fmt"
)

func main() {
	utils.LoadEnvironment()
	job.SetElkClient()

	////	init logger
	log_record.InitLogger()
	log_record.InitDetailLogger()

	if global.EnvConfig.INFORMATION.Execute_cron {
		utils.LoadCrontab()
		wg := new(sync.WaitGroup)
		num := 1
		wg.Add(num)
		wg.Wait()
	} else if !global.EnvConfig.INFORMATION.Execute_cron {
		job.Action_controll()
	}

}





// func main1() {
// 	fmt.Println("start")
// 	utils.LoadEnvironment()
	
// 	fmt.Println(global.EnvConfig.ES.URL)
// 	job.SetElkClient()
// 	// job.Allocation1([]string{"logstash-bimap-test01"})
// 	// job.CatNodes1()
// 	// job.Nodetest()
// 	job.NodeRoleDetermination("w")
// 	// job.CatIndices()
// 	// Action_test()
// 	job.CatShards()
// 	// job.Test111()
// 	// job.Allocation([]string{"logstash-bimap-test01"},"include","_tier_preference","data_hot")
// 	// job.ListTest()
// 	// job.Action_controll()
// 	// job.CatIndices()
// 	// job.CatNodes()
// 	// job.CatIndices_withPattern([]string{"logstash-atm-sn-20231223"})
// }
