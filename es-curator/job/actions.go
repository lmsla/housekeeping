package job

import (
	"es-curator/global"
	"es-curator/structs"
	"fmt"
	"strconv"
	"time"
)

func logAction(action string, description string) {
	actionMsg := fmt.Sprintf("Action: %s, Description: %s", action, description)
	global.Logger.Infow(actionMsg, "logType", "Procedures")
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
		var filter_record []string
		logAction(ActionList[actions].Action, ActionList[actions].Description)

		FilterList := ActionList[actions].Filters
		processFilters(FilterList, &filter_record, &role)

		filter_info := fmt.Sprintf("use these filter :%s", filter_record)
		global.Logger.Infow(filter_info, "logType", "Procedures")

		if len(role) == 0 {
			comparelist = Filter_of_filter(filter_record, FilterList)
		} else {
			comparelist = Filters_With_node(role, filter_record, FilterList)
		}
		comparelist = RemoveDuplicates(comparelist)

		// 根據不同的 action 類型進行處理
		switch ActionList[actions].Action {
		case "allocation":
			handleAllocation(comparelist, ActionList[actions])
		case "forcemerge":
			handleForceMerge(comparelist, ActionList[actions])
		case "delete_indices":
			handleDeleteIndices(comparelist)
		case "close":
			handleClose(comparelist)
		case "open":
			handleOpen(comparelist)
		}
		delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)
		global.Logger.Infow(delaymsg)
		time.Sleep(time.Duration(ActionList[actions].Options.Delay) * time.Second)
	}
}

func handleAllocation(comparelist []string, action structs.Actiond) {
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
				logTestMode(index_onebyone, "Allocation")
			} else {
				Allocation(index_onebyone, action.Options.AllocationType, action.Options.Key, action.Options.Value)
				logExecutionMode(index_onebyone, "Allocation")
			}
		}
		Node_relocating_checking()
	} else {
		global.Logger.Infow("No match indices", "logType", "Procedures")
	}
}

func handleForceMerge(comparelist []string, action structs.Actiond) {
	if comparelist != nil {
		detailMsg := fmt.Sprintf("forcemerge these indices :%s, segment num :%v", comparelist, action.Options.MaxNumSegment)
		global.Logger.Infow(detailMsg, "logType", "Procedures")

		for _, index := range comparelist {
			index_onebyone := []string{index}
			if global.EnvConfig.INFORMATION.Test_mode {
				logTestMode(index_onebyone, "Forcemerge")
			} else {
				ForceMerge(index_onebyone, action.Options.MaxNumSegment)
				logExecutionMode(index_onebyone, "Forcemerge")
			}
		}
	} else {
		global.Logger.Infow("No match indices", "logType", "Procedures")
	}
}

func handleDeleteIndices(comparelist []string) {
	if comparelist != nil {
		detailMsg := fmt.Sprintf("Delete these indices : %s, Number of indices : %d", comparelist, len(comparelist))
		global.Logger.Infow(detailMsg, "logType", "Procedures")

		for _, index := range comparelist {
			index_onebyone := []string{index}
			if global.EnvConfig.INFORMATION.Test_mode {
				logTestMode(index_onebyone, "Delete")
			} else {
				DeleteIndex(index_onebyone)
				logExecutionMode(index_onebyone, "Delete")
			}
		}
	} else {
		global.Logger.Infow("No match indices", "logType", "Procedures")
	}
}

func handleClose(comparelist []string) {
	if comparelist != nil {
		detailMsg := fmt.Sprintf("Close these indices :%s", comparelist)
		global.Logger.Infow(detailMsg, "logType", "Procedures")

		for _, index := range comparelist {
			index_onebyone := []string{index}
			if global.EnvConfig.INFORMATION.Test_mode {
				logTestMode(index_onebyone, "Close")
			} else {
				CloseIndices(index_onebyone)
				logExecutionMode(index_onebyone, "Close")
			}
		}
	} else {
		global.Logger.Infow("No match indices", "logType", "Procedures")
	}
}

func handleOpen(comparelist []string) {
	if comparelist != nil {
		detailMsg := fmt.Sprintf("Open these indices :%s", comparelist)
		global.Logger.Infow(detailMsg, "logType", "Procedures")

		for _, index := range comparelist {
			index_onebyone := []string{index}
			if global.EnvConfig.INFORMATION.Test_mode {
				logTestMode(index_onebyone, "Open")
			} else {
				OpenIndices(index_onebyone)
				logExecutionMode(index_onebyone, "Open")
			}
		}
	} else {
		global.Logger.Infow("No match indices", "logType", "Procedures")
	}
}

func logTestMode(index_onebyone []string, action string) {
	indicesInfo := CatIndices_withPattern(index_onebyone)
	i := indicesInfo[0]
	timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
	CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
	individual_Msg := fmt.Sprintf("%s has already processed in test mode", index_onebyone)

	global.Detail_Logger.Infow(individual_Msg, "mode", "Test_mode", "logType", "Detail", "action", action, "pri", i.Pri, "Rep", i.Rep, "DocCount", i.DocsCount, "DocDelete", i.DocsDeleted, "StoreSize", i.StoreSize, "PriStoreSize", i.PriStoreSize, "CreationDate", CreationDate)
}

func logExecutionMode(index_onebyone []string, action string) {
	indicesInfo := CatIndices_withPattern(index_onebyone)
	i := indicesInfo[0]
	timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
	CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
	individual_Msg := fmt.Sprintf("%s has already processed in execution mode", index_onebyone)

	global.Detail_Logger.Infow(individual_Msg, "mode", "execution_mode", "logType", "Detail", "action", action, "pri", i.Pri, "Rep", i.Rep, "DocCount", i.DocsCount, "DocDelete", i.DocsDeleted, "StoreSize", i.StoreSize, "PriStoreSize", i.PriStoreSize, "CreationDate", CreationDate)
}
