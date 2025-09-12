# Universal Service Monitoring

一個通用的 Go 服務監控指標收集系統，基於 BiMAP Housekeeping 的監控實踐開發。

## 特性

- 🔄 **通用性**: 支援多種服務類型 (Web API, 批次處理, 消息處理)
- 📊 **多重匯出**: 支援 Console, Prometheus, InfluxDB 等多種匯出格式
- 🔌 **插件化**: 可擴展的健康檢查和自定義指標插件
- ⚙️ **配置驅動**: 透過 YAML 配置檔案定義監控行為
- 🎯 **動態切換**: 支援運行時動態啟用/停用匯出器
- 🔔 **智能警報**: 內建閾值檢查和多渠道通知

## 快速開始

### 安裝

```bash
go get github.com/bimap-org/universal-monitoring@latest
```

### 基本使用

```go
package main

import (
    "context"
    "time"
    
    "github.com/bimap-org/universal-monitoring/pkg/config"
    "github.com/bimap-org/universal-monitoring/pkg/metrics"
)

func main() {
    // 載入配置
    cfg, err := config.LoadFromFile("universal-metrics.yml")
    if err != nil {
        panic(err)
    }
    
    // 初始化監控收集器
    collector, err := metrics.NewUniversalCollector(cfg)
    if err != nil {
        panic(err)
    }
    defer collector.Close()
    
    // 啟動監控
    ctx := context.Background()
    go collector.Start(ctx)
    
    // 記錄操作
    collector.RecordOperation("process_data", true, 250*time.Millisecond)
    
    // 記錄錯誤
    collector.RecordError("validation_error", "Invalid input", "warning")
    
    // 你的業務邏輯...
}
```

## 文檔

- [快速開始指南](docs/getting-started.md)
- [配置參考](docs/configuration-guide.md)
- [API 文檔](docs/api-reference.md)
- [使用範例](examples/)

## 貢獻

歡迎提交 Issue 和 Pull Request！

## 許可證

MIT License