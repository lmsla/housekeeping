package job

import (
	"fmt"
	"housekeeping/internal/global"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	// "math"
)

// ✅ HIGH-008 修復: 使用 sync.Pool 重用切片，減少內存分配和 GC 壓力
var (
	// stringSlicePool 用於重用字符串切片
	stringSlicePool = sync.Pool{
		New: func() interface{} {
			// 預分配 2000 容量，適用於大型 ES 集群
			s := make([]string, 0, 2000)
			return &s
		},
	}

	// mapPool 用於重用 map[string]string
	mapPool = sync.Pool{
		New: func() interface{} {
			m := make(map[string]string, 2000)
			return &m
		},
	}
)

func timetransform(unit string, unit_count int) string {
	var benchmarkDate string
	if unit == "years" {
		benchmarkDate = time.Now().AddDate(-unit_count, -0, -0).Format("2006-01-02 15:04:05")
	} else if unit == "months" {
		benchmarkDate = time.Now().AddDate(-0, -unit_count, -0).Format("2006-01-02 15:04:05")
	} else if unit == "days" {
		benchmarkDate = time.Now().AddDate(-0, -0, -unit_count).Format("2006-01-02 15:04:05")
	}
	return benchmarkDate
}

func FilterType_age(source string, direction string, unit string, unit_count int) (indiceslist []string) {

	if source == "creation_date" {
		indicesinfo := metadataProvider.CatIndices()
		var indices []string
		var benchmarkDate = timetransform(unit, unit_count)

		// // 時間往前推
		// var benchmarkDate string
		// if unit == "years" {
		// 	benchmarkDate = time.Now().AddDate(-unit_count, -0, -0).Format("2006-01-02 15:04:05")
		// } else if unit == "months" {
		// 	benchmarkDate = time.Now().AddDate(-0, -unit_count, -0).Format("2006-01-02 15:04:05")
		// } else if unit == "days" {
		// 	benchmarkDate = time.Now().AddDate(-0, -0, -unit_count).Format("2006-01-02 15:04:05")
		// }

		benchmarkDateT, error := time.Parse("2006-01-02 15:04:05", benchmarkDate)
		if error != nil {
			fmt.Println(error)
			return
		}
		for data := range indicesinfo {
			// 將indices的 creation date 由 unixtime 轉為 "2006-01-02"的格式
			timestamp, _ := strconv.ParseInt(indicesinfo[data].CreationDate, 10, 64)
			CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")

			CreationDateT, error := time.Parse("2006-01-02 15:04:05", CreationDate)
			if error != nil {
				fmt.Println(error)
				return
			}

			// 滿足 direction = "older" 及 產生日期 (creation date) 在 基準日期(benchmark Date)之前的 index
			if CreationDateT.Before(benchmarkDateT) && direction == "older" {
				// fmt.Println(indicesinfo[data].Index, "date is:", CreationDate, indicesinfo[data].CreationDate)
				indices = append(indices, indicesinfo[data].Index)

				// 滿足 direction = "younger" 及 產生日期 (creation date) 在 基準日期(benchmark Date)之後的 index
			} else if CreationDateT.After(benchmarkDateT) && direction == "younger" {
				// fmt.Println(indicesinfo[data].Index, "date is:", CreationDate, indicesinfo[data].CreationDate)
				indices = append(indices, indicesinfo[data].Index)
			}
		}
		// fmt.Println(indices)
		return indices
	}
	return
}

func FilterType_age_range(source string, direction string, unit string, range_from int, range_to int) (indiceslist []string) {

	if source == "creation_date" {
		indicesinfo := metadataProvider.CatIndices()
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
		RangeFromDateT, error := time.Parse("2006-01-02 15:04:05", RangeFromDate)
		if error != nil {
			fmt.Println(error)
			return
		}
		RangeToDateT, error := time.Parse("2006-01-02 15:04:05", RangeToDate)
		if error != nil {
			fmt.Println(error)
			return
		}

		// log_record.Logrecord("Details", fmt.Sprintf("Date Range From :%s ,Range To :%s", RangeFromDate, RangeToDate))
		global.Logger.Infow(fmt.Sprintf("Date Range From :%s ,Range To :%s", RangeFromDate, RangeToDate), "type", "Details")

		for data := range indicesinfo {
			// 將indices的 creation date 由 unixtime 轉為 "2006-01-02"的格式
			timestamp, _ := strconv.ParseInt(indicesinfo[data].CreationDate, 10, 64)
			CreationDate := time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
			// NowDateTime := time.Now().Format("2006-01-02")
			CreationDateT, error := time.Parse("2006-01-02 15:04:05", CreationDate)
			if error != nil {
				fmt.Println(error)
				return
			}

			// 滿足 direction = "range" 及 產生日期 (creation date) 在 (RangeFrom Date)之後 及在(RangeTo Date)之前的index
			if CreationDateT.After(RangeFromDateT) && CreationDateT.Before(RangeToDateT) && direction == "range" {
				// fmt.Println(indicesinfo[data].Index, "date is:", CreationDate, indicesinfo[data].CreationDate)
				indices = append(indices, indicesinfo[data].Index)
			}
		}
		return indices
	}
	return
}

