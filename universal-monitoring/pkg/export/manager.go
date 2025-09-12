package export

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bimap-org/universal-monitoring/pkg/metrics"
)

type MetricExporter interface {
	Export(metrics *metrics.UniversalMetrics) ([]byte, error)
	HealthCheck() error
	Configure(config map[string]interface{}) error
	Close() error
}

type ExporterManager struct {
	exporters    map[string]MetricExporter
	healthStatus map[string]bool
	config       *ExporterManagerConfig
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

type ExporterManagerConfig struct {
	HealthCheckInterval   time.Duration `yaml:"health_check_interval"`
	AutoFailover          bool          `yaml:"auto_failover"`
	EnableHealthCheck     bool          `yaml:"enable_health_check"`
	ExportTimeout         time.Duration `yaml:"export_timeout"`
	MaxRetries           int           `yaml:"max_retries"`
	RetryDelay           time.Duration `yaml:"retry_delay"`
	BackupExporters      map[string]string `yaml:"backup_exporters"`
}

type ExportResult struct {
	ExporterName string
	Success      bool
	Error        error
	Duration     time.Duration
	DataSize     int
}

func NewExporterManager(config *ExporterManagerConfig) *ExporterManager {
	if config == nil {
		config = &ExporterManagerConfig{
			HealthCheckInterval: 30 * time.Second,
			AutoFailover:       true,
			EnableHealthCheck:  true,
			ExportTimeout:      10 * time.Second,
			MaxRetries:         3,
			RetryDelay:         1 * time.Second,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	manager := &ExporterManager{
		exporters:    make(map[string]MetricExporter),
		healthStatus: make(map[string]bool),
		config:       config,
		ctx:          ctx,
		cancel:       cancel,
	}

	if config.EnableHealthCheck {
		manager.startHealthCheck()
	}

	return manager
}

func (em *ExporterManager) RegisterExporter(name string, exporter MetricExporter) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if _, exists := em.exporters[name]; exists {
		return fmt.Errorf("exporter %s already registered", name)
	}

	em.exporters[name] = exporter
	em.healthStatus[name] = true

	log.Printf("Registered exporter: %s", name)
	return nil
}

func (em *ExporterManager) UnregisterExporter(name string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	exporter, exists := em.exporters[name]
	if !exists {
		return fmt.Errorf("exporter %s not found", name)
	}

	if err := exporter.Close(); err != nil {
		log.Printf("Error closing exporter %s: %v", name, err)
	}

	delete(em.exporters, name)
	delete(em.healthStatus, name)

	log.Printf("Unregistered exporter: %s", name)
	return nil
}

func (em *ExporterManager) EnableExporter(name string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if _, exists := em.exporters[name]; !exists {
		return fmt.Errorf("exporter %s not found", name)
	}

	em.healthStatus[name] = true
	log.Printf("Enabled exporter: %s", name)
	return nil
}

func (em *ExporterManager) DisableExporter(name string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if _, exists := em.exporters[name]; !exists {
		return fmt.Errorf("exporter %s not found", name)
	}

	em.healthStatus[name] = false
	log.Printf("Disabled exporter: %s", name)
	return nil
}

func (em *ExporterManager) ExportMetrics(metrics *metrics.UniversalMetrics) []ExportResult {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var results []ExportResult
	var wg sync.WaitGroup

	resultCh := make(chan ExportResult, len(em.exporters))

	for name, exporter := range em.exporters {
		if !em.healthStatus[name] {
			continue
		}

		wg.Add(1)
		go func(exporterName string, exp MetricExporter) {
			defer wg.Done()
			result := em.exportToSingle(exporterName, exp, metrics)
			resultCh <- result
		}(name, exporter)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for result := range resultCh {
		results = append(results, result)
		
		if !result.Success && em.config.AutoFailover {
			em.handleFailover(result.ExporterName)
		}
	}

	return results
}

func (em *ExporterManager) exportToSingle(name string, exporter MetricExporter, metrics *metrics.UniversalMetrics) ExportResult {
	start := time.Now()
	
	ctx, cancel := context.WithTimeout(em.ctx, em.config.ExportTimeout)
	defer cancel()

	var lastErr error
	for attempt := 0; attempt <= em.config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ExportResult{
				ExporterName: name,
				Success:      false,
				Error:        fmt.Errorf("export timeout after %v", time.Since(start)),
				Duration:     time.Since(start),
			}
		default:
		}

		data, err := exporter.Export(metrics)
		duration := time.Since(start)

		if err == nil {
			return ExportResult{
				ExporterName: name,
				Success:      true,
				Duration:     duration,
				DataSize:     len(data),
			}
		}

		lastErr = err
		if attempt < em.config.MaxRetries {
			select {
			case <-time.After(em.config.RetryDelay):
			case <-ctx.Done():
				break
			}
		}
	}

	return ExportResult{
		ExporterName: name,
		Success:      false,
		Error:        fmt.Errorf("failed after %d retries: %v", em.config.MaxRetries, lastErr),
		Duration:     time.Since(start),
	}
}

func (em *ExporterManager) handleFailover(failedExporter string) {
	backup, exists := em.config.BackupExporters[failedExporter]
	if !exists {
		return
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	if _, exists := em.exporters[backup]; exists && !em.healthStatus[backup] {
		em.healthStatus[backup] = true
		log.Printf("Activated backup exporter %s for failed exporter %s", backup, failedExporter)
	}
}

func (em *ExporterManager) startHealthCheck() {
	em.wg.Add(1)
	go func() {
		defer em.wg.Done()
		ticker := time.NewTicker(em.config.HealthCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-em.ctx.Done():
				return
			case <-ticker.C:
				em.performHealthCheck()
			}
		}
	}()
}

func (em *ExporterManager) performHealthCheck() {
	em.mu.RLock()
	exporters := make(map[string]MetricExporter)
	for name, exporter := range em.exporters {
		exporters[name] = exporter
	}
	em.mu.RUnlock()

	for name, exporter := range exporters {
		go func(exporterName string, exp MetricExporter) {
			healthy := em.checkExporterHealth(exp)
			
			em.mu.Lock()
			oldStatus := em.healthStatus[exporterName]
			em.healthStatus[exporterName] = healthy
			em.mu.Unlock()

			if oldStatus != healthy {
				status := "unhealthy"
				if healthy {
					status = "healthy"
				}
				log.Printf("Exporter %s status changed to %s", exporterName, status)

				if !healthy && em.config.AutoFailover {
					em.handleFailover(exporterName)
				}
			}
		}(name, exporter)
	}
}

func (em *ExporterManager) checkExporterHealth(exporter MetricExporter) bool {
	ctx, cancel := context.WithTimeout(em.ctx, 5*time.Second)
	defer cancel()

	done := make(chan bool, 1)
	go func() {
		err := exporter.HealthCheck()
		done <- (err == nil)
	}()

	select {
	case healthy := <-done:
		return healthy
	case <-ctx.Done():
		return false
	}
}

func (em *ExporterManager) GetHealthStatus() map[string]bool {
	em.mu.RLock()
	defer em.mu.RUnlock()

	status := make(map[string]bool)
	for name, health := range em.healthStatus {
		status[name] = health
	}
	return status
}

func (em *ExporterManager) GetExporterNames() []string {
	em.mu.RLock()
	defer em.mu.RUnlock()

	names := make([]string, 0, len(em.exporters))
	for name := range em.exporters {
		names = append(names, name)
	}
	return names
}

func (em *ExporterManager) ReloadConfig(config *ExporterManagerConfig) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	oldInterval := em.config.HealthCheckInterval
	em.config = config

	if em.config.HealthCheckInterval != oldInterval {
		log.Printf("Health check interval changed from %v to %v", 
			oldInterval, em.config.HealthCheckInterval)
	}

	return nil
}

