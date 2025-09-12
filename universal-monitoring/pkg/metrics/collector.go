package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"time"
)

// NewUniversalCollector 建立通用指標收集器
func NewUniversalCollector(config *MetricsConfig) (*UniversalMetrics, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	um := &UniversalMetrics{
		ServiceName:    config.Service.Name,
		ServiceType:    config.Service.Type,
		ServiceVersion: config.Service.Version,
		StartTime:      time.Now(),
		Config:         config,

		// 初始化通用指標
		Resource: &ResourceMetrics{},
		Dependencies: make(map[string]*DependencyHealth),
		Errors: &ErrorMetrics{
			ErrorsByType:     make(map[string]*ErrorStats),
			ErrorsBySeverity: make(map[string]*ErrorStats),
			HTTPStatusCodes:  make(map[int]int64),
			RecentErrors:     make([]ErrorEvent, 0),
		},

		// 初始化自定義指標
		CustomMetrics:    make(map[string]interface{}),
		OperationMetrics: make(map[string]*OperationStats),
	}

	// 如果啟用請求指標模組
	if config.Service.Type == "web-api" {
		um.Requests = &RequestMetrics{
			EndpointStats: make(map[string]*EndpointMetrics),
		}
	}

	// 註冊配置中的自定義操作
	for _, opConfig := range config.Operations {
		um.RegisterOperation(opConfig.Name, opConfig.DisplayName)
	}

	// 初始化 Console 匯出器
	if config.Exporters.Console.Enabled {
		um.consoleExporter = NewConsoleExporter(config.Exporters.Console)
	}

	return um, nil
}

// Start 啟動指標收集
func (um *UniversalMetrics) Start(ctx context.Context) {
	if !um.Config.Metrics.Enabled {
		return
	}

	// 啟動資源監控
	if um.Config.Modules.ResourceMonitoring.Enabled {
		go um.startResourceMonitoring(ctx)
	}

	// 啟動依賴健康檢查
	if um.Config.Modules.DependencyHealth.Enabled {
		go um.startHealthChecking(ctx)
	}

	// 啟動定期摘要輸出
	if um.Config.Metrics.SummaryInterval > 0 {
		go um.startSummaryOutput(ctx)
	}
}

// startResourceMonitoring 啟動資源監控
func (um *UniversalMetrics) startResourceMonitoring(ctx context.Context) {
	ticker := time.NewTicker(um.Config.Modules.ResourceMonitoring.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			um.UpdateResourceMetrics()
			// 每次資源更新時也匯出指標（用於即時監控）
			um.ExportMetrics()
		}
	}
}

// startHealthChecking 啟動健康檢查
func (um *UniversalMetrics) startHealthChecking(ctx context.Context) {
	ticker := time.NewTicker(um.Config.Modules.DependencyHealth.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			um.checkAllDependencies()
		}
	}
}

// startSummaryOutput 啟動定期摘要輸出
func (um *UniversalMetrics) startSummaryOutput(ctx context.Context) {
	ticker := time.NewTicker(um.Config.Metrics.SummaryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			um.PrintSummary()
			um.ExportMetrics()
		}
	}
}

// UpdateResourceMetrics 更新資源指標
func (um *UniversalMetrics) UpdateResourceMetrics() {
	um.mu.Lock()
	defer um.mu.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	um.Resource.MemoryUsage = int64(m.Alloc)
	um.Resource.HeapInuse = int64(m.HeapInuse)
	um.Resource.HeapSys = int64(m.HeapSys)
	um.Resource.HeapObjects = m.HeapObjects
	um.Resource.GoroutineCount = runtime.NumGoroutine()
	um.Resource.GCPauseTime = int64(m.PauseNs[(m.NumGC+255)%256])
	um.Resource.GCRuns = m.NumGC
	um.Resource.LastUpdateTime = time.Now()
}

// RegisterOperation 註冊業務操作
func (um *UniversalMetrics) RegisterOperation(name, displayName string) {
	um.mu.Lock()
	defer um.mu.Unlock()

	if _, exists := um.OperationMetrics[name]; !exists {
		um.OperationMetrics[name] = &OperationStats{
			Name:        name,
			DisplayName: displayName,
		}
	}
}

