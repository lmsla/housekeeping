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
			if ActionList[actions].Options.DisableAction == true {
				// Action_allocation_indices()
				delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)
				log_record.Logrecord("Info", delaymsg)
				time.Sleep( time.Duration(ActionList[actions].Options.Delay) * time.Second)
			}
		case "forcemerge":
			if ActionList[actions].Options.DisableAction == true {
				// Action_forcemerge_indices()
				delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)
				log_record.Logrecord("Info", delaymsg)
				time.Sleep( time.Duration(ActionList[actions].Options.Delay) * time.Second)
			}
		case "delete_indices":
			if ActionList[actions].Options.DisableAction == true {
				// Action_delete_indices()
				delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)
				log_record.Logrecord("Info", delaymsg)
				time.Sleep( time.Duration(ActionList[actions].Options.Delay) * time.Second)
			}
		case "close":
			if ActionList[actions].Options.DisableAction == true {
				// Action_close_indices()
				delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)
				log_record.Logrecord("Info", delaymsg)
				time.Sleep( time.Duration(ActionList[actions].Options.Delay) * time.Second)
			}
		case "open":
			if ActionList[actions].Options.DisableAction == true {
				// Action_open_indices()
				delaymsg := fmt.Sprintf("Pausing for %v seconds before continuing...", ActionList[actions].Options.Delay)
				log_record.Logrecord("Info", delaymsg)
				time.Sleep( time.Duration(ActionList[actions].Options.Delay) * time.Second)
			}
		}
	}
	// for actions := range ActionList {
	// 	if ActionList[actions].Action == "allocation" {
	// 		// Action_allocation_indices()
	// 		time.Sleep(5 * time.Second)
	// 		fmt.Println("allocation wait fot 5 s ")
	// 	} else if ActionList[actions].Action == "forcemerge" {
	// 		// Action_forcemerge_indices()
	// 		time.Sleep(5 * time.Second)
	// 		fmt.Println("forcemerge wait fot 5 s ")
	// 	} else if ActionList[actions].Action == "delete_indices" {
	// 		fmt.Println("delete_indices wait fot 5 s ")
	// 		time.Sleep(5 * time.Second)
	// 		// Action_delete_indices()
	// 	} else if ActionList[actions].Action == "close" {
	// 		time.Sleep(5 * time.Second)
	// 		fmt.Println("close indices wait fot 5 s ")
	// 	}
	// }
	// Action_open_indices()
	// Action_close_indices()
	// Action_delete_indices()
	// Action_forcemerge_indices()
	// Action_allocation_indices()
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
			for filtertype := range FilterList {
				// fmt.Println("action:", ActionList[actions].Action, "description:", ActionList[actions].Description, "filtertype:", FilterList[filtertype].Filtertype)
				if FilterList[filtertype].Filtertype == "age" {
					agelist = FilterType_age(FilterList[filtertype].Source, FilterList[filtertype].Direction, FilterList[filtertype].Unit, FilterList[filtertype].Unit_count)
				} else if FilterList[filtertype].Filtertype == "pattern" {
					patternlist = FilterType_pattern(FilterList[filtertype].Kind, FilterList[filtertype].Value)
				} else if FilterList[filtertype].Filtertype == "space" {
					spacelist = FilterType_space(FilterList[filtertype].Disk_space)
				}
				comparelist = Indicesmapping(spacelist, agelist, patternlist)

			}
			// fmt.Println("agelist:", agelist)
			// fmt.Println("patternlist:", patternlist)
			// fmt.Println("spacelist:", spacelist)
			// fmt.Println("compare:", comparelist)
			detailMsg := fmt.Sprintf("Open these indices :%s", comparelist)
			log_record.Logrecord("Details", detailMsg)
			//// Close function write from here
			OpenIndices(comparelist)
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
			for filtertype := range FilterList {
				// fmt.Println("action:", ActionList[actions].Action, "description:", ActionList[actions].Description, "filtertype:", FilterList[filtertype].Filtertype)
				if FilterList[filtertype].Filtertype == "age" {
					agelist = FilterType_age(FilterList[filtertype].Source, FilterList[filtertype].Direction, FilterList[filtertype].Unit, FilterList[filtertype].Unit_count)
				} else if FilterList[filtertype].Filtertype == "pattern" {
					patternlist = FilterType_pattern(FilterList[filtertype].Kind, FilterList[filtertype].Value)
				} else if FilterList[filtertype].Filtertype == "space" {
					spacelist = FilterType_space(FilterList[filtertype].Disk_space)
				}
				comparelist = Indicesmapping(spacelist, agelist, patternlist)

			}
			// fmt.Println("agelist:", agelist)
			// fmt.Println("patternlist:", patternlist)
			// fmt.Println("spacelist:", spacelist)
			// fmt.Println("compare:", comparelist)
			detailMsg := fmt.Sprintf("Close these indices :%s", comparelist)
			log_record.Logrecord("Details", detailMsg)
			//// Close function write from here
			CloseIndices(comparelist)
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
			for filtertype := range FilterList {
				// fmt.Println("action:", ActionList[actions].Action, "description:", ActionList[actions].Description, "filtertype:", FilterList[filtertype].Filtertype)
				if FilterList[filtertype].Filtertype == "age" {
					agelist = FilterType_age(FilterList[filtertype].Source, FilterList[filtertype].Direction, FilterList[filtertype].Unit, FilterList[filtertype].Unit_count)
				} else if FilterList[filtertype].Filtertype == "pattern" {
					patternlist = FilterType_pattern(FilterList[filtertype].Kind, FilterList[filtertype].Value)
				} else if FilterList[filtertype].Filtertype == "space" {
					spacelist = FilterType_space(FilterList[filtertype].Disk_space)
				}
				comparelist = Indicesmapping(spacelist, agelist, patternlist)

			}
			// fmt.Println("agelist:", agelist)
			// fmt.Println("patternlist:", patternlist)
			// fmt.Println("spacelist:", spacelist)
			// fmt.Println("compare:", comparelist)
			detailMsg := fmt.Sprintf("Delete these indices :%s", comparelist)
			log_record.Logrecord("Details", detailMsg)
			//// delete function write from here
			DeleteIndex(comparelist)

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
			for filtertype := range FilterList {
				// fmt.Println("action:", ActionList[actions].Action, "description:", ActionList[actions].Description, "filtertype:", FilterList[filtertype].Filtertype)
				if FilterList[filtertype].Filtertype == "age" {
					agelist = FilterType_age(FilterList[filtertype].Source, FilterList[filtertype].Direction, FilterList[filtertype].Unit, FilterList[filtertype].Unit_count)
				} else if FilterList[filtertype].Filtertype == "pattern" {
					patternlist = FilterType_pattern(FilterList[filtertype].Kind, FilterList[filtertype].Value)
				} else if FilterList[filtertype].Filtertype == "space" {
					spacelist = FilterType_space(FilterList[filtertype].Disk_space)
				}
				comparelist = Indicesmapping(spacelist, agelist, patternlist)

			}
			MaxNumSegments = ActionList[actions].Options.MaxNumSegment
			// MaxNumSegments := ActionList[actions].Options.MaxNumSegments
			// fmt.Println("agelist:", agelist)
			// fmt.Println("patternlist:", patternlist)
			// fmt.Println("spacelist:", spacelist)
			// fmt.Println("compare:", comparelist)
			fmt.Println(MaxNumSegments)
			fmt.Println(ActionList[actions].Options.Key)
			fmt.Println(ActionList[actions].Options.Delay)
			fmt.Println(ActionList[actions].Options.MaxNumSegment)
			fmt.Println(ActionList[actions].Options.TimeoutOverride)
			fmt.Println(ActionList[actions].Description)
			detailMsg := fmt.Sprintf("forcemerge these indices :%s", comparelist)
			msg := fmt.Sprintf("segement num :%v", MaxNumSegments)
			log_record.Logrecord("segement", msg)
			log_record.Logrecord("Details", detailMsg)

			//// Close function write from here
			ForceMerge(comparelist, MaxNumSegments)

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
			for filtertype := range FilterList {
				// fmt.Println("action:", ActionList[actions].Action, "description:", ActionList[actions].Description, "filtertype:", FilterList[filtertype].Filtertype)
				if FilterList[filtertype].Filtertype == "age" {
					agelist = FilterType_age(FilterList[filtertype].Source, FilterList[filtertype].Direction, FilterList[filtertype].Unit, FilterList[filtertype].Unit_count)
				} else if FilterList[filtertype].Filtertype == "pattern" {
					patternlist = FilterType_pattern(FilterList[filtertype].Kind, FilterList[filtertype].Value)
				} else if FilterList[filtertype].Filtertype == "space" {
					spacelist = FilterType_space(FilterList[filtertype].Disk_space)
				}
				comparelist = Indicesmapping(spacelist, agelist, patternlist)

			}

			detailMsg := fmt.Sprintf("allocation these indices :%s", comparelist)
			log_record.Logrecord("Details", detailMsg)

			//// allocation function write from here
			Allocation(comparelist, allocationtype, key, value)

			Node_relocating_checking()

		}
	}

}
