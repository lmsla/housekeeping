package job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"es-curator/global"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func newTrue() *bool {
	b := true
	return &b
}

type CatIndice []struct {
	Health       string `json:"health"`
	Status       string `json:"status"`
	Index        string `json:"index"`
	UUID         string `json:"uuid"`
	Pri          string `json:"pri"`
	Rep          string `json:"rep"`
	DocsCount    string `json:"docs.count"`
	DocsDeleted  string `json:"docs.deleted"`
	StoreSize    string `json:"store.size"`
	PriStoreSize string `json:"pri.store.size"`
	CreationDate string `json:"creation.date"`
}

type IndicesInfo struct {
	Health       string `json:"health"`
	Status       string `json:"status"`
	Index        string `json:"index"`
	UUID         string `json:"uuid"`
	Pri          string `json:"pri"`
	Rep          string `json:"rep"`
	DocsCount    string `json:"docs.count"`
	DocsDeleted  string `json:"docs.deleted"`
	StoreSize    string `json:"store.size"`
	PriStoreSize string `json:"pri.store.size"`
	CreationDate string `json:"creation.date"`
	Shard        string `json:"shard"`
	// CreationDate	time.Time
}

type CatClusterHealth struct {
	ClusterName                 string  `json:"cluster_name"`
	Status                      string  `json:"status"`
	TimedOut                    bool    `json:"timed_out"`
	NumberOfNodes               int     `json:"number_of_nodes"`
	NumberOfDataNodes           int     `json:"number_of_data_nodes"`
	ActivePrimaryShards         int     `json:"active_primary_shards"`
	ActiveShards                int     `json:"active_shards"`
	RelocatingShards            int     `json:"relocating_shards"`
	InitializingShards          int     `json:"initializing_shards"`
	UnassignedShards            int     `json:"unassigned_shards"`
	DelayedUnassignedShards     int     `json:"delayed_unassigned_shards"`
	NumberOfPendingTasks        int     `json:"number_of_pending_tasks"`
	NumberOfInFlightFetch       int     `json:"number_of_in_flight_fetch"`
	TaskMaxWaitingInQueueMillis int     `json:"task_max_waiting_in_queue_millis"`
	ActiveShardsPercentAsNumber float64 `json:"active_shards_percent_as_number"`
}

func ClusterHealth() CatClusterHealth {
	return ClusterHealthWithRetry(3)
}

