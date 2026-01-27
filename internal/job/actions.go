package job

import (
	"housekeeping/internal/global"
	"housekeeping/internal/metrics"
	"housekeeping/internal/structs"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FilterInfo 包含過濾器處理的結果
type FilterInfo struct {
	FilterRecord []string
	Role         []string
	Filters      []structs.Filter
}

// ActionExecutor 統一處理 Action 執行邏輯，保留完整的測試模式功能
type ActionExecutor struct {
	UUID   string
	Action string
}

// ExecuteOnIndices 核心執行邏輯 - 完全保留原有的測試模式邏輯，新增指標收集
func (ae *ActionExecutor) ExecuteOnIndices(indices []string, operation func([]string)) {
	if len(indices) == 0 {
		global.Logger.Infow("No match indices", "logType", "Procedures")
		return
	}

	// 保留原有的逐個處理邏輯
	for _, index := range indices {
		index_onebyone := []string{index}

		// 記錄操作開始時間
		startTime := time.Now()
		var success bool = true

		// 完全保留原有的測試模式分支邏輯
		if global.EnvConfig.INFORMATION.TestMode {
			logTestMode(ae.UUID, index_onebyone, ae.Action)
			success = true // 測試模式總是成功
		} else {
			// 在實際操作中加入錯誤捕獲
			func() {
				defer func() {
					if r := recover(); r != nil {
						success = false
						global.Logger.Error(fmt.Sprintf("Operation %s failed on index %s: %v", ae.Action, index, r))
					}
				}()

				// 先記錄日誌
				logExecutionMode(ae.UUID, index_onebyone, ae.Action)
				// 再執行操作
				operation(index_onebyone) // 只在非測試模式執行實際操作

			}()
		}

		// 記錄操作指標
		duration := time.Since(startTime)

		// 記錄到本地指標系統
		if metrics.GlobalMetrics != nil {
			metrics.GlobalMetrics.RecordOperation(ae.Action, success, duration)
		}

		// 記錄操作詳細日誌
		statusMsg := "success"
		if !success {
			statusMsg = "failed"
		}
		global.Logger.Infow(
			fmt.Sprintf("Operation %s on index %s: %s (duration: %v)", ae.Action, index, statusMsg, duration),
			"logType", "Metrics",
			"action", ae.Action,
			"index", index,
			"success", success,
			"duration_ms", duration.Nanoseconds()/1e6,
		)
	}
}

// LogActionSummary 統一的 Action 摘要日誌記錄
func (ae *ActionExecutor) LogActionSummary(actionDesc string, indices []string) {
	detailMsg := fmt.Sprintf("%s these indices: %s, Number of indices: %d", actionDesc, indices, len(indices))
	global.Logger.Infow(detailMsg, "logType", "Procedures")
}

// LogActionSummaryWithParams 帶參數的 Action 摘要日誌記錄
func (ae *ActionExecutor) LogActionSummaryWithParams(actionDesc string, indices []string, params string) {
	detailMsg := fmt.Sprintf("%s these indices: %s, %s", actionDesc, indices, params)
	global.Logger.Infow(detailMsg, "logType", "Procedures")
}

// ProcessedAction 包含處理後的 Action 信息
type ProcessedAction struct {
	UUID      string
	Action    structs.Actiond
	IndexList []string
}

func logAction(uuid string, action string, description string) {
	actionMsg := fmt.Sprintf("Action: %s, Description: %s", action, description)

	global.Logger.Infow(actionMsg, "logType", "Procedures")

	if global.EnvConfig.Log.ToES && global.EnvConfig.INFORMATION.TestMode {
		ProceduresLogToES(uuid, "Test mode", description, action)
	} else if global.EnvConfig.Log.ToES && !global.EnvConfig.INFORMATION.TestMode {
		ProceduresLogToES(uuid, "Execution mode", description, action)
	}
}

// 採集 action 下的所有 filter
func processFilters(FilterList []structs.Filter, filterRecord *[]string, role *[]string) {

	for filtertype := range FilterList {
		switch FilterList[filtertype].Filtertype {
		case "age":
			*filterRecord = append(*filterRecord, "age")
		case "pattern":
			*filterRecord = append(*filterRecord, "pattern")
		case "space":
			*filterRecord = append(*filterRecord, "space")
		case "node_role":
			*filterRecord = append(*filterRecord, "node_role")
			*role = FilterList[filtertype].Value
		case "water_level":
			*filterRecord = append(*filterRecord, "water_level")
		}
	}
}

// generateExecutionID 生成唯一的執行 ID
func generateExecutionID() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), strconv.FormatInt(time.Now().Unix(), 36))
}

