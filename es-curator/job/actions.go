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
		}
	}
}

func Action_controll() {
	var comparelist, filter_record, role []string

	ActionList := global.ActionStruct.Actions
	for actions := range ActionList {
		// 控制 action 的 disabled
		if ActionList[actions].Options.DisableAction {
			continue
		}

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
	if comparelist != nil {
		detailMsg := fmt.Sprintf("allocation these indices :%s to %s ", comparelist, action.Options.Value)
		global.Logger.Infow(detailMsg, "type", "Details")

		for _, index := range comparelist {
			index_onebyone := []string{index}
			if global.EnvConfig.INFORMATION.Test_mode {
				logTestMode(index_onebyone)
			} else {
				Allocation(index_onebyone, action.Options.AllocationType, action.Options.Key, action.Options.Value)
				logExecutionMode(index_onebyone)
			}
		}
		Node_relocating_checking()
	} else {
		global.Logger.Infow("no match indices", "logType", "Procedures")
	}
}

func handleForceMerge(comparelist []string, action structs.Actiond) {
	if comparelist != nil {
		detailMsg := fmt.Sprintf("forcemerge these indices :%s, segment num :%v", comparelist, action.Options.MaxNumSegment)
		global.Logger.Infow(detailMsg, "logType", "Procedures")

		for _, index := range comparelist {
			index_onebyone := []string{index}
			if global.EnvConfig.INFORMATION.Test_mode {
				logTestMode(index_onebyone)
			} else {
				ForceMerge(index_onebyone, action.Options.MaxNumSegment)
				logExecutionMode(index_onebyone)
			}
		}
	} else {
		global.Logger.Infow("no match indices", "logType", "Procedures")
	}
}

func handleDeleteIndices(comparelist []string) {
	if comparelist != nil {
		detailMsg := fmt.Sprintf("Delete these indices : %s, Number of indices : %d", comparelist, len(comparelist))
		global.Logger.Infow(detailMsg, "logType", "Procedures")

		for _, index := range comparelist {
			index_onebyone := []string{index}
			if global.EnvConfig.INFORMATION.Test_mode {
				logTestMode(index_onebyone)
			} else {
				DeleteIndex(index_onebyone)
				logExecutionMode(index_onebyone)
			}
		}
	} else {
		global.Logger.Infow("no match indices", "logType", "Procedures")
	}
}

func handleClose(comparelist []string) {
	if comparelist != nil {
		detailMsg := fmt.Sprintf("Close these indices :%s", comparelist)
		global.Logger.Infow(detailMsg, "logType", "Procedures")

		for _, index := range comparelist {
			index_onebyone := []string{index}
			if global.EnvConfig.INFORMATION.Test_mode {
				logTestMode(index_onebyone)
			} else {
				CloseIndices(index_onebyone)
				logExecutionMode(index_onebyone)
			}
		}
	} else {
		global.Logger.Infow("no match indices", "logType", "Procedures")
	}
}

func handleOpen(comparelist []string) {
	if comparelist != nil {
		detailMsg := fmt.Sprintf("Open these indices :%s", comparelist)
		global.Logger.Infow(detailMsg, "logType", "Procedures")

		for _, index := range comparelist {
			index_onebyone := []string{index}
			if global.EnvConfig.INFORMATION.Test_mode {
				logTestMode(index_onebyone)
			} else {
				OpenIndices(index_onebyone)
				logExecutionMode(index_onebyone)
			}
		}
	} else {
		global.Logger.Infow("no match indices", "logType", "Procedures")
	}
}

func logTestMode(index_onebyone []string) {
	indicesInfo := CatIndices_withPattern(index_onebyone)
	i := indicesInfo[0]
	timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
	CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
	individual_Msg := fmt.Sprintf("%s has already processed in test mode", index_onebyone)

	global.Detail_Logger.Infow(individual_Msg, "mode", "Test_mode", "logType", "Detail", "pri", i.Pri, "Rep", i.Rep, "DocCount", i.DocsCount, "DocDelete", i.DocsDeleted, "StoreSize", i.StoreSize, "PriStoreSize", i.PriStoreSize, "CreationDate", CreationDate)
}

func logExecutionMode(index_onebyone []string) {
	indicesInfo := CatIndices_withPattern(index_onebyone)
	i := indicesInfo[0]
	timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
	CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
	individual_Msg := fmt.Sprintf("%s has already processed in execution mode", index_onebyone)

	global.Detail_Logger.Infow(individual_Msg, "mode", "execution_mode", "logType", "Detail", "pri", i.Pri, "Rep", i.Rep, "DocCount", i.DocsCount, "DocDelete", i.DocsDeleted, "StoreSize", i.StoreSize, "PriStoreSize", i.PriStoreSize, "CreationDate", CreationDate)
}
