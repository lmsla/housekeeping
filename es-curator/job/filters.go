package job

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"time"
	// "es-curator/global"
)

func FilterType_age(source string, direction string, unit string, unit_count int) (indiceslist []string) {

	if source == "creation_date" {
		indicesinfo := CatIndices()
		var indices []string
		for data := range indicesinfo {
			// 將indices的 creation date 由 unixtime 轉為 "2006-01-02"的格式
			timestamp, _ := strconv.ParseInt(indicesinfo[data].CreationDate, 10, 64)
			CreationDate := time.UnixMilli(timestamp).Format("2006-01-02")
			// 時間往前推
			var benchmarkDate string
			if unit == "years" {
				benchmarkDate = time.Now().AddDate(-unit_count, -0, -0).Format("2006-01-02")
			} else if unit == "months" {
				benchmarkDate = time.Now().AddDate(-0, -unit_count, -0).Format("2006-01-02")
			} else if unit == "days" {
				benchmarkDate = time.Now().AddDate(-0, -0, -unit_count).Format("2006-01-02")
			}
			// timeadjust := time.Now().AddDate(-0,-0,-global.EnvConfig.DeleteIndices.Filters.Unit_count).Format("2006-01-02")
			CreationDateT, error := time.Parse("2006-01-02", CreationDate)
			if error != nil {
				fmt.Println(error)
				return
			}
			benchmarkDateT, error := time.Parse("2006-01-02", benchmarkDate)
			if error != nil {
				fmt.Println(error)
				return
			}
			// 滿足 direction = "older" 及 產生日期 (creation date) 在 基準日期(benchmark Date)之前的 index
			if CreationDateT.Before(benchmarkDateT) == true && direction == "older" {
				// fmt.Println(indicesinfo[data].Index, "date is:", CreationDate, indicesinfo[data].CreationDate)
				indices = append(indices, indicesinfo[data].Index)
				// DeleteIndex([]string{indicesinfo[data].Index})

				// 滿足 direction = "younger" 及 產生日期 (creation date) 在 基準日期(benchmark Date)之後的 index
			} else if CreationDateT.After(benchmarkDateT) == true && direction == "younger" {
				// fmt.Println(indicesinfo[data].Index, "date is:", CreationDate, indicesinfo[data].CreationDate)
				indices = append(indices, indicesinfo[data].Index)
				// DeleteIndex([]string{indicesinfo[data].Index})

			}

		}
		fmt.Println(indices)
		return indices
	}
	return
}

func FilterType_pattern(kind string, value string) (indiceslist []string) {
	indicesinfo := CatIndices()
	var indices []string
	if kind == "prefix" {
		fmt.Println("prefix")
		for data := range indicesinfo {
			matchstring := fmt.Sprintf("^%s.*$", value)
			matchbool, err := regexp.MatchString(matchstring, indicesinfo[data].Index)
			if err != nil {
				panic("wow")
			}
			if matchbool == true {
				indices = append(indices, indicesinfo[data].Index)
				// fmt.Println(indicesinfo[data].Index, matchbool)
			}
		}

	} else if kind == "suffix" {
		for data := range indicesinfo {
			matchstring := fmt.Sprintf("%s.*$", value)
			matchbool, err := regexp.MatchString(matchstring, indicesinfo[data].Index)
			if err != nil {
				panic("suffix")
			}
			if matchbool == true {
				indices = append(indices, indicesinfo[data].Index)
				// fmt.Println(indicesinfo[data].Index, matchbool)
			}
		}
	} else if kind == "regex" {
		for data := range indicesinfo {
			matchbool, err := regexp.MatchString(value, indicesinfo[data].Index)
			if err != nil {
				panic("suffix")
			}
			if matchbool == true {
				indices = append(indices, indicesinfo[data].Index)
				// fmt.Println(indicesinfo[data].Index, matchbool)
			}
		}
	}
	fmt.Println(indices)
	return indices

}

func FilterType_space(disk_space int) (indiceslist []string) {
	// comparelist := []string{"logstash-department-iis-20221224", "logstash-department-iis-20221229", "logstash-department-iis-20221230", "logstash-department-iis-20221231", "logstash-department-iis-20230101", "logstash-department-iis-20221225"}
	indicesinfo := CatIndices()
	var creationDateSlice []string
	var indexSizemap map[string]string
	indexSizemap = make(map[string]string)
	var creationdate_NameMap map[string]string
	creationdate_NameMap = make(map[string]string)
	for data := range indicesinfo {
		indexSizemap[indicesinfo[data].Index] = indicesinfo[data].StoreSize
		creationdate_NameMap[indicesinfo[data].CreationDate] = indicesinfo[data].Index
		creationDateSlice = append(creationDateSlice, indicesinfo[data].CreationDate)
		fmt.Println(indicesinfo[data].Index, "size", indicesinfo[data].StoreSize, "date", indicesinfo[data].CreationDate)
	}
	// 按 index 的 create_date 排序
	sort.Strings(creationDateSlice)
	// fmt.Println("sort of creationDateSlice:", creationDateSlice)
	// fmt.Println("indexSizemap", indexSizemap)
	// 把 index_name 塞到 slice 中
	var indexSortbycreation []string
	for date := range creationDateSlice {
		indexSortbycreation = append(indexSortbycreation, creationdate_NameMap[creationDateSlice[date]])
	}
	// fmt.Println("indexSortbycreation:", indexSortbycreation)

	/// 倒序 - 從 newest create 的 index 開始加總 disk_space
	var indexSortbycreationAsc []string
	for index := range indexSortbycreation {
		name := indexSortbycreation[len(indexSortbycreation)-index-1]
		indexSortbycreationAsc = append(indexSortbycreationAsc, name)

	}
	fmt.Println("asc:", indexSortbycreationAsc)
	var finalIndexList []string
	total := 0
	for bytes := range indexSortbycreationAsc {
		bytesnum, err := strconv.Atoi(indexSizemap[indexSortbycreationAsc[bytes]])
		if err != nil {
			fmt.Println("Error during conversion")
			return
		}
		finalIndexList = append(finalIndexList, indexSortbycreationAsc[bytes])
		// 加總 index storage
		total += bytesnum
		if total > disk_space*1024*1024 {
			break
		}

	}
	reserveIndexList := finalIndexList[:len(finalIndexList)-1]
	finalIndexList = indexSortbycreationAsc[len(reserveIndexList):]
	// fmt.Println("reserveIndexList:", reserveIndexList)
	// fmt.Println("final_list:", finalIndexList)
	// fmt.Println(total)
	return finalIndexList
}
