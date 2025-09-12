package config

import (
	"fmt"
	"os"
	"time"

	"github.com/bimap-org/universal-monitoring/pkg/metrics"
	"gopkg.in/yaml.v3"
)

// LoadFromFile 從檔案載入配置
func LoadFromFile(filename string) (*metrics.MetricsConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	// 展開環境變數
	expandedData := os.ExpandEnv(string(data))

	var config metrics.MetricsConfig
	if err := yaml.Unmarshal([]byte(expandedData), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 應用預設值
	applyDefaults(&config)

	// 驗證配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// LoadWebAPIDefaults 載入 Web API 服務的預設配置
func LoadWebAPIDefaults() *metrics.MetricsConfig {
	return &metrics.MetricsConfig{
		Service: metrics.ServiceConfig{
			Name:        "web-service",
			Type:        "web-api",
			Version:     "1.0.0",
			Environment: "development",
		},
		Metrics: metrics.MetricsSettings{
			Enabled:            true,
			CollectionInterval: 30 * time.Second,
			SummaryInterval:    5 * time.Minute,
			RetentionDuration:  24 * time.Hour,
		},
		Modules: metrics.ModulesConfig{
			ResourceMonitoring: metrics.ResourceModuleConfig{
				Enabled:            true,
				CollectionInterval: 10 * time.Second,
				CPUMonitoring:      true,
				MemoryMonitoring:   true,
				GCMonitoring:       true,
			},
			ErrorTracking: metrics.ErrorModuleConfig{
				Enabled:           true,
				MaxRecentErrors:   100,
				ErrorSamplingRate: 1.0,
			},
			DependencyHealth: metrics.HealthModuleConfig{
				Enabled:       true,
				CheckInterval: 30 * time.Second,
				Timeout:       5 * time.Second,
			},
		},
		Exporters: metrics.ExportersConfig{
			Console: metrics.ConsoleExporterConfig{
				Enabled: true,
				Format:  "json",
				Level:   "info",
				Output:  "stdout",
			},
		},
	}
}

// LoadBatchJobDefaults 載入批次處理服務的預設配置
func LoadBatchJobDefaults() *metrics.MetricsConfig {
	config := LoadWebAPIDefaults()
	config.Service.Type = "batch-job"
	config.Metrics.SummaryInterval = 1 * time.Minute // 批次任務更頻繁的摘要
	return config
}

// applyDefaults 應用預設值
func applyDefaults(config *metrics.MetricsConfig) {
	// 服務預設值
	if config.Service.Version == "" {
		config.Service.Version = "1.0.0"
	}
	if config.Service.Environment == "" {
		config.Service.Environment = "development"
	}

	// 指標設定預設值
	if config.Metrics.CollectionInterval == 0 {
		config.Metrics.CollectionInterval = 30 * time.Second
	}
	if config.Metrics.SummaryInterval == 0 {
		config.Metrics.SummaryInterval = 5 * time.Minute
	}
	if config.Metrics.RetentionDuration == 0 {
		config.Metrics.RetentionDuration = 24 * time.Hour
	}

	// 資源監控預設值
	if config.Modules.ResourceMonitoring.CollectionInterval == 0 {
		config.Modules.ResourceMonitoring.CollectionInterval = 10 * time.Second
	}

	// 錯誤追蹤預設值
	if config.Modules.ErrorTracking.MaxRecentErrors == 0 {
		config.Modules.ErrorTracking.MaxRecentErrors = 100
	}
	if config.Modules.ErrorTracking.ErrorSamplingRate == 0 {
		config.Modules.ErrorTracking.ErrorSamplingRate = 1.0
	}

	// 健康檢查預設值
	if config.Modules.DependencyHealth.CheckInterval == 0 {
		config.Modules.DependencyHealth.CheckInterval = 30 * time.Second
	}
	if config.Modules.DependencyHealth.Timeout == 0 {
		config.Modules.DependencyHealth.Timeout = 5 * time.Second
	}

	// 匯出器預設值
	if config.Exporters.Console.Format == "" {
		config.Exporters.Console.Format = "json"
	}
	if config.Exporters.Console.Level == "" {
		config.Exporters.Console.Level = "info"
	}
	if config.Exporters.Console.Output == "" {
		config.Exporters.Console.Output = "stdout"
	}
}

// validateConfig 驗證配置
func validateConfig(config *metrics.MetricsConfig) error {
	// 必填欄位檢查
	if config.Service.Name == "" {
		return fmt.Errorf("service.name is required")
	}
	if config.Service.Type == "" {
		return fmt.Errorf("service.type is required")
	}

	// 服務類型檢查
	validTypes := map[string]bool{
		"web-api":           true,
		"batch-job":         true,
		"message-processor": true,
	}
	if !validTypes[config.Service.Type] {
		return fmt.Errorf("invalid service.type: %s", config.Service.Type)
	}

	// 時間間隔檢查
	if config.Metrics.CollectionInterval < time.Second {
		return fmt.Errorf("metrics.collection_interval must be at least 1 second")
	}

	// 匯出器格式檢查
	validFormats := map[string]bool{
		"json": true, "text": true, "structured": true,
	}
	if config.Exporters.Console.Enabled && !validFormats[config.Exporters.Console.Format] {
		return fmt.Errorf("invalid console exporter format: %s", config.Exporters.Console.Format)
	}

	// 日誌級別檢查
	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}
	if config.Exporters.Console.Enabled && !validLevels[config.Exporters.Console.Level] {
		return fmt.Errorf("invalid console exporter level: %s", config.Exporters.Console.Level)
	}

	return nil
}

// CreateDefaultConfig 創建預設配置檔案
func CreateDefaultConfig(filename string, serviceType string) error {
	var config *metrics.MetricsConfig

	switch serviceType {
	case "web-api":
		config = LoadWebAPIDefaults()
	case "batch-job":
		config = LoadBatchJobDefaults()
	default:
		config = LoadWebAPIDefaults()
		config.Service.Type = serviceType
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}