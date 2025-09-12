package export

import (
	"fmt"
	"sort"
	"strings"
	"github.com/bimap-org/universal-monitoring/pkg/metrics"
)

type PrometheusExporter struct {
	namespace    string
	subsystem    string
	labels       map[string]string
	metricsFilter *MetricsFilter
}

type PrometheusExporterConfig struct {
	Namespace     string            `yaml:"namespace"`
	Subsystem     string            `yaml:"subsystem"`
	Labels        map[string]string `yaml:"labels"`
	MetricsFilter *MetricsFilter    `yaml:"metrics_filter"`
}

type MetricsFilter struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}

func NewPrometheusExporter(config PrometheusExporterConfig) *PrometheusExporter {
	if config.Namespace == "" {
		config.Namespace = "service"
	}
	
	if config.Labels == nil {
		config.Labels = make(map[string]string)
	}

	return &PrometheusExporter{
		namespace:     config.Namespace,
		subsystem:     config.Subsystem,
		labels:        config.Labels,
		metricsFilter: config.MetricsFilter,
	}
}

func (e *PrometheusExporter) Export(m *metrics.UniversalMetrics) ([]byte, error) {
	var lines []string
	
	serviceLabels := e.buildServiceLabels(m)
	
	if m.Resource != nil {
		lines = append(lines, e.exportResourceMetrics(m.Resource, serviceLabels)...)
	}
	
	if len(m.Dependencies) > 0 {
		lines = append(lines, e.exportDependencyMetrics(m.Dependencies, serviceLabels)...)
	}
	
	if m.Errors != nil {
		lines = append(lines, e.exportErrorMetrics(m.Errors, serviceLabels)...)
	}
	
	if len(m.OperationMetrics) > 0 {
		lines = append(lines, e.exportOperationMetrics(m.OperationMetrics, serviceLabels)...)
	}
	
	filteredLines := e.applyMetricsFilter(lines)
	
	return []byte(strings.Join(filteredLines, "\n") + "\n"), nil
}

func (e *PrometheusExporter) buildServiceLabels(m *metrics.UniversalMetrics) map[string]string {
	labels := make(map[string]string)
	
	for k, v := range e.labels {
		labels[k] = v
	}
	
	labels["service"] = m.ServiceName
	labels["service_type"] = m.ServiceType
	labels["version"] = m.ServiceVersion
	
	return labels
}

func (e *PrometheusExporter) exportResourceMetrics(resource *metrics.ResourceMetrics, labels map[string]string) []string {
	var lines []string
	labelStr := e.formatLabels(labels)
	
	lines = append(lines, e.buildMetric(
		"resource_memory_usage_bytes",
		"gauge",
		"Current memory usage in bytes",
		resource.MemoryUsage,
		labelStr,
	))
	
	lines = append(lines, e.buildMetric(
		"resource_heap_inuse_bytes", 
		"gauge",
		"Heap memory in use in bytes",
		resource.HeapInuse,
		labelStr,
	))
	
	lines = append(lines, e.buildMetric(
		"resource_heap_sys_bytes",
		"gauge", 
		"Heap memory obtained from system in bytes",
		resource.HeapSys,
		labelStr,
	))
	
	lines = append(lines, e.buildMetric(
		"resource_heap_objects_total",
		"gauge",
		"Number of heap objects",
		resource.HeapObjects,
		labelStr,
	))
	
	lines = append(lines, e.buildMetric(
		"resource_goroutines_total",
		"gauge",
		"Number of goroutines",
		resource.GoroutineCount,
		labelStr,
	))
	
	lines = append(lines, e.buildMetric(
		"resource_gc_pause_time_nanoseconds",
		"gauge",
		"GC pause time in nanoseconds",
		resource.GCPauseTime,
		labelStr,
	))
	
	lines = append(lines, e.buildMetric(
		"resource_gc_runs_total",
		"counter",
		"Total number of GC runs",
		resource.GCRuns,
		labelStr,
	))
	
	return lines
}

func (e *PrometheusExporter) exportDependencyMetrics(dependencies map[string]*metrics.DependencyHealth, labels map[string]string) []string {
	var lines []string
	
	for name, dep := range dependencies {
		depLabels := make(map[string]string)
		for k, v := range labels {
			depLabels[k] = v
		}
		depLabels["dependency"] = name
		depLabels["dependency_type"] = dep.Type
		depLabels["endpoint"] = dep.Endpoint
		
		labelStr := e.formatLabels(depLabels)
		
		var statusValue float64
		switch dep.Status {
		case "connected":
			statusValue = 1
		case "degraded":
			statusValue = 0.5
		default:
			statusValue = 0
		}
		
		lines = append(lines, e.buildMetric(
			"dependency_status",
			"gauge",
			"Dependency connection status (1=connected, 0.5=degraded, 0=disconnected)",
			statusValue,
			labelStr,
		))
		
		lines = append(lines, e.buildMetric(
			"dependency_response_time_nanoseconds",
			"gauge",
			"Dependency response time in nanoseconds",
			dep.ResponseTime.Nanoseconds(),
			labelStr,
		))
		
		lines = append(lines, e.buildMetric(
			"dependency_avg_response_time_nanoseconds",
			"gauge",
			"Dependency average response time in nanoseconds",
			dep.AvgResponseTime.Nanoseconds(),
			labelStr,
		))
		
		lines = append(lines, e.buildMetric(
			"dependency_connection_attempts_total",
			"counter",
			"Total dependency connection attempts",
			dep.ConnectionAttempts,
			labelStr,
		))
		
		lines = append(lines, e.buildMetric(
			"dependency_successful_connections_total",
			"counter",
			"Total successful dependency connections",
			dep.SuccessfulConnections,
			labelStr,
		))
		
		lines = append(lines, e.buildMetric(
			"dependency_failed_connections_total",
			"counter",
			"Total failed dependency connections",
			dep.FailedConnections,
			labelStr,
		))
	}
	
	return lines
}

