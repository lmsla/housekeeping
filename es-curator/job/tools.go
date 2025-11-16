package job

import (
	"es-curator/global"
	"strings"
	// "es-curator/log_record"
	"es-curator/structs"
	"fmt"
	// "reflect"
	"sort"
	"strconv"
	"time"
)

// 取得兩個 list中相同的元素 method1
func Intersection1(a, b []string) []string {
	m := make(map[string]bool)
	for _, item := range a {
		m[item] = true
	}
	var intersection []string
	for _, item := range b {
		if m[item] {
			intersection = append(intersection, item)
		}
	}
	return intersection
}

// 取得兩個 list中相同的元素 method2 (優化版本 - 使用 hash map, O(n) 複雜度)
func Indicesmapping2(list1 []string, list2 []string) []string {
	// 使用高效的 Intersection1 實現 (O(n) 複雜度)
	return Intersection1(list1, list2)
}

// Indicesmapping2_bak 原始實現備份 (O(n²) 複雜度) - 保留以防萬一
func Indicesmapping2_bak(list1 []string, list2 []string) []string {
	var compareList []string
	if len(list1) != 0 && len(list2) != 0 {
		for list1data := range list1 {
			for list2data := range list2 {
				if list1[list1data] == list2[list2data] {
					compareList = append(compareList, list1[list1data])
				}
			}
		}
	}
	return compareList
}

func Node_relocating_checking() {
	i := 1
	for i <= 10000 {
		indicesinfo := ClusterHealth()
		time.Sleep(5 * time.Second)
		// a := fmt.Printf("%s",indicesinfo["relocating_shards"])
		if indicesinfo.RelocatingShards != 0 {
			fmt.Printf("relocating_shards : %d shards not finished\n", indicesinfo.RelocatingShards)
			continue
		} else if indicesinfo.RelocatingShards == 0 {
			fmt.Println("relocating_shards : 0")
			break
		}
	}
	fmt.Println("check finish")

}

func Diff(a, b []string) (added []string, removed []string) {
	m := make(map[string]bool)
	for _, item := range a {
		m[item] = true
	}
	for _, item := range b {
		if _, ok := m[item]; !ok {
			added = append(added, item)
		} else {
			delete(m, item)
		}
	}
	for item := range m {
		removed = append(removed, item)
	}
	return added, removed
}

func Filter_of_filter(filterRecord []string, FilterList []structs.Filter) []string {

	var agelist, patternlist, spacelist, water_level_list, patternListPre, compareList []string

	for filtertype := range FilterList {
		if FilterList[filtertype].Filtertype == "pattern" {
			patternListPre = FilterType_pattern(FilterList[filtertype].Kind, FilterList[filtertype].Value)
		}
	}

	if containsBothParams(filterRecord, "age", "space") {
		global.Logger.Error("Can't use age & space at the same time")
	} else if containsBothParams(filterRecord, "age", "water_level") {
		global.Logger.Error("Can't use age & water_level at the same time")
	} else if containsBothParams(filterRecord, "space", "water_level") {
		global.Logger.Error("Can't use space & water_level at the same time")
	} else {
		for filtertype := range FilterList {
			if FilterList[filtertype].Filtertype == "age" && FilterList[filtertype].Direction != "range" {
				agelist = FilterType_age(FilterList[filtertype].Source, FilterList[filtertype].Direction, FilterList[filtertype].Unit, FilterList[filtertype].UnitCount)
			} else if FilterList[filtertype].Filtertype == "age" && FilterList[filtertype].Direction == "range" {
				agelist = FilterType_age_range(FilterList[filtertype].Source, FilterList[filtertype].Direction, FilterList[filtertype].Unit, FilterList[filtertype].RangeFrom, FilterList[filtertype].RangeTo)
			} else if FilterList[filtertype].Filtertype == "pattern" {
				patternlist = FilterType_pattern(FilterList[filtertype].Kind, FilterList[filtertype].Value)
			} else if FilterList[filtertype].Filtertype == "space" {
				spacelist = FilterType_space(patternListPre, FilterList[filtertype].DiskSpace)
			} else if FilterList[filtertype].Filtertype == "water_level" {
				water_level_list = FilterType_waterLevel(patternListPre, FilterList[filtertype].UpperLimit, FilterList[filtertype].LowerLimit)
			}

		}
	}

	agelist = RemoveDuplicates(agelist)
	patternlist = RemoveDuplicates(patternlist)
	spacelist = RemoveDuplicates(spacelist)
	water_level_list = RemoveDuplicates(water_level_list)

	compareList = ResolveCompareListNonRole(filterRecord,agelist,patternlist,spacelist,water_level_list)

	// fmt.Println("filter_of_filter's compare: ", compareList)
	return compareList
}