// logFilterInfoToES 將過濾器信息記錄到 ES
func logFilterInfoToES(uuid string, action structs.Actiond, filterInfo string) {
	if global.EnvConfig.Log.ToES && global.EnvConfig.INFORMATION.TestMode {
		ProceduresLogToES(uuid, "Test mode", filterInfo, action.Action)
	} else if global.EnvConfig.Log.ToES && !global.EnvConfig.INFORMATION.TestMode {
		ProceduresLogToES(uuid, "Execution mode", filterInfo, action.Action)
	}
}

// processFiltersAndLog 處理過濾器並記錄日誌
func processFiltersAndLog(uuid string, action structs.Actiond) FilterInfo {
	var filterRecord, role []string

	logAction(uuid, action.Action, action.Description)
	processFilters(action.Filters, &filterRecord, &role)

	filterInfo := fmt.Sprintf("use these filter :%s", filterRecord)
	global.Logger.Infow(filterInfo, "logType", "Procedures")

	logFilterInfoToES(uuid, action, filterInfo)

	return FilterInfo{
		FilterRecord: filterRecord,
		Role:         role,
		Filters:      action.Filters,
	}
}

// getFilteredIndices 獲取過濾後的索引清單
func getFilteredIndices(info FilterInfo) []string {
	var comparelist []string

	if len(info.Role) == 0 {
		comparelist = Filter_of_filter(info.FilterRecord, info.Filters)
	} else {
		comparelist = Filters_With_node(info.Role, info.FilterRecord, info.Filters)
	}

	return RemoveDuplicates(comparelist)
}

// applyDelay 執行延遲等待
func applyDelay(delaySeconds int) {
	if delaySeconds > 0 {
		delayMsg := fmt.Sprintf("Pausing for %v seconds before continuing...", delaySeconds)
		global.Logger.Infow(delayMsg)
		time.Sleep(time.Duration(delaySeconds) * time.Second)
	}
}

// executeAction 執行指定的 Action
func executeAction(uuid string, indexList []string, action structs.Actiond) {
	switch action.Action {
	case "allocation":
		handleAllocation(uuid, indexList, action)
	case "forcemerge":
		handleForceMerge(uuid, indexList, action)
	case "delete_indices":
		handleDeleteIndices(uuid, indexList)
	case "close":
		handleClose(uuid, indexList)
	case "open":
		handleOpen(uuid, indexList)
	case "rollover":
		handleRollover(uuid, action)
	}
}

// executeActionWithDelay 執行 Action 並應用延遲
func executeActionWithDelay(processed ProcessedAction) {
	executeAction(processed.UUID, processed.IndexList, processed.Action)
	applyDelay(processed.Action.Options.Delay)
}

// processAction 處理單個 Action (過濾器 + 索引獲取)
func processAction(uuid string, action structs.Actiond) ProcessedAction {
	// 處理過濾器並記錄日誌
	filterInfo := processFiltersAndLog(uuid, action)

	// 獲取過濾後的索引清單
	indexList := getFilteredIndices(filterInfo)

	return ProcessedAction{
		UUID:      uuid,
		Action:    action,
		IndexList: indexList,
	}
}

