package job

import (
	"es-curator/log_record"
	"es-curator/structs"
	"fmt"
	"reflect"
	"sort"
	"time"
)

func Indicesmapping3(list1 []string, list2 []string, list3 []string) []string {
	var compareList []string
	if list1 != nil && list2 != nil && list3 != nil {
		///123
		for list1data := range list1 {
			for list2data := range list2 {
				for list3data := range list3 {
					if list2[list2data] == list1[list1data] && list3[list3data] == list1[list1data] {
						compareList = append(compareList, list1[list1data])
					}
				}

			}
		}
	}

	return compareList
}

// 取得兩個 list中相同的元素 method1
func Intersection(a, b []string) []string {
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

// 取得兩個 list中相同的元素 method2
func Indicesmapping2(list1 []string, list2 []string) []string {
	var compareList []string
	if list1 != nil && list2 != nil {
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

func Indicesmapping(list1, list2, list3 []string) []string {
	var compareList []string
	switch {
	///123
	case list1 != nil && list2 != nil && list3 != nil:
		for list1data := range list1 {
			for list2data := range list2 {
				for list3data := range list3 {
					if list2[list2data] == list1[list1data] && list3[list3data] == list1[list1data] {
						compareList = append(compareList, list1[list1data])
					}
				}

			}
		}
	/// 12
	case list1 != nil && list2 != nil && list3 == nil:
		for list1data := range list1 {
			for list2data := range list2 {
				if list1[list1data] == list2[list2data] {
					compareList = append(compareList, list1[list1data])
				}
			}
		}
	/// 23
	case list1 == nil && list2 != nil && list3 != nil:
		for list2data := range list2 {
			for list3data := range list3 {
				if list2[list2data] == list3[list3data] {
					compareList = append(compareList, list2[list2data])
				}
			}
		}
	/// 13
	case list1 != nil && list2 == nil && list3 != nil:
		for list1data := range list1 {
			for list3data := range list3 {
				if list1[list1data] == list3[list3data] {
					compareList = append(compareList, list1[list1data])
				}
			}
		}

	case list1 != nil && list2 == nil && list3 == nil:
		compareList = list1
	case list1 == nil && list2 != nil && list3 == nil:
		compareList = list2
	case list1 == nil && list2 == nil && list3 != nil:
		compareList = list3
	}

	return compareList
}

func Node_relocating_checking() {
	i := 1
	for i <= 10000 {
		indicesinfo := ClusterHealth()
		time.Sleep(3 * time.Second)
		// a := fmt.Printf("%s",indicesinfo["relocating_shards"])
		if indicesinfo.RelocatingShards != 0 {
			fmt.Println("relocating_shards : not finished")
			continue
		} else if indicesinfo.RelocatingShards == 0 {
			fmt.Println("relocating_shards : 0")
			break
		}
	}
	fmt.Println("check finish1")

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

func Filter_of_filter(filter_record []string, FilterList []structs.Filter) []string {
	var agelist, patternlist, spacelist, water_level_list, patternListPre, compareList []string
	// aps := []string{"age", "pattern", "space"}
	ap := []string{"age", "pattern"}
	// as := []string{"age", "space"}
	sp := []string{"space", "pattern"}
	wp := []string{"water_level", "pattern"}
	a := []string{"age"}
	p := []string{"pattern"}
	s := []string{"space"}
	w := []string{"water_level"}
	sort.Strings(ap)
	sort.Strings(wp)
	sort.Strings(sp)
	// sort.Strings(aps)
	sort.Strings(filter_record)

	// fmt.Println("sp:",sp)
	// fmt.Println("filter_record: ", filter_record)

	arr1 := []string{"space", "pattern", "age"}
	arr2 := []string{"space", "age"}
	arr3 := []string{"space", "water_level"}
	arr4 := []string{"space", "water_level", "pattern"}
	arr5 := []string{"space", "water_level", "pattern", "age"}
	sort.Strings(arr1)
	sort.Strings(arr2)
	sort.Strings(arr3)
	sort.Strings(arr4)
	sort.Strings(arr5)
	sort.Strings(filter_record)

	// fmt.Println("filterList: ", FilterList)

	for filtertype := range FilterList {
		if FilterList[filtertype].Filtertype == "pattern" {
			for _, pattern := range FilterList[filtertype].Value {
				patternListPre = append(patternListPre, pattern+"*")
			}

		}
	}

	if fmt.Sprint(filter_record) == fmt.Sprint(arr1) || fmt.Sprint(filter_record) == fmt.Sprint(arr2) {
		log_record.Logrecord("ERROR", "Can't use age & space at the same time")
	} else if fmt.Sprint(filter_record) == fmt.Sprint(arr3) || fmt.Sprint(filter_record) == fmt.Sprint(arr4) {
		log_record.Logrecord("ERROR", "Can't use space & water_level at the same time")
	} else if fmt.Sprint(filter_record) == fmt.Sprint(arr5) {
		log_record.Logrecord("ERROR", "age can't use with space or water_level at the same time")
	} else {
		for filtertype := range FilterList {
			if FilterList[filtertype].Filtertype == "age" && FilterList[filtertype].Direction != "range" {
				agelist = FilterType_age(FilterList[filtertype].Source, FilterList[filtertype].Direction, FilterList[filtertype].Unit, FilterList[filtertype].Unit_count)
			} else if FilterList[filtertype].Filtertype == "age" && FilterList[filtertype].Direction == "range" {
				agelist = FilterType_age_range(FilterList[filtertype].Source, FilterList[filtertype].Direction, FilterList[filtertype].Unit, FilterList[filtertype].Range_From, FilterList[filtertype].Range_To)
			} else if FilterList[filtertype].Filtertype == "pattern" {
				patternlist = FilterType_pattern(FilterList[filtertype].Kind, FilterList[filtertype].Value)
			} else if FilterList[filtertype].Filtertype == "space" {
				spacelist = FilterType_space(patternListPre, FilterList[filtertype].Disk_space)
			} else if FilterList[filtertype].Filtertype == "water_level" {
				water_level_list = FilterType_waterLevel(patternListPre, FilterList[filtertype].Upper_limit, FilterList[filtertype].Lower_limit)
			}

		}
	}

	switch {

	case reflect.DeepEqual(filter_record, ap):
		compareList = Indicesmapping2(agelist, patternlist)

	case reflect.DeepEqual(filter_record, sp):
		compareList = Indicesmapping2(patternlist, spacelist)

	case reflect.DeepEqual(filter_record, wp):
		compareList = Indicesmapping2(patternlist, water_level_list)

	case reflect.DeepEqual(filter_record, a):
		compareList = agelist

	case reflect.DeepEqual(filter_record, p):
		compareList = patternlist

	case reflect.DeepEqual(filter_record, s):
		compareList = spacelist

	case reflect.DeepEqual(filter_record, w):
		compareList = water_level_list

	}
	fmt.Println("filter_of_filter's compare: ", compareList)
	return compareList
}

func Filter_of_filter_bak(filter_record, agelist, patternlist, spacelist []string) []string {
	var compareList []string
	aps := []string{"age", "pattern", "space"}
	ap := []string{"age", "pattern"}
	as := []string{"age", "space"}
	sp := []string{"space", "pattern"}
	a := []string{"age"}
	p := []string{"pattern"}
	s := []string{"space"}
	sort.Strings(ap)
	sort.Strings(as)
	sort.Strings(sp)
	sort.Strings(aps)
	sort.Strings(filter_record)

	// fmt.Println("sp:",sp)
	fmt.Println("filter_record: ", filter_record)

	switch {
	case reflect.DeepEqual(filter_record, aps):
		compareList = Indicesmapping3(agelist, patternlist, spacelist)

	case reflect.DeepEqual(filter_record, ap):
		compareList = Indicesmapping2(agelist, patternlist)

	case reflect.DeepEqual(filter_record, as):
		compareList = Indicesmapping2(agelist, spacelist)

	case reflect.DeepEqual(filter_record, sp):
		compareList = Indicesmapping2(patternlist, spacelist)

	case reflect.DeepEqual(filter_record, a):
		compareList = agelist

	case reflect.DeepEqual(filter_record, p):
		compareList = patternlist

	case reflect.DeepEqual(filter_record, s):
		compareList = spacelist

	}
	fmt.Println("filter_of_filter's compare: ", compareList)
	return compareList
}
