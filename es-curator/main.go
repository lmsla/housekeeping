package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"es-curator/global"
	"es-curator/job"
	"es-curator/log_record"
	"es-curator/utils"
)

func main() {
	// 載入環境配置
	if err := utils.LoadEnvironment(); err != nil {
		fmt.Printf("Failed to load environment: %v\n", err)
		os.Exit(1)
	}

	// init logger
	log_record.InitLogger()
	log_record.InitDetailLogger()
	log_record.InitStderrLogger()
	
	// 調試信息：檢查配置是否正確讀取
	fmt.Printf("=== 配置調試信息 ===\n")
	fmt.Printf("Test mode: %v\n", global.EnvConfig.INFORMATION.TestMode)
	fmt.Printf("Log path: %s\n", global.EnvConfig.Log.Path)
	fmt.Printf("Execute cron: %v\n", global.EnvConfig.INFORMATION.ExecuteCron)
	fmt.Printf("================\n")

	// 初始化 Elasticsearch 客戶端
	if err := job.SetElkClient(); err != nil {

		global.Logger.Error(err)
		global.Stderr_logger.Fatalf("初始化 Elasticsearch 客戶端失敗: %v", err)
	}

	// 創建可取消的 context 來優雅關閉
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 設置信號處理，支援優雅關閉
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)


	job.CatCluster(ctx)

	if global.EnvConfig.INFORMATION.ExecuteCron {
		utils.LoadCrontab()
		// 等待信號或無限等待
		go func() {
			<-sigChan
			global.Logger.Info("Received shutdown signal, gracefully shutting down...")
			cancel()
		}()
		select {} // 保持程序運行等待定時任務
	} else if !global.EnvConfig.INFORMATION.ExecuteCron {
		job.Action_controll()
	}
}

