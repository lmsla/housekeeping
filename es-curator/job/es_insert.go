package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
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
		log.Fatalf("Error encoding data: %s", err)
	}

	// build index request
	req := esapi.IndexRequest{
		Index: fmt.Sprintf("housekeeping_detail-%s", time.Now().Format("200601")), // 生成带日期的索引名称
		// DocumentID: "1",          // 可選，設置文檔 ID
		Body:    &buf,
		Refresh: "true", // 刷新索引，使數據立即可用
	}

	// execute request
	res, err := req.Do(context.Background(), es)
	if err != nil {
		global.Logger.Error(err.Error())
	}
	defer res.Body.Close()

	// print request result
	if res.IsError() {
		log.Printf("Error response: %s", res.String())
	} else {
		var response map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
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
		log.Fatalf("Error encoding data: %s", err)
	}

	// build index request
	req := esapi.IndexRequest{
		Index: fmt.Sprintf("housekeeping_detail-%s", time.Now().Format("200601")), // 生成带日期的索引名称
		// DocumentID: "1",          // 可選，設置文檔 ID
		Body:    &buf,
		Refresh: "true", // 刷新索引，使數據立即可用
	}

	// execute request
	res, err := req.Do(context.Background(), es)
	if err != nil {
		global.Logger.Error(err.Error())
	}
	defer res.Body.Close()

	// print request result
	if res.IsError() {
		log.Printf("Error response: %s", res.String())
	} else {
		var response map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		}
	}
	// return res.Body

}



func ClusterLogToES(data map[string]interface{}) {

	// convert data to JSON
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		log.Fatalf("Error encoding data: %s", err)
	}

	// build index request
	req := esapi.IndexRequest{
		Index: fmt.Sprintf("housekeeping_cluster_health-%s", time.Now().Format("20060102")), // 生成带日期的索引名称
		// DocumentID: "1",          // 可選，設置文檔 ID
		Body:    &buf,
		Refresh: "true", // 刷新索引，使數據立即可用
	}

	// execute request
	res, err := req.Do(context.Background(), es)
	if err != nil {
		global.Logger.Error(err.Error())
	}
	defer res.Body.Close()

	// print request result
	if res.IsError() {
		log.Printf("Error response: %s", res.String())
	} else {
		var response map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		}
	}
	// return res.Body

}