func Filters_With_node(role []string, filterRecord []string, FilterList []structs.Filter) []string {
	var agelist, patternlist_tmp, patternlist, patternListPre, spacelist_tmp, spacelist, water_level_list, compareList []string

	for filtertype := range FilterList {
		if FilterList[filtertype].Filtertype == "pattern" {
			patternListPre = FilterType_pattern(FilterList[filtertype].Kind, FilterList[filtertype].Value)
		}
	}

	for filtertype := range FilterList {
		if FilterList[filtertype].Filtertype == "node_role" {
			role = FilterList[filtertype].Value
		}
	}

	if containsBothParams(filterRecord, "age", "space") {
		global.Logger.Error("Can't use age & space at the same time")
	} else if containsBothParams(filterRecord, "age", "water_level") {
		global.Logger.Error("Can't use age & water_level at the same time")
	} else if containsBothParams(filterRecord, "space", "water_level") {
		global.Logger.Error("Can't use space & water_level at the same time")
	} else {
		// 取得符合 node role 的 node names
		nodeNames := NodeRoleDetermination(role[0])

		for _, nodeName := range nodeNames {

			for filtertype := range FilterList {
				if FilterList[filtertype].Filtertype == "age" && FilterList[filtertype].Direction != "range" {
					agelist = FilterType_age_node(nodeName, FilterList[filtertype].Source, FilterList[filtertype].Direction, FilterList[filtertype].Unit, FilterList[filtertype].UnitCount)
				} else if FilterList[filtertype].Filtertype == "age" && FilterList[filtertype].Direction == "range" {
					agelist = FilterType_age_range_node(nodeName, FilterList[filtertype].Source, FilterList[filtertype].Direction, FilterList[filtertype].Unit, FilterList[filtertype].RangeFrom, FilterList[filtertype].RangeTo)
				} else if FilterList[filtertype].Filtertype == "pattern" {
					patternlist_tmp = FilterType_pattern_role(nodeName, FilterList[filtertype].Kind, FilterList[filtertype].Value)
					patternlist = append(patternlist, patternlist_tmp...)
				} else if FilterList[filtertype].Filtertype == "water_level" {
					water_level_list = FilterType_waterLevel_role(nodeName, patternListPre, FilterList[filtertype].UpperLimit, FilterList[filtertype].LowerLimit)
					// }  else if FilterList[filtertype].Filtertype == "space" {
					// 	spacelist_tmp = FilterType_space_role(nodeName, patternListPre, FilterList[filtertype].Disk_space)
					// 	spacelist = append(spacelist, spacelist_tmp...)
					// }
				}

			}

		}
		// 帶 node role 時，space list 另外處理
		for filtertype := range FilterList {
			if FilterList[filtertype].Filtertype == "space" {
				spacelist_tmp = FilterType_space_role(nodeNames, patternListPre, FilterList[filtertype].DiskSpace)
				fmt.Println("spacelist_tmp", spacelist_tmp)
				spacelist = append(spacelist, spacelist_tmp...)
			}
		}
	}

	agelist = RemoveDuplicates(agelist)
	patternlist = RemoveDuplicates(patternlist)
	spacelist = RemoveDuplicates(spacelist)
	water_level_list = RemoveDuplicates(water_level_list)

	compareList = ResolveCompareList(filterRecord,agelist,patternlist,spacelist,water_level_list)

	// fmt.Println("filter_of_filter's compare: ", compareList)
	return compareList
}

func containsBothParams(params []string, param1, param2 string) bool {
	hasParam1 := false
	hasParam2 := false

	for _, param := range params {
		if param == param1 {
			hasParam1 = true
		} else if param == param2 {
			hasParam2 = true
		}
		if hasParam1 && hasParam2 {
			return true
		}
	}
	return false
}

