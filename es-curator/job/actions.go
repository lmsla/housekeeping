package job

import (
	"es-curator/global"
	"es-curator/log_record"
	"fmt"
	"time"
)

func Action_controll() {
	ActionList := global.ActionStruct.Actions
	for actions := range ActionList {
		switch ActionList[actions].Action {
		case "allocation":
			if ActionList[actions].Options.DisableAction == false {
				Action_allocation_indices()
				delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)
				log_record.Logrecord("Info", delaymsg)
				time.Sleep(time.Duration(ActionList[actions].Options.Delay) * time.Second)
			}
		case "forcemerge":
			if ActionList[actions].Options.DisableAction == false {
				Action_forcemerge_indices()
				delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)
				log_record.Logrecord("Info", delaymsg)
				time.Sleep(time.Duration(ActionList[actions].Options.Delay) * time.Second)
			}
		case "delete_indices":
			if ActionList[actions].Options.DisableAction == false {
				Action_delete_indices()
				delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)
				log_record.Logrecord("Info", delaymsg)
				time.Sleep(time.Duration(ActionList[actions].Options.Delay) * time.Second)
			}
		case "close":
			if ActionList[actions].Options.DisableAction == false {
				Action_close_indices()
				delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)
				log_record.Logrecord("Info", delaymsg)
				time.Sleep(time.Duration(ActionList[actions].Options.Delay) * time.Second)
			}
		case "open":
			if ActionList[actions].Options.DisableAction == false {
				Action_open_indices()
				delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)
				log_record.Logrecord("Info", delaymsg)
				time.Sleep(time.Duration(ActionList[actions].Options.Delay) * time.Second)
			}
		}
	}

}

////------- Open indices -------////
func Action_open_indices() {
	ActionList := global.ActionStruct.Actions

	for actions := range ActionList {
		// fmt.Println(global.ActionStruct.Actions[actions].Action)
		if ActionList[actions].Action == "open" {
			actionMsg := fmt.Sprintf("Action: %s,Description: %s", ActionList[actions].Action, ActionList[actions].Description)
			log_record.Logrecord("Actions", actionMsg)
			FilterList := ActionList[actions].Filters
			var agelist []string
			var patternlist []string
			var spacelist []string
			var comparelist []string
			var filter_record []string
			for filtertype := range FilterList {
				// fmt.Println("action:", ActionList[actions].Action, "description:", ActionList[actions].Description, "filtertype:", FilterList[filtertype].Filtertype)
				if FilterList[filtertype].Filtertype == "age" {
					filter_record = append(filter_record, "age")
					agelist = FilterType_age(FilterList[filtertype].Source, FilterList[filtertype].Direction, FilterList[filtertype].Unit, FilterList[filtertype].Unit_count)
				} else if FilterList[filtertype].Filtertype == "pattern" {
					filter_record = append(filter_record, "pattern")
					patternlist = FilterType_pattern(FilterList[filtertype].Kind, FilterList[filtertype].Value)
				} else if FilterList[filtertype].Filtertype == "space" {
					spacelist = FilterType_space(FilterList[filtertype].Disk_space)
					filter_record = append(filter_record, "space")
				}
				// comparelist = Indicesmapping3(spacelist, agelist, patternlist)

			}
			comparelist = Filter_of_filter(filter_record, agelist, patternlist, spacelist)

			filter_info := fmt.Sprintf("use these filter :%s", filter_record)
			log_record.Logrecord("Details", filter_info)

			if comparelist != nil {
				detailMsg := fmt.Sprintf("Open these indices :%s", comparelist)
				log_record.Logrecord("Details", detailMsg)

				var index_onebyone []string
				for index := range comparelist {
					index_onebyone := append(index_onebyone, comparelist[index])

					if global.EnvConfig.INFORMATION.Test_mode == true {
						individual_Msg := fmt.Sprintf("%s has already opened", index_onebyone)
						log_record.Logrecord("Test Mode Details", individual_Msg)
					} else if global.EnvConfig.INFORMATION.Test_mode == false {
						OpenIndices(index_onebyone)
						individual_Msg := fmt.Sprintf("%s has already opened", index_onebyone)
						log_record.Logrecord("Details", individual_Msg)
					}
				}
				//// open function write from here
				// OpenIndices(comparelist)
			} else if comparelist == nil {
				log_record.Logrecord("Details", "no match indices")
			}

		}
	}
}