func (e *PrometheusExporter) exportErrorMetrics(errors *metrics.ErrorMetrics, labels map[string]string) []string {
	var lines []string
	labelStr := e.formatLabels(labels)
	
	lines = append(lines, e.buildMetric(
		"errors_total",
		"counter",
		"Total number of errors",
		errors.TotalErrors,
		labelStr,
	))
	
	lines = append(lines, e.buildMetric(
		"error_rate",
		"gauge",
		"Current error rate",
		errors.ErrorRate,
		labelStr,
	))
	
	for errorType, stats := range errors.ErrorsByType {
		typeLabels := make(map[string]string)
		for k, v := range labels {
			typeLabels[k] = v
		}
		typeLabels["error_type"] = errorType
		typeLabelStr := e.formatLabels(typeLabels)
		
		lines = append(lines, e.buildMetric(
			"errors_by_type_total",
			"counter",
			"Total errors by type",
			stats.Count,
			typeLabelStr,
		))
	}
	
	for severity, stats := range errors.ErrorsBySeverity {
		severityLabels := make(map[string]string)
		for k, v := range labels {
			severityLabels[k] = v
		}
		severityLabels["severity"] = severity
		severityLabelStr := e.formatLabels(severityLabels)
		
		lines = append(lines, e.buildMetric(
			"errors_by_severity_total",
			"counter",
			"Total errors by severity",
			stats.Count,
			severityLabelStr,
		))
	}
	
	for statusCode, count := range errors.HTTPStatusCodes {
		statusLabels := make(map[string]string)
		for k, v := range labels {
			statusLabels[k] = v
		}
		statusLabels["status_code"] = fmt.Sprintf("%d", statusCode)
		statusLabelStr := e.formatLabels(statusLabels)
		
		lines = append(lines, e.buildMetric(
			"http_status_codes_total",
			"counter",
			"HTTP status code counts",
			count,
			statusLabelStr,
		))
	}
	
	return lines
}

func (e *PrometheusExporter) exportOperationMetrics(operations map[string]*metrics.OperationStats, labels map[string]string) []string {
	var lines []string
	
	for opName, stats := range operations {
		opLabels := make(map[string]string)
		for k, v := range labels {
			opLabels[k] = v
		}
		opLabels["operation"] = opName
		labelStr := e.formatLabels(opLabels)
		
		lines = append(lines, e.buildMetric(
			"operation_total",
			"counter",
			"Total operations executed",
			stats.Total,
			labelStr,
		))
		
		lines = append(lines, e.buildMetric(
			"operation_success_total",
			"counter",
			"Total successful operations",
			stats.Success,
			labelStr,
		))
		
		lines = append(lines, e.buildMetric(
			"operation_failure_total",
			"counter",
			"Total failed operations",
			stats.Failed,
			labelStr,
		))
		
		lines = append(lines, e.buildMetric(
			"operation_avg_duration_milliseconds",
			"gauge",
			"Average operation duration in milliseconds",
			stats.AvgTime,
			labelStr,
		))
		
		lines = append(lines, e.buildMetric(
			"operation_total_duration_milliseconds",
			"counter",
			"Total operation duration in milliseconds",
			stats.TotalTime,
			labelStr,
		))
	}
	
	return lines
}

func (e *PrometheusExporter) buildMetric(name, metricType, help string, value interface{}, labels string) string {
	fullName := e.buildMetricName(name)
	
	helpLine := fmt.Sprintf("# HELP %s %s", fullName, help)
	typeLine := fmt.Sprintf("# TYPE %s %s", fullName, metricType)
	valueLine := fmt.Sprintf("%s%s %v", fullName, labels, value)
	
	return fmt.Sprintf("%s\n%s\n%s", helpLine, typeLine, valueLine)
}

func (e *PrometheusExporter) buildMetricName(name string) string {
	parts := []string{e.namespace}
	
	if e.subsystem != "" {
		parts = append(parts, e.subsystem)
	}
	
	parts = append(parts, name)
	
	return strings.Join(parts, "_")
}

func (e *PrometheusExporter) formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	
	var pairs []string
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	for _, k := range keys {
		v := labels[k]
		pairs = append(pairs, fmt.Sprintf(`%s="%s"`, k, strings.ReplaceAll(v, `"`, `\"`)))
	}
	
	return "{" + strings.Join(pairs, ",") + "}"
}

func (e *PrometheusExporter) applyMetricsFilter(lines []string) []string {
	if e.metricsFilter == nil {
		return lines
	}
	
	var filtered []string
	
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		
		if e.shouldIncludeMetric(line) {
			filtered = append(filtered, line)
		}
	}
	
	return filtered
}

func (e *PrometheusExporter) shouldIncludeMetric(line string) bool {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return false
	}
	
	metricName := strings.Split(parts[0], "{")[0]
	
	if len(e.metricsFilter.Exclude) > 0 {
		for _, pattern := range e.metricsFilter.Exclude {
			if e.matchesPattern(metricName, pattern) {
				return false
			}
		}
	}
	
	if len(e.metricsFilter.Include) > 0 {
		for _, pattern := range e.metricsFilter.Include {
			if e.matchesPattern(metricName, pattern) {
				return true
			}
		}
		return false
	}
	
	return true
}

func (e *PrometheusExporter) matchesPattern(name, pattern string) bool {
	if strings.Contains(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(name, prefix)
	}
	return name == pattern
}

func (e *PrometheusExporter) Close() error {
	return nil
}