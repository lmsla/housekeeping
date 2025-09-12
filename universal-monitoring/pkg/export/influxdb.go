package export

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bimap-org/universal-monitoring/pkg/metrics"
)

type InfluxDBExporter struct {
	measurement   string
	tags          map[string]string
	batchSize     int
	precisionUnit string
}

type InfluxDBExporterConfig struct {
	Measurement   string            `yaml:"measurement"`
	Tags          map[string]string `yaml:"tags"`
	BatchSize     int               `yaml:"batch_size"`
	PrecisionUnit string            `yaml:"precision_unit"`
}

func NewInfluxDBExporter(config InfluxDBExporterConfig) *InfluxDBExporter {
	if config.Measurement == "" {
		config.Measurement = "service_metrics"
	}
	
	if config.BatchSize == 0 {
		config.BatchSize = 1000
	}
	
	if config.PrecisionUnit == "" {
		config.PrecisionUnit = "ns"
	}
	
	if config.Tags == nil {
		config.Tags = make(map[string]string)
	}
	
	return &InfluxDBExporter{
		measurement:   config.Measurement,
		tags:          config.Tags,
		batchSize:     config.BatchSize,
		precisionUnit: config.PrecisionUnit,
	}
}

func (e *InfluxDBExporter) Export(m *metrics.UniversalMetrics) ([]byte, error) {
	var buffer bytes.Buffer
	timestamp := time.Now().UnixNano()
	
	baseTags := e.buildBaseTags(m)
	
	if m.Resource != nil {
		lines := e.exportResourceMetrics(m.Resource, baseTags, timestamp)
		buffer.WriteString(strings.Join(lines, "\n"))
		buffer.WriteString("\n")
	}
	
	if len(m.Dependencies) > 0 {
		lines := e.exportDependencyMetrics(m.Dependencies, baseTags, timestamp)
		buffer.WriteString(strings.Join(lines, "\n"))
		buffer.WriteString("\n")
	}
	
	if m.Errors != nil {
		lines := e.exportErrorMetrics(m.Errors, baseTags, timestamp)
		buffer.WriteString(strings.Join(lines, "\n"))
		buffer.WriteString("\n")
	}
	
	if len(m.OperationMetrics) > 0 {
		lines := e.exportOperationMetrics(m.OperationMetrics, baseTags, timestamp)
		buffer.WriteString(strings.Join(lines, "\n"))
		buffer.WriteString("\n")
	}
	
	return buffer.Bytes(), nil
}

func (e *InfluxDBExporter) buildBaseTags(m *metrics.UniversalMetrics) map[string]string {
	tags := make(map[string]string)
	
	for k, v := range e.tags {
		tags[k] = v
	}
	
	tags["service"] = m.ServiceName
	tags["service_type"] = m.ServiceType
	tags["version"] = m.ServiceVersion
	tags["metric_type"] = "resource"
	
	return tags
}

func (e *InfluxDBExporter) exportResourceMetrics(resource *metrics.ResourceMetrics, baseTags map[string]string, timestamp int64) []string {
	var lines []string
	
	tags := make(map[string]string)
	for k, v := range baseTags {
		tags[k] = v
	}
	tags["metric_type"] = "resource"
	
	fields := map[string]interface{}{
		"memory_usage":     resource.MemoryUsage,
		"heap_inuse":       resource.HeapInuse,
		"heap_sys":         resource.HeapSys,
		"heap_objects":     resource.HeapObjects,
		"goroutine_count":  resource.GoroutineCount,
		"gc_pause_time":    resource.GCPauseTime,
		"gc_runs":          resource.GCRuns,
	}
	
	line := e.buildLineProtocol(e.measurement, tags, fields, timestamp)
	lines = append(lines, line)
	
	return lines
}

func (e *InfluxDBExporter) exportDependencyMetrics(dependencies map[string]*metrics.DependencyHealth, baseTags map[string]string, timestamp int64) []string {
	var lines []string
	
	for name, dep := range dependencies {
		tags := make(map[string]string)
		for k, v := range baseTags {
			tags[k] = v
		}
		tags["metric_type"] = "dependency"
		tags["dependency"] = name
		tags["dependency_type"] = dep.Type
		tags["endpoint"] = e.sanitizeTagValue(dep.Endpoint)
		
		var statusValue float64
		switch dep.Status {
		case "connected":
			statusValue = 1
		case "degraded":
			statusValue = 0.5
		default:
			statusValue = 0
		}
		
		fields := map[string]interface{}{
			"status":                  statusValue,
			"status_text":             dep.Status,
			"response_time":           dep.ResponseTime.Nanoseconds(),
			"avg_response_time":       dep.AvgResponseTime.Nanoseconds(),
			"connection_attempts":     dep.ConnectionAttempts,
			"successful_connections":  dep.SuccessfulConnections,
			"failed_connections":      dep.FailedConnections,
			"last_check_timestamp":    dep.LastCheckTime.Unix(),
			"last_success_timestamp":  dep.LastSuccessTime.Unix(),
			"last_failure_timestamp":  dep.LastFailureTime.Unix(),
		}
		
		line := e.buildLineProtocol(e.measurement, tags, fields, timestamp)
		lines = append(lines, line)
	}
	
	return lines
}