////------- Close indices -------////
func Action_close_indices() {
	ActionList := global.ActionStruct.Actions

	for actions := range ActionList {
		// fmt.Println(global.ActionStruct.Actions[actions].Action)
		if ActionList[actions].Action == "close" {
			actionMsg := fmt.Sprintf("Action: %s,Description: %s", ActionList[actions].Action, ActionList[actions].Description)
			log_record.Logrecord("Actions", actionMsg)
			FilterList := ActionList[actions].Filters
			var agelist []string
			var patternlist []string
			var spacelist []string
			var comparelist []string
			var filter_record []string
			for filtertype := range FilterList {
				// fmt.Println("action:", ActionList[actions].Action, "description:", ActionList[actions].Description, "filtertype:", FilterList[filtertype].Filtertype)
				if FilterList[filtertype].Filtertype == "age" {
					filter_record = append(filter_record, "age")
					agelist = FilterType_age(FilterList[filtertype].Source, FilterList[filtertype].Direction, FilterList[filtertype].Unit, FilterList[filtertype].Unit_count)
				} else if FilterList[filtertype].Filtertype == "pattern" {
					filter_record = append(filter_record, "pattern")
					patternlist = FilterType_pattern(FilterList[filtertype].Kind, FilterList[filtertype].Value)
				} else if FilterList[filtertype].Filtertype == "space" {
					filter_record = append(filter_record, "space")
					spacelist = FilterType_space(FilterList[filtertype].Disk_space)
				}
				// comparelist = Indicesmapping3(spacelist, agelist, patternlist)

			}
			comparelist = Filter_of_filter(filter_record, agelist, patternlist, spacelist)

			filter_info := fmt.Sprintf("use these filter :%s", filter_record)
			log_record.Logrecord("Details", filter_info)

			if comparelist != nil {
				detailMsg := fmt.Sprintf("Close these indices :%s", comparelist)
				log_record.Logrecord("Details", detailMsg)

				var index_onebyone []string
				for index := range comparelist {
					index_onebyone := append(index_onebyone, comparelist[index])

					if global.EnvConfig.INFORMATION.Test_mode == true {
						individual_Msg := fmt.Sprintf("%s has already closed", index_onebyone)
						log_record.Logrecord("Test Mode Details", individual_Msg)

					} else if global.EnvConfig.INFORMATION.Test_mode == false {
						CloseIndices(index_onebyone)
						individual_Msg := fmt.Sprintf("%s has already closed", index_onebyone)
						log_record.Logrecord("Details", individual_Msg)
					}
				}
				//// Close function write from here
				// CloseIndices(comparelist)
			} else if comparelist == nil {
				log_record.Logrecord("Details", "no match indices")
			}

		}
	}

}