// Action_controll 主要的 Action 控制函數 - 重構後版本
func Action_controll() {
	ActionList := global.ActionStruct.Actions

	// 🚨 安全檢查：確認配置正確讀取
	if global.EnvConfig.Log.Path == "" {
		global.Logger.Error("CRITICAL: Configuration appears to be invalid - aborting execution")
		fmt.Printf("🚨 SAFETY ABORT: Configuration validation failed\n")
		return
	}

	fmt.Printf("🔒 執行模式確認: Test mode = %v\n", global.EnvConfig.INFORMATION.TestMode)

	for _, action := range ActionList {
		// 跳過被禁用的 Action
		if action.Options.DisableAction {
			continue
		}

		// 生成執行 ID
		executionID := generateExecutionID()
		global.Logger.Infow(fmt.Sprintf("執行 ID: %s", executionID), "logType", "Procedures")

		// 根據 action type 決定處理路徑
		if action.Action == "rollover" {
			// Rollover 不需要 filter，直接執行
			logAction(executionID, action.Action, action.Description)
			processedAction := ProcessedAction{
				UUID:      executionID,
				Action:    action,
				IndexList: nil, // Rollover 不依賴索引列表
			}
			executeActionWithDelay(processedAction)
		} else {
			// 其他 action 走標準的 filter 處理流程
			processedAction := processAction(executionID, action)
			executeActionWithDelay(processedAction)
		}
	}
}

func handleAllocation(uuid string, comparelist []string, action structs.Actiond) {
	var indices_on_node []string

	// 轉換 node role value
	roleMapping := map[string]string{
		"data_hot":  "h",
		"data_warm": "w",
		"data_cold": "c",
	}

	convertedValue, exists := roleMapping[action.Options.Value]
	if !exists {
		convertedValue = action.Options.Value // 保持原值，防止未知值影響
	}

	// 找出目前存放在指定 node role 下的 index 並去除重複 (一個index 可能會有多個 shards)
	nodeNames := NodeRoleDetermination(convertedValue)

	for _, node := range nodeNames {
		indices_on_node = append(indices_on_node, CatIndicesbyNodeName(node)...)
	}
	indices_on_node = RemoveDuplicates(indices_on_node)

	filtered_comparelist := []string{}
	compareSet := make(map[string]struct{})
	// 將 indices_on_node 轉換為 Set，提高查找效率
	for _, index := range indices_on_node {
		compareSet[index] = struct{}{}
	}

	// 只保留 comparelist 中不在 indices_on_node 的索引
	for _, index := range comparelist {
		if _, exists := compareSet[index]; !exists {
			filtered_comparelist = append(filtered_comparelist, index)
		}
	}

	if len(filtered_comparelist) != 0 {
		executor := &ActionExecutor{UUID: uuid, Action: action.Action}
		params := fmt.Sprintf("to %s", action.Options.Value)
		executor.LogActionSummaryWithParams("allocation", filtered_comparelist, params)

		// 創建包含所有 allocation 參數的操作函數
		operation := func(indices []string) {
			Allocation(indices, action.Options.AllocationType, action.Options.Key, action.Options.Value)
		}
		executor.ExecuteOnIndices(filtered_comparelist, operation)
		Node_relocating_checking()
	} else {
		global.Logger.Infow("No match indices (all indices already on target node)", "logType", "Procedures")
	}
}

func handleForceMerge(uuid string, comparelist []string, action structs.Actiond) {
	executor := &ActionExecutor{UUID: uuid, Action: action.Action}

	if comparelist != nil {
		params := fmt.Sprintf("segment num: %v", action.Options.MaxNumSegment)
		executor.LogActionSummaryWithParams("forcemerge", comparelist, params)

		// 創建一個包含 MaxNumSegment 參數的操作函數
		operation := func(indices []string) {
			ForceMerge(indices, action.Options.MaxNumSegment)
		}
		executor.ExecuteOnIndices(comparelist, operation)
	} else {
		global.Logger.Infow("No match indices", "logType", action.Action)
	}
}

