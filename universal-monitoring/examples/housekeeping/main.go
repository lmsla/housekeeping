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

// 模擬 BiMAP Housekeeping 的整合範例
func main() {
	fmt.Println("🏠 Universal Monitoring - BiMAP Housekeeping 整合範例")

	// 1. 載入監控配置
	cfg, err := config.LoadFromFile("universal-metrics.yml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化通用監控收集器
	universalMetrics, err := metrics.NewUniversalCollector(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize universal metrics: %v", err)
	}
	defer universalMetrics.Close()

	// 3. 註冊 ES 依賴
	universalMetrics.AddDependency("elasticsearch", "search_engine", "https://es.example.com:9200")

	// 4. 啟動監控收集
	ctx := context.Background()
	go universalMetrics.Start(ctx)

	fmt.Println("✅ 通用監控系統已啟動")

	// 5. 模擬 housekeeping 操作
	simulateHousekeepingOperations(universalMetrics)

	// 6. 輸出最終摘要
	fmt.Println("\n📊 Housekeeping 任務完成，正在生成監控摘要...")
	time.Sleep(1 * time.Second)
	universalMetrics.PrintSummary()
}

// simulateHousekeepingOperations 模擬 housekeeping 操作
func simulateHousekeepingOperations(metrics *metrics.UniversalMetrics) {
	fmt.Println("\n🔧 開始執行 Housekeeping 操作...")

	// 模擬 ES 健康檢查
	fmt.Println("🔍 檢查 Elasticsearch 連線狀態")
	esHealthy := true
	responseTime := 163 * time.Millisecond
	metrics.RecordDependencyHealth("elasticsearch", esHealthy, responseTime, "green")
	fmt.Printf("  ES 狀態: %v | 響應時間: %v\n", 
		map[bool]string{true: "✅ 連線正常", false: "❌ 連線失敗"}[esHealthy], responseTime)

	// 模擬索引操作
	operations := []struct {
		name        string
		displayName string
		count       int
		successRate float32
	}{
		{"allocation", "索引分配", 3, 0.95},
		{"delete_indices", "刪除索引", 2, 1.0},
		{"forcemerge", "強制合併", 1, 1.0},
		{"close", "關閉索引", 1, 1.0},
	}

	for _, op := range operations {
		fmt.Printf("\n📋 執行 %s 操作:\n", op.displayName)

		for i := 1; i <= op.count; i++ {
			startTime := time.Now()

			// 模擬操作執行時間 (根據操作類型調整)
			var operationTime time.Duration
			switch op.name {
			case "allocation":
				operationTime = time.Duration(rand.Intn(2000)+500) * time.Millisecond // 0.5-2.5秒
			case "delete_indices":
				operationTime = time.Duration(rand.Intn(1000)+200) * time.Millisecond // 0.2-1.2秒
			case "forcemerge":
				operationTime = time.Duration(rand.Intn(5000)+2000) * time.Millisecond // 2-7秒
			default:
				operationTime = time.Duration(rand.Intn(500)+100) * time.Millisecond // 0.1-0.6秒
			}

			time.Sleep(operationTime)

			// 根據成功率決定是否成功
			success := rand.Float32() < op.successRate
			if !success {
				errorMsg := fmt.Sprintf("%s 操作第 %d 次執行失敗", op.displayName, i)
				metrics.RecordError("operation_failed", errorMsg, "warning")
			}

			// 記錄操作結果
			duration := time.Since(startTime)
			metrics.RecordOperation(op.name, success, duration)

			// 模擬索引名稱
			indexName := fmt.Sprintf("logstash-%s-%d", 
				[]string{"ubuntu", "windows", "forti"}[rand.Intn(3)], 
				20250100+rand.Intn(10))

			fmt.Printf("  索引 %s: %v (耗時: %v)\n", indexName,
				map[bool]string{true: "✅ 成功", false: "❌ 失敗"}[success],
				duration.Truncate(time.Millisecond))

			// 小間隔
			time.Sleep(100 * time.Millisecond)
		}
	}

	// 更新自定義指標
	metrics.UpdateCustomMetric("processed_indices", 7)
	metrics.UpdateCustomMetric("freed_space_gb", 15.7)
	metrics.UpdateCustomMetric("execution_duration_min", 3.2)

	fmt.Println("\n✅ Housekeeping 操作執行完成")
}