package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"es-curator/global"
	"es-curator/job"
	"es-curator/log_record"
	localmetrics "es-curator/metrics"
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
	
	// 創建可取消的 context 來優雅關閉
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 初始化舊的本地指標收集器
	localmetrics.InitMetrics()
	global.Logger.Info("本地指標收集器已初始化")
	
	
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

	// 設置信號處理，支援優雅關閉
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)


	job.CatCluster(ctx)

	// 啟動資源指標週期更新 (每 30 秒)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if localmetrics.GlobalMetrics != nil {
					localmetrics.GlobalMetrics.UpdateResourceMetrics()
				}
			}
		}
	}()

	if global.EnvConfig.INFORMATION.ExecuteCron {
		utils.LoadCrontab()
		// 等待信號或無限等待
		go func() {
			<-sigChan
			global.Logger.Info("Received shutdown signal, gracefully shutting down...")
			printMetricsSummary()
			cancel()
		}()
		// 等待取消信號
		<-ctx.Done()
		global.Logger.Info("Application shutdown completed")
	} else if !global.EnvConfig.INFORMATION.ExecuteCron {
		job.Action_controll()
		// 執行完畢後顯示指標摘要
		printMetricsSummary()
	}
}

// printMetricsSummary 輸出指標摘要
func printMetricsSummary() {
	// 保留舊的本地指標摘要
	
	// 保留舊的本地指標摘要（向後兼容）
	if localmetrics.GlobalMetrics == nil {
		return
	}
	
	global.Logger.Info("=== 執行指標摘要 ===")
	
	// 操作統計
	opMetrics := localmetrics.GlobalMetrics.GetOperationMetrics()
	global.Logger.Infow("操作統計", 
		"總操作數", opMetrics.TotalOperations,
		"成功操作數", opMetrics.SuccessfulOps, 
		"失敗操作數", opMetrics.FailedOps)
	
	// 各類型操作統計
	if opMetrics.OperationsByType != nil {
		if opMetrics.OperationsByType.Delete.Total > 0 {
			global.Logger.Infow("刪除操作", 
				"總數", opMetrics.OperationsByType.Delete.Total,
				"成功", opMetrics.OperationsByType.Delete.Success,
				"失敗", opMetrics.OperationsByType.Delete.Failed,
				"平均時間ms", opMetrics.OperationsByType.Delete.AvgTime)
		}
		if opMetrics.OperationsByType.Allocation.Total > 0 {
			global.Logger.Infow("分配操作", 
				"總數", opMetrics.OperationsByType.Allocation.Total,
				"成功", opMetrics.OperationsByType.Allocation.Success,
				"失敗", opMetrics.OperationsByType.Allocation.Failed,
				"平均時間ms", opMetrics.OperationsByType.Allocation.AvgTime)
		}
		if opMetrics.OperationsByType.ForceMerge.Total > 0 {
			global.Logger.Infow("強制合併操作", 
				"總數", opMetrics.OperationsByType.ForceMerge.Total,
				"成功", opMetrics.OperationsByType.ForceMerge.Success,
				"失敗", opMetrics.OperationsByType.ForceMerge.Failed,
				"平均時間ms", opMetrics.OperationsByType.ForceMerge.AvgTime)
		}
		if opMetrics.OperationsByType.Close.Total > 0 {
			global.Logger.Infow("關閉操作", 
				"總數", opMetrics.OperationsByType.Close.Total,
				"成功", opMetrics.OperationsByType.Close.Success,
				"失敗", opMetrics.OperationsByType.Close.Failed,
				"平均時間ms", opMetrics.OperationsByType.Close.AvgTime)
		}
		if opMetrics.OperationsByType.Open.Total > 0 {
			global.Logger.Infow("開啟操作", 
				"總數", opMetrics.OperationsByType.Open.Total,
				"成功", opMetrics.OperationsByType.Open.Success,
				"失敗", opMetrics.OperationsByType.Open.Failed,
				"平均時間ms", opMetrics.OperationsByType.Open.AvgTime)
		}
		if opMetrics.OperationsByType.Rollover.Total > 0 {
			global.Logger.Infow("Rollover操作", 
				"總數", opMetrics.OperationsByType.Rollover.Total,
				"成功", opMetrics.OperationsByType.Rollover.Success,
				"失敗", opMetrics.OperationsByType.Rollover.Failed,
				"平均時間ms", opMetrics.OperationsByType.Rollover.AvgTime)
		}
	}
	
	// ES 健康狀態
	esHealth := localmetrics.GlobalMetrics.GetESHealthMetrics()
	global.Logger.Infow("ES 連線狀態", 
		"連線狀態", esHealth.ConnectionStatus,
		"叢集健康", esHealth.ClusterHealth,
		"響應時間ms", esHealth.ResponseTime.Nanoseconds()/1e6,
		"連線嘗試次數", esHealth.ConnectionAttempts,
		"成功連線次數", esHealth.SuccessfulConnections,
		"失敗連線次數", esHealth.FailedConnections)
	
	// 資源使用
	resourceMetrics := localmetrics.GlobalMetrics.GetResourceMetrics()
	global.Logger.Infow("資源使用狀況", 
		"記憶體使用MB", resourceMetrics.MemoryUsage/1024/1024,
		"堆記憶體MB", resourceMetrics.HeapInuse/1024/1024,
		"Goroutine數", resourceMetrics.GoroutineCount)
}

