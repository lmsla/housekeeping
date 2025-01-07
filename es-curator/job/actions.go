package job

import (
	"es-curator/global"
	// "es-curator/log_record"
	"fmt"
	"strconv"
	"time"
)

func Action_controll() {
	var comparelist, filter_record, role []string

	ActionList := global.ActionStruct.Actions
	for actions := range ActionList {
		switch ActionList[actions].Action {
		case "allocation":
			if !ActionList[actions].Options.DisableAction {

				global.Logger.Infow(fmt.Sprintf("Action %s Start,", ActionList[actions].Action), "logType", "Procedures")

				actionMsg := fmt.Sprintf("Action: %s,Description: %s", ActionList[actions].Action, ActionList[actions].Description)
				global.Logger.Infow(actionMsg, "logType", "Procedures")

				allocationtype := ActionList[actions].Options.AllocationType
				key := ActionList[actions].Options.Key
				value := ActionList[actions].Options.Value

				FilterList := ActionList[actions].Filters

				for filtertype := range FilterList {
					if FilterList[filtertype].Filtertype == "age" {
						filter_record = append(filter_record, "age")
					} else if FilterList[filtertype].Filtertype == "pattern" {
						filter_record = append(filter_record, "pattern")
					} else if FilterList[filtertype].Filtertype == "space" {
						filter_record = append(filter_record, "space")
					} else if FilterList[filtertype].Filtertype == "node_role" {
						filter_record = append(filter_record, "node_role")
						role = FilterList[filtertype].Value
					}
				}
				filter_info := fmt.Sprintf("use these filter :%s", filter_record)
				global.Logger.Infow(filter_info, "logType", "Procedures")

				if len(role) == 0 {
					comparelist = Filter_of_filter(filter_record, FilterList)
				} else {

					comparelist = Filters_With_node(role, filter_record, FilterList)
				}

				if comparelist != nil {
					detailMsg := fmt.Sprintf("allocation these indices :%s to %s ", comparelist, value)
					global.Logger.Infow(detailMsg, "type", "Details")

					var index_onebyone []string
					for index := range comparelist {
						index_onebyone := append(index_onebyone, comparelist[index])
						if global.EnvConfig.INFORMATION.Test_mode {
							indicesInfo := CatIndices_withPattern(index_onebyone)
							i := indicesInfo[0]
							timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
							CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
							individual_Msg := fmt.Sprintf("%s has already allocated ", index_onebyone)

							global.Detail_Logger.Infow(individual_Msg, "mode", "Test_mode", "logType", "Detail", "pri", i.Pri, "Rep", i.Rep, "DocCount", i.DocsCount, "DocDelete", i.DocsDeleted, "StoreSize", i.StoreSize, "PriStoreSize", i.PriStoreSize, "CreationDate", CreationDate)

						} else if !global.EnvConfig.INFORMATION.Test_mode {
							Allocation(index_onebyone, allocationtype, key, value)
							indicesInfo := CatIndices_withPattern(index_onebyone)
							i := indicesInfo[0]
							timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
							CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")

							individual_Msg := fmt.Sprintf("%s has already allocated ", index_onebyone)

							global.Detail_Logger.Infow(individual_Msg, "mode", "execution_mode", "logType", "Detail", "pri", i.Pri, "Rep", i.Rep, "DocCount", i.DocsCount, "DocDelete", i.DocsDeleted, "StoreSize", i.StoreSize, "PriStoreSize", i.PriStoreSize, "CreationDate", CreationDate)
						}

					}
					Node_relocating_checking()
				} else {
					global.Logger.Infow("no match indices", "logType", "Procedures")
				}
				delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)
				global.Logger.Infow(delaymsg)
				time.Sleep(time.Duration(ActionList[actions].Options.Delay) * time.Second)
				
				global.Logger.Infow(fmt.Sprintf("Action %s End,", ActionList[actions].Action), "logType", "Procedures")
			}
		case "forcemerge":
			if !ActionList[actions].Options.DisableAction {
				Node_relocating_checking()
				actionMsg := fmt.Sprintf("Action: %s,Description: %s", ActionList[actions].Action, ActionList[actions].Description)
				global.Logger.Infow(actionMsg, "logType", "Procedures")

				FilterList := ActionList[actions].Filters

				var MaxNumSegments int

				for filtertype := range FilterList {
					if FilterList[filtertype].Filtertype == "age" {
						filter_record = append(filter_record, "age")
					} else if FilterList[filtertype].Filtertype == "pattern" {
						filter_record = append(filter_record, "pattern")
					} else if FilterList[filtertype].Filtertype == "space" {
						filter_record = append(filter_record, "space")
					} else if FilterList[filtertype].Filtertype == "node_role" {
						filter_record = append(filter_record, "node_role")
						role = FilterList[filtertype].Value
					}
				}
				filter_info := fmt.Sprintf("use these filter :%s", filter_record)
				global.Logger.Infow(filter_info, "logType", "Procedures")

				if len(role) == 0 {
					comparelist = Filter_of_filter(filter_record, FilterList)
				} else {
					comparelist = Filters_With_node(role, filter_record, FilterList)
				}

				MaxNumSegments = ActionList[actions].Options.MaxNumSegment
				if comparelist != nil {
					detailMsg := fmt.Sprintf("forcemerge these indices :%s,segement num :%v", comparelist, MaxNumSegments)

					global.Logger.Infow(detailMsg, "logType", "Procedures")

					var index_onebyone []string
					for index := range comparelist {
						index_onebyone := append(index_onebyone, comparelist[index])

						if global.EnvConfig.INFORMATION.Test_mode {
							indicesInfo := CatIndices_withPattern(index_onebyone)
							i := indicesInfo[0]
							timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
							CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
							individual_Msg := fmt.Sprintf("%s has already forcemerged", index_onebyone)
							global.Detail_Logger.Infow(individual_Msg, "mode", "Test_mode", "logType", "Detail", "pri", i.Pri, "Rep", i.Rep, "DocCount", i.DocsCount, "DocDelete", i.DocsDeleted, "StoreSize", i.StoreSize, "PriStoreSize", i.PriStoreSize, "CreationDate", CreationDate)

						} else if !global.EnvConfig.INFORMATION.Test_mode {
							ForceMerge(index_onebyone, MaxNumSegments)
							indicesInfo := CatIndices_withPattern(index_onebyone)
							i := indicesInfo[0]
							timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
							CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
							individual_Msg := fmt.Sprintf("%s has already forcemerged", index_onebyone)

							global.Detail_Logger.Infow(individual_Msg, "mode", "execution_mode", "logType", "Detail", "pri", i.Pri, "Rep", i.Rep, "DocCount", i.DocsCount, "DocDelete", i.DocsDeleted, "StoreSize", i.StoreSize, "PriStoreSize", i.PriStoreSize, "CreationDate", CreationDate)
						}
					}

				} else {
					global.Logger.Infow("no match indices", "logType", "Procedures")
				}
				delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)
				global.Logger.Infow(delaymsg)
				time.Sleep(time.Duration(ActionList[actions].Options.Delay) * time.Second)
			}
		case "delete_indices":
			if !ActionList[actions].Options.DisableAction {

				actionMsg := fmt.Sprintf("Action: %s,Description: %s", ActionList[actions].Action, ActionList[actions].Description)

				global.Logger.Infow(actionMsg, "logType", "Procedures")
				FilterList := ActionList[actions].Filters

				var comparelist []string
				var filter_record []string

				for filtertype := range FilterList {
					if FilterList[filtertype].Filtertype == "age" {
						filter_record = append(filter_record, "age")
					} else if FilterList[filtertype].Filtertype == "pattern" {
						filter_record = append(filter_record, "pattern")
					} else if FilterList[filtertype].Filtertype == "space" {
						filter_record = append(filter_record, "space")
					} else if FilterList[filtertype].Filtertype == "water_level" {
						filter_record = append(filter_record, "water_level")
					} else if FilterList[filtertype].Filtertype == "node_role" {
						filter_record = append(filter_record, "node_role")
						role = FilterList[filtertype].Value
					}

				}
				filter_info := fmt.Sprintf("use these filter :%s", filter_record)

				global.Logger.Infow(filter_info, "logType", "Procedures")

				if len(role) == 0 {
					comparelist = Filter_of_filter(filter_record, FilterList)
				} else {
					comparelist = Filters_With_node(role, filter_record, FilterList)
				}

				//// delete function start write from here
				if comparelist != nil {
					detailMsg := fmt.Sprintf("Delete these indices : %s,Number of indices : %d", comparelist,len(comparelist))

					global.Logger.Infow(detailMsg, "logType", "Procedures")

					var index_onebyone []string
					for index := range comparelist {
						index_onebyone := append(index_onebyone, comparelist[index])
						if global.EnvConfig.INFORMATION.Test_mode {
							indicesInfo := CatIndices_withPattern(index_onebyone)
							i := indicesInfo[0]
							timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
							CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
							individual_Msg := fmt.Sprintf("%s has already deleted", index_onebyone)

							global.Detail_Logger.Infow(individual_Msg, "mode", "Test_mode", "logType", "Detail", "pri", i.Pri, "Rep", i.Rep, "DocCount", i.DocsCount, "DocDelete", i.DocsDeleted, "StoreSize", i.StoreSize, "PriStoreSize", i.PriStoreSize, "CreationDate", CreationDate)
						} else if !global.EnvConfig.INFORMATION.Test_mode {
							indicesInfo := CatIndices_withPattern(index_onebyone)
							i := indicesInfo[0]
							timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
							CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
							individual_Msg := fmt.Sprintf("%s has already deleted", index_onebyone)

							global.Detail_Logger.Infow(individual_Msg, "mode", "execution_mode", "logType", "Detail", "pri", i.Pri, "Rep", i.Rep, "DocCount", i.DocsCount, "DocDelete", i.DocsDeleted, "StoreSize", i.StoreSize, "PriStoreSize", i.PriStoreSize, "CreationDate", CreationDate)
							DeleteIndex(index_onebyone)
						}

					}
				} else {

					global.Logger.Infow("no match indices", "logType", "Procedures")
				}
				delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)

				global.Logger.Infow(delaymsg)
				time.Sleep(time.Duration(ActionList[actions].Options.Delay) * time.Second)
			}
		case "close":
			if !ActionList[actions].Options.DisableAction {

				actionMsg := fmt.Sprintf("Action: %s,Description: %s", ActionList[actions].Action, ActionList[actions].Description)

				global.Logger.Infow(actionMsg, "logType", "Procedures")
				FilterList := ActionList[actions].Filters
				var comparelist []string
				var filter_record []string
				for filtertype := range FilterList {
					if FilterList[filtertype].Filtertype == "age" {
						filter_record = append(filter_record, "age")
					} else if FilterList[filtertype].Filtertype == "pattern" {
						filter_record = append(filter_record, "pattern")
					} else if FilterList[filtertype].Filtertype == "space" {
						filter_record = append(filter_record, "space")
					} else if FilterList[filtertype].Filtertype == "node_role" {
						filter_record = append(filter_record, "node_role")
						role = FilterList[filtertype].Value
					}

				}

				filter_info := fmt.Sprintf("use these filter :%s", filter_record)

				global.Logger.Infow(filter_info, "logType", "Procedures")

				if len(role) == 0 {
					comparelist = Filter_of_filter(filter_record, FilterList)
				} else {
					comparelist = Filters_With_node(role, filter_record, FilterList)
				}

				if comparelist != nil {
					detailMsg := fmt.Sprintf("Close these indices :%s", comparelist)

					global.Logger.Infow(detailMsg, "logType", "Procedures")

					var index_onebyone []string
					for index := range comparelist {
						index_onebyone := append(index_onebyone, comparelist[index])

						if global.EnvConfig.INFORMATION.Test_mode {
							indicesInfo := CatIndices_withPattern(index_onebyone)
							i := indicesInfo[0]
							timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
							CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
							individual_Msg := fmt.Sprintf("%s has already closed", index_onebyone)

							global.Detail_Logger.Infow(individual_Msg, "mode", "Test_mode", "logType", "Detail", "pri", i.Pri, "Rep", i.Rep, "DocCount", i.DocsCount, "DocDelete", i.DocsDeleted, "StoreSize", i.StoreSize, "PriStoreSize", i.PriStoreSize, "CreationDate", CreationDate)

						} else if !global.EnvConfig.INFORMATION.Test_mode {
							CloseIndices(index_onebyone)
							indicesInfo := CatIndices_withPattern(index_onebyone)
							i := indicesInfo[0]
							timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
							CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
							individual_Msg := fmt.Sprintf("%s has already closed", index_onebyone)

							global.Detail_Logger.Infow(individual_Msg, "mode", "execution_mode", "logType", "Detail", "pri", i.Pri, "Rep", i.Rep, "DocCount", i.DocsCount, "DocDelete", i.DocsDeleted, "StoreSize", i.StoreSize, "PriStoreSize", i.PriStoreSize, "CreationDate", CreationDate)
						}
					}
				} else {

					global.Logger.Infow("no match indices", "logType", "Procedures")
				}

				delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)

				global.Logger.Infow(delaymsg)

				time.Sleep(time.Duration(ActionList[actions].Options.Delay) * time.Second)
			}
		case "open":
			if !ActionList[actions].Options.DisableAction {

				actionMsg := fmt.Sprintf("Action: %s,Description: %s", ActionList[actions].Action, ActionList[actions].Description)

				global.Logger.Infow(actionMsg, "logType", "Procedures")
				FilterList := ActionList[actions].Filters
				var comparelist []string
				var filter_record []string
				for filtertype := range FilterList {
					if FilterList[filtertype].Filtertype == "age" {
						filter_record = append(filter_record, "age")
					} else if FilterList[filtertype].Filtertype == "pattern" {
						filter_record = append(filter_record, "pattern")
					} else if FilterList[filtertype].Filtertype == "space" {
						filter_record = append(filter_record, "space")
					} else if FilterList[filtertype].Filtertype == "node_role" {
						filter_record = append(filter_record, "node_role")
						role = FilterList[filtertype].Value
					}
				}
				filter_info := fmt.Sprintf("use these filter :%s", filter_record)

				global.Logger.Infow(filter_info, "logType", "Procedures")
				if len(role) == 0 {
					comparelist = Filter_of_filter(filter_record, FilterList)
				} else {
					comparelist = Filters_With_node(role, filter_record, FilterList)
				}

				if comparelist != nil {
					detailMsg := fmt.Sprintf("Open these indices :%s", comparelist)

					global.Logger.Infow(detailMsg, "logType", "Procedures")

					var index_onebyone []string
					for index := range comparelist {
						index_onebyone := append(index_onebyone, comparelist[index])
						if global.EnvConfig.INFORMATION.Test_mode {
							indicesInfo := CatIndices_withPattern(index_onebyone)
							i := indicesInfo[0]
							timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
							CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
							individual_Msg := fmt.Sprintf("%s has already opened", index_onebyone)

							global.Detail_Logger.Infow(individual_Msg, "mode", "Test_mode", "logType", "Detail", "pri", i.Pri, "Rep", i.Rep, "DocCount", i.DocsCount, "DocDelete", i.DocsDeleted, "StoreSize", i.StoreSize, "PriStoreSize", i.PriStoreSize, "CreationDate", CreationDate)
						} else if !global.EnvConfig.INFORMATION.Test_mode {
							OpenIndices(index_onebyone)
							indicesInfo := CatIndices_withPattern(index_onebyone)
							i := indicesInfo[0]
							timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
							CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
							individual_Msg := fmt.Sprintf("%s has already opened", index_onebyone)

							global.Detail_Logger.Infow(individual_Msg, "mode", "execution_mode", "logType", "Detail", "pri", i.Pri, "Rep", i.Rep, "DocCount", i.DocsCount, "DocDelete", i.DocsDeleted, "StoreSize", i.StoreSize, "PriStoreSize", i.PriStoreSize, "CreationDate", CreationDate)
						}
					}
				} else {

					global.Logger.Infow("no match indices", "logType", "Procedures")
				}
				delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)

				global.Logger.Infow(delaymsg)
				time.Sleep(time.Duration(ActionList[actions].Options.Delay) * time.Second)
			}
		}
	}

}
