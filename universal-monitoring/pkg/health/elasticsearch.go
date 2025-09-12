package health

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ElasticsearchHealthChecker Elasticsearch 健康檢查器
type ElasticsearchHealthChecker struct {
	config   HealthConfig
	client   *http.Client
	endpoint string
}

// NewElasticsearchHealthChecker 創建 Elasticsearch 健康檢查器
func NewElasticsearchHealthChecker(config HealthConfig) (HealthChecker, error) {
	// 解析 endpoint，支援多個 URL
	endpoint := config.Endpoint
	if endpoint == "" {
		return nil, fmt.Errorf("elasticsearch endpoint is required")
	}

	// 如果是多個 URL，取第一個
	urls := strings.Split(endpoint, ",")
	mainURL := strings.TrimSpace(urls[0])

	// 確保 URL 格式正確
	if !strings.HasPrefix(mainURL, "http") {
		mainURL = "https://" + mainURL
	}

	// 健康檢查端點
	healthEndpoint := config.Check.Endpoint
	if healthEndpoint == "" {
		healthEndpoint = "/_cluster/health"
	}

	// 確保端點以 / 開頭
	if !strings.HasPrefix(healthEndpoint, "/") {
		healthEndpoint = "/" + healthEndpoint
	}

	checker := &ElasticsearchHealthChecker{
		config:   config,
		endpoint: mainURL + healthEndpoint,
		client: &http.Client{
			Timeout: config.Check.Timeout,
		},
	}

	return checker, nil
}

// CheckHealth 檢查 Elasticsearch 健康狀態
func (e *ElasticsearchHealthChecker) CheckHealth(ctx context.Context) HealthResult {
	start := time.Now()
	result := HealthResult{
		Connected:    false,
		ResponseTime: 0,
		Status:       "unhealthy",
		Message:      "",
		Metadata:     make(map[string]interface{}),
	}

	// 創建請求
	req, err := http.NewRequestWithContext(ctx, "GET", e.endpoint, nil)
	if err != nil {
		result.Message = fmt.Sprintf("failed to create request: %v", err)
		return result
	}

	// 執行請求
	resp, err := e.client.Do(req)
	result.ResponseTime = time.Since(start)

	if err != nil {
		result.Message = fmt.Sprintf("connection failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	// 讀取響應
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Message = fmt.Sprintf("failed to read response: %v", err)
		return result
	}

	// 檢查 HTTP 狀態碼
	if resp.StatusCode != http.StatusOK {
		result.Message = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))
		return result
	}

	// 解析集群健康狀態
	var healthResp ElasticsearchHealthResponse
	if err := json.Unmarshal(body, &healthResp); err != nil {
		result.Message = fmt.Sprintf("failed to parse health response: %v", err)
		return result
	}

	// 設定基本狀態
	result.Connected = true
	result.Metadata["cluster_name"] = healthResp.ClusterName
	result.Metadata["number_of_nodes"] = healthResp.NumberOfNodes
	result.Metadata["number_of_data_nodes"] = healthResp.NumberOfDataNodes
	result.Metadata["active_primary_shards"] = healthResp.ActivePrimaryShards
	result.Metadata["active_shards"] = healthResp.ActiveShards
	result.Metadata["unassigned_shards"] = healthResp.UnassignedShards
	result.Metadata["pending_tasks"] = healthResp.PendingTasks

	// 根據集群狀態設定健康狀態
	switch strings.ToLower(healthResp.Status) {
	case "green":
		result.Status = "healthy"
		result.Message = fmt.Sprintf("Cluster %s is healthy (green)", healthResp.ClusterName)
	case "yellow":
		result.Status = "degraded"
		result.Message = fmt.Sprintf("Cluster %s has some issues (yellow) - %d unassigned shards", 
			healthResp.ClusterName, healthResp.UnassignedShards)
	case "red":
		result.Status = "unhealthy"
		result.Message = fmt.Sprintf("Cluster %s has serious issues (red) - %d unassigned shards", 
			healthResp.ClusterName, healthResp.UnassignedShards)
	default:
		result.Status = "unhealthy"
		result.Message = fmt.Sprintf("Unknown cluster status: %s", healthResp.Status)
	}

	return result
}

// GetDependencyInfo 獲取依賴信息
func (e *ElasticsearchHealthChecker) GetDependencyInfo() DependencyInfo {
	return DependencyInfo{
		Name:     e.config.Name,
		Type:     e.config.Type,
		Endpoint: e.config.Endpoint,
	}
}

// Configure 配置健康檢查器
func (e *ElasticsearchHealthChecker) Configure(config map[string]interface{}) error {
	// 支援運行時配置更新
	if timeout, exists := config["timeout"]; exists {
		if timeoutDur, ok := timeout.(time.Duration); ok {
			e.client.Timeout = timeoutDur
		}
	}
	return nil
}

// Close 關閉健康檢查器
func (e *ElasticsearchHealthChecker) Close() error {
	// HTTP 客戶端不需要特別關閉
	return nil
}

// ElasticsearchHealthResponse Elasticsearch 健康檢查響應
type ElasticsearchHealthResponse struct {
	ClusterName                 string `json:"cluster_name"`
	Status                      string `json:"status"`
	TimedOut                    bool   `json:"timed_out"`
	NumberOfNodes               int    `json:"number_of_nodes"`
	NumberOfDataNodes           int    `json:"number_of_data_nodes"`
	ActivePrimaryShards         int    `json:"active_primary_shards"`
	ActiveShards                int    `json:"active_shards"`
	RelocatingShards            int    `json:"relocating_shards"`
	InitializingShards          int    `json:"initializing_shards"`
	UnassignedShards            int    `json:"unassigned_shards"`
	DelayedUnassignedShards     int    `json:"delayed_unassigned_shards"`
	PendingTasks                int    `json:"number_of_pending_tasks"`
	NumberOfInFlightFetch       int    `json:"number_of_in_flight_fetch"`
	TaskMaxWaitingInQueueMillis int    `json:"task_max_waiting_in_queue_millis"`
	ActiveShardsPercentAsNumber float64 `json:"active_shards_percent_as_number"`
}