func (em *ExporterManager) ConfigureExporter(name string, config map[string]interface{}) error {
	em.mu.RLock()
	exporter, exists := em.exporters[name]
	em.mu.RUnlock()

	if !exists {
		return fmt.Errorf("exporter %s not found", name)
	}

	return exporter.Configure(config)
}

func (em *ExporterManager) Close() error {
	em.cancel()
	em.wg.Wait()

	em.mu.Lock()
	defer em.mu.Unlock()

	var lastErr error
	for name, exporter := range em.exporters {
		if err := exporter.Close(); err != nil {
			log.Printf("Error closing exporter %s: %v", name, err)
			lastErr = err
		}
	}

	log.Println("ExporterManager closed")
	return lastErr
}

type ConsoleExporter struct {
	format string
}

func NewConsoleExporter(format string) *ConsoleExporter {
	if format == "" {
		format = "json"
	}
	return &ConsoleExporter{format: format}
}

func (c *ConsoleExporter) Export(metrics *metrics.UniversalMetrics) ([]byte, error) {
	switch c.format {
	case "prometheus":
		exporter := NewPrometheusExporter(PrometheusExporterConfig{})
		return exporter.Export(metrics)
	case "influxdb":
		exporter := NewInfluxDBExporter(InfluxDBExporterConfig{})
		return exporter.Export(metrics)
	default:
		return []byte(fmt.Sprintf("Service: %s, Memory: %d bytes, Goroutines: %d\n", 
			metrics.ServiceName, metrics.Resource.MemoryUsage, metrics.Resource.GoroutineCount)), nil
	}
}

func (c *ConsoleExporter) HealthCheck() error {
	return nil
}

func (c *ConsoleExporter) Configure(config map[string]interface{}) error {
	if format, ok := config["format"].(string); ok {
		c.format = format
	}
	return nil
}

func (c *ConsoleExporter) Close() error {
	return nil
}