////------- Delete indices -------////
func Action_delete_indices() {
	ActionList := global.ActionStruct.Actions

	for actions := range ActionList {
		// fmt.Println(global.ActionStruct.Actions[actions].Action)
		if ActionList[actions].Action == "delete_indices" {
			actionMsg := fmt.Sprintf("Action: %s,Description: %s", ActionList[actions].Action, ActionList[actions].Description)
			log_record.Logrecord("Actions", actionMsg)
			FilterList := ActionList[actions].Filters
			var agelist []string
			var patternlist []string
			var spacelist []string
			var comparelist []string
			var filter_record []string
			for filtertype := range FilterList {
				// fmt.Println("action:", ActionList[actions].Action, "description:", ActionList[actions].Description, "filtertype:", FilterList[filtertype].Filtertype)
				if FilterList[filtertype].Filtertype == "age" {
					filter_record = append(filter_record, "age")
					agelist = FilterType_age(FilterList[filtertype].Source, FilterList[filtertype].Direction, FilterList[filtertype].Unit, FilterList[filtertype].Unit_count)
				} else if FilterList[filtertype].Filtertype == "pattern" {
					filter_record = append(filter_record, "pattern")
					patternlist = FilterType_pattern(FilterList[filtertype].Kind, FilterList[filtertype].Value)
				} else if FilterList[filtertype].Filtertype == "space" {
					filter_record = append(filter_record, "space")
					spacelist = FilterType_space(FilterList[filtertype].Disk_space)
				}
				// comparelist = Indicesmapping3(agelist, patternlist,spacelist )

			}
			comparelist = Filter_of_filter(filter_record, agelist, patternlist, spacelist)

			// fmt.Println("filter_record:", filter_record)
			// fmt.Println("agelist:", agelist)
			// fmt.Println("patternlist:", patternlist)
			// fmt.Println("spacelist:", spacelist)
			// fmt.Println("compare:", comparelist)
			filter_info := fmt.Sprintf("use these filter :%s", filter_record)
			log_record.Logrecord("Details", filter_info)
			//// delete function write from here
			if comparelist != nil {
				detailMsg := fmt.Sprintf("Delete these indices :%s", comparelist)
				log_record.Logrecord("Details", detailMsg)

				var index_onebyone []string
				for index := range comparelist {
					index_onebyone := append(index_onebyone, comparelist[index])
					if global.EnvConfig.INFORMATION.Test_mode == true {
						individual_Msg := fmt.Sprintf("%s has already deleted", index_onebyone)
						log_record.Logrecord("Test Mode Details", individual_Msg)

					} else if global.EnvConfig.INFORMATION.Test_mode == false {
						DeleteIndex(index_onebyone)
						individual_Msg := fmt.Sprintf("%s has already deleted", index_onebyone)
						log_record.Logrecord("Details", individual_Msg)
					}

				}
				// DeleteIndex(comparelist)
			} else if comparelist == nil {
				log_record.Logrecord("Details", "no match indices")
			}

		}
	}
}

////------- Forcemerge indices -------////
func Action_forcemerge_indices() {
	Node_relocating_checking()

	ActionList := global.ActionStruct.Actions

	for actions := range ActionList {
		// fmt.Println(global.ActionStruct.Actions[actions].Action)
		if ActionList[actions].Action == "forcemerge" {
			actionMsg := fmt.Sprintf("Action: %s,Description: %s", ActionList[actions].Action, ActionList[actions].Description)
			log_record.Logrecord("Actions", actionMsg)
			FilterList := ActionList[actions].Filters
			var agelist []string
			var patternlist []string
			var spacelist []string
			var comparelist []string
			var MaxNumSegments int
			var filter_record []string
			for filtertype := range FilterList {
				// fmt.Println("action:", ActionList[actions].Action, "description:", ActionList[actions].Description, "filtertype:", FilterList[filtertype].Filtertype)
				if FilterList[filtertype].Filtertype == "age" {
					filter_record = append(filter_record, "age")
					agelist = FilterType_age(FilterList[filtertype].Source, FilterList[filtertype].Direction, FilterList[filtertype].Unit, FilterList[filtertype].Unit_count)
				} else if FilterList[filtertype].Filtertype == "pattern" {
					filter_record = append(filter_record, "pattern")
					patternlist = FilterType_pattern(FilterList[filtertype].Kind, FilterList[filtertype].Value)
				} else if FilterList[filtertype].Filtertype == "space" {
					filter_record = append(filter_record, "space")
					spacelist = FilterType_space(FilterList[filtertype].Disk_space)
				}
				// comparelist = Indicesmapping(spacelist, agelist, patternlist)

			}
			comparelist = Filter_of_filter(filter_record, agelist, patternlist, spacelist)
			MaxNumSegments = ActionList[actions].Options.MaxNumSegment
			// fmt.Println(MaxNumSegments)
			// fmt.Println(ActionList[actions].Options.Key)
			// fmt.Println(ActionList[actions].Options.Delay)
			// fmt.Println(ActionList[actions].Options.MaxNumSegment)
			// fmt.Println(ActionList[actions].Options.TimeoutOverride)
			// fmt.Println(ActionList[actions].Description)
			// msg := fmt.Sprintf("segement num :%v", MaxNumSegments)
			// log_record.Logrecord("segement", msg)

			if comparelist != nil {
				detailMsg := fmt.Sprintf("forcemerge these indices :%s,segement num :%v", comparelist, MaxNumSegments)
				log_record.Logrecord("Details", detailMsg)

				var index_onebyone []string
				for index := range comparelist {
					index_onebyone := append(index_onebyone, comparelist[index])

					if global.EnvConfig.INFORMATION.Test_mode == true {
						individual_Msg := fmt.Sprintf("%s has already forcemerged", index_onebyone)
						log_record.Logrecord("Test Mode Details", individual_Msg)

					} else if global.EnvConfig.INFORMATION.Test_mode == false {
						ForceMerge(index_onebyone, MaxNumSegments)
						individual_Msg := fmt.Sprintf("%s has already forcemerged", index_onebyone)
						log_record.Logrecord("Details", individual_Msg)
					}
				}
				// ForceMerge(comparelist, MaxNumSegments)
			} else if comparelist == nil {
				log_record.Logrecord("Details", "no match indices")
			}

		}
	}

}

