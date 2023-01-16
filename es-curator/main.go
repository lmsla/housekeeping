package main

import (
	"es-curator/job"
	"es-curator/utils"
	"es-curator/global"
	"sync"
)

func main() {
	utils.LoadEnvironment()
	job.SetElkClient()
	// fmt.Println(es.Info())
	// job.OpenIndices()
	// job.CreateIndex()
	// job.CatIndices()
	// job.Action_delete_index()
	// job.Action_delete_index_pattern()
	// job.Test()
	// job.Testsum()
	// job.FilterType_space(3)
	// job.Action_delete_indices()
	// job.Action_open_indices()
	// job.Job1()
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
