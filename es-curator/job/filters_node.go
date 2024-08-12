package job

import (
	// "es-curator/log_record"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"es-curator/global"
	// "math"
)

func FilterType_age_node(nodeName string, source string, direction string, unit string, unit_count int) (indiceslist []string) {

	indicesinfo := CatIndices()

	match := MatchIndexBetweenNodeNCluster(indicesinfo, nodeName)

	if source == "creation_date" {

		var indices []string
		// 時間往前推
		var benchmarkDate string
		if unit == "years" {
			benchmarkDate = time.Now().AddDate(-unit_count, -0, -0).Format("2006-01-02 15:04:05")
		} else if unit == "months" {
			benchmarkDate = time.Now().AddDate(-0, -unit_count, -0).Format("2006-01-02 15:04:05")
		} else if unit == "days" {
			benchmarkDate = time.Now().AddDate(-0, -0, -unit_count).Format("2006-01-02 15:04:05")
		}

		benchmarkDateT, error := time.Parse("2006-01-02 15:04:05", benchmarkDate)
		if error != nil {
			fmt.Println(error)
			return
		}
		// for data := range indicesinfo {
		for data := range match {
			// 將indices的 creation date 由 unixtime 轉為 "2006-01-02"的格式
			timestamp, _ := strconv.ParseInt(match[data].CreationDate, 10, 64)
			CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")

			CreationDateT, error := time.Parse("2006-01-02 15:04:05", CreationDate)
			if error != nil {
				fmt.Println(error)
				return
			}

			// 滿足 direction = "older" 及 產生日期 (creation date) 在 基準日期(benchmark Date)之前的 index
			if CreationDateT.Before(benchmarkDateT) && direction == "older" {
				// fmt.Println(match[data].Index, "date is:", CreationDate, match[data].CreationDate)
				indices = append(indices, match[data].Index)

				// 滿足 direction = "younger" 及 產生日期 (creation date) 在 基準日期(benchmark Date)之後的 index
			} else if CreationDateT.After(benchmarkDateT) && direction == "younger" {
				// fmt.Println(match[data].Index, "date is:", CreationDate, match[data].CreationDate)
				indices = append(indices, match[data].Index)

				// 滿足 direction = "range" 及 產生日期 (creation date) 在 基準日期(benchmark Date)之後 及產生日期 (creation date) 在現在日期之前的 index
			}
		}
		// fmt.Println("last",indices)
		return indices
	}
	return
}

func FilterType_age_range_node(nodeName string, source string, direction string, unit string, range_from int, range_to int) (indiceslist []string) {

	indicesinfo := CatIndices()

	match := MatchIndexBetweenNodeNCluster(indicesinfo, nodeName)

	if source == "creation_date" {
		var indices []string
		// 時間往前推
		var RangeFromDate string
		var RangeToDate string
		if unit == "years" {
			RangeFromDate = time.Now().AddDate(-range_from, -0, -0).Format("2006-01-02 15:04:05")
			RangeToDate = time.Now().AddDate(-range_to, -0, -0).Format("2006-01-02 15:04:05")
		} else if unit == "months" {
			RangeFromDate = time.Now().AddDate(-0, -range_from, -0).Format("2006-01-02 15:04:05")
			RangeToDate = time.Now().AddDate(-0, -range_to, -0).Format("2006-01-02 15:04:05")
		} else if unit == "days" {
			RangeFromDate = time.Now().AddDate(-0, -0, -range_from).Format("2006-01-02 15:04:05")
			RangeToDate = time.Now().AddDate(-0, -0, -range_to).Format("2006-01-02 15:04:05")
		}
		RangeFromDateT, err := time.Parse("2006-01-02 15:04:05", RangeFromDate)
		if err != nil {
			// fmt.Println(error)
			global.Logger.Error(err.Error())
			return
		}
		RangeToDateT, err := time.Parse("2006-01-02 15:04:05", RangeToDate)
		if err != nil {
			// fmt.Println(error)
			global.Logger.Error(err.Error())
			return
		}
		// fmt.Println("RangeFromDateT", RangeFromDateT)
		// fmt.Println("RangeToDateT", RangeToDateT)
		// log_record.Logrecord("Details", fmt.Sprintf("Date Range From :%s ,Range To :%s", RangeFromDate, RangeToDate))
		global.Logger.Infow(fmt.Sprintf("Date Range From :%s ,Range To :%s", RangeFromDate, RangeToDate),"type","Details")

		for data := range match {
			// 將indices的 creation date 由 unixtime 轉為 "2006-01-02"的格式
			timestamp, _ := strconv.ParseInt(match[data].CreationDate, 10, 64)
			CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
			// NowDateTime := time.Now().Format("2006-01-02")

			// timeadjust := time.Now().AddDate(-0,-0,-global.EnvConfig.DeleteIndices.Filters.Unit_count).Format("2006-01-02")
			CreationDateT, error := time.Parse("2006-01-02 15:04:05", CreationDate)
			if error != nil {
				fmt.Println(error)

				return
			}

			// 滿足 direction = "range" 及 產生日期 (creation date) 在 (RangeFrom Date)之前 及在(RangeTo Date)之後的index
			if CreationDateT.After(RangeFromDateT) && CreationDateT.Before(RangeToDateT) && direction == "range" {
				// fmt.Println(match[data].Index, "date is:", CreationDate, match[data].CreationDate)
				indices = append(indices, match[data].Index)
			}
		}

		// fmt.Println("indices", indices)
		return indices
	}
	return
}