func FilterType_pattern(kind string, value []string) (indiceslist []string) {
	indicesinfo := metadataProvider.CatIndices()
	var indices []string
	var systemIndicesExcluded []string

	if kind == "prefix" {
		for data := range indicesinfo {
			indexName := indicesinfo[data].Index

			// 🔒 安全檢查：跳過系統 index
			if isSystemIndex(indexName) {
				systemIndicesExcluded = append(systemIndicesExcluded, indexName)
				continue
			}

			for _, pattern := range value {
				matchstring := fmt.Sprintf("^%s.*$", pattern)
				matchbool, err := regexp.MatchString(matchstring, indexName)
				if err != nil {
					global.Logger.Error(err.Error())
				}
				if matchbool {
					indices = append(indices, indexName)
					break // 已匹配，無需再檢查其他 pattern
				}
			}
		}
	} else if kind == "suffix" {
		for data := range indicesinfo {
			indexName := indicesinfo[data].Index

			// 🔒 安全檢查：跳過系統 index
			if isSystemIndex(indexName) {
				systemIndicesExcluded = append(systemIndicesExcluded, indexName)
				continue
			}

			for _, pattern := range value {
				matchstring := fmt.Sprintf("%s.*$", pattern)
				matchbool, err := regexp.MatchString(matchstring, indexName)
				if err != nil {
					global.Logger.Error(err.Error())
				}
				if matchbool {
					indices = append(indices, indexName)
					break // 已匹配，無需再檢查其他 pattern
				}
			}
		}
	} else if kind == "regex" {
		for data := range indicesinfo {
			indexName := indicesinfo[data].Index

			// 🔒 安全檢查：跳過系統 index
			if isSystemIndex(indexName) {
				systemIndicesExcluded = append(systemIndicesExcluded, indexName)
				continue
			}

			for _, pattern := range value {
				matchbool, err := regexp.MatchString(pattern, indexName)
				if err != nil {
					global.Logger.Error(err.Error())
				}
				if matchbool {
					indices = append(indices, indexName)
					break // 已匹配，無需再檢查其他 pattern
				}
			}
		}
	}

	// 📊 記錄 pattern filter 執行結果
	if len(systemIndicesExcluded) > 0 {
		global.Logger.Infow("🔒 System indices excluded from pattern filter",
			"excluded_count", len(systemIndicesExcluded),
			"examples", getFirstN(systemIndicesExcluded, 3))
	}

	if len(indices) == 0 {
		global.Logger.Warnw("⚠️  Pattern filter matched NO indices (system indices excluded)",
			"kind", kind,
			"patterns", value,
			"total_indices_checked", len(indicesinfo),
			"system_indices_excluded", len(systemIndicesExcluded))
	} else {
		global.Logger.Infow("Pattern filter matched indices",
			"kind", kind,
			"patterns", value,
			"matched_count", len(indices),
			"system_indices_excluded", len(systemIndicesExcluded))
	}

	return indices
}

// getFirstN 返回 slice 的前 n 個元素，用於日誌記錄
func getFirstN(slice []string, n int) []string {
	if len(slice) <= n {
		return slice
	}
	return slice[:n]
}

