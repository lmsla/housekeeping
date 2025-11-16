package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"
	// "github.com/elastic/go-elasticsearch/v8"
	"es-curator/global"
	"strconv"
	"strings"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func LogToES(uuid, individual_Msg, mode, action, Pri, Rep, DocCount, DocsDeleted, StoreSize, PriStoreSize, CreationDate, index string) {

	pri, _ := strconv.Atoi(Pri)
	rep, _ := strconv.Atoi(Rep)
	docCount, _ := strconv.Atoi(DocCount)
	docsDeleted, _ := strconv.Atoi(DocsDeleted)
	storeSize, _ := strconv.Atoi(StoreSize)
	priStoreSize, _ := strconv.Atoi(PriStoreSize)
	// 從 index 字串中取出 "-20" 前的內容

	var indexPrefix string
	if idx := strings.Index(index, "-20"); idx != -1 {
		indexPrefix = index[:idx]
	} else {
		indexPrefix = index
	}


	data := map[string]interface{}{
		"@timestamp":   time.Now().Format(time.RFC3339),
		"msg":          individual_Msg,
		"mode":         mode,
		"type":         "Detail",
		"index":        index,
		"action":       action,
		"Pri":          pri,
		"Rep":          rep,
		"DocCount":     docCount,
		"StoreSize":    storeSize*1024,
		"PriStoreSize": priStoreSize*1024,
		"CreationDate": CreationDate,
		"DocsDeleted":  docsDeleted,
		"uuid":         uuid,
		"logName":      indexPrefix,
	}

	// convert data to JSON
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		global.Logger.Error("Error encoding data for LogToES", "error", err, "index", index)
		return
	}

	// build index request
	req := esapi.IndexRequest{
		Index: fmt.Sprintf("housekeeping_detail-%s", time.Now().Format("200601")), // 生成带日期的索引名称
		// DocumentID: "1",          // 可選，設置文檔 ID
		Body:    &buf,
		Refresh: "true", // 刷新索引，使數據立即可用
	}

	// execute request
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, es)
	if err != nil {
		global.Logger.Error("Error executing ES request in LogToES", "error", err, "index", index)
		return
	}
	defer res.Body.Close()

	// print request result
	if res.IsError() {
		global.Logger.Error("Error response from ES in LogToES", "response", res.String(), "index", index)
	} else {
		var response map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			global.Logger.Error("Error parsing the response body in LogToES", "error", err, "index", index)
		}
	}
	// return res.Body

}


func ProceduresLogToES(uuid,mode,description, action string) {

	// pri, _ := strconv.Atoi(Pri)
	// rep, _ := strconv.Atoi(Rep)
	// docCount, _ := strconv.Atoi(DocCount)
	// docsDeleted, _ := strconv.Atoi(DocsDeleted)
	// storeSize, _ := strconv.Atoi(StoreSize)
	// priStoreSize, _ := strconv.Atoi(PriStoreSize)

	data := map[string]interface{}{
		"@timestamp":   time.Now().Format(time.RFC3339),
		"msg":          description,
		"mode":         mode,
		"type":         "Procedures",
		"action":       action,
		"uuid":         uuid,
	}

	// convert data to JSON
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		global.Logger.Error("Error encoding data for ProceduresLogToES", "error", err, "action", action)
		return
	}

	// build index request
	req := esapi.IndexRequest{
		Index: fmt.Sprintf("housekeeping_detail-%s", time.Now().Format("200601")), // 生成带日期的索引名称
		// DocumentID: "1",          // 可選，設置文檔 ID
		Body:    &buf,
		Refresh: "true", // 刷新索引，使數據立即可用
	}

	// execute request
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, es)
	if err != nil {
		global.Logger.Error("Error executing ES request in ProceduresLogToES", "error", err, "action", action)
		return
	}
	defer res.Body.Close()

	// print request result
	if res.IsError() {
		global.Logger.Error("Error response from ES in ProceduresLogToES", "response", res.String(), "action", action)
	} else {
		var response map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			global.Logger.Error("Error parsing the response body in ProceduresLogToES", "error", err, "action", action)
		}
	}
	// return res.Body

}



func ClusterLogToES(data map[string]interface{}) {

	// convert data to JSON
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		global.Logger.Error("Error encoding data for ClusterLogToES", "error", err)
		return
	}

	// build index request
	req := esapi.IndexRequest{
		Index: fmt.Sprintf("housekeeping_cluster_health-%s", time.Now().Format("200601")), // 生成带日期的索引名称
		// DocumentID: "1",          // 可選，設置文檔 ID
		Body:    &buf,
		Refresh: "true", // 刷新索引，使數據立即可用
	}

	// execute request
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, es)
	if err != nil {
		global.Logger.Error("Error executing ES request in ClusterLogToES", "error", err)
		return
	}
	defer res.Body.Close()

	// print request result
	if res.IsError() {
		global.Logger.Error("Error response from ES in ClusterLogToES", "response", res.String())
	} else {
		var response map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			global.Logger.Error("Error parsing the response body in ClusterLogToES", "error", err)
		}
	}
	// return res.Body

}