func handleDeleteIndices(uuid string, comparelist []string) {
	executor := &ActionExecutor{UUID: uuid, Action: "delete_indices"}

	if comparelist != nil {
		// 🔒 P0-002 修復：系統 index 最終防護層
		// 即使 filter 層級已經過濾，這裡作為最後一道防線再次檢查
		systemIndices := []string{}
		safeIndices := []string{}

		for _, idx := range comparelist {
			if isSystemIndex(idx) {
				systemIndices = append(systemIndices, idx)
			} else {
				safeIndices = append(safeIndices, idx)
			}
		}

		// 如果發現系統 index，拒絕整個操作
		if len(systemIndices) > 0 {
			global.Logger.Errorw("🚨 CRITICAL: Attempting to delete SYSTEM indices - OPERATION ABORTED",
				"system_indices", systemIndices,
				"system_indices_count", len(systemIndices),
				"total_indices_in_list", len(comparelist),
				"uuid", uuid,
				"action", "delete_indices_blocked",
				"logType", "Procedures")
			global.Logger.Errorw("⛔ DELETE operation has been BLOCKED to prevent system index deletion",
				"blocked_count", len(systemIndices),
				"safe_count", len(safeIndices),
				"logType", "Procedures")
			return // 中止操作，不刪除任何 index
		}

		// 🔒 安全警告：大量刪除操作
		if len(comparelist) > 100 {
			global.Logger.Warnw("⚠️  Large deletion operation detected",
				"index_count", len(comparelist),
				"uuid", uuid,
				"warning", "Deleting more than 100 indices - please ensure this is intentional",
				"logType", "Procedures")
		}

		executor.LogActionSummary("Delete", comparelist)
		executor.ExecuteOnIndices(comparelist, DeleteIndex)
	} else {
		global.Logger.Infow("No match indices", "logType", "Procedures")
	}
}

func handleClose(uuid string, comparelist []string) {
	executor := &ActionExecutor{UUID: uuid, Action: "close"}

	if comparelist != nil {
		executor.LogActionSummary("Close", comparelist)
		executor.ExecuteOnIndices(comparelist, CloseIndices)
	} else {
		global.Logger.Infow("No match indices", "logType", "Procedures")
	}
}

func handleOpen(uuid string, comparelist []string) {
	executor := &ActionExecutor{UUID: uuid, Action: "open"}

	if comparelist != nil {
		executor.LogActionSummary("Open", comparelist)
		executor.ExecuteOnIndices(comparelist, OpenIndices)
	} else {
		global.Logger.Infow("No match indices", "logType", "Procedures")
	}
}