// // 將 pattern list 切分
func chunkSlice(slice []string, chunkSize int) [][]string {
	var chunks [][]string
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

func FilterType_space(patternlist []string, disk_space int) (indiceslist []string) {
	fmt.Println("disk_space", disk_space)

	// ✅ HIGH-008 修復: 從池中獲取切片，減少內存分配
	creationDateSlicePtr := stringSlicePool.Get().(*[]string)
	indexSortbycreationPtr := stringSlicePool.Get().(*[]string)
	indexSortbycreationAscPtr := stringSlicePool.Get().(*[]string)
	aggregateBytesPtr := stringSlicePool.Get().(*[]string)
	finalIndexListPtr := stringSlicePool.Get().(*[]string)

	indexSizemapPtr := mapPool.Get().(*map[string]string)
	creationdateNameMapPtr := mapPool.Get().(*map[string]string)

	// ✅ 確保歸還給池
	defer func() {
		// 清空切片但保留容量
		*creationDateSlicePtr = (*creationDateSlicePtr)[:0]
		*indexSortbycreationPtr = (*indexSortbycreationPtr)[:0]
		*indexSortbycreationAscPtr = (*indexSortbycreationAscPtr)[:0]
		*aggregateBytesPtr = (*aggregateBytesPtr)[:0]
		*finalIndexListPtr = (*finalIndexListPtr)[:0]

		// ✅ 明確清理 map 後歸還
		for k := range *indexSizemapPtr {
			delete(*indexSizemapPtr, k)
		}
		for k := range *creationdateNameMapPtr {
			delete(*creationdateNameMapPtr, k)
		}

		stringSlicePool.Put(creationDateSlicePtr)
		stringSlicePool.Put(indexSortbycreationPtr)
		stringSlicePool.Put(indexSortbycreationAscPtr)
		stringSlicePool.Put(aggregateBytesPtr)
		stringSlicePool.Put(finalIndexListPtr)
		mapPool.Put(indexSizemapPtr)
		mapPool.Put(creationdateNameMapPtr)
	}()

	// 解引用指針
	creationDateSlice := *creationDateSlicePtr
	indexSortbycreation := *indexSortbycreationPtr
	indexSortbycreationAsc := *indexSortbycreationAscPtr
	aggregate_bytes := *aggregateBytesPtr
	finalIndexList := *finalIndexListPtr
	indexSizemap := *indexSizemapPtr
	creationdate_NameMap := *creationdateNameMapPtr

	var indicesinfo CatIndice

	// 🔒 P0-001 修復：禁止無 pattern filter 的 space filter
	// 原因：會選中所有 index（包括 .kibana, .security 等系統 index），非常危險
	if len(patternlist) < 1 {
		global.Logger.Errorw("🚨 CRITICAL: space filter without pattern filter is FORBIDDEN",
			"reason", "Would affect ALL indices including system indices (.kibana, .security, etc.)",
			"risk", "May cause permanent data loss and ES cluster failure",
			"suggestion", "Add a pattern filter to specify target indices",
			"action", "space_filter_blocked")
		return []string{} // 返回空列表，拒絕執行
	} else {
		chunks := chunkSlice(patternlist, 10)
		for _, chunk := range chunks {
			indicesinfo1 := metadataProvider.CatIndicesWithPattern(chunk)
			indicesinfo = append(indicesinfo, indicesinfo1...)
		}
	}

	// ✅ 收集索引信息到 map 和 slice
	for data := range indicesinfo {
		indexSizemap[indicesinfo[data].Index] = indicesinfo[data].StoreSize
		creationdate_NameMap[indicesinfo[data].CreationDate] = indicesinfo[data].Index
		creationDateSlice = append(creationDateSlice, indicesinfo[data].CreationDate)
	}

	if creationDateSlice == nil {
		// 如果撈不到 index 則返回一個空的list
		finalIndexList = append(finalIndexList, "")
	} else {
		// 按 index 的 create_date 排序
		sort.Strings(creationDateSlice)

		// 把 index_name 塞到 slice 中
		for date := range creationDateSlice {
			indexSortbycreation = append(indexSortbycreation, creationdate_NameMap[creationDateSlice[date]])
		}

		/// 倒序 - 從 newest create 的 index 開始加總 disk_space
		for index := range indexSortbycreation {
			name := indexSortbycreation[len(indexSortbycreation)-index-1]
			indexSortbycreationAsc = append(indexSortbycreationAsc, name)
		}

		total := 0

		for bytes := range indexSortbycreationAsc {
			var bytesnum int
			if indexSizemap[indexSortbycreationAsc[bytes]] == "" {
				bytesnum = 0
			} else {
				bytesint, err := strconv.Atoi(indexSizemap[indexSortbycreationAsc[bytes]])
				if err != nil {
					global.Logger.Error(err.Error())
					return
				}
				bytesnum = bytesint
			}

			// 加總 index storage
			total += bytesnum
			if total > disk_space*1024*1024 {
				break
			}
			aggregate_bytes = append(aggregate_bytes, indexSortbycreationAsc[bytes])
		}

		_, removed := Diff(indexSortbycreationAsc, aggregate_bytes)
		finalIndexList = removed
	}

	// ✅ 創建返回值的副本（避免返回池中的切片）
	result := make([]string, len(finalIndexList))
	copy(result, finalIndexList)
	return result
}

func FilterType_waterLevel(patternlist []string, upper_limit int, lower_limit int) (indiceslist []string) {
	// ✅ HIGH-008 修復: 從池中獲取切片和 map，減少內存分配
	creationDateSlicePtr := stringSlicePool.Get().(*[]string)
	aggregateBytesPtr := stringSlicePool.Get().(*[]string)
	indexSortbycreationPtr := stringSlicePool.Get().(*[]string)

	indexSizemapPtr := mapPool.Get().(*map[string]string)
	creationdateNameMapPtr := mapPool.Get().(*map[string]string)

	// ✅ 確保歸還給池
	defer func() {
		// 清空切片但保留容量
		*creationDateSlicePtr = (*creationDateSlicePtr)[:0]
		*aggregateBytesPtr = (*aggregateBytesPtr)[:0]
		*indexSortbycreationPtr = (*indexSortbycreationPtr)[:0]

		// ✅ 明確清理 map 後歸還
		for k := range *indexSizemapPtr {
			delete(*indexSizemapPtr, k)
		}
		for k := range *creationdateNameMapPtr {
			delete(*creationdateNameMapPtr, k)
		}

		stringSlicePool.Put(creationDateSlicePtr)
		stringSlicePool.Put(aggregateBytesPtr)
		stringSlicePool.Put(indexSortbycreationPtr)
		mapPool.Put(indexSizemapPtr)
		mapPool.Put(creationdateNameMapPtr)
	}()

	// 解引用指針
	creationDateSlice := *creationDateSlicePtr
	aggregate_bytes := *aggregateBytesPtr
	indexSortbycreation := *indexSortbycreationPtr
	indexSizemap := *indexSizemapPtr
	creationdate_NameMap := *creationdateNameMapPtr

	var indicesinfo CatIndice

	// 🔒 P0-001 修復：禁止無 pattern filter 的 water_level filter
	// 原因：會選中所有 index（包括 .kibana, .security 等系統 index），非常危險
	if len(patternlist) < 1 {
		global.Logger.Errorw("🚨 CRITICAL: water_level filter without pattern filter is FORBIDDEN",
			"reason", "Would affect ALL indices including system indices (.kibana, .security, etc.)",
			"risk", "May cause permanent data loss and ES cluster failure",
			"suggestion", "Add a pattern filter to specify target indices",
			"action", "water_level_filter_blocked")
		return []string{} // 返回空列表，拒絕執行
	} else {
		// 分批處理避免 HTTP URL 過長 (4096 bytes 限制)
		chunks := chunkSlice(patternlist, 10)
		for _, chunk := range chunks {
			indicesinfo1 := metadataProvider.CatIndicesWithPattern(chunk)
			indicesinfo = append(indicesinfo, indicesinfo1...)
		}
	}

	/// 統計各個 Node 的 Average Water Level
	nodesinfo := metadataProvider.CatNodes()
	water_level := 0.00
	AllDiskTotal := 0.00
	for _, data := range nodesinfo {
		DiskUsedPercent, err := strconv.ParseFloat(data.DiskUsedPercent, 32)
		if err != nil {
			global.Logger.Error(err.Error())
			return
		}
		DiskTotalstr := strings.TrimSuffix(data.DiskTotal, "gb")
		DiskTotal, err := strconv.ParseFloat(DiskTotalstr, 32)
		if err != nil {
			global.Logger.Error(err.Error())
			return
		}
		AllDiskTotal += DiskTotal
		water_level += DiskUsedPercent
	}

	AnerageLevel := water_level / float64(len(nodesinfo))
	msg := fmt.Sprintf("Average Water Level: %f", AnerageLevel)
	global.Logger.Infow(msg)

	var diskKbToClean float64

	// 觸發 upper_limit 才進行動作
	if AnerageLevel >= float64(upper_limit) {
		diskToCleanPercentage := float64(upper_limit) - float64(lower_limit)
		diskToClean := AllDiskTotal * (diskToCleanPercentage / 100)
		diskKbToClean = diskToClean * 1024 * 1024

		global.Logger.Infow(fmt.Sprintf("Estimated to Release %f GB of Disk Space", diskToClean), "type", "Details")

		// ✅ 收集索引信息到 map 和 slice
		for _, data := range indicesinfo {
			indexSizemap[data.Index] = data.StoreSize
			creationdate_NameMap[data.CreationDate] = data.Index
			creationDateSlice = append(creationDateSlice, data.CreationDate)
		}

		if creationDateSlice != nil {
			// 按 index 的 create_date 排序
			sort.Strings(creationDateSlice)

			// 把 index_name 塞到 slice 中
			for date := range creationDateSlice {
				indexSortbycreation = append(indexSortbycreation, creationdate_NameMap[creationDateSlice[date]])
			}

			total := 0
			for _, data := range indexSortbycreation {
				bytesnum, err := strconv.Atoi(indexSizemap[data])
				if err != nil {
					global.Logger.Error(err.Error())
					return
				}
				aggregate_bytes = append(aggregate_bytes, data)
				// 加總 index storage
				total += bytesnum
				if float64(total) > diskKbToClean {
					break
				}
			}
			global.Logger.Infow(fmt.Sprintf("Actually released %d kbs", total), "type", "Details")
		}
	} else {
		global.Logger.Infow("Water Level doesn't exceed upper limit")
	}

	// ✅ 創建返回值的副本（避免返回池中的切片）
	result := make([]string, len(aggregate_bytes))
	copy(result, aggregate_bytes)
	return result
}