func FilterType_pattern_role(nodeName string, kind string, value []string) (indiceslist []string) {
	indicesinfo := CatIndices()
	match := MatchIndexBetweenNodeNCluster(indicesinfo,nodeName)

	// indicesinfo := CatIndices()
	var indices []string
	if kind == "prefix" {
		// fmt.Println("prefix")
		for data := range match {
			for _, pattern := range value {
				matchstring := fmt.Sprintf("^%s.*$", pattern)
				matchbool, err := regexp.MatchString(matchstring, match[data].Index)
				if err != nil {
					// log_record.Logrecord("ERROR ", "filter prefix error"+err.Error())
					global.Logger.Error(err.Error())
				}
				if matchbool {
					indices = append(indices, match[data].Index)
					// fmt.Println(indicesinfo[data].Index, matchbool)
				}
			}
		}
	} else if kind == "suffix" {
		for data := range match {
			for _, pattern := range value {
				matchstring := fmt.Sprintf("%s.*$", pattern)
				matchbool, err := regexp.MatchString(matchstring, match[data].Index)
				if err != nil {
					// log_record.Logrecord("ERROR ", "filter suffix error"+err.Error())
					global.Logger.Error(err.Error())
					// panic("suffix")
				}
				if matchbool {
					indices = append(indices, match[data].Index)
					// fmt.Println(indicesinfo[data].Index, matchbool)
				}

			}
		}
	} else if kind == "regex" {
		for data := range match {
			for _, pattern := range value {
				matchbool, err := regexp.MatchString(pattern, match[data].Index)
				if err != nil {
					// log_record.Logrecord("ERROR ", "filter regex error"+err.Error())
					global.Logger.Error(err.Error())
					// panic("regex")
				}
				if matchbool {
					indices = append(indices, match[data].Index)
					// fmt.Println(indicesinfo[data].Index, matchbool)
				}
			}
		}
	}
	indices = RemoveDuplicates(indices)
	return indices

}

