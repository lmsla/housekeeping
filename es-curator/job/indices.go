package job

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"

	"context"
	"encoding/json"
	"es-curator/global"
	"io"
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

		res, err := req.Do(context.Background(), es)
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
			ResponseStatusCheck(res,"ClusterHealth")
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
	res, err := req.Do(context.Background(), es)
	if err != nil {
		global.Logger.Error("CatIndices request failed: ", err.Error())
	}

	ResponseStatusCheck(res,"CatIndices")

	resString, _ := io.ReadAll(res.Body)
	var s CatIndice
	json.Unmarshal(resString, &s)
	defer res.Body.Close()
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
	res, err := req.Do(context.Background(), es)
	if err != nil {
		global.Logger.Error(err.Error())
	}
	resString, err := io.ReadAll(res.Body)
	if err != nil {
		global.Logger.Error("CatIndices_withPattern request failed: ", err.Error())
	}

	ResponseStatusCheck(res,"CatIndices_withPattern")

	var s CatIndice
	json.Unmarshal(resString, &s)
	defer res.Body.Close()
	return s
}

// ---------- open,close,delete,forcemerge ---------- //

func OpenIndices(Index []string) {
	req := esapi.IndicesOpenRequest{
		Index: Index,
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		global.Logger.Error("OpenIndices request failed: ", err.Error())
	}
	ResponseStatusCheck(res,"OpenIndices")
	defer res.Body.Close()
	log.Println(res)
}

func CloseIndices(Index []string) {
	req := esapi.IndicesCloseRequest{
		Index: Index,
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		global.Logger.Error("CloseIndices request failed: ", err.Error())
	}
	ResponseStatusCheck(res,"CloseIndices")
	defer res.Body.Close()
	log.Println(res)
}

func CreateIndex() {
	req := esapi.IndicesCreateRequest{
		Index: "logstash-bimap-test01",
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		global.Logger.Error("CreateIndices request failed: ", err.Error())
	}
	ResponseStatusCheck(res,"CreateIndices")
	defer res.Body.Close()
	log.Println(res)
}

func DeleteIndex(Index []string) {
	req := esapi.IndicesDeleteRequest{
		// Index: []string{"test_index"},
		Index: Index,
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		global.Logger.Error("DeleteIndex request failed: ", err.Error())
	}
	ResponseStatusCheck(res,"DeleteIndex")
	defer res.Body.Close()
	log.Println(res)
}

func IndicesStatus() {
	req := esapi.IndicesStatsRequest{
		Index: []string{"logstash-imperva-20221208"},
		// Metric: []string{"_all"},
		Pretty: true,
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		global.Logger.Error("IndicesStatus request failed: ", err.Error())
	}
	ResponseStatusCheck(res,"IndicesStatus")
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
	res, err := req.Do(context.Background(), es)
	if err != nil {
		global.Logger.Error("ForceMerge request failed: ", err.Error())
	}

	ResponseStatusCheck(res,"ForceMerge")

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
	res, err := req.Do(context.Background(), es)

	if err != nil {
		global.Logger.Error("Allocation request failed: ", err.Error())
		return
	}

	ResponseStatusCheck(res,"Allocation")

	defer res.Body.Close()

	log.Println(res)
}


func ResponseStatusCheck(res *esapi.Response,action string) {
		// 解析 ES API 回應，確保狀態碼是 2xx
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			resBody, _ := io.ReadAll(res.Body)
	
			var formattedError map[string]interface{}
			if err := json.Unmarshal(resBody, &formattedError); err == nil {
				// 解析 reason
				if errMap, ok := formattedError["error"].(map[string]interface{}); ok {
					if reason, ok := errMap["reason"].(string); ok {
							global.Logger.Error(fmt.Sprintf("%s API failed with status [%d], Reason: %s", action,res.StatusCode,reason))
					}
				}
			} else {
				global.Logger.Error(fmt.Sprintf("%s API failed with status [%d]: %s", action, res.StatusCode, string(resBody)))
			}
			return
		}
}