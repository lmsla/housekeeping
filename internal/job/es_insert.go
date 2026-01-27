package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"housekeeping/internal/global"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func LogToES(uuid, individual_Msg, mode, action, Pri, Rep, DocCount, DocsDeleted, StoreSize, PriStoreSize, CreationDate, index string) {

	pri, _ := strconv.Atoi(Pri)
	rep, _ := strconv.Atoi(Rep)
	docCount, _ := strconv.Atoi(DocCount)
	docsDeleted, _ := strconv.Atoi(DocsDeleted)
	storeSize, _ := strconv.Atoi(StoreSize)
	priStoreSize, _ := strconv.Atoi(PriStoreSize)

	// 從 index 字串中取出第一個 "-數字" 前的內容
	// 例如: "logs-app-2024.01.01" → "logs-app"
	//       "metrics-001" → "metrics"
	//       "logstash-20241126" → "logstash"
	var indexPrefix string
	re := regexp.MustCompile(`-\d`)
	if loc := re.FindStringIndex(index); loc != nil {
		indexPrefix = index[:loc[0]]
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
		Body: &buf,
		// Refresh: "true", // 移除強制刷新以提升效能，ES 將使用預設的自動刷新機制（約1秒）
	}

	// execute request
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, global.Elasticsearch)
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
		Body: &buf,
		// Refresh: "true", // 移除強制刷新以提升效能，ES 將使用預設的自動刷新機制（約1秒）
	}

	// execute request
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, global.Elasticsearch)
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
		Body: &buf,
		// Refresh: "true", // 移除強制刷新以提升效能，ES 將使用預設的自動刷新機制（約1秒）
	}

	// execute request
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, global.Elasticsearch)
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