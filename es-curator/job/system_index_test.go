package job

import "testing"

func TestIsSystemIndex(t *testing.T) {
	tests := []struct {
		name      string
		indexName string
		want      bool
	}{
		// 系統 index (應該返回 true)
		{
			name:      "Kibana index",
			indexName: ".kibana",
			want:      true,
		},
		{
			name:      "Kibana versioned index",
			indexName: ".kibana_7.10.0",
			want:      true,
		},
		{
			name:      "Security index",
			indexName: ".security-7",
			want:      true,
		},
		{
			name:      "Monitoring index",
			indexName: ".monitoring-es-7-2024.01.01",
			want:      true,
		},
		{
			name:      "Tasks index",
			indexName: ".tasks",
			want:      true,
		},
		{
			name:      "ILM history index",
			indexName: "ilm-history-5-000001",
			want:      true,
		},
		{
			name:      "Kibana sample data",
			indexName: "kibana_sample_data_logs",
			want:      true,
		},
		{
			name:      "Data streams system index",
			indexName: ".ds-logs-2024.01.01",
			want:      true,
		},
		// 一般 index (應該返回 false)
		{
			name:      "Regular logs index",
			indexName: "logs-app-2024.01.01",
			want:      false,
		},
		{
			name:      "Application index",
			indexName: "myapp-index-001",
			want:      false,
		},
		{
			name:      "Logstash index",
			indexName: "logstash-2024.01.01",
			want:      false,
		},
		{
			name:      "Index starting with dot-like but not dot",
			indexName: "dot.index",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSystemIndex(tt.indexName)
			if got != tt.want {
				t.Errorf("isSystemIndex(%q) = %v, want %v", tt.indexName, got, tt.want)
			}
		})
	}
}
