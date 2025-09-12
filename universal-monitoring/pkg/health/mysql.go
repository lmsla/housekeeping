package health

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

// MySQLHealthChecker MySQL 健康檢查器
type MySQLHealthChecker struct {
	config  HealthConfig
	address string
	timeout time.Duration
}

// NewMySQLHealthChecker 創建 MySQL 健康檢查器
func NewMySQLHealthChecker(config HealthConfig) (HealthChecker, error) {
	// 從 DSN 中提取地址信息
	address := config.Endpoint
	timeout := config.Check.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	// 簡單解析 MySQL DSN 格式: user:password@tcp(host:port)/database
	if strings.Contains(address, "@tcp(") {
		start := strings.Index(address, "@tcp(") + 5
		end := strings.Index(address[start:], ")")
		if end > 0 {
			address = address[start : start+end]
		}
	}

	checker := &MySQLHealthChecker{
		config:  config,
		address: address,
		timeout: timeout,
	}

	return checker, nil
}

// CheckHealth 檢查 MySQL 健康狀態
func (m *MySQLHealthChecker) CheckHealth(ctx context.Context) HealthResult {
	start := time.Now()
	result := HealthResult{
		Connected:    false,
		ResponseTime: 0,
		Status:       "unhealthy",
		Message:      "",
		Metadata:     make(map[string]any),
	}

	// 測試 TCP 連線
	conn, err := m.connect(ctx)
	result.ResponseTime = time.Since(start)

	if err != nil {
		result.Message = fmt.Sprintf("connection failed: %v", err)
		return result
	}
	defer conn.Close()

	// 連線成功
	result.Connected = true
	result.Status = "healthy"
	result.Message = "MySQL TCP connection successful"
	result.Metadata["address"] = m.address
	result.Metadata["connection_type"] = "tcp"

	// 如果連線成功，嘗試讀取一些數據來確認 MySQL 服務正在運行
	if err := m.checkMySQLProtocol(conn); err != nil {
		result.Status = "degraded"
		result.Message = fmt.Sprintf("MySQL protocol check failed: %v", err)
	}

	return result
}

// connect 建立 TCP 連線
func (m *MySQLHealthChecker) connect(ctx context.Context) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: m.timeout,
	}
	return dialer.DialContext(ctx, "tcp", m.address)
}

// checkMySQLProtocol 檢查 MySQL 協議
func (m *MySQLHealthChecker) checkMySQLProtocol(conn net.Conn) error {
	// 設定讀取超時
	conn.SetReadDeadline(time.Now().Add(m.timeout))
	
	// 讀取 MySQL 的初始握手包
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read MySQL handshake: %v", err)
	}

	// 檢查是否是 MySQL 協議
	if n < 5 {
		return fmt.Errorf("invalid MySQL handshake packet size: %d", n)
	}

	// MySQL 握手包格式檢查 (簡化版)
	// 第一個字節應該是包長度的低位
	// 第5個字節應該是協議版本 (通常是10)
	if n >= 5 && buffer[4] == 10 {
		return nil // 看起來像 MySQL 協議
	}

	return fmt.Errorf("does not appear to be MySQL protocol")
}


// GetDependencyInfo 獲取依賴信息
func (m *MySQLHealthChecker) GetDependencyInfo() DependencyInfo {
	return DependencyInfo{
		Name:     m.config.Name,
		Type:     m.config.Type,
		Endpoint: m.address,
	}
}

// Configure 配置健康檢查器
func (m *MySQLHealthChecker) Configure(config map[string]any) error {
	// 支援運行時配置更新
	if address, exists := config["address"]; exists {
		if addressStr, ok := address.(string); ok {
			m.address = addressStr
		}
	}
	if timeout, exists := config["timeout"]; exists {
		if timeoutDur, ok := timeout.(time.Duration); ok {
			m.timeout = timeoutDur
		}
	}
	return nil
}

// Close 關閉健康檢查器
func (m *MySQLHealthChecker) Close() error {
	return nil
}