////------- Allocation indices -------////
func Action_allocation_indices() {
	ActionList := global.ActionStruct.Actions

	for actions := range ActionList {
		// fmt.Println(global.ActionStruct.Actions[actions].Action)
		if ActionList[actions].Action == "allocation" {
			actionMsg := fmt.Sprintf("Action: %s,Description: %s", ActionList[actions].Action, ActionList[actions].Description)
			allocationtype := ActionList[actions].Options.AllocationType
			key := ActionList[actions].Options.Key
			value := ActionList[actions].Options.Value
			log_record.Logrecord("Actions", actionMsg)
			FilterList := ActionList[actions].Filters
			var agelist []string
			var patternlist []string
			var spacelist []string
			var comparelist []string
			var filter_record []string
			for filtertype := range FilterList {
				// fmt.Println("action:", ActionList[actions].Action, "description:", ActionList[actions].Description, "filtertype:", FilterList[filtertype].Filtertype)
				if FilterList[filtertype].Filtertype == "age" {
					filter_record = append(filter_record, "age")
					agelist = FilterType_age(FilterList[filtertype].Source, FilterList[filtertype].Direction, FilterList[filtertype].Unit, FilterList[filtertype].Unit_count)
				} else if FilterList[filtertype].Filtertype == "pattern" {
					filter_record = append(filter_record, "pattern")
					patternlist = FilterType_pattern(FilterList[filtertype].Kind, FilterList[filtertype].Value)
				} else if FilterList[filtertype].Filtertype == "space" {
					filter_record = append(filter_record, "space")
					spacelist = FilterType_space(FilterList[filtertype].Disk_space)
				}
				// comparelist = Indicesmapping3(spacelist, agelist, patternlist)

			}
			comparelist = Filter_of_filter(filter_record, agelist, patternlist, spacelist)
			fmt.Println(allocationtype, key)

			if comparelist != nil {
				detailMsg := fmt.Sprintf("allocation these indices :%s to %s ", comparelist, value)
				log_record.Logrecord("Details", detailMsg)

				var index_onebyone []string
				for index := range comparelist {
					index_onebyone := append(index_onebyone, comparelist[index])
					if global.EnvConfig.INFORMATION.Test_mode == true {
						individual_Msg := fmt.Sprintf("%s has already allocated", index_onebyone)
						log_record.Logrecord("Test Mode Details", individual_Msg)

					} else if global.EnvConfig.INFORMATION.Test_mode == false {
						Allocation(index_onebyone, allocationtype, key, value)
						individual_Msg := fmt.Sprintf("%s has already allocated", index_onebyone)
						log_record.Logrecord("Details", individual_Msg)
					}

				}
				//// allocation function write from here
				// Allocation(comparelist, allocationtype, key, value)
				Node_relocating_checking()
			} else if comparelist == nil {
				log_record.Logrecord("Details", "no match indices")
			}

		}
	}

}