// RecordOperation 記錄操作
func (um *UniversalMetrics) RecordOperation(operationName string, success bool, duration time.Duration) {
	um.mu.RLock()
	opStats, exists := um.OperationMetrics[operationName]
	um.mu.RUnlock()

	if !exists {
		// 自動註冊未知操作
		um.RegisterOperation(operationName, operationName)
		um.mu.RLock()
		opStats = um.OperationMetrics[operationName]
		um.mu.RUnlock()
	}

	opStats.mu.Lock()
	defer opStats.mu.Unlock()

	atomic.AddInt64(&opStats.Total, 1)
	atomic.AddInt64(&opStats.TotalTime, duration.Nanoseconds()/1e6) // 轉換為毫秒

	if success {
		atomic.AddInt64(&opStats.Success, 1)
	} else {
		atomic.AddInt64(&opStats.Failed, 1)
	}

	// 更新平均時間
	if opStats.Total > 0 {
		opStats.AvgTime = float64(opStats.TotalTime) / float64(opStats.Total)
	}
}

// RecordError 記錄錯誤
func (um *UniversalMetrics) RecordError(errorType, message, severity string) {
	if !um.Config.Modules.ErrorTracking.Enabled {
		return
	}

	um.Errors.mu.Lock()
	defer um.Errors.mu.Unlock()

	atomic.AddInt64(&um.Errors.TotalErrors, 1)

	// 按類型統計
	if _, exists := um.Errors.ErrorsByType[errorType]; !exists {
		um.Errors.ErrorsByType[errorType] = &ErrorStats{}
	}
	stats := um.Errors.ErrorsByType[errorType]
	stats.Count++
	stats.LastOccurred = time.Now()
	stats.Description = message
	if stats.FirstOccurred.IsZero() {
		stats.FirstOccurred = time.Now()
	}

	// 按嚴重程度統計
	if _, exists := um.Errors.ErrorsBySeverity[severity]; !exists {
		um.Errors.ErrorsBySeverity[severity] = &ErrorStats{}
	}
	severityStats := um.Errors.ErrorsBySeverity[severity]
	severityStats.Count++
	severityStats.LastOccurred = time.Now()
	if severityStats.FirstOccurred.IsZero() {
		severityStats.FirstOccurred = time.Now()
	}

	// 記錄最近錯誤
	event := ErrorEvent{
		Timestamp: time.Now(),
		Type:      errorType,
		Message:   message,
		Severity:  severity,
		Context:   make(map[string]interface{}),
	}

	um.Errors.RecentErrors = append(um.Errors.RecentErrors, event)

	// 保持最近錯誤數量限制
	maxRecentErrors := um.Config.Modules.ErrorTracking.MaxRecentErrors
	if len(um.Errors.RecentErrors) > maxRecentErrors {
		um.Errors.RecentErrors = um.Errors.RecentErrors[len(um.Errors.RecentErrors)-maxRecentErrors:]
	}

	um.Errors.LastErrorTime = time.Now()
}

// AddDependency 添加依賴
func (um *UniversalMetrics) AddDependency(name, depType, endpoint string) {
	um.mu.Lock()
	defer um.mu.Unlock()

	um.Dependencies[name] = &DependencyHealth{
		Name:     name,
		Type:     depType,
		Endpoint: endpoint,
		Status:   "unknown",
	}
}

// RecordDependencyHealth 記錄依賴健康狀態
func (um *UniversalMetrics) RecordDependencyHealth(name string, connected bool, responseTime time.Duration, status string) {
	um.mu.Lock()
	defer um.mu.Unlock()

	dep, exists := um.Dependencies[name]
	if !exists {
		return
	}

	atomic.AddInt64(&dep.ConnectionAttempts, 1)

	if connected {
		atomic.AddInt64(&dep.SuccessfulConnections, 1)
		dep.Status = status
		dep.LastSuccessTime = time.Now()
	} else {
		atomic.AddInt64(&dep.FailedConnections, 1)
		dep.Status = "disconnected"
		dep.LastFailureTime = time.Now()
	}

	dep.ResponseTime = responseTime
	dep.LastCheckTime = time.Now()

	// 計算平均響應時間
	if dep.SuccessfulConnections > 0 {
		// 簡化版平均計算，實際中可能需要更精確的算法
		if dep.AvgResponseTime == 0 {
			dep.AvgResponseTime = responseTime
		} else {
			dep.AvgResponseTime = (dep.AvgResponseTime + responseTime) / 2
		}
	}
}

