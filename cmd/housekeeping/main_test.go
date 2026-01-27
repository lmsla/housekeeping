package main

import (
	"housekeeping/internal/global"
	"housekeeping/internal/job"
	"housekeeping/internal/log_record"
	"housekeeping/internal/utils"
	// "sync"
	"testing"
	// "fmt"

)

func TestMain(t *testing.T) {
	utils.LoadEnvironment()

	//// init logger
	log_record.InitLogger()
	log_record.InitDetailLogger()
	log_record.InitStderrLogger()

	// 初始化 Elasticsearch 客戶端
	if err := job.SetElkClient(); err != nil {

		global.Logger.Error(err)
		global.Stderr_logger.Fatalf("初始化 Elasticsearch 客戶端失敗: %v", err)
	}
	job.CatNodes()
	// job.SetElkClient()
	// job.NodeRoleDetermination("h")
	// if global.EnvConfig.INFORMATION.ExecuteCron {
	// 	utils.LoadCrontab()
	// 	wg := new(sync.WaitGroup)
	// 	num := 1
	// 	wg.Add(num)
	// 	wg.Wait()
	// } else if !global.EnvConfig.INFORMATION.ExecuteCron {
	// 	job.Action_controll()
	// }

	// var indices_on_node []string

	// nodeNames := job.NodeRoleDetermination("h")
	// for _, node := range nodeNames {
	// 	indices_on_node = append(indices_on_node, job.CatIndicesbyNodeName(node)...)
	// }
	// fmt.Println(indices_on_node)
	// job.CatShardsbyNodeName("es04")
	// job.FilterType_space_role("es05",[]string{},3)

}


	// 使用不同的函式來記錄不同日誌等級的訊息
	// log.Info("CCUCSIE Plus is running")
	// log.Warn("Some warning occurred")
	// log.Error("Some error occurred")
	// log.Debug("Debug message")