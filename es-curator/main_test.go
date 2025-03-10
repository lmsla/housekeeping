package main

import (
	"es-curator/global"
	"es-curator/job"
	"es-curator/log_record"
	"es-curator/utils"
	"sync"
	"testing"

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
	// job.SetElkClient()

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


	// 使用不同的函式來記錄不同日誌等級的訊息
	// log.Info("CCUCSIE Plus is running")
	// log.Warn("Some warning occurred")
	// log.Error("Some error occurred")
	// log.Debug("Debug message")