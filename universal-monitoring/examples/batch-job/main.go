package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/bimap-org/universal-monitoring/pkg/config"
	"github.com/bimap-org/universal-monitoring/pkg/metrics"
)

func main() {
	fmt.Println("🚀 Universal Monitoring - Batch Job 範例")

	// 1. 載入配置 (使用預設配置)
	cfg := config.LoadBatchJobDefaults()
	cfg.Service.Name = "example-batch-job"
	cfg.Service.Version = "1.0.0"

	// 2. 初始化通用監控收集器
	collector, err := metrics.NewUniversalCollector(cfg)
	if err != nil {
		log.Fatalf("Failed to create collector: %v", err)
	}
	defer collector.Close()

	// 3. 註冊自定義業務操作
	collector.RegisterOperation("process_data", "資料處理")
	collector.RegisterOperation("send_report", "發送報告")
	collector.RegisterOperation("cleanup_temp", "清理臨時檔案")

	// 4. 添加依賴監控
	collector.AddDependency("database", "mysql", "10.99.1.135:3306")
	collector.AddDependency("file_storage", "filesystem", "/data")

	// 5. 啟動監控
	ctx := context.Background()
	go collector.Start(ctx)

	fmt.Println("✅ 監控系統已啟動")

	// 6. 模擬批次任務執行
	simulateBatchJob(collector)

	// 7. 輸出最終摘要
	fmt.Println("\n📊 任務執行完成，正在生成摘要...")
	time.Sleep(1 * time.Second)
	collector.PrintSummary()
}

// simulateBatchJob 模擬批次任務執行
func simulateBatchJob(collector *metrics.UniversalMetrics) {
	fmt.Println("\n🔄 開始執行批次任務...")

	// 模擬資料處理
	fmt.Println("📈 階段 1: 資料處理")
	for i := 1; i <= 5; i++ {
		startTime := time.Now()

		// 模擬處理時間
		processingTime := time.Duration(rand.Intn(500)+100) * time.Millisecond
		time.Sleep(processingTime)

		// 模擬偶爾的失敗
		success := rand.Float32() > 0.1 // 90% 成功率
		if !success {
			collector.RecordError("data_processing_error", fmt.Sprintf("處理批次 %d 失敗", i), "warning")
		}

		// 記錄操作結果
		duration := time.Since(startTime)
		collector.RecordOperation("process_data", success, duration)

		fmt.Printf("  批次 %d: %v (耗時: %v)\n", i, 
			map[bool]string{true: "✅ 成功", false: "❌ 失敗"}[success], 
			duration.Truncate(time.Millisecond))

		// 模擬依賴健康狀態變化
		if i == 3 {
			// 模擬資料庫連線問題
			collector.RecordDependencyHealth("database", false, 2*time.Second, "timeout")
		} else {
			collector.RecordDependencyHealth("database", true, 50*time.Millisecond, "connected")
		}
	}

	// 模擬發送報告
	fmt.Println("\n📧 階段 2: 發送報告")
	for i := 1; i <= 2; i++ {
		startTime := time.Now()
		time.Sleep(time.Duration(rand.Intn(200)+50) * time.Millisecond)
		
		success := rand.Float32() > 0.05 // 95% 成功率
		duration := time.Since(startTime)
		collector.RecordOperation("send_report", success, duration)

		fmt.Printf("  報告 %d: %v (耗時: %v)\n", i, 
			map[bool]string{true: "✅ 已發送", false: "❌ 發送失敗"}[success], 
			duration.Truncate(time.Millisecond))
	}

	// 模擬清理作業
	fmt.Println("\n🧹 階段 3: 清理臨時檔案")
	startTime := time.Now()
	time.Sleep(100 * time.Millisecond)
	duration := time.Since(startTime)
	collector.RecordOperation("cleanup_temp", true, duration)
	fmt.Printf("  清理作業: ✅ 完成 (耗時: %v)\n", duration.Truncate(time.Millisecond))

	// 更新自定義指標
	collector.UpdateCustomMetric("processed_records", 1000)
	collector.UpdateCustomMetric("generated_reports", 2)
	collector.UpdateCustomMetric("cleaned_files", 15)

	fmt.Println("✅ 批次任務執行完成")
}