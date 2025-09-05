package main

import (
	"es-curator/global"
	"es-curator/job"
	"es-curator/log_record"
	"es-curator/utils"
	// "log"
)

func main() {
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

	job.CatCluster()
	// job.SetElkClient()

	if global.EnvConfig.INFORMATION.Execute_cron {
		utils.LoadCrontab()
		select {} // 保持程序運行等待定時任務
	} else if !global.EnvConfig.INFORMATION.Execute_cron {
		job.Action_controll()
	}
}

// // 新增一個函數，用於檢查 Elasticsearch 集群的健康狀態
// func checkElasticsearchHealth() error {
// 	// 連接 Elasticsearch 集群
// 	esClient, err := elasticsearch.NewClient(elasticsearch.SetURL(global.EnvConfig.Elasticsearch.URL))
// 	if err != nil {
// 		return err
// 	}

// 	// 檢查 Elasticsearch 集群的健康狀態
// 	res, err := esClient.Cluster.Health()
// 	if err != nil {
// 		return err
// 	}

// 	// 檢查 Elasticsearch 集群的健康狀態是否為綠色
// 	if res.Status != "green" {
// 		return fmt.Errorf("Elasticsearch 集群的健康狀態為 %s，不是綠色", res.Status)
// 	}
// 	return nil
// }