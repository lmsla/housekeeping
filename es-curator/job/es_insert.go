package job

import (
	"bytes"
	"context"
	"encoding/json"

	"log"
	"time"

	// "github.com/elastic/go-elasticsearch/v8"
	"es-curator/global"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func CatIndices_withPattern1() {
	data := map[string]interface{}{
		"@timestamp": time.Now().Format(time.RFC3339),
		"user":      "johndoe",
		"message":   "Hello, Elasticsearch!sssss",
	}

	// 將資料轉換為 JSON
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		log.Fatalf("Error encoding data: %s", err)
	}

	// 構建索引請求
	req := esapi.IndexRequest{
		Index:      "test-index1", // 替換為你的索引名稱
		// DocumentID: "1",          // 可選，設置文檔 ID
		Body:       &buf,
		Refresh:    "true", // 刷新索引，使數據立即可用
	}

	// 執行請求
	res, err := req.Do(context.Background(), es)
	if err != nil {
		global.Logger.Error(err.Error())
	}
	defer res.Body.Close()

	// 打印請求結果
	if res.IsError() {
		log.Printf("Error response: %s", res.String())
	} else {
		var response map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			log.Printf("Document indexed successfully: %s", response["result"])
		}
	}
	// return res.Body

}
