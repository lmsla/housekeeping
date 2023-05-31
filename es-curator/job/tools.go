package job

import (
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


//取得兩個 list中相同的元素 method1
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

//取得兩個 list中相同的元素 method2
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

func Filter_of_filter(filter_record, agelist, patternlist, spacelist []string) []string{
	var compareList []string
	aps := []string{"age", "pattern", "space"}
	ap := []string{"age", "pattern"}
	as := []string{"age", "space"}
	sp := []string{"space","pattern"}
	a := []string{"age"}
	p := []string{"pattern"}
	s := []string{"space"}
	sort.Strings(ap)
	sort.Strings(as)
	sort.Strings(sp)
	sort.Strings(aps)
	sort.Strings(filter_record)

	// fmt.Println("sp:",sp)
	fmt.Println("filter_record: ",filter_record)

	switch {
	case reflect.DeepEqual(filter_record, aps) == true:
		compareList = Indicesmapping3(agelist, patternlist, spacelist)

	case reflect.DeepEqual(filter_record, ap) == true:
		compareList = Indicesmapping2(agelist, patternlist)

	case reflect.DeepEqual(filter_record,as) == true:
		compareList = Indicesmapping2(agelist,spacelist)

	case reflect.DeepEqual(filter_record,sp) == true:
		compareList = Indicesmapping2(patternlist,spacelist)

	case reflect.DeepEqual(filter_record,a) == true:
		compareList = agelist
	
	case reflect.DeepEqual(filter_record,p) == true:
		compareList = patternlist
	
	case reflect.DeepEqual(filter_record,s) == true:
		compareList = spacelist
	
	}
	fmt.Println("filter_of_filter's compare: ",compareList)
	return compareList
}
