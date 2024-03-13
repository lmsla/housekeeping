package job

import (
	"es-curator/global"
	"es-curator/log_record"
	"fmt"
	"strconv"
	"time"
)

func Action_controll() {
	ActionList := global.ActionStruct.Actions
	for actions := range ActionList {
		switch ActionList[actions].Action {
		case "allocation":
			if !ActionList[actions].Options.DisableAction {
				// Action_allocation_indices()
				actionMsg := fmt.Sprintf("Action: %s,Description: %s", ActionList[actions].Action, ActionList[actions].Description)
				allocationtype := ActionList[actions].Options.AllocationType
				key := ActionList[actions].Options.Key
				value := ActionList[actions].Options.Value
				log_record.Logrecord("Actions", actionMsg)
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
					}
				}
				filter_info := fmt.Sprintf("use these filter :%s", filter_record)
				log_record.Logrecord("Details", filter_info)
				comparelist = Filter_of_filter(filter_record, FilterList)

				fmt.Println(allocationtype, key)
				if comparelist != nil {
					detailMsg := fmt.Sprintf("allocation these indices :%s to %s ", comparelist, value)
					log_record.Logrecord("Details", detailMsg)

					var index_onebyone []string
					for index := range comparelist {
						index_onebyone := append(index_onebyone, comparelist[index])
						if global.EnvConfig.INFORMATION.Test_mode {
							indicesInfo := CatIndices_withPattern(index_onebyone)
							i := indicesInfo[0]
							timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
							CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
							individual_Msg := fmt.Sprintf("%s has already allocated ,Pri: %s,Rep: %s,DocCount: %s, DocDelete: %s,StoreSize: %s,PriStoreSize: %s,CreationDate: %s", index_onebyone, i.Pri, i.Rep, i.DocsCount, i.DocsDeleted, i.StoreSize, i.PriStoreSize, CreationDate)

							log_record.ActionDetailrecord("Test Mode Details", individual_Msg)

						} else if !global.EnvConfig.INFORMATION.Test_mode {
							Allocation(index_onebyone, allocationtype, key, value)
							indicesInfo := CatIndices_withPattern(index_onebyone)
							i := indicesInfo[0]
							timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
							CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
							individual_Msg := fmt.Sprintf("%s has already allocated ,Pri: %s,Rep: %s,DocCount: %s, DocDelete: %s,StoreSize: %s,PriStoreSize: %s,CreationDate: %s", index_onebyone, i.Pri, i.Rep, i.DocsCount, i.DocsDeleted, i.StoreSize, i.PriStoreSize, CreationDate)
							log_record.ActionDetailrecord("Details", individual_Msg)
						}

					}
					Node_relocating_checking()
				} else {
					log_record.Logrecord("Details", "no match indices")
				}
				delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)
				log_record.Logrecord("Info", delaymsg)
				time.Sleep(time.Duration(ActionList[actions].Options.Delay) * time.Second)
			}
		case "forcemerge":
			if !ActionList[actions].Options.DisableAction {
				// Action_forcemerge_indices()
				Node_relocating_checking()
				actionMsg := fmt.Sprintf("Action: %s,Description: %s", ActionList[actions].Action, ActionList[actions].Description)
				log_record.Logrecord("Actions", actionMsg)
				FilterList := ActionList[actions].Filters
				var comparelist []string
				var MaxNumSegments int
				var filter_record []string
				for filtertype := range FilterList {
					if FilterList[filtertype].Filtertype == "age" {
						filter_record = append(filter_record, "age")
					} else if FilterList[filtertype].Filtertype == "pattern" {
						filter_record = append(filter_record, "pattern")
					} else if FilterList[filtertype].Filtertype == "space" {
						filter_record = append(filter_record, "space")
					}
				}
				filter_info := fmt.Sprintf("use these filter :%s", filter_record)
				log_record.Logrecord("Details", filter_info)
				comparelist = Filter_of_filter(filter_record, FilterList)

				MaxNumSegments = ActionList[actions].Options.MaxNumSegment
				if comparelist != nil {
					detailMsg := fmt.Sprintf("forcemerge these indices :%s,segement num :%v", comparelist, MaxNumSegments)
					log_record.Logrecord("Details", detailMsg)

					var index_onebyone []string
					for index := range comparelist {
						index_onebyone := append(index_onebyone, comparelist[index])

						if global.EnvConfig.INFORMATION.Test_mode {
							indicesInfo := CatIndices_withPattern(index_onebyone)
							i := indicesInfo[0]
							timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
							CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
							individual_Msg := fmt.Sprintf("%s has already forcemerged ,Pri: %s,Rep: %s,DocCount: %s, DocDelete: %s,StoreSize: %s,PriStoreSize: %s,CreationDate: %s", index_onebyone, i.Pri, i.Rep, i.DocsCount, i.DocsDeleted, i.StoreSize, i.PriStoreSize, CreationDate)
							log_record.ActionDetailrecord("Test Mode Details", individual_Msg)

						} else if !global.EnvConfig.INFORMATION.Test_mode {
							ForceMerge(index_onebyone, MaxNumSegments)
							indicesInfo := CatIndices_withPattern(index_onebyone)
							i := indicesInfo[0]
							timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
							CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
							individual_Msg := fmt.Sprintf("%s has already forcemerged ,Pri: %s,Rep: %s,DocCount: %s, DocDelete: %s,StoreSize: %s,PriStoreSize: %s,CreationDate: %s", index_onebyone, i.Pri, i.Rep, i.DocsCount, i.DocsDeleted, i.StoreSize, i.PriStoreSize, CreationDate)
							log_record.ActionDetailrecord("Details", individual_Msg)
						}
					}
					// ForceMerge(comparelist, MaxNumSegments)
				} else {
					log_record.Logrecord("Details", "no match indices")
				}
				delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)
				log_record.Logrecord("Info", delaymsg)
				time.Sleep(time.Duration(ActionList[actions].Options.Delay) * time.Second)
			}
		case "delete_indices":
			if !ActionList[actions].Options.DisableAction {
				// Action_delete_indices()
				actionMsg := fmt.Sprintf("Action: %s,Description: %s", ActionList[actions].Action, ActionList[actions].Description)
				log_record.Logrecord("Actions", actionMsg)
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
					}

				}
				filter_info := fmt.Sprintf("use these filter :%s", filter_record)
				log_record.Logrecord("Details", filter_info)
				comparelist = Filter_of_filter(filter_record, FilterList)

				//// delete function write from here
				if comparelist != nil {
					detailMsg := fmt.Sprintf("Delete these indices :%s", comparelist)
					log_record.Logrecord("Details", detailMsg)

					var index_onebyone []string
					for index := range comparelist {
						index_onebyone := append(index_onebyone, comparelist[index])
						if global.EnvConfig.INFORMATION.Test_mode {
							indicesInfo := CatIndices_withPattern(index_onebyone)
							i := indicesInfo[0]
							timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
							CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
							individual_Msg := fmt.Sprintf("%s has already deleted ,Pri: %s,Rep: %s,DocCount: %s, DocDelete: %s,StoreSize: %s,PriStoreSize: %s,CreationDate: %s", index_onebyone, i.Pri, i.Rep, i.DocsCount, i.DocsDeleted, i.StoreSize, i.PriStoreSize, CreationDate)
							log_record.ActionDetailrecord("Test Mode Details", individual_Msg)
						} else if !global.EnvConfig.INFORMATION.Test_mode {
							indicesInfo := CatIndices_withPattern(index_onebyone)
							i := indicesInfo[0]
							timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
							CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
							individual_Msg := fmt.Sprintf("%s has already deleted ,Pri: %s,Rep: %s,DocCount: %s, DocDelete: %s,StoreSize: %s,PriStoreSize: %s,CreationDate: %s", index_onebyone, i.Pri, i.Rep, i.DocsCount, i.DocsDeleted, i.StoreSize, i.PriStoreSize, CreationDate)
							log_record.ActionDetailrecord("Details", individual_Msg)
							DeleteIndex(index_onebyone)
						}

					}
				} else {
					log_record.Logrecord("Details", "no match indices")
				}
				delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)
				log_record.Logrecord("Info", delaymsg)
				time.Sleep(time.Duration(ActionList[actions].Options.Delay) * time.Second)
			}
		case "close":
			if !ActionList[actions].Options.DisableAction {
				// Action_close_indices()
				actionMsg := fmt.Sprintf("Action: %s,Description: %s", ActionList[actions].Action, ActionList[actions].Description)
				log_record.Logrecord("Actions", actionMsg)
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
					}

				}

				filter_info := fmt.Sprintf("use these filter :%s", filter_record)
				log_record.Logrecord("Details", filter_info)
				comparelist = Filter_of_filter(filter_record, FilterList)

				if comparelist != nil {
					detailMsg := fmt.Sprintf("Close these indices :%s", comparelist)
					log_record.Logrecord("Details", detailMsg)

					var index_onebyone []string
					for index := range comparelist {
						index_onebyone := append(index_onebyone, comparelist[index])

						if global.EnvConfig.INFORMATION.Test_mode {
							indicesInfo := CatIndices_withPattern(index_onebyone)
							i := indicesInfo[0]
							timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
							CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
							individual_Msg := fmt.Sprintf("%s has already closed ,Pri: %s,Rep: %s,DocCount: %s, DocDelete: %s,StoreSize: %s,PriStoreSize: %s,CreationDate: %s", index_onebyone, i.Pri, i.Rep, i.DocsCount, i.DocsDeleted, i.StoreSize, i.PriStoreSize, CreationDate)
							log_record.ActionDetailrecord("Test Mode Details", individual_Msg)

						} else if !global.EnvConfig.INFORMATION.Test_mode {
							CloseIndices(index_onebyone)
							indicesInfo := CatIndices_withPattern(index_onebyone)
							i := indicesInfo[0]
							timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
							CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
							individual_Msg := fmt.Sprintf("%s has already closed ,Pri: %s,Rep: %s,DocCount: %s, DocDelete: %s,StoreSize: %s,PriStoreSize: %s,CreationDate: %s", index_onebyone, i.Pri, i.Rep, i.DocsCount, i.DocsDeleted, i.StoreSize, i.PriStoreSize, CreationDate)
							log_record.ActionDetailrecord("Details", individual_Msg)
						}
					}
				} else {
					log_record.Logrecord("Details", "no match indices")
				}

				delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)
				log_record.Logrecord("Info", delaymsg)
				time.Sleep(time.Duration(ActionList[actions].Options.Delay) * time.Second)
			}
		case "open":
			if !ActionList[actions].Options.DisableAction {
				// Action_open_indices()
				actionMsg := fmt.Sprintf("Action: %s,Description: %s", ActionList[actions].Action, ActionList[actions].Description)
				log_record.Logrecord("Actions", actionMsg)
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
					}
				}
				filter_info := fmt.Sprintf("use these filter :%s", filter_record)
				log_record.Logrecord("Details", filter_info)
				comparelist = Filter_of_filter(filter_record, FilterList)

				if comparelist != nil {
					detailMsg := fmt.Sprintf("Open these indices :%s", comparelist)
					log_record.Logrecord("Details", detailMsg)

					var index_onebyone []string
					for index := range comparelist {
						index_onebyone := append(index_onebyone, comparelist[index])
						if global.EnvConfig.INFORMATION.Test_mode {
							indicesInfo := CatIndices_withPattern(index_onebyone)
							i := indicesInfo[0]
							timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
							CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
							individual_Msg := fmt.Sprintf("%s has already opened ,Pri: %s,Rep: %s,DocCount: %s, DocDelete: %s,StoreSize: %s,PriStoreSize: %s,CreationDate: %s", index_onebyone, i.Pri, i.Rep, i.DocsCount, i.DocsDeleted, i.StoreSize, i.PriStoreSize, CreationDate)
							log_record.ActionDetailrecord("Test Mode Details", individual_Msg)
						} else if !global.EnvConfig.INFORMATION.Test_mode {
							OpenIndices(index_onebyone)
							indicesInfo := CatIndices_withPattern(index_onebyone)
							i := indicesInfo[0]
							timestamp, _ := strconv.ParseInt(i.CreationDate, 10, 64)
							CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
							individual_Msg := fmt.Sprintf("%s has already opened ,Pri: %s,Rep: %s,DocCount: %s, DocDelete: %s,StoreSize: %s,PriStoreSize: %s,CreationDate: %s", index_onebyone, i.Pri, i.Rep, i.DocsCount, i.DocsDeleted, i.StoreSize, i.PriStoreSize, CreationDate)
							log_record.ActionDetailrecord("Details", individual_Msg)
						}
					}
				} else {
					log_record.Logrecord("Details", "no match indices")
				}
				delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)
				log_record.Logrecord("Info", delaymsg)
				time.Sleep(time.Duration(ActionList[actions].Options.Delay) * time.Second)
			}
		}
	}

}
