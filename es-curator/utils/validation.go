package utils

import (
	"es-curator/global"
	"es-curator/structs"
	"fmt"
	"net/url"
	"strings"
)

// ValidateConfig 驗證完整配置
func ValidateConfig() error {
	if err := validateESConfig(); err != nil {
		return fmt.Errorf("ES 配置驗證失敗: %w", err)
	}

	if err := validateLogConfig(); err != nil {
		return fmt.Errorf("日誌配置驗證失敗: %w", err)
	}

	if err := validateInformationConfig(); err != nil {
		return fmt.Errorf("服務配置驗證失敗: %w", err)
	}

	if err := validateActions(); err != nil {
		return fmt.Errorf("Action 配置驗證失敗: %w", err)
	}

	return nil
}

// validateESConfig 驗證 ES 連接配置
func validateESConfig() error {
	cfg := global.EnvConfig.ES

	// 驗證 URL 非空且格式正確
	if len(cfg.URL) == 0 {
		return fmt.Errorf("es.url 不能為空")
	}

	for i, urlStr := range cfg.URL {
		if _, err := url.Parse(urlStr); err != nil {
			return fmt.Errorf("es.url[%d] 格式無效: %s - %v", i, urlStr, err)
		}
		if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
			return fmt.Errorf("es.url[%d] 必須以 http:// 或 https:// 開頭: %s", i, urlStr)
		}
	}

	// 驗證帳號密碼
	if cfg.SourceAccount == "" {
		return fmt.Errorf("es.sourceAccount 不能為空")
	}
	if cfg.SourcePassword == "" {
		return fmt.Errorf("es.sourcePassword 不能為空")
	}

	// 驗證重試次數範圍
	if cfg.MaxRetries < 0 || cfg.MaxRetries > 20 {
		return fmt.Errorf("es.max_retries 必須在 0-20 之間，當前值: %d", cfg.MaxRetries)
	}

	return nil
}

// validateLogConfig 驗證日誌配置
func validateLogConfig() error {
	cfg := global.EnvConfig.Log

	if cfg.Path == "" {
		return fmt.Errorf("log.path 不能為空")
	}

	if cfg.MaxSize <= 0 {
		return fmt.Errorf("log.maxSize 必須 > 0，當前值: %d", cfg.MaxSize)
	}

	if cfg.MaxBackups < 0 {
		return fmt.Errorf("log.maxBackups 不能為負數，當前值: %d", cfg.MaxBackups)
	}

	if cfg.MaxAge <= 0 {
		return fmt.Errorf("log.maxAge 必須 > 0，當前值: %d", cfg.MaxAge)
	}

	// 健康檢查間隔必須 > 0
	if cfg.HealthCheckInterval <= 0 {
		return fmt.Errorf("log.health_check_interval 必須 > 0，當前值: %d", cfg.HealthCheckInterval)
	}

	if cfg.HealthCheckInterval > 3600 {
		if global.Logger != nil {
			global.Logger.Warnw("⚠️  health_check_interval 過大，建議設置在 60-300 秒之間",
				"current", cfg.HealthCheckInterval)
		}
	}

	return nil
}

// validateInformationConfig 驗證服務配置
func validateInformationConfig() error {
	cfg := global.EnvConfig.INFORMATION

	// 如果啟用 cron，必須有有效的 period
	if cfg.ExecuteCron {
		if cfg.Period == "" {
			return fmt.Errorf("execute_cron 為 true 時，period 不能為空")
		}

		// 基本的 cron 表達式格式檢查（5 或 6 個欄位）
		fields := strings.Fields(cfg.Period)
		if len(fields) != 5 && len(fields) != 6 {
			return fmt.Errorf("period 格式錯誤，cron 表達式應該有 5 或 6 個欄位，當前有 %d 個: %s", len(fields), cfg.Period)
		}
	}

	return nil
}

// validateActions 驗證 Actions 配置
func validateActions() error {
	actions := global.ActionStruct.Actions

	if len(actions) == 0 {
		return fmt.Errorf("必須至少配置一個 action")
	}

	validActions := map[string]bool{
		"delete_indices": true,
		"allocation":     true,
		"forcemerge":     true,
		"close":          true,
		"open":           true,
		"rollover":       true,
	}

	for i, action := range actions {
		// 驗證 action type
		if !validActions[action.Action] {
			return fmt.Errorf("action[%d]: 無效的 action 類型 '%s'，有效值: %v",
				i, action.Action, getValidActionTypes(validActions))
		}

		// 驗證 rollover 必要選項
		if action.Action == "rollover" {
			if action.Options.RolloverAlias == "" {
				return fmt.Errorf("action[%d]: rollover action 需要 rollover_alias 選項", i)
			}
			if action.Options.MaxSize == "" && action.Options.MaxAge == "" && action.Options.MaxDocs == 0 {
				return fmt.Errorf("action[%d]: rollover action 至少需要 max_size、max_age 或 max_docs 其中之一", i)
			}
		}

		// 驗證 allocation 必要選項
		if action.Action == "allocation" {
			if action.Options.Key == "" {
				return fmt.Errorf("action[%d]: allocation action 需要 key 選項", i)
			}
			if action.Options.Value == "" {
				return fmt.Errorf("action[%d]: allocation action 需要 value 選項", i)
			}
		}

		// 驗證 forcemerge 必要選項
		if action.Action == "forcemerge" {
			if action.Options.MaxNumSegment <= 0 {
				return fmt.Errorf("action[%d]: forcemerge action 的 max_num_segment 必須 > 0，當前值: %d", i, action.Options.MaxNumSegment)
			}
		}

		// 驗證過濾器
		if err := validateFilters(i, action); err != nil {
			return err
		}
	}

	return nil
}

