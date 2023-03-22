package job

import (
	"es-curator/global"
	"es-curator/log_record"
	"fmt"
)


func Action_controll() {
	// Action_open_indices()
	// Action_close_indices()
	// Action_delete_indices()
	// Action_forcemerge_indices()
	Action_allocation_indices()

}





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

func Action_forcemerge_indices() {
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
			ForceMerge(comparelist,MaxNumSegments)
		}
	}

}



func Action_allocation_indices() {
	ActionList := global.ActionStruct.Actions

	for actions := range ActionList {
		// fmt.Println(global.ActionStruct.Actions[actions].Action)
		if ActionList[actions].Action == "allocation" {
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
			detailMsg := fmt.Sprintf("allocation these indices :%s", comparelist)
			log_record.Logrecord("Details", detailMsg)

			//// Close function write from here
			Allocation1(comparelist)
		}
	}

}