func (e *InfluxDBExporter) exportErrorMetrics(errors *metrics.ErrorMetrics, baseTags map[string]string, timestamp int64) []string {
	var lines []string
	
	tags := make(map[string]string)
	for k, v := range baseTags {
		tags[k] = v
	}
	tags["metric_type"] = "error"
	
	fields := map[string]interface{}{
		"total_errors":        errors.TotalErrors,
		"error_rate":          errors.ErrorRate,
		"last_error_timestamp": errors.LastErrorTime.Unix(),
	}
	
	line := e.buildLineProtocol(e.measurement, tags, fields, timestamp)
	lines = append(lines, line)
	
	for errorType, stats := range errors.ErrorsByType {
		typeTags := make(map[string]string)
		for k, v := range baseTags {
			typeTags[k] = v
		}
		typeTags["metric_type"] = "error_by_type"
		typeTags["error_type"] = errorType
		
		typeFields := map[string]interface{}{
			"count":          stats.Count,
			"first_occurred": stats.FirstOccurred.Unix(),
			"last_occurred":  stats.LastOccurred.Unix(),
			"description":    stats.Description,
		}
		
		typeLine := e.buildLineProtocol(e.measurement, typeTags, typeFields, timestamp)
		lines = append(lines, typeLine)
	}
	
	for severity, stats := range errors.ErrorsBySeverity {
		severityTags := make(map[string]string)
		for k, v := range baseTags {
			severityTags[k] = v
		}
		severityTags["metric_type"] = "error_by_severity"
		severityTags["severity"] = severity
		
		severityFields := map[string]interface{}{
			"count":          stats.Count,
			"first_occurred": stats.FirstOccurred.Unix(),
			"last_occurred":  stats.LastOccurred.Unix(),
			"description":    stats.Description,
		}
		
		severityLine := e.buildLineProtocol(e.measurement, severityTags, severityFields, timestamp)
		lines = append(lines, severityLine)
	}
	
	for statusCode, count := range errors.HTTPStatusCodes {
		statusTags := make(map[string]string)
		for k, v := range baseTags {
			statusTags[k] = v
		}
		statusTags["metric_type"] = "http_status"
		statusTags["status_code"] = fmt.Sprintf("%d", statusCode)
		
		statusFields := map[string]interface{}{
			"count": count,
		}
		
		statusLine := e.buildLineProtocol(e.measurement, statusTags, statusFields, timestamp)
		lines = append(lines, statusLine)
	}
	
	return lines
}

func (e *InfluxDBExporter) exportOperationMetrics(operations map[string]*metrics.OperationStats, baseTags map[string]string, timestamp int64) []string {
	var lines []string
	
	for opName, stats := range operations {
		tags := make(map[string]string)
		for k, v := range baseTags {
			tags[k] = v
		}
		tags["metric_type"] = "operation"
		tags["operation"] = opName
		
		fields := map[string]interface{}{
			"total_count":    stats.Total,
			"success_count":  stats.Success,
			"failure_count":  stats.Failed,
			"avg_duration":   stats.AvgTime,
			"total_duration": stats.TotalTime,
			"success_rate":   float64(stats.Success) / float64(stats.Total) * 100,
		}
		
		line := e.buildLineProtocol(e.measurement, tags, fields, timestamp)
		lines = append(lines, line)
	}
	
	return lines
}

func (e *InfluxDBExporter) buildLineProtocol(measurement string, tags map[string]string, fields map[string]interface{}, timestamp int64) string {
	var parts []string
	
	parts = append(parts, e.escapeMeasurement(measurement))
	
	if len(tags) > 0 {
		tagStr := e.formatTags(tags)
		parts[0] += "," + tagStr
	}
	
	fieldStr := e.formatFields(fields)
	parts = append(parts, fieldStr)
	
	parts = append(parts, strconv.FormatInt(timestamp, 10))
	
	return strings.Join(parts, " ")
}

func (e *InfluxDBExporter) formatTags(tags map[string]string) string {
	var pairs []string
	
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	for _, k := range keys {
		v := tags[k]
		pairs = append(pairs, fmt.Sprintf("%s=%s", 
			e.escapeTagKey(k), 
			e.escapeTagValue(v)))
	}
	
	return strings.Join(pairs, ",")
}

func (e *InfluxDBExporter) formatFields(fields map[string]interface{}) string {
	var pairs []string
	
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	for _, k := range keys {
		v := fields[k]
		pairs = append(pairs, fmt.Sprintf("%s=%s", 
			e.escapeFieldKey(k), 
			e.formatFieldValue(v)))
	}
	
	return strings.Join(pairs, ",")
}

func (e *InfluxDBExporter) formatFieldValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, e.escapeStringValue(v))
	case int:
		return fmt.Sprintf("%di", v)
	case int64:
		return fmt.Sprintf("%di", v)
	case uint64:
		return fmt.Sprintf("%di", v)
	case float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf(`"%v"`, v)
	}
}

func (e *InfluxDBExporter) escapeMeasurement(s string) string {
	s = strings.ReplaceAll(s, ",", `\,`)
	s = strings.ReplaceAll(s, " ", `\ `)
	return s
}

func (e *InfluxDBExporter) escapeTagKey(s string) string {
	s = strings.ReplaceAll(s, ",", `\,`)
	s = strings.ReplaceAll(s, "=", `\=`)
	s = strings.ReplaceAll(s, " ", `\ `)
	return s
}

func (e *InfluxDBExporter) escapeTagValue(s string) string {
	s = strings.ReplaceAll(s, ",", `\,`)
	s = strings.ReplaceAll(s, "=", `\=`)
	s = strings.ReplaceAll(s, " ", `\ `)
	return s
}

func (e *InfluxDBExporter) escapeFieldKey(s string) string {
	s = strings.ReplaceAll(s, ",", `\,`)
	s = strings.ReplaceAll(s, "=", `\=`)
	s = strings.ReplaceAll(s, " ", `\ `)
	return s
}

func (e *InfluxDBExporter) escapeStringValue(s string) string {
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

func (e *InfluxDBExporter) sanitizeTagValue(s string) string {
	if len(s) > 100 {
		s = s[:97] + "..."
	}
	
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	
	return s
}

func (e *InfluxDBExporter) Close() error {
	return nil
}