func handleRollover(uuid string, action structs.Actiond) {
	// Rollover 不需要索引清單，直接基於 alias 操作
	if action.Options.RolloverAlias == "" {
		global.Logger.Errorw("Rollover alias is required but not specified",
			"uuid", uuid, "action", action.Action)
		return
	}

	// 記錄 rollover 操作詳情
	conditions := []string{}
	if action.Options.MaxSize != "" {
		conditions = append(conditions, fmt.Sprintf("max_size: %s", action.Options.MaxSize))
	}
	if action.Options.MaxDocs > 0 {
		conditions = append(conditions, fmt.Sprintf("max_docs: %d", action.Options.MaxDocs))
	}
	if action.Options.MaxAge != "" {
		conditions = append(conditions, fmt.Sprintf("max_age: %s", action.Options.MaxAge))
	}

	conditionsStr := "no conditions"
	if len(conditions) > 0 {
		conditionsStr = fmt.Sprintf("[%s]", strings.Join(conditions, ", "))
	}

	detailMsg := fmt.Sprintf("Rollover alias '%s' with conditions: %s", action.Options.RolloverAlias, conditionsStr)
	global.Logger.Infow(detailMsg, "logType", "Procedures", "uuid", uuid)

	// 記錄操作開始時間進行指標追蹤
	startTime := time.Now()
	success := true

	// 完全保留原有的測試模式分支邏輯
	if global.EnvConfig.INFORMATION.TestMode {
		global.Logger.Infow(fmt.Sprintf("Rollover alias '%s' processed in test mode", action.Options.RolloverAlias),
			"logType", "Procedures", "mode", "TestMode", "uuid", uuid)
		success = true // 測試模式總是成功
	} else {
		// 實際執行 rollover 操作
		func() {
			defer func() {
				if r := recover(); r != nil {
					success = false
					global.Logger.Error(fmt.Sprintf("Rollover operation failed on alias %s: %v", action.Options.RolloverAlias, r))
				}
			}()

			global.Logger.Infow(fmt.Sprintf("Executing rollover on alias '%s'", action.Options.RolloverAlias),
				"logType", "Procedures", "uuid", uuid)

			if err := Rollover(
				action.Options.RolloverAlias,
				action.Options.MaxSize,
				action.Options.MaxDocs,
				action.Options.MaxAge,
				action.Options.NewIndexName,
			); err != nil {
				success = false
			}
		}()
	}

	// 記錄操作指標
	duration := time.Since(startTime)

	// 記錄到本地指標系統
	if metrics.GlobalMetrics != nil {
		metrics.GlobalMetrics.RecordOperation(action.Action, success, duration)
	}

	// 記錄操作詳細日誌
	statusMsg := "success"
	if !success {
		statusMsg = "failed"
	}
	global.Logger.Infow(
		fmt.Sprintf("Rollover operation on alias %s: %s (duration: %v)", action.Options.RolloverAlias, statusMsg, duration),
		"logType", "Metrics",
		"action", action.Action,
		"alias", action.Options.RolloverAlias,
		"success", success,
		"duration_ms", duration.Nanoseconds()/1e6,
		"uuid", uuid,
	)
}

func logTestMode(uuid string, index_onebyone []string, action string) {
	indicesInfo := CatIndices_withPattern(index_onebyone)
	i := indicesInfo[0]
	timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
	CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
	individual_Msg := fmt.Sprintf("%s has already processed in test mode", index_onebyone[0])

	global.Detail_Logger.Infow(individual_Msg, "mode", "TestMode", "logType", "Detail", "action", action, "pri", i.Pri, "Rep", i.Rep, "DocCount", i.DocsCount, "DocDelete", i.DocsDeleted, "StoreSize", i.StoreSize, "PriStoreSize", i.PriStoreSize, "CreationDate", CreationDate)
	mode := "Test mode"
	if global.EnvConfig.Log.ToES {
		LogToES(uuid, individual_Msg, mode, action, i.Pri, i.Rep, i.DocsCount, i.DocsDeleted, i.StoreSize, i.PriStoreSize, CreationDate, index_onebyone[0])
	}
}

func logExecutionMode(uuid string, index_onebyone []string, action string) {
	indicesInfo := CatIndices_withPattern(index_onebyone)
	if len(indicesInfo) == 0 {
		global.Logger.Errorw("Failed to get index info for logging",
			"logType", "Detail",
			"index", index_onebyone[0],
			"action", action,
			"uuid", uuid)
		return
	}
	i := indicesInfo[0]
	timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
	CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
	individual_Msg := fmt.Sprintf("%s has been processed in execution mode", index_onebyone[0])
	mode := "Execution mode"
	global.Detail_Logger.Infow(individual_Msg, "mode", "execution_mode", "logType", "Detail", "action", action, "pri", i.Pri, "Rep", i.Rep, "DocCount", i.DocsCount, "DocDelete", i.DocsDeleted, "StoreSize", i.StoreSize, "PriStoreSize", i.PriStoreSize, "CreationDate", CreationDate)
	if global.EnvConfig.Log.ToES {
		LogToES(uuid, individual_Msg, mode, action, i.Pri, i.Rep, i.DocsCount, i.DocsDeleted, i.StoreSize, i.PriStoreSize, CreationDate, index_onebyone[0])
	}
}
