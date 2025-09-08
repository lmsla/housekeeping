package job

import (
	"es-curator/global"
	"es-curator/structs"
	"fmt"
	"strconv"
	"time"
)

// FilterInfo 包含過濾器處理的結果
type FilterInfo struct {
	FilterRecord []string
	Role         []string
	Filters      []structs.Filter
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
		
		// 處理 Action (過濾器處理 + 索引獲取)
		processedAction := processAction(executionID, action)
		
		// 執行 Action 並應用延遲
		executeActionWithDelay(processedAction)
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

	comparelist = filtered_comparelist

	if len(comparelist) != 0 {
		detailMsg := fmt.Sprintf("allocation these indices :%s to %s ", comparelist, action.Options.Value)
		global.Logger.Infow(detailMsg, "type", "Details")

		for _, index := range comparelist {
			index_onebyone := []string{index}
			if global.EnvConfig.INFORMATION.TestMode {
				logTestMode(uuid, index_onebyone, action.Action)
			} else {
				logExecutionMode(uuid, index_onebyone, action.Action)
				Allocation(index_onebyone, action.Options.AllocationType, action.Options.Key, action.Options.Value)
			}
		}
		Node_relocating_checking()
	} else {
		global.Logger.Infow("No match indices", "logType", "Procedures")
	}
}

func handleForceMerge(uuid string, comparelist []string, action structs.Actiond) {
	if comparelist != nil {
		detailMsg := fmt.Sprintf("forcemerge these indices :%s, segment num :%v", comparelist, action.Options.MaxNumSegment)
		global.Logger.Infow(detailMsg, "logType", "Procedures")

		for _, index := range comparelist {
			index_onebyone := []string{index}
			if global.EnvConfig.INFORMATION.TestMode {
				logTestMode(uuid, index_onebyone, action.Action)
			} else {
				logExecutionMode(uuid, index_onebyone, action.Action)
				ForceMerge(index_onebyone, action.Options.MaxNumSegment)
			}
		}
	} else {
		global.Logger.Infow("No match indices", "logType", action.Action)
	}
}

func handleDeleteIndices(uuid string, comparelist []string) {
	if comparelist != nil {
		detailMsg := fmt.Sprintf("Delete these indices : %s, Number of indices : %d", comparelist, len(comparelist))
		global.Logger.Infow(detailMsg, "logType", "Procedures")

		for _, index := range comparelist {
			index_onebyone := []string{index}
			if global.EnvConfig.INFORMATION.TestMode {
				logTestMode(uuid, index_onebyone, "delete_indices")
			} else {
				logExecutionMode(uuid, index_onebyone, "delete_indices")
				DeleteIndex(index_onebyone)
			}
		}
	} else {
		global.Logger.Infow("No match indices", "logType", "Procedures")
	}
}

func handleClose(uuid string, comparelist []string) {
	if comparelist != nil {
		detailMsg := fmt.Sprintf("Close these indices :%s", comparelist)
		global.Logger.Infow(detailMsg, "logType", "Procedures")

		for _, index := range comparelist {
			index_onebyone := []string{index}
			if global.EnvConfig.INFORMATION.TestMode {
				logTestMode(uuid, index_onebyone, "close")
			} else {
				logExecutionMode(uuid, index_onebyone, "close")
				CloseIndices(index_onebyone)
			}
		}
	} else {
		global.Logger.Infow("No match indices", "logType", "Procedures")
	}
}

func handleOpen(uuid string, comparelist []string) {
	if comparelist != nil {
		detailMsg := fmt.Sprintf("Open these indices :%s", comparelist)
		global.Logger.Infow(detailMsg, "logType", "Procedures")

		for _, index := range comparelist {
			index_onebyone := []string{index}
			if global.EnvConfig.INFORMATION.TestMode {
				logTestMode(uuid, index_onebyone, "open")
			} else {
				logExecutionMode(uuid, index_onebyone, "open")
				OpenIndices(index_onebyone)
			}
		}
	} else {
		global.Logger.Infow("No match indices", "logType", "Procedures")
	}
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
	i := indicesInfo[0]
	timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
	CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
	individual_Msg := fmt.Sprintf("%s has already processed in execution mode", index_onebyone[0])
	mode := "Execution mode"
	global.Detail_Logger.Infow(individual_Msg, "mode", "execution_mode", "logType", "Detail", "action", action, "pri", i.Pri, "Rep", i.Rep, "DocCount", i.DocsCount, "DocDelete", i.DocsDeleted, "StoreSize", i.StoreSize, "PriStoreSize", i.PriStoreSize, "CreationDate", CreationDate)
	if global.EnvConfig.Log.ToES {
		LogToES(uuid, individual_Msg, mode, action, i.Pri, i.Rep, i.DocsCount, i.DocsDeleted, i.StoreSize, i.PriStoreSize, CreationDate, index_onebyone[0])
	}
}
