# 快速開始指南

## 安裝

```bash
go get github.com/bimap-org/universal-monitoring@latest
```

## 基本使用

### 1. 建立監控配置檔案

```yaml
# universal-metrics.yml
service:
  name: "my-service"
  type: "web-api"
  version: "1.0.0"
  environment: "development"

metrics:
  enabled: true
  collection_interval: 30s
  summary_interval: 5m

universal_modules:
  resource_monitoring:
    enabled: true
    collection_interval: 10s
  error_tracking:
    enabled: true
    max_recent_errors: 100
  dependency_health:
    enabled: true
    check_interval: 30s

custom_operations:
  - name: "process_user"
    display_name: "處理用戶資料"
    type: "timer"

exporters:
  console:
    enabled: true
    format: "json"
    level: "info"
```

### 2. 在代碼中整合

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/bimap-org/universal-monitoring/pkg/config"
    "github.com/bimap-org/universal-monitoring/pkg/metrics"
)

func main() {
    // 載入配置
    cfg, err := config.LoadFromFile("universal-metrics.yml")
    if err != nil {
        log.Fatal(err)
    }
    
    // 初始化監控收集器
    collector, err := metrics.NewUniversalCollector(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer collector.Close()
    
    // 添加依賴監控
    collector.AddDependency("database", "mysql", "localhost:3306")
    
    // 啟動監控
    ctx := context.Background()
    go collector.Start(ctx)
    
    // 你的業務邏輯
    processUser(collector)
}

func processUser(collector *metrics.UniversalMetrics) {
    startTime := time.Now()
    
    // 模擬業務邏輯
    time.Sleep(100 * time.Millisecond)
    success := true
    
    // 記錄操作
    duration := time.Since(startTime)
    collector.RecordOperation("process_user", success, duration)
    
    // 記錄錯誤 (如果有的話)
    if !success {
        collector.RecordError("user_processing", "處理失敗", "error")
    }
}
```

## 支援的服務類型

### Web API 服務
```go
cfg := config.LoadWebAPIDefaults()
cfg.Service.Name = "my-api"
```

### 批次處理服務  
```go
cfg := config.LoadBatchJobDefaults()
cfg.Service.Name = "my-batch-job"
```

## 查看監控數據

運行程式後，會看到類似的監控輸出：

```
=== my-service 服務指標摘要 ===
服務類型: web-api | 版本: 1.0.0 | 運行時間: 1m30s

📊 操作統計:
  處理用戶資料: 總數=100, 成功=98 (98.0%), 平均時間=125.50ms
  整體成功率: 98.00% (98/100)

🔗 依賴狀態:
  database (mysql): connected | 響應時間: 45ms | 成功率: 100.0%

💾 資源使用:
  記憶體: 25.3 MB | 堆: 12.8 MB | Goroutines: 15
```

## 下一步

- [配置參考](configuration-guide.md) - 了解詳細的配置選項
- [API 文檔](api-reference.md) - 查看完整的 API 介面  
- [範例](../examples/) - 查看更多使用範例