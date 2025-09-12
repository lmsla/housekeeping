package health

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type HTTPHealthChecker struct {
	config         HealthConfig
	client         *http.Client
	endpoint       string
	method         string
	expectedStatus int
	headers        map[string]string
}

func NewHTTPHealthChecker(config HealthConfig) (HealthChecker, error) {
	endpoint := config.Check.Endpoint
	if endpoint == "" {
		endpoint = config.Endpoint
	}

	method := "GET"
	expectedStatus := 200
	headers := make(map[string]string)

	if settings := config.Settings; settings != nil {
		if m, ok := settings["method"].(string); ok && m != "" {
			method = strings.ToUpper(m)
		}
		if status, ok := settings["expected_status"].(int); ok && status > 0 {
			expectedStatus = status
		}
		if statusStr, ok := settings["expected_status"].(string); ok {
			if parsed, err := strconv.Atoi(statusStr); err == nil {
				expectedStatus = parsed
			}
		}
		if h, ok := settings["headers"].(map[string]interface{}); ok {
			for k, v := range h {
				if vStr, ok := v.(string); ok {
					headers[k] = vStr
				}
			}
		}
	}

	client := &http.Client{
		Timeout: config.Check.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 2,
		},
	}

	checker := &HTTPHealthChecker{
		config:         config,
		client:         client,
		endpoint:       endpoint,
		method:         method,
		expectedStatus: expectedStatus,
		headers:        headers,
	}

	return checker, nil
}

func (h *HTTPHealthChecker) CheckHealth(ctx context.Context) HealthResult {
	start := time.Now()
	result := HealthResult{
		Connected:    false,
		ResponseTime: 0,
		Status:       "unhealthy",
		Message:      "",
		Metadata:     make(map[string]interface{}),
	}

	req, err := http.NewRequestWithContext(ctx, h.method, h.endpoint, nil)
	if err != nil {
		result.ResponseTime = time.Since(start)
		result.Message = fmt.Sprintf("failed to create request: %v", err)
		return result
	}

	for key, value := range h.headers {
		req.Header.Set(key, value)
	}

	resp, err := h.client.Do(req)
	result.ResponseTime = time.Since(start)

	if err != nil {
		result.Message = fmt.Sprintf("request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.Connected = true
	result.Metadata["status_code"] = resp.StatusCode
	result.Metadata["content_length"] = resp.ContentLength
	result.Metadata["content_type"] = resp.Header.Get("Content-Type")

	if resp.StatusCode == h.expectedStatus {
		result.Status = "healthy"
		result.Message = fmt.Sprintf("HTTP %s successful, got expected status %d", h.method, resp.StatusCode)
	} else {
		result.Status = "degraded"
		result.Message = fmt.Sprintf("HTTP %s returned status %d, expected %d", h.method, resp.StatusCode, h.expectedStatus)
	}

	body, err := io.ReadAll(resp.Body)
	if err == nil && len(body) > 0 {
		bodyStr := string(body)
		if len(bodyStr) > 500 {
			bodyStr = bodyStr[:500] + "..."
		}
		result.Metadata["response_body"] = bodyStr
	}

	return result
}

func (h *HTTPHealthChecker) GetDependencyInfo() DependencyInfo {
	return DependencyInfo{
		Name:     h.config.Name,
		Type:     h.config.Type,
		Endpoint: h.endpoint,
	}
}

func (h *HTTPHealthChecker) Configure(config map[string]interface{}) error {
	if endpoint, exists := config["endpoint"]; exists {
		if endpointStr, ok := endpoint.(string); ok {
			h.endpoint = endpointStr
		}
	}
	if method, exists := config["method"]; exists {
		if methodStr, ok := method.(string); ok {
			h.method = strings.ToUpper(methodStr)
		}
	}
	if expectedStatus, exists := config["expected_status"]; exists {
		if statusInt, ok := expectedStatus.(int); ok {
			h.expectedStatus = statusInt
		} else if statusStr, ok := expectedStatus.(string); ok {
			if parsed, err := strconv.Atoi(statusStr); err == nil {
				h.expectedStatus = parsed
			}
		}
	}
	if headers, exists := config["headers"]; exists {
		if headersMap, ok := headers.(map[string]interface{}); ok {
			h.headers = make(map[string]string)
			for k, v := range headersMap {
				if vStr, ok := v.(string); ok {
					h.headers[k] = vStr
				}
			}
		}
	}
	return nil
}

func (h *HTTPHealthChecker) Close() error {
	return nil
}