func ClusterHealthWithRetry(maxRetries int) CatClusterHealth {
	var lastErr error

	for retry := 0; retry <= maxRetries; retry++ {
		if retry > 0 {
			// 指數退避，延遲 2^retry 秒
			backoffDelay := time.Duration(1<<uint(retry-1)) * time.Second
			global.Logger.Infow("Retrying ClusterHealth request", "attempt", retry+1, "delay", backoffDelay.String())
			time.Sleep(backoffDelay)
		}

		req := esapi.ClusterHealthRequest{
			Index: []string{"*"},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		res, err := req.Do(ctx, global.Elasticsearch)
		cancel()
		if err != nil {
			lastErr = err
			global.Logger.Error("ClusterHealth request failed: ", err.Error())
			if retry < maxRetries {
				continue
			}
			return CatClusterHealth{}
		}

		// 檢查 HTTP 狀態碼
		if res.StatusCode >= 400 {
			defer res.Body.Close()
			if retry < maxRetries {
				continue
			}
			ResponseStatusCheck(res, "ClusterHealth")
			return CatClusterHealth{}
		}

		defer res.Body.Close()
		// Parse the response
		resString, err := io.ReadAll(res.Body)
		if err != nil {
			lastErr = err
			global.Logger.Error("ClusterHealth read response body failed: ", err.Error())
			if retry < maxRetries {
				continue
			}
			return CatClusterHealth{}
		}

		var s CatClusterHealth
		if err := json.Unmarshal(resString, &s); err != nil {
			lastErr = err
			global.Logger.Error("ClusterHealth unmarshal failed: ", err.Error())
			if retry < maxRetries {
				continue
			}
			return CatClusterHealth{}
		}

		// 成功時記錄日誌
		if retry > 0 {
			global.Logger.Infow("ClusterHealth request succeeded after retries", "attempt", retry+1)
		}
		return s
	}

	// 所有重試都失敗
	global.Logger.Errorw("ClusterHealth failed after all retries", "maxRetries", maxRetries, "lastError", lastErr)
	return CatClusterHealth{}
}

func CatIndices() CatIndice {
	req := esapi.CatIndicesRequest{
		ExpandWildcards: "open,closed",
		// ExpandWildcards: "hidden",
		Format: "json",
		Bytes:  "kb",
		H:      []string{"health", "status", "index", "uuid", "pri", "rep", "docs.count", "docs.deleted", "store.size", "pri.store.size", "creation.date"},
		V:      newTrue(),
		Pretty: true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, global.Elasticsearch)
	if err != nil {
		global.Logger.Error("CatIndices request failed: ", err.Error())
		return CatIndice{}
	}
	defer res.Body.Close()

	ResponseStatusCheck(res, "CatIndices")

	resString, err := io.ReadAll(res.Body)
	if err != nil {
		global.Logger.Error("Failed to read response body in CatIndices", "error", err)
		return CatIndice{}
	}

	var s CatIndice
	if err := json.Unmarshal(resString, &s); err != nil {
		global.Logger.Error("Failed to unmarshal JSON in CatIndices", "error", err)
		return CatIndice{}
	}
	return s
}

func CatIndices_withPattern(index_list []string) CatIndice {
	req := esapi.CatIndicesRequest{
		Index:           index_list,
		ExpandWildcards: "open",
		Format:          "json",
		Bytes:           "kb",
		H:               []string{"health", "status", "index", "uuid", "pri", "rep", "docs.count", "docs.deleted", "store.size", "pri.store.size", "creation.date"},
		V:               newTrue(),
		Pretty:          true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, global.Elasticsearch)
	if err != nil {
		global.Logger.Error("CatIndices_withPattern request failed: ", err.Error())
		return CatIndice{}
	}
	defer res.Body.Close()

	ResponseStatusCheck(res, "CatIndices_withPattern")

	resString, err := io.ReadAll(res.Body)
	if err != nil {
		global.Logger.Error("Failed to read response body in CatIndices_withPattern", "error", err)
		return CatIndice{}
	}

	var s CatIndice
	if err := json.Unmarshal(resString, &s); err != nil {
		global.Logger.Error("Failed to unmarshal JSON in CatIndices_withPattern", "error", err)
		return CatIndice{}
	}
	return s
}

// ---------- open,close,delete,forcemerge ---------- //

func OpenIndices(Index []string) {
	req := esapi.IndicesOpenRequest{
		Index: Index,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, global.Elasticsearch)
	if err != nil {
		global.Logger.Error("OpenIndices request failed: ", err.Error())
		return
	}
	ResponseStatusCheck(res, "OpenIndices")
	defer res.Body.Close()
	log.Println(res)
}

func CloseIndices(Index []string) {
	req := esapi.IndicesCloseRequest{
		Index: Index,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, global.Elasticsearch)
	if err != nil {
		global.Logger.Error("CloseIndices request failed: ", err.Error())
		return
	}
	ResponseStatusCheck(res, "CloseIndices")
	defer res.Body.Close()
	log.Println(res)
}

func CreateIndex() {
	req := esapi.IndicesCreateRequest{
		Index: "logstash-bimap-test01",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, global.Elasticsearch)
	if err != nil {
		global.Logger.Error("CreateIndices request failed: ", err.Error())
		return
	}
	ResponseStatusCheck(res, "CreateIndices")
	defer res.Body.Close()
	log.Println(res)
}

func DeleteIndex(Index []string) {
	req := esapi.IndicesDeleteRequest{
		// Index: []string{"test_index"},
		Index: Index,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, global.Elasticsearch)
	if err != nil {
		global.Logger.Error("DeleteIndex request failed: ", err.Error())
		return
	}
	ResponseStatusCheck(res, "DeleteIndex")
	defer res.Body.Close()
	log.Println(res)
}

func IndicesStatus() {
	req := esapi.IndicesStatsRequest{
		Index: []string{"logstash-imperva-20221208"},
		// Metric: []string{"_all"},
		Pretty: true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, global.Elasticsearch)
	if err != nil {
		global.Logger.Error("IndicesStatus request failed: ", err.Error())
		return
	}
	ResponseStatusCheck(res, "IndicesStatus")
	defer res.Body.Close()
	log.Println(res)
}

func ForceMerge(Index []string, MaxNumSegments int) {
	a := true
	req := esapi.IndicesForcemergeRequest{
		Index:             Index,
		MaxNumSegments:    &MaxNumSegments,
		WaitForCompletion: &a,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, global.Elasticsearch)
	if err != nil {
		global.Logger.Error("ForceMerge request failed: ", err.Error())
		return
	}

	ResponseStatusCheck(res, "ForceMerge")

	defer res.Body.Close()
	log.Println(res)
}

func Allocation(Index []string, AllocationType string, key string, value string) {

	body := fmt.Sprintf(`{"settings":{"index.routing.allocation":{"%s":{"%s":"%s"}}}}`, AllocationType, key, value)
	// a := `{"settings":{"index.routing.allocation":{"include":{"_tier_preference":"data_hot"},"exclude":{"_tier_preference":""},"require":{"_tier_preference":""}}}}`
	req := esapi.IndicesPutSettingsRequest{
		Index: Index,
		Body:  strings.NewReader(body),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, global.Elasticsearch)

	if err != nil {
		global.Logger.Error("Allocation request failed: ", err.Error())
		return
	}

	ResponseStatusCheck(res, "Allocation")

	defer res.Body.Close()

	log.Println(res)
}

func ResponseStatusCheck(res *esapi.Response, action string) {
	// 檢查 response 是否為 nil（防止 panic）
	if res == nil {
		global.Logger.Warn(fmt.Sprintf("%s: Response is nil, skipping status check", action))
		return
	}

	// 解析 ES API 回應，確保狀態碼是 2xx
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		resBody, _ := io.ReadAll(res.Body)

		var formattedError map[string]interface{}
		if err := json.Unmarshal(resBody, &formattedError); err == nil {
			// 解析 reason
			if errMap, ok := formattedError["error"].(map[string]interface{}); ok {
				if reason, ok := errMap["reason"].(string); ok {
					global.Logger.Error(fmt.Sprintf("%s API failed with status [%d], Reason: %s", action, res.StatusCode, reason))
				}
			}
		} else {
			global.Logger.Error(fmt.Sprintf("%s API failed with status [%d]: %s", action, res.StatusCode, string(resBody)))
		}
		return
	}
}

