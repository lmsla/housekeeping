package job

import (
	"es-curator/global"
	"es-curator/structs"
	"fmt"
	"strconv"
	"time"
)

func logAction(uuid string, action string, description string) {
	actionMsg := fmt.Sprintf("Action: %s, Description: %s", action, description)

	global.Logger.Infow(actionMsg, "logType", "Procedures")

	if global.EnvConfig.Log.ToES && global.EnvConfig.INFORMATION.Test_mode {
		ProceduresLogToES(uuid, "Test mode", description, action)
	} else if global.EnvConfig.Log.ToES && !global.EnvConfig.INFORMATION.Test_mode {
		ProceduresLogToES(uuid, "Execution mode", description, action)
	}
}

// 採集 action 下的所有 filter
func processFilters(FilterList []structs.Filter, filter_record *[]string, role *[]string) {

	for filtertype := range FilterList {
		switch FilterList[filtertype].Filtertype {
		case "age":
			*filter_record = append(*filter_record, "age")
		case "pattern":
			*filter_record = append(*filter_record, "pattern")
		case "space":
			*filter_record = append(*filter_record, "space")
		case "node_role":
			*filter_record = append(*filter_record, "node_role")
			*role = FilterList[filtertype].Value
		case "water_level":
			*filter_record = append(*filter_record, "water_level")
		}
	}
}

func Action_controll() {
	var comparelist, role []string

	ActionList := global.ActionStruct.Actions
	for actions := range ActionList {
		// 控制 action 的 disabled
		if ActionList[actions].Options.DisableAction {
			continue
		}
		// 產生 UUID 作為此次執行的唯一識別碼
		uuid := fmt.Sprintf("%d-%s", time.Now().UnixNano(), strconv.FormatInt(time.Now().Unix(), 36))
		global.Logger.Infow(fmt.Sprintf("執行 ID: %s", uuid), "logType", "Procedures")

		var filter_record []string
		logAction(uuid, ActionList[actions].Action, ActionList[actions].Description)

		FilterList := ActionList[actions].Filters
		processFilters(FilterList, &filter_record, &role)

		filter_info := fmt.Sprintf("use these filter :%s", filter_record)
		global.Logger.Infow(filter_info, "logType", "Procedures")

		if global.EnvConfig.Log.ToES && global.EnvConfig.INFORMATION.Test_mode {
			ProceduresLogToES(uuid,"Test mode",filter_info, ActionList[actions].Action)
		} else if global.EnvConfig.Log.ToES && !global.EnvConfig.INFORMATION.Test_mode {
			ProceduresLogToES(uuid,"Execution mode",filter_info, ActionList[actions].Action)
		}

		if len(role) == 0 {
			comparelist = Filter_of_filter(filter_record, FilterList)
		} else {
			comparelist = Filters_With_node(role, filter_record, FilterList)
		}
		comparelist = RemoveDuplicates(comparelist)

		// 根據不同的 action 類型進行處理
		switch ActionList[actions].Action {
		case "allocation":
			handleAllocation(uuid, comparelist, ActionList[actions])
		case "forcemerge":
			handleForceMerge(uuid, comparelist, ActionList[actions])
		case "delete_indices":
			handleDeleteIndices(uuid, comparelist)
		case "close":
			handleClose(uuid, comparelist)
		case "open":
			handleOpen(uuid, comparelist)
		}
		delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)
		global.Logger.Infow(delaymsg)
		time.Sleep(time.Duration(ActionList[actions].Options.Delay) * time.Second)
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
	// fmt.Println("nodeNames", nodeNames)

	for _, node := range nodeNames {
		indices_on_node = append(indices_on_node, CatIndicesbyNodeName(node)...)
	}
	indices_on_node = RemoveDuplicates(indices_on_node)

	filtered_comparelist := []string{}
	compareSet := make(map[string]struct{})
	// fmt.Println("indices_on_node", indices_on_node)
	// fmt.Println("comparelist", comparelist)
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
	// fmt.Println("filtered_comparelist", filtered_comparelist)

	comparelist = filtered_comparelist

	if len(comparelist) != 0 {
		detailMsg := fmt.Sprintf("allocation these indices :%s to %s ", comparelist, action.Options.Value)
		global.Logger.Infow(detailMsg, "type", "Details")

		for _, index := range comparelist {
			index_onebyone := []string{index}
			if global.EnvConfig.INFORMATION.Test_mode {
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
			if global.EnvConfig.INFORMATION.Test_mode {
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
			if global.EnvConfig.INFORMATION.Test_mode {
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
			if global.EnvConfig.INFORMATION.Test_mode {
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
			if global.EnvConfig.INFORMATION.Test_mode {
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

	global.Detail_Logger.Infow(individual_Msg, "mode", "Test_mode", "logType", "Detail", "action", action, "pri", i.Pri, "Rep", i.Rep, "DocCount", i.DocsCount, "DocDelete", i.DocsDeleted, "StoreSize", i.StoreSize, "PriStoreSize", i.PriStoreSize, "CreationDate", CreationDate)
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
