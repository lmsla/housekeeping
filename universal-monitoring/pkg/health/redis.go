package health

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type RedisHealthChecker struct {
	config  HealthConfig
	address string
	command string
	timeout time.Duration
}

func NewRedisHealthChecker(config HealthConfig) (HealthChecker, error) {
	command := config.Check.Command
	if command == "" {
		command = "PING"
	}

	address := config.Endpoint
	timeout := config.Check.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	checker := &RedisHealthChecker{
		config:  config,
		address: address,
		command: command,
		timeout: timeout,
	}

	return checker, nil
}

func (r *RedisHealthChecker) CheckHealth(ctx context.Context) HealthResult {
	start := time.Now()
	result := HealthResult{
		Connected:    false,
		ResponseTime: 0,
		Status:       "unhealthy",
		Message:      "",
		Metadata:     make(map[string]interface{}),
	}

	conn, err := r.connect(ctx)
	result.ResponseTime = time.Since(start)

	if err != nil {
		result.Message = fmt.Sprintf("connection failed: %v", err)
		return result
	}
	defer conn.Close()

	response, err := r.sendCommand(conn, r.command)
	result.ResponseTime = time.Since(start)

	if err != nil {
		result.Message = fmt.Sprintf("command failed: %v", err)
		return result
	}

	if strings.ToUpper(r.command) == "PING" {
		if strings.ToUpper(response) == "PONG" {
			result.Connected = true
			result.Status = "healthy"
			result.Message = "Redis PING successful"
		} else {
			result.Message = fmt.Sprintf("unexpected PING response: %s", response)
			return result
		}
	} else {
		result.Connected = true
		result.Status = "healthy"
		result.Message = fmt.Sprintf("Redis command %s successful", r.command)
		result.Metadata["command_response"] = response
	}

	if infoResp, err := r.sendCommand(conn, "INFO server"); err == nil {
		infoData := r.parseInfo(infoResp)
		
		if version, exists := infoData["redis_version"]; exists {
			result.Metadata["version"] = version
		}
		if mode, exists := infoData["redis_mode"]; exists {
			result.Metadata["mode"] = mode
		}
		if usedMemory, exists := infoData["used_memory"]; exists {
			result.Metadata["used_memory"] = usedMemory
		}
		if usedMemoryHuman, exists := infoData["used_memory_human"]; exists {
			result.Metadata["used_memory_human"] = usedMemoryHuman
		}
		if connectedClients, exists := infoData["connected_clients"]; exists {
			result.Metadata["connected_clients"] = connectedClients
		}

		if role, exists := infoData["role"]; exists {
			result.Metadata["role"] = role
		}
	}

	if dbSizeResp, err := r.sendCommand(conn, "DBSIZE"); err == nil {
		if size, err := strconv.Atoi(strings.TrimSpace(dbSizeResp)); err == nil {
			result.Metadata["db_keys"] = size
		}
	}

	return result
}

func (r *RedisHealthChecker) connect(ctx context.Context) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: r.timeout,
	}
	return dialer.DialContext(ctx, "tcp", r.address)
}

func (r *RedisHealthChecker) sendCommand(conn net.Conn, command string) (string, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	cmd := fmt.Sprintf("*%d\r\n", len(parts))
	for _, part := range parts {
		cmd += fmt.Sprintf("$%d\r\n%s\r\n", len(part), part)
	}

	if _, err := conn.Write([]byte(cmd)); err != nil {
		return "", err
	}

	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		return "", err
	}

	response := string(buffer[:n])
	
	if strings.HasPrefix(response, "+") {
		return strings.TrimSpace(response[1:]), nil
	} else if strings.HasPrefix(response, ":") {
		return strings.TrimSpace(response[1:]), nil
	} else if strings.HasPrefix(response, "$") {
		lines := strings.Split(response, "\r\n")
		if len(lines) > 2 {
			return lines[1], nil
		}
		return strings.TrimSpace(response[1:]), nil
	} else if strings.HasPrefix(response, "-") {
		return "", fmt.Errorf("redis error: %s", strings.TrimSpace(response[1:]))
	}

	return strings.TrimSpace(response), nil
}

func (r *RedisHealthChecker) parseInfo(info string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(info, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	
	return result
}

func (r *RedisHealthChecker) GetDependencyInfo() DependencyInfo {
	return DependencyInfo{
		Name:     r.config.Name,
		Type:     r.config.Type,
		Endpoint: r.config.Endpoint,
	}
}

func (r *RedisHealthChecker) Configure(config map[string]interface{}) error {
	if command, exists := config["command"]; exists {
		if commandStr, ok := command.(string); ok {
			r.command = commandStr
		}
	}
	return nil
}

func (r *RedisHealthChecker) Close() error {
	return nil
}