package metrics

import (
	"os"
	"sync"
	"time"
)

// UniversalMetrics 通用指標收集器
type UniversalMetrics struct {
	// 服務基本信息
	ServiceName    string
	ServiceType    string
	ServiceVersion string
	StartTime      time.Time

	// 通用指標模組
	Resource     *ResourceMetrics
	Dependencies map[string]*DependencyHealth
	Errors       *ErrorMetrics

	// 可選指標模組
	Requests *RequestMetrics // Web/API 服務

	// 自定義業務指標
	CustomMetrics    map[string]interface{}
	OperationMetrics map[string]*OperationStats

	// 控制和配置
	Config *MetricsConfig
	mu     sync.RWMutex
	
	// 匯出相關
	metricsFile *os.File
	consoleExporter *ConsoleExporter
}

// ResourceMetrics 系統資源監控
type ResourceMetrics struct {
	// 記憶體指標
	MemoryUsage int64 `json:"memory_usage_bytes"`
	HeapInuse   int64 `json:"heap_inuse_bytes"`
	HeapSys     int64 `json:"heap_sys_bytes"`
	HeapObjects uint64 `json:"heap_objects"`

	// CPU 指標
	CPUUsage float64 `json:"cpu_usage_percent"`
	CPUTime  int64   `json:"cpu_time_nanoseconds"`

	// 運行時指標
	GoroutineCount int    `json:"goroutine_count"`
	GCPauseTime    int64  `json:"gc_pause_time_ns"`
	GCRuns         uint32 `json:"gc_runs"`

	// 更新時間
	LastUpdateTime time.Time `json:"last_update_time"`
}

// DependencyHealth 外部依賴健康檢查
type DependencyHealth struct {
	// 基本信息
	Name     string `json:"name"`
	Type     string `json:"type"`
	Endpoint string `json:"endpoint"`

	// 連線狀態
	Status          string        `json:"status"` // connected, disconnected, degraded
	ResponseTime    time.Duration `json:"response_time_ms"`
	AvgResponseTime time.Duration `json:"avg_response_time_ms"`

	// 統計數據
	ConnectionAttempts    int64 `json:"connection_attempts"`
	SuccessfulConnections int64 `json:"successful_connections"`
	FailedConnections     int64 `json:"failed_connections"`

	// 時間信息
	LastCheckTime   time.Time `json:"last_check_time"`
	LastSuccessTime time.Time `json:"last_success_time"`
	LastFailureTime time.Time `json:"last_failure_time"`
}

// ErrorMetrics 錯誤追蹤
type ErrorMetrics struct {
	// 總體錯誤統計
	TotalErrors int64   `json:"total_errors"`
	ErrorRate   float64 `json:"error_rate"`

	// 按類型分類
	ErrorsByType     map[string]*ErrorStats `json:"errors_by_type"`
	ErrorsBySeverity map[string]*ErrorStats `json:"errors_by_severity"`

	// HTTP 服務專用
	HTTPStatusCodes map[int]int64 `json:"http_status_codes"`

	// 最近錯誤
	RecentErrors  []ErrorEvent `json:"recent_errors"`
	LastErrorTime time.Time    `json:"last_error_time"`

	mu sync.RWMutex
}

// ErrorStats 錯誤統計
type ErrorStats struct {
	Count         int64     `json:"count"`
	FirstOccurred time.Time `json:"first_occurred"`
	LastOccurred  time.Time `json:"last_occurred"`
	Description   string    `json:"description"`
}

// ErrorEvent 錯誤事件
type ErrorEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"`
	Message   string                 `json:"message"`
	Severity  string                 `json:"severity"`
	Context   map[string]interface{} `json:"context"`
}

// RequestMetrics 請求處理指標
type RequestMetrics struct {
	// 總體請求統計
	TotalRequests      int64 `json:"total_requests"`
	SuccessfulRequests int64 `json:"successful_requests"`
	FailedRequests     int64 `json:"failed_requests"`

	// 性能指標
	AvgResponseTime time.Duration `json:"avg_response_time_ms"`
	MinResponseTime time.Duration `json:"min_response_time_ms"`
	MaxResponseTime time.Duration `json:"max_response_time_ms"`
	P95ResponseTime time.Duration `json:"p95_response_time_ms"`
	P99ResponseTime time.Duration `json:"p99_response_time_ms"`

	// 按端點統計
	EndpointStats map[string]*EndpointMetrics `json:"endpoint_stats"`

	// 流量統計
	RequestsPerSecond float64 `json:"requests_per_second"`
	ThroughputMBps    float64 `json:"throughput_mbps"`

	mu sync.RWMutex
}

