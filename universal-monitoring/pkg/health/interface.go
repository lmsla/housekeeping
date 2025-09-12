package health

import (
	"context"
	"time"
)

// HealthChecker 健康檢查介面
type HealthChecker interface {
	// 檢查健康狀態
	CheckHealth(ctx context.Context) HealthResult

	// 獲取依賴信息
	GetDependencyInfo() DependencyInfo

	// 配置健康檢查
	Configure(config map[string]any) error

	// 清理資源
	Close() error
}

// HealthResult 健康檢查結果
type HealthResult struct {
	Connected    bool          `json:"connected"`
	ResponseTime time.Duration `json:"response_time"`
	Status       string        `json:"status"` // "healthy", "degraded", "unhealthy"
	Message      string        `json:"message"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// DependencyInfo 依賴信息
type DependencyInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Version  string `json:"version,omitempty"`
	Endpoint string `json:"endpoint"`
}

// HealthConfig 健康檢查配置
type HealthConfig struct {
	Name     string                 `yaml:"name"`
	Type     string                 `yaml:"type"`
	Driver   string                 `yaml:"driver"`
	Endpoint string                 `yaml:"endpoint"`
	Settings map[string]any `yaml:"settings,omitempty"`
	Check    HealthCheckSettings    `yaml:"health_check"`
}

// HealthCheckSettings 健康檢查設定
type HealthCheckSettings struct {
	Enabled  bool          `yaml:"enabled"`
	Query    string        `yaml:"query,omitempty"`
	Command  string        `yaml:"command,omitempty"`
	Endpoint string        `yaml:"endpoint,omitempty"`
	Timeout  time.Duration `yaml:"timeout"`
}

// Registry 健康檢查插件註冊表
type Registry struct {
	checkers map[string]func(config HealthConfig) (HealthChecker, error)
}

// NewRegistry 創建新的註冊表
func NewRegistry() *Registry {
	r := &Registry{
		checkers: make(map[string]func(config HealthConfig) (HealthChecker, error)),
	}

	// 註冊內建健康檢查器
	r.Register("mysql", NewMySQLHealthChecker)
	r.Register("redis", NewRedisHealthChecker)
	r.Register("http", NewHTTPHealthChecker)
	r.Register("elasticsearch", NewElasticsearchHealthChecker)

	return r
}

// Register 註冊健康檢查器
func (r *Registry) Register(driverType string, factory func(config HealthConfig) (HealthChecker, error)) {
	r.checkers[driverType] = factory
}

// Create 創建健康檢查器
func (r *Registry) Create(config HealthConfig) (HealthChecker, error) {
	factory, exists := r.checkers[config.Driver]
	if !exists {
		return nil, &UnsupportedDriverError{Driver: config.Driver}
	}

	return factory(config)
}

// UnsupportedDriverError 不支援的驅動程式錯誤
type UnsupportedDriverError struct {
	Driver string
}

func (e *UnsupportedDriverError) Error() string {
	return "unsupported health check driver: " + e.Driver
}

// Manager 健康檢查管理器
type Manager struct {
	registry *Registry
	checkers map[string]HealthChecker
}

// NewManager 創建健康檢查管理器
func NewManager() *Manager {
	return &Manager{
		registry: NewRegistry(),
		checkers: make(map[string]HealthChecker),
	}
}

// AddHealthChecker 添加健康檢查器
func (m *Manager) AddHealthChecker(config HealthConfig) error {
	if !config.Check.Enabled {
		return nil
	}

	checker, err := m.registry.Create(config)
	if err != nil {
		return err
	}

	m.checkers[config.Name] = checker
	return nil
}

// CheckAll 檢查所有依賴的健康狀態
func (m *Manager) CheckAll(ctx context.Context) map[string]HealthResult {
	results := make(map[string]HealthResult)

	for name, checker := range m.checkers {
		result := checker.CheckHealth(ctx)
		results[name] = result
	}

	return results
}

// Close 關閉所有健康檢查器
func (m *Manager) Close() error {
	for _, checker := range m.checkers {
		if err := checker.Close(); err != nil {
			// 記錄錯誤但繼續關閉其他檢查器
			continue
		}
	}
	return nil
}