// checkAllDependencies 檢查所有依賴的健康狀態
func (um *UniversalMetrics) checkAllDependencies() {
	// 這裡需要實際的健康檢查實現
	// 目前只是一個占位符，實際實現會在 health 套件中
	for _, dep := range um.Dependencies {
		// 模擬健康檢查
		_ = dep
		// TODO: 實際的健康檢查邏輯
	}
}

// UpdateCustomMetric 更新自定義指標
func (um *UniversalMetrics) UpdateCustomMetric(name string, value interface{}) {
	um.mu.Lock()
	defer um.mu.Unlock()

	um.CustomMetrics[name] = value
}

// GetAllMetrics 獲取所有指標
func (um *UniversalMetrics) GetAllMetrics() map[string]interface{} {
	um.mu.RLock()
	defer um.mu.RUnlock()

	result := map[string]interface{}{
		"service": map[string]interface{}{
			"name":         um.ServiceName,
			"type":         um.ServiceType,
			"version":      um.ServiceVersion,
			"start_time":   um.StartTime,
			"uptime":       time.Since(um.StartTime),
			"environment":  um.Config.Environment,
		},
		"resource":      um.Resource,
		"dependencies":  um.Dependencies,
		"errors":        um.Errors,
		"operations":    um.OperationMetrics,
		"custom":        um.CustomMetrics,
		"timestamp":     time.Now(),
	}

	if um.Requests != nil {
		result["requests"] = um.Requests
	}

	return result
}

// PrintSummary 輸出指標摘要
func (um *UniversalMetrics) PrintSummary() {
	fmt.Printf("\n=== %s 服務指標摘要 ===\n", um.ServiceName)
	fmt.Printf("服務類型: %s | 版本: %s | 運行時間: %v\n",
		um.ServiceType, um.ServiceVersion, time.Since(um.StartTime).Truncate(time.Second))

	// 操作統計
	fmt.Println("\n📊 操作統計:")
	totalOps := int64(0)
	totalSuccess := int64(0)
	for _, stats := range um.OperationMetrics {
		totalOps += stats.Total
		totalSuccess += stats.Success
		if stats.Total > 0 {
			successRate := float64(stats.Success) / float64(stats.Total) * 100
			fmt.Printf("  %s: 總數=%d, 成功=%d (%.1f%%), 平均時間=%.2fms\n",
				stats.DisplayName, stats.Total, stats.Success, successRate, stats.AvgTime)
		}
	}

	// 整體成功率
	if totalOps > 0 {
		overallRate := float64(totalSuccess) / float64(totalOps) * 100
		fmt.Printf("  整體成功率: %.2f%% (%d/%d)\n", overallRate, totalSuccess, totalOps)
	}

	// 依賴狀態
	fmt.Println("\n🔗 依賴狀態:")
	for depName, dep := range um.Dependencies {
		successRate := float64(0)
		if dep.ConnectionAttempts > 0 {
			successRate = float64(dep.SuccessfulConnections) / float64(dep.ConnectionAttempts) * 100
		}
		fmt.Printf("  %s (%s): %s | 響應時間: %v | 成功率: %.1f%%\n",
			depName, dep.Type, dep.Status, dep.ResponseTime.Truncate(time.Millisecond), successRate)
	}

	// 資源使用
	fmt.Println("\n💾 資源使用:")
	fmt.Printf("  記憶體: %.2f MB | 堆: %.2f MB | Goroutines: %d\n",
		float64(um.Resource.MemoryUsage)/1024/1024,
		float64(um.Resource.HeapInuse)/1024/1024,
		um.Resource.GoroutineCount)

	// 錯誤統計
	if um.Errors.TotalErrors > 0 {
		fmt.Println("\n❌ 錯誤統計:")
		fmt.Printf("  總錯誤數: %d | 最後錯誤時間: %v\n",
			um.Errors.TotalErrors, um.Errors.LastErrorTime.Format("15:04:05"))
	}

	fmt.Println("=" + fmt.Sprintf("%*s", 50, "="))
}

// ExportMetrics 匯出指標到配置的匯出器
func (um *UniversalMetrics) ExportMetrics() {
	if um.consoleExporter != nil {
		metricsData := um.GetAllMetrics()
		if err := um.consoleExporter.Export(metricsData); err != nil {
			fmt.Printf("Console export error: %v\n", err)
		}
	}
}

