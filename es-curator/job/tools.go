package job

import (
	"time"
	"fmt"
)

func Indicesmapping(list1 []string, list2 []string, list3 []string) []string {
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