// EndpointMetrics 端點指標
type EndpointMetrics struct {
	Path            string        `json:"path"`
	Method          string        `json:"method"`
	TotalRequests   int64         `json:"total_requests"`
	SuccessRequests int64         `json:"success_requests"`
	FailedRequests  int64         `json:"failed_requests"`
	AvgResponseTime time.Duration `json:"avg_response_time_ms"`
}

// OperationStats 業務操作統計
type OperationStats struct {
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Total       int64   `json:"total"`
	Success     int64   `json:"success"`
	Failed      int64   `json:"failed"`
	AvgTime     float64 `json:"avg_time_ms"`
	TotalTime   int64   `json:"total_time_ms"`
	mu          sync.RWMutex
}

// MetricsConfig 配置結構
type MetricsConfig struct {
	Service     ServiceConfig     `yaml:"service"`
	Metrics     MetricsSettings   `yaml:"metrics"`
	Modules     ModulesConfig     `yaml:"universal_modules"`
	Operations  []OperationConfig `yaml:"custom_operations"`
	Exporters   ExportersConfig   `yaml:"exporters"`
	Environment string            `yaml:"environment"`
}

// ServiceConfig 服務配置
type ServiceConfig struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
}

// MetricsSettings 指標設定
type MetricsSettings struct {
	Enabled            bool          `yaml:"enabled"`
	CollectionInterval time.Duration `yaml:"collection_interval"`
	SummaryInterval    time.Duration `yaml:"summary_interval"`
	RetentionDuration  time.Duration `yaml:"retention_duration"`
}

// ModulesConfig 模組配置
type ModulesConfig struct {
	ResourceMonitoring ResourceModuleConfig `yaml:"resource_monitoring"`
	ErrorTracking      ErrorModuleConfig    `yaml:"error_tracking"`
	DependencyHealth   HealthModuleConfig   `yaml:"dependency_health"`
}

// ResourceModuleConfig 資源監控模組配置
type ResourceModuleConfig struct {
	Enabled            bool          `yaml:"enabled"`
	CollectionInterval time.Duration `yaml:"collection_interval"`
	CPUMonitoring      bool          `yaml:"cpu_monitoring"`
	MemoryMonitoring   bool          `yaml:"memory_monitoring"`
	GCMonitoring       bool          `yaml:"gc_monitoring"`
}

// ErrorModuleConfig 錯誤追蹤模組配置
type ErrorModuleConfig struct {
	Enabled           bool    `yaml:"enabled"`
	MaxRecentErrors   int     `yaml:"max_recent_errors"`
	ErrorSamplingRate float64 `yaml:"error_sampling_rate"`
}

// HealthModuleConfig 健康檢查模組配置
type HealthModuleConfig struct {
	Enabled       bool          `yaml:"enabled"`
	CheckInterval time.Duration `yaml:"check_interval"`
	Timeout       time.Duration `yaml:"timeout"`
}

// OperationConfig 操作配置
type OperationConfig struct {
	Name        string    `yaml:"name"`
	DisplayName string    `yaml:"display_name"`
	Type        string    `yaml:"type"` // counter, timer, histogram
	Buckets     []float64 `yaml:"buckets,omitempty"`
}

// ExportersConfig 匯出器配置
type ExportersConfig struct {
	Console    ConsoleExporterConfig `yaml:"console"`
	Prometheus PrometheusConfig      `yaml:"prometheus"`
	InfluxDB   InfluxDBConfig        `yaml:"influxdb"`
}

// ConsoleExporterConfig Console 匯出器配置
type ConsoleExporterConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Format   string `yaml:"format"` // json, text, structured
	Level    string `yaml:"level"`  // debug, info, warn, error
	Output   string `yaml:"output"` // stdout, file, both
	FilePath string `yaml:"file_path"`
}

// PrometheusConfig Prometheus 配置
type PrometheusConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Endpoint  string   `yaml:"endpoint"`
	Namespace string   `yaml:"namespace"`
	Include   []string `yaml:"include,omitempty"`
	Exclude   []string `yaml:"exclude,omitempty"`
}

// InfluxDBConfig InfluxDB 配置
type InfluxDBConfig struct {
	Enabled         bool          `yaml:"enabled"`
	URL             string        `yaml:"url"`
	Database        string        `yaml:"database"`
	Username        string        `yaml:"username"`
	Password        string        `yaml:"password"`
	BatchSize       int           `yaml:"batch_size"`
	FlushInterval   time.Duration `yaml:"flush_interval"`
	RetentionPolicy string        `yaml:"retention_policy"`
}

// ConsoleExporter Console 匯出器
type ConsoleExporter struct {
	Config ConsoleExporterConfig
	file   *os.File
}