// NewConsoleExporter 創建 Console 匯出器
func NewConsoleExporter(config ConsoleExporterConfig) *ConsoleExporter {
	exporter := &ConsoleExporter{
		Config: config,
	}
	
	// 如果需要輸出到檔案，開啟檔案
	if config.Output == "file" || config.Output == "both" {
		// 確保目錄存在
		if err := os.MkdirAll(filepath.Dir(config.FilePath), 0755); err != nil {
			fmt.Printf("Failed to create directory for metrics file: %v\n", err)
		} else {
			file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				fmt.Printf("Failed to open metrics file %s: %v\n", config.FilePath, err)
			} else {
				exporter.file = file
			}
		}
	}
	
	return exporter
}

// Export 匯出指標
func (ce *ConsoleExporter) Export(metricsData map[string]interface{}) error {
	if !ce.Config.Enabled {
		return nil
	}
	
	var logEntry string
	
	switch ce.Config.Format {
	case "json":
		// 將時間戳加入 JSON 數據中
		metricsData["exported_at"] = time.Now().Format("2006-01-02T15:04:05Z07:00")
		// 使用 compact JSON (一行一個 JSON 物件，JSON Lines 格式)
		compactData, jsonErr := json.Marshal(metricsData)
		if jsonErr != nil {
			return fmt.Errorf("failed to marshal JSON: %w", jsonErr)
		}
		logEntry = string(compactData) + "\n"
	case "text":
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		output := ce.formatAsText(metricsData)
		logEntry = fmt.Sprintf("[%s] %s\n", timestamp, output)
	default:
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		output := fmt.Sprintf("%+v", metricsData)
		logEntry = fmt.Sprintf("[%s] %s\n", timestamp, output)
	}
	
	// 輸出到 stdout
	if ce.Config.Output == "stdout" || ce.Config.Output == "both" {
		fmt.Print(logEntry)
	}
	
	// 輸出到檔案
	if (ce.Config.Output == "file" || ce.Config.Output == "both") && ce.file != nil {
		_, err := ce.file.WriteString(logEntry)
		if err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
		ce.file.Sync() // 確保立即寫入
	}
	
	return nil
}

// formatAsText 格式化為文字輸出
func (ce *ConsoleExporter) formatAsText(data map[string]interface{}) string {
	output := "=== Universal Metrics ===\n"
	
	if service, ok := data["service"].(map[string]interface{}); ok {
		output += fmt.Sprintf("Service: %v (Type: %v, Version: %v)\n", 
			service["name"], service["type"], service["version"])
	}
	
	if resource, ok := data["resource"].(map[string]interface{}); ok {
		output += fmt.Sprintf("Memory: %.2fMB, Goroutines: %v\n", 
			getFloat64(resource, "memory_usage_bytes")/1024/1024,
			getInt64(resource, "goroutine_count"))
	}
	
	if ops, ok := data["operations"].(map[string]*OperationStats); ok {
		output += "Operations:\n"
		for name, stats := range ops {
			if stats.Total > 0 {
				successRate := float64(stats.Success) / float64(stats.Total) * 100
				output += fmt.Sprintf("  %s: %d total, %.1f%% success, %.2fms avg\n", 
					name, stats.Total, successRate, stats.AvgTime)
			}
		}
	}
	
	if deps, ok := data["dependencies"].(map[string]*DependencyHealth); ok && len(deps) > 0 {
		output += "Dependencies:\n"
		for name, dep := range deps {
			output += fmt.Sprintf("  %s: %s\n", name, dep.Status)
		}
	}
	
	return output
}

// getFloat64 安全地從 map 中獲取 float64 值
func getFloat64(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int64:
			return float64(v)
		case int:
			return float64(v)
		}
	}
	return 0
}

// getInt64 安全地從 map 中獲取 int64 值
func getInt64(m map[string]interface{}, key string) int64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case float64:
			return int64(v)
		}
	}
	return 0
}

// Close 關閉收集器
func (um *UniversalMetrics) Close() error {
	// 匯出最終指標
	um.ExportMetrics()
	
	// 關閉 console exporter
	if um.consoleExporter != nil && um.consoleExporter.file != nil {
		um.consoleExporter.file.Close()
	}
	
	return nil
}