func MatchIndexBetweenNodeNCluster1(indicesinfo CatIndice, nodeNames []string) map[string]IndicesInfo {

	match := make(map[string]IndicesInfo)
	//// 取得每一個 node 存放的 shards

	for _, indices := range indicesinfo {
		for _, nodeName := range nodeNames {
			shardsinfo := CatShardsbyNodeName(nodeName)

			for _, data := range shardsinfo {
				matchKey := fmt.Sprintf("%s-%s-%s", nodeName, data.Shard, data.Index)
				if indices.Index == data.Index {
					match[matchKey] = IndicesInfo{
						Index:        data.Index,
						DocsCount:    data.Docs,
						StoreSize:    data.Store,
						CreationDate: indices.CreationDate,
						Shard:        data.Shard,
					}
				}
			}
		}
	}
	return match
}

func MatchIndexBetweenNodeNCluster(indicesinfo CatIndice, nodeName string) map[string]IndicesInfo {

	match := make(map[string]IndicesInfo)
	//// 取得每一個 node 存放的 shards

	shardsinfo := CatShardsbyNodeName(nodeName)

	for _, indices := range indicesinfo {
		for i, data := range shardsinfo {
			if indices.Index == data.Index {
				match[strconv.Itoa(i)] = IndicesInfo{
					Index:        data.Index,
					DocsCount:    data.Docs,
					StoreSize:    data.Store,
					CreationDate: indices.CreationDate,
					Shard:        data.Shard,
				}
			}
		}
	}

	return match
}

// 用來對 index list 去重
func RemoveDuplicates(arr []string) []string {
	uniqueMap := make(map[string]bool) // 用於存儲唯一元素
	var uniqueArr []string

	for _, item := range arr {
		if _, exists := uniqueMap[item]; !exists {
			uniqueMap[item] = true
			uniqueArr = append(uniqueArr, item)
		}
	}

	return uniqueArr
}

// 用來對 index list 去重
func RemoveDuplicates1(arr []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, data := range arr {
		if _, ok := seen[data]; !ok {
			result = append(result, data)
			seen[data] = true
		}
	}
	return result
}


func ResolveCompareListNonRole(filterRecord []string, agelist, patternlist, spacelist, water_level_list []string) []string {
	// 對條件進行排序，確保順序一致性
	sort.Strings(filterRecord)
	key := strings.Join(filterRecord, "-") // 建立唯一 key，如 "age-pattern-node_role"

	// 建立條件組合與對應邏輯的映射表 取 list 交集
	actionMap := map[string]func() []string{
		"age": func() []string {
			return agelist
		},
		"pattern": func() []string {
			return patternlist
		},
		"space": func() []string {
			return spacelist
		},
		"water_level": func() []string {
			return water_level_list
		},
		"age-pattern": func() []string {
			return Indicesmapping2(agelist, patternlist)
		},
		"pattern-space": func() []string {
			return Indicesmapping2(patternlist, spacelist)
		},
		"pattern-water_level": func() []string {
			return Indicesmapping2(patternlist, water_level_list)
		},
	}

	// 依據 key 執行對應邏輯
	if action, ok := actionMap[key]; ok {
		return action()
	} else {
		global.Logger.Error("Undefined filter combination,Please check filters again.")
		return nil
	}
}




// water_level 應該考慮不與 node_role 並用
func ResolveCompareList(filterRecord []string, agelist, patternlist, spacelist, water_level_list []string) []string {
	// 對條件進行排序，確保順序一致性
	sort.Strings(filterRecord)
	key := strings.Join(filterRecord, "-") // 建立唯一 key，如 "age-pattern-node_role"


	// 建立條件組合與對應邏輯的映射表 取 list 交集
	actionMap := map[string]func() []string{
		"age-node_role": func() []string {
			return agelist
		},
		"node_role-pattern": func() []string {
			return patternlist
		},
		"node_role-space": func() []string {
			return spacelist
		},
		"node_role-water_level": func() []string {
			return water_level_list
		},
		"age-node_role-pattern": func() []string {
			return Indicesmapping2(agelist, patternlist)
		},
		"node_role-pattern-space": func() []string {
			return Indicesmapping2(patternlist, spacelist)
		},
		"node_role-pattern-water_level": func() []string {
			return Indicesmapping2(patternlist, water_level_list)
		},
	}

	// 依據 key 執行對應邏輯
	if action, ok := actionMap[key]; ok {
		return action()
	} else {
		global.Logger.Error("Undefined filter combination,Please check filters again.")
		return nil
	}
}


