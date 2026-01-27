package metrics

import (
	"sync"
	"time"
	"sync/atomic"
	"runtime"
)

// MetricsCollector 指標收集器
type MetricsCollector struct {
	operationMetrics   *OperationMetrics
	performanceMetrics *PerformanceMetrics
	resourceMetrics    *ResourceMetrics
	esHealthMetrics    *ESHealthMetrics
	mu                 sync.RWMutex
}

// OperationMetrics 操作統計指標
type OperationMetrics struct {
	TotalOperations   int64 `json:"total_operations"`
	SuccessfulOps     int64 `json:"successful_ops"`
	FailedOps         int64 `json:"failed_ops"`
	OperationsByType  *OpTypeMetrics `json:"operations_by_type"`
}

// OpTypeMetrics 按操作類型統計
type OpTypeMetrics struct {
	Delete        *OpStats `json:"delete"`
	Close         *OpStats `json:"close"`
	Open          *OpStats `json:"open"`
	Allocation    *OpStats `json:"allocation"`
	ForceMerge    *OpStats `json:"force_merge"`
	Rollover      *OpStats `json:"rollover"`
}

// OpStats 單一操作類型統計
type OpStats struct {
	Total     int64   `json:"total"`
	Success   int64   `json:"success"`
	Failed    int64   `json:"failed"`
	AvgTime   float64 `json:"avg_time_ms"`
	TotalTime int64   `json:"total_time_ms"`
}

// PerformanceMetrics 性能指標
type PerformanceMetrics struct {
	ActionDuration    time.Duration `json:"action_duration_ms"`
	FilterDuration    time.Duration `json:"filter_duration_ms"`
	ESOperationTime   time.Duration `json:"es_operation_time_ms"`
	TotalProcessTime  time.Duration `json:"total_process_time_ms"`
	LastUpdateTime    time.Time     `json:"last_update_time"`
}

// ResourceMetrics 資源使用指標
type ResourceMetrics struct {
	CPUUsage         float64   `json:"cpu_usage_percent"`
	MemoryUsage      int64     `json:"memory_usage_bytes"`
	GoroutineCount   int       `json:"goroutine_count"`
	HeapInuse        int64     `json:"heap_inuse_bytes"`
	HeapSys          int64     `json:"heap_sys_bytes"`
	LastUpdateTime   time.Time `json:"last_update_time"`
}

// ESHealthMetrics ES 健康狀態指標
type ESHealthMetrics struct {
	ConnectionStatus    string        `json:"connection_status"`
	ResponseTime        time.Duration `json:"response_time_ms"`
	ClusterHealth       string        `json:"cluster_health"`
	LastCheckTime       time.Time     `json:"last_check_time"`
	ConnectionAttempts  int64         `json:"connection_attempts"`
	SuccessfulConnections int64       `json:"successful_connections"`
	FailedConnections   int64         `json:"failed_connections"`
}

// 全域指標收集器實例
var GlobalMetrics *MetricsCollector
var once sync.Once

// InitMetrics 初始化指標收集器
func InitMetrics() *MetricsCollector {
	once.Do(func() {
		GlobalMetrics = &MetricsCollector{
			operationMetrics: &OperationMetrics{
				OperationsByType: &OpTypeMetrics{
					Delete:     &OpStats{},
					Close:      &OpStats{},
					Open:       &OpStats{},
					Allocation: &OpStats{},
					ForceMerge: &OpStats{},
					Rollover:   &OpStats{},
				},
			},
			performanceMetrics: &PerformanceMetrics{},
			resourceMetrics:    &ResourceMetrics{},
			esHealthMetrics:    &ESHealthMetrics{
				ConnectionStatus: "unknown",
				ClusterHealth:   "unknown",
			},
		}
	})
	return GlobalMetrics
}