func FilterType_space_role(nodeName string,patternlist []string, disk_space int) (indiceslist []string) {
	fmt.Println("nodeName ",nodeName)
	var indicesinfo CatIndice
	if len(patternlist) < 1 {
		indicesinfo = CatIndices()
	} else {
		indicesinfo = CatIndices_withPattern(patternlist)
	}


	match := MatchIndexBetweenNodeNCluster(indicesinfo,nodeName)


	var creationDateSlice []string

	var indexSizemap, creationdate_NameMap , onlyIndexName map[string]string
	indexSizemap = make(map[string]string)
	// var creationdate_NameMap map[string]string
	creationdate_NameMap = make(map[string]string)
	onlyIndexName = make(map[string]string)
	for i,data := range match {
		fmt.Println("i: ",i,"data+i: ",data.Index+i)
		onlyIndexName[data.Index+i] = data.Index
		//// 用 index name+i 做 key map size
		indexSizemap[data.Index+i] = data.StoreSize
		//// 用 CreationDate + Shard 做 key map index name+i
		creationdate_NameMap[data.CreationDate+data.Shard] = data.Index+i
		//// 用 CreationDate + Shard 組成的 array 
		creationDateSlice = append(creationDateSlice,data.CreationDate+data.Shard)
		// fmt.Println(data.Index, "size:", data.StoreSize, "date:", data.CreationDate)
	}

	// fmt.Println("indexSizemap",indexSizemap)
	// fmt.Println("creationdate_NameMap",creationdate_NameMap)
	// fmt.Println("creationDateSlice",creationDateSlice)


	var finalIndexList []string
	var aggregate_bytes []string
	if creationDateSlice == nil {
		//// 如果撈不到 index 則返回一個空的list
		// finalIndexList = append(finalIndexList, "")
		finalIndexList = nil
	} else {
		// 按 index 的 create_date 排序
		sort.Strings(creationDateSlice)
		//// 把 index_name 塞到 slice 中
		var indexSortbycreation []string
		for date := range creationDateSlice {
			indexSortbycreation = append(indexSortbycreation, creationdate_NameMap[creationDateSlice[date]])
		}
		/// 以 create_date 後整理好的 index name + i
		// fmt.Println("indexSortbycreation:", indexSortbycreation)

		//// 倒序 - 從 newest create 的 index 開始加總 disk_space
		var indexSortbycreationAsc []string
		for index := range indexSortbycreation {
			name := indexSortbycreation[len(indexSortbycreation)-index-1]
			indexSortbycreationAsc = append(indexSortbycreationAsc, name)

		}
		//// indexSortbycreationAsc - 按新到舊排序 index name + i
		// fmt.Println("asc:", indexSortbycreationAsc)

		total := 0

		for bytes := range indexSortbycreationAsc {
			// fmt.Println("bytes",bytes)
			var bytesnum int
			// fmt.Println("indexSortbycreationAsc[bytes]"+indexSortbycreationAsc[bytes])
			// fmt.Println("indexSizemap[indexSortbycreationAsc[bytes]]"+indexSizemap[indexSortbycreationAsc[bytes]])
			if indexSizemap[indexSortbycreationAsc[bytes]] == "" {
				bytesnum = 0
				// total += bytesnum
			} else {

				bytesint, err := strconv.Atoi(indexSizemap[indexSortbycreationAsc[bytes]])
				if err != nil {
					// log_record.Logrecord("ERROR ", "Error during conversion "+err.Error())
					global.Logger.Error(err.Error())
					return
				}
				// total += bytesnum
				bytesnum = bytesint
			}

			// 加總 index storage
			total += bytesnum
			// fmt.Println("total_in",total)

			if total > disk_space*1024*1024 {
				break
			}
			aggregate_bytes = append(aggregate_bytes, indexSortbycreationAsc[bytes])
			// fmt.Println("aggregate_bytes",aggregate_bytes)
		}
		// fmt.Println("final_list:", finalIndexList)
		// fmt.Println(total)

		_, removed := Diff(indexSortbycreationAsc, aggregate_bytes)
		// finalIndexList = removed
		// fmt.Println("added: ", added)
		// fmt.Println("removed: ", removed)

		for _,data := range removed {
			finalIndexList = append(finalIndexList, onlyIndexName[data])
		}
		
		finalIndexList = RemoveDuplicates(finalIndexList)
		// fmt.Println("finalIndexList: ", finalIndexList)
	}
	return finalIndexList
}