// validateFilters 驗證過濾器配置
func validateFilters(actionIndex int, action structs.Actiond) error {
	hasAge := false
	hasSpace := false
	hasWaterLevel := false
	hasPattern := false

	for j, filter := range action.Filters {
		switch filter.Filtertype {
		case "age":
			hasAge = true

			// 驗證 direction 有效性
			if filter.Direction != "older" && filter.Direction != "younger" && filter.Direction != "range" {
				return fmt.Errorf("action[%d].filters[%d]: age filter 的 direction 必須是 'older', 'younger' 或 'range'，當前值: '%s'", actionIndex, j, filter.Direction)
			}

			// range 模式的特殊驗證
			if filter.Direction == "range" {
				// range 模式需要 range_from 和 range_to
				if filter.RangeFrom <= 0 {
					return fmt.Errorf("action[%d].filters[%d]: direction='range' 時，range_from 必須 > 0，當前值: %d", actionIndex, j, filter.RangeFrom)
				}
				if filter.RangeTo <= 0 {
					return fmt.Errorf("action[%d].filters[%d]: direction='range' 時，range_to 必須 > 0，當前值: %d", actionIndex, j, filter.RangeTo)
				}
				// range_from 是起始時間（較遠），range_to 是結束時間（較近）
				// 例如：range_from=20 (20天前) 到 range_to=1 (1天前)
				if filter.RangeFrom <= filter.RangeTo {
					return fmt.Errorf("action[%d].filters[%d]: range_from (%d) 必須 > range_to (%d)，因為 range_from 是較遠的過去時間點", actionIndex, j, filter.RangeFrom, filter.RangeTo)
				}
			} else {
				// older/younger 模式需要 unit_count
				if filter.UnitCount <= 0 {
					return fmt.Errorf("action[%d].filters[%d]: direction='%s' 時，unit_count 必須 > 0，當前值: %d", actionIndex, j, filter.Direction, filter.UnitCount)
				}
			}

			// 驗證 unit 有效性
			validUnits := map[string]bool{
				"seconds": true,
				"minutes": true,
				"hours":   true,
				"days":    true,
				"weeks":   true,
				"months":  true,
				"years":   true, // 支持 years
			}
			if !validUnits[filter.Unit] {
				return fmt.Errorf("action[%d].filters[%d]: age filter 的 unit 無效 '%s'，有效值: seconds, minutes, hours, days, weeks, months, years", actionIndex, j, filter.Unit)
			}

		case "space":
			hasSpace = true
			if filter.DiskSpace <= 0 {
				return fmt.Errorf("action[%d].filters[%d]: space filter 的 disk_space 必須 > 0，當前值: %d", actionIndex, j, filter.DiskSpace)
			}

		case "water_level":
			hasWaterLevel = true
			// water_level 使用 UpperLimit 存儲閾值
			if filter.UpperLimit <= 0 || filter.UpperLimit > 100 {
				return fmt.Errorf("action[%d].filters[%d]: water_level 的 upper_limit 必須在 1-100 之間，當前值: %d", actionIndex, j, filter.UpperLimit)
			}

		case "pattern":
			hasPattern = true
			if len(filter.Value) == 0 {
				return fmt.Errorf("action[%d].filters[%d]: pattern filter 的 value 不能為空", actionIndex, j)
			}
			// 驗證 kind 有效性
			validKinds := map[string]bool{
				"prefix": true,
				"suffix": true,
				"regex":  true,
			}
			if filter.Kind != "" && !validKinds[filter.Kind] {
				return fmt.Errorf("action[%d].filters[%d]: pattern filter 的 kind 無效 '%s'，有效值: prefix, suffix, regex", actionIndex, j, filter.Kind)
			}

		case "node_role":
			// node_role 驗證
			if len(filter.Value) == 0 {
				return fmt.Errorf("action[%d].filters[%d]: node_role filter 的 value 不能為空", actionIndex, j)
			}
			validRoles := map[string]bool{
				"h": true, // hot
				"w": true, // warm
				"c": true, // cold
			}
			// 驗證每個角色值
			for _, role := range filter.Value {
				if !validRoles[role] {
					return fmt.Errorf("action[%d].filters[%d]: node_role filter 的 value 無效 '%s'，有效值: h (hot), w (warm), c (cold)", actionIndex, j, role)
				}
			}

		default:
			return fmt.Errorf("action[%d].filters[%d]: 未知的 filtertype '%s'", actionIndex, j, filter.Filtertype)
		}
	}

	// 檢查互斥過濾器
	if hasAge && hasSpace {
		return fmt.Errorf("action[%d]: age 和 space 過濾器不能同時使用", actionIndex)
	}
	if hasAge && hasWaterLevel {
		return fmt.Errorf("action[%d]: age 和 water_level 過濾器不能同時使用", actionIndex)
	}
	if hasSpace && hasWaterLevel {
		return fmt.Errorf("action[%d]: space 和 water_level 過濾器不能同時使用", actionIndex)
	}

	// 非 rollover action 建議有 pattern 過濾器（安全措施）
	if action.Action != "rollover" && !hasPattern {
		if global.Logger != nil {
			global.Logger.Warnw("⚠️  建議為所有非 rollover action 添加 pattern 過濾器以避免意外操作系統索引",
				"action_index", actionIndex,
				"action_type", action.Action)
		}
	}

	return nil
}

// getValidActionTypes 取得有效的 action 類型列表
func getValidActionTypes(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