// RecordOperation 記錄操作
func (mc *MetricsCollector) RecordOperation(opType string, success bool, duration time.Duration) {
	atomic.AddInt64(&mc.operationMetrics.TotalOperations, 1)
	
	if success {
		atomic.AddInt64(&mc.operationMetrics.SuccessfulOps, 1)
	} else {
		atomic.AddInt64(&mc.operationMetrics.FailedOps, 1)
	}
	
	// 記錄按類型統計
	var opStats *OpStats
	switch opType {
	case "delete_indices":
		opStats = mc.operationMetrics.OperationsByType.Delete
	case "close":
		opStats = mc.operationMetrics.OperationsByType.Close
	case "open":
		opStats = mc.operationMetrics.OperationsByType.Open
	case "allocation":
		opStats = mc.operationMetrics.OperationsByType.Allocation
	case "forcemerge":
		opStats = mc.operationMetrics.OperationsByType.ForceMerge
	case "rollover":
		opStats = mc.operationMetrics.OperationsByType.Rollover
	default:
		return
	}
	
	atomic.AddInt64(&opStats.Total, 1)
	atomic.AddInt64(&opStats.TotalTime, duration.Nanoseconds()/1e6) // 轉換為毫秒

	if success {
		atomic.AddInt64(&opStats.Success, 1)
	} else {
		atomic.AddInt64(&opStats.Failed, 1)
	}

	// 更新平均時間 - 使用互斥鎖保護非原子操作
	mc.mu.Lock()
	total := atomic.LoadInt64(&opStats.Total)
	if total > 0 {
		totalTime := atomic.LoadInt64(&opStats.TotalTime)
		opStats.AvgTime = float64(totalTime) / float64(total)
	}
	mc.mu.Unlock()
}

// RecordESHealth 記錄 ES 健康狀態
func (mc *MetricsCollector) RecordESHealth(connected bool, responseTime time.Duration, clusterHealth string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	atomic.AddInt64(&mc.esHealthMetrics.ConnectionAttempts, 1)
	
	if connected {
		atomic.AddInt64(&mc.esHealthMetrics.SuccessfulConnections, 1)
		mc.esHealthMetrics.ConnectionStatus = "connected"
	} else {
		atomic.AddInt64(&mc.esHealthMetrics.FailedConnections, 1)
		mc.esHealthMetrics.ConnectionStatus = "disconnected"
	}
	
	mc.esHealthMetrics.ResponseTime = responseTime
	mc.esHealthMetrics.ClusterHealth = clusterHealth
	mc.esHealthMetrics.LastCheckTime = time.Now()
}

// UpdateResourceMetrics 更新資源使用指標
func (mc *MetricsCollector) UpdateResourceMetrics() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	mc.resourceMetrics.MemoryUsage = int64(m.Alloc)
	mc.resourceMetrics.HeapInuse = int64(m.HeapInuse)
	mc.resourceMetrics.HeapSys = int64(m.HeapSys)
	mc.resourceMetrics.GoroutineCount = runtime.NumGoroutine()
	mc.resourceMetrics.LastUpdateTime = time.Now()
}

// GetOperationMetrics 獲取操作指標
func (mc *MetricsCollector) GetOperationMetrics() OperationMetrics {
	return *mc.operationMetrics
}

// GetResourceMetrics 獲取資源指標
func (mc *MetricsCollector) GetResourceMetrics() ResourceMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return *mc.resourceMetrics
}

// GetESHealthMetrics 獲取 ES 健康指標
func (mc *MetricsCollector) GetESHealthMetrics() ESHealthMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return *mc.esHealthMetrics
}

// GetAllMetrics 獲取所有指標
func (mc *MetricsCollector) GetAllMetrics() map[string]interface{} {
	return map[string]interface{}{
		"operations": mc.GetOperationMetrics(),
		"resource":   mc.GetResourceMetrics(),
		"es_health":  mc.GetESHealthMetrics(),
		"timestamp":  time.Now(),
	}
}