func FilterType_waterLevel_role(nodeName string, patternlist []string, upper_limit int, lower_limit int) (indiceslist []string) {

	var creationDateSlice, aggregate_bytes []string
	var indexSizemap, creationdate_NameMap map[string]string
	var indicesinfo CatIndice
	if len(patternlist) < 1 {
		indicesinfo = CatIndices()
	} else {
		indicesinfo = CatIndices_withPattern(patternlist)
	}


	match := MatchIndexBetweenNodeNCluster(indicesinfo,nodeName)

	/// 統計各個 Node 的 Average Water Level
	nodesinfo := CatNodesWithNodeName(nodeName)
	water_level := 0.00
	AllDiskTotal := 0.00
	for _, data := range nodesinfo {
		DiskUsedPercent, err := strconv.ParseFloat(data.DiskUsedPercent, 32)
		if err != nil {
			// log_record.Logrecord("ERROR", "Error during conversion DiskUsedPercent str"+err.Error())
			global.Logger.Error(err.Error())
			return
		}
		DiskTotalstr := strings.TrimSuffix(data.DiskTotal, "gb")
		DiskTotal, err := strconv.ParseFloat(DiskTotalstr, 32)
		if err != nil {
			// log_record.Logrecord("ERROR", "Error during conversion disk total str"+err.Error())
			global.Logger.Error(err.Error())
			return
		}
		AllDiskTotal += DiskTotal
		water_level += DiskUsedPercent
	}

	AnerageLevel := water_level / float64(len(nodesinfo))
	msg := fmt.Sprintf("Average Water Level: %f", AnerageLevel)
	// log_record.Logrecord("INFO", msg)
	global.Logger.Infow(msg)

	// for _, data := range nodesinfo {
	// 	fmt.Println("disk total", data.DiskTotal)
	// }
	var diskKbToClean float64

	// 觸發 upper_limit 才進行動作
	if AnerageLevel >= float64(upper_limit) {

		diskToCleanPercentage := float64(upper_limit) - float64(lower_limit)
		diskToClean := AllDiskTotal * (diskToCleanPercentage / 100)
		diskKbToClean = diskToClean * 1024 * 1024

		// log_record.Logrecord("Details", fmt.Sprintf("Disk Space to Clean %f gb", diskToClean))
		global.Logger.Infow(fmt.Sprintf("Disk Space to Clean %f gb", diskToClean),"type","Details")

		indexSizemap = make(map[string]string)
		creationdate_NameMap = make(map[string]string)

		for _, data := range match {
			indexSizemap[data.Index] = data.StoreSize
			creationdate_NameMap[data.CreationDate] = data.Index
			creationDateSlice = append(creationDateSlice, data.CreationDate)
			// fmt.Println(data.Index, "size", data.StoreSize, "date", data.CreationDate)
		}

		if creationDateSlice != nil {
			// 按 index 的 create_date 排序
			sort.Strings(creationDateSlice)
			// fmt.Println("sort of creationDateSlice:", creationDateSlice)
			// fmt.Println("indexSizemap", indexSizemap)
			// 把 index_name 塞到 slice 中
			var indexSortbycreation []string
			for date := range creationDateSlice {
				indexSortbycreation = append(indexSortbycreation, creationdate_NameMap[creationDateSlice[date]])
			}

			total := 0
			for _, data := range indexSortbycreation {

				bytesnum, err := strconv.Atoi(indexSizemap[data])
				if err != nil {
					// log_record.Logrecord("ERROR ", "Error during conversion"+err.Error())
					global.Logger.Error(err.Error())
					fmt.Println("Error during conversion 163")
					return
				}
				aggregate_bytes = append(aggregate_bytes, data)
				// 加總 index storage
				total += bytesnum
				if float64(total) > diskKbToClean {
					break
				}
			}
			// log_record.Logrecord("Details", fmt.Sprintf("Total Delete kbs %d", total))
			global.Logger.Infow(fmt.Sprintf("Total Delete kbs %d", total),"type","Details")
		}
	} else {
		// log_record.Logrecord("INFO ", "Water Level doesn't exceed upper limit")
		global.Logger.Infow("Water Level doesn't exceed upper limit")
	}
	return aggregate_bytes
}