// Rollover 執行索引 rollover 操作
func Rollover(alias string, maxSize string, maxDocs int64, maxAge string, newIndexName string) error {
	// 構建 rollover 條件
	conditions := make(map[string]interface{})

	if maxSize != "" {
		conditions["max_size"] = maxSize
	}
	if maxDocs > 0 {
		conditions["max_docs"] = maxDocs
	}
	if maxAge != "" {
		conditions["max_age"] = maxAge
	}

	// 構建請求體
	rolloverBody := map[string]interface{}{
		"conditions": conditions,
	}

	// 如果指定了新索引名稱模式，添加到請求中
	if newIndexName != "" {
		// 可以根據需要添加索引設定或映射
		rolloverBody["settings"] = map[string]interface{}{
			"index.number_of_shards":   1,
			"index.number_of_replicas": 1,
		}
	}

	bodyJSON, err := json.Marshal(rolloverBody)
	if err != nil {
		global.Logger.Error("Failed to marshal rollover body: ", err.Error())
		return err
	}

	// 執行 rollover 請求
	req := esapi.IndicesRolloverRequest{
		Alias:  alias,
		Body:   strings.NewReader(string(bodyJSON)),
		Pretty: true,
	}

	// 如果有指定新索引名稱，加入請求中
	if newIndexName != "" {
		req.NewIndex = newIndexName
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, global.Elasticsearch)
	if err != nil {
		global.Logger.Error("Rollover request failed: ", err.Error())
		return err
	}
	defer res.Body.Close()

	// 讀取 body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		global.Logger.Error("Failed to read rollover response body: ", err.Error())
		return err
	}

	// 檢查回應狀態
	if res.IsError() {
		var errResponse map[string]interface{}
		if json.Unmarshal(body, &errResponse) == nil {
			if errMap, ok := errResponse["error"].(map[string]interface{}); ok {
				if reason, ok := errMap["reason"].(string); ok {
					errMsg := fmt.Sprintf("Rollover API failed with status [%d], Reason: %s", res.StatusCode, reason)
					global.Logger.Error(errMsg)
					return errors.New(errMsg)
				}
			}
		}
		errMsg := fmt.Sprintf("Rollover API failed with status [%d]: %s", res.StatusCode, string(body))
		global.Logger.Error(errMsg)
		return errors.New(errMsg)
	}

	// 解析回應以獲取詳細信息
	var rolloverResponse map[string]interface{}
	if err := json.Unmarshal(body, &rolloverResponse); err == nil {
		if rolledOver, ok := rolloverResponse["rolled_over"].(bool); ok && rolledOver {
			if oldIndex, ok := rolloverResponse["old_index"].(string); ok {
				if newIndex, ok := rolloverResponse["new_index"].(string); ok {
					global.Logger.Infow("Rollover successful",
						"alias", alias,
						"old_index", oldIndex,
						"new_index", newIndex,
						"logType", "Procedures")
				}
			}
		} else {
			global.Logger.Infow("Rollover conditions not met",
				"alias", alias,
				"logType", "Procedures")
		}
	}

	log.Println(res)
	return nil
}
