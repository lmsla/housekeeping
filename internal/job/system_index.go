package job

import "strings"

// systemIndexPrefixes 定義系統 index 的黑名單前綴
// 這些 index 是 Elasticsearch 內部使用的，不應該被刪除、關閉或修改
var systemIndexPrefixes = []string{
	".",                  // 所有以 . 開頭的 index (.kibana, .security, .monitoring-*, .tasks, etc.)
	"ilm-history-",       // ILM (Index Lifecycle Management) 歷史記錄
	"kibana_sample_",     // Kibana 範例數據
	".ds-",               // Data streams 系統 index
}

// isSystemIndex 檢查給定的 index 名稱是否為系統 index
//
// 參數:
//   indexName - 要檢查的 index 名稱
//
// 回傳:
//   true 如果是系統 index，false 如果不是
//
// 範例:
//   isSystemIndex(".kibana")           -> true
//   isSystemIndex(".security-7")       -> true
//   isSystemIndex("ilm-history-5-001") -> true
//   isSystemIndex("logs-app-2024.01")  -> false
func isSystemIndex(indexName string) bool {
	for _, prefix := range systemIndexPrefixes {
		if strings.HasPrefix(indexName, prefix) {
			return true
		}
	}
	return false
}
