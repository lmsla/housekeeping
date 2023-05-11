package job

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"log"
	"strings"
	// "sync"
	// "net/http"
	// "time"
	// "crypto/tls"
	"es-curator/log_record"
	"context"
	"encoding/json"
	"io"
)

// type CatIndicesRequest struct {
// 	Index []string

// 	Bytes                   string
// 	ExpandWildcards         string
// 	Format                  string
// 	H                       []string
// 	Health                  string
// 	Help                    *bool
// 	IncludeUnloadedSegments *bool
// 	Local                   *bool
// 	MasterTimeout           time.Duration
// 	Pri                     *bool
// 	S                       []string
// 	Time                    string
// 	V                       *bool

// 	Pretty     bool
// 	Human      bool
// 	ErrorTrace bool
// 	FilterPath []string

// 	Header http.Header
// 	// contains filtered or unexported fields
// }

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
	req := esapi.ClusterHealthRequest{
		Index: []string{"*"},
	}

	res, err := req.Do(context.Background(), es)
	if err != nil {
		log_record.Logrecord("ERROR ","cluster health error" + err.Error())
		// panic(err)
	}
	// fmt.Println(res)

	defer res.Body.Close()
	// Parse the response
	resString, err := io.ReadAll(res.Body)
	var s CatClusterHealth
	json.Unmarshal(resString, &s)
	defer res.Body.Close()

	return s

}

func CatIndices() CatIndice {
	req := esapi.CatIndicesRequest{
		ExpandWildcards: "open,closed",
		Format:          "json",
		Bytes:           "kb",
		H:               []string{"health", "status", "index", "uuid", "pri", "rep", "docs.count", "docs.deleted", "store.size", "pri.store.size", "creation.date"},
		V:               newTrue(),
		Pretty:          true,
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		log_record.Logrecord("ERROR ","cat index error" + err.Error())
		// panic(err)
	}
	// log.Println(res)
	resString, err := io.ReadAll(res.Body)
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
		log_record.Logrecord("ERROR ","open index error" + err.Error())
		// panic(err)
	}

	defer res.Body.Close()
	log.Println(res)
}

func CloseIndices(Index []string) {
	req := esapi.IndicesCloseRequest{
		Index: Index,
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		log_record.Logrecord("ERROR ","close index error" + err.Error())
		// panic(err)
	}

	defer res.Body.Close()
	log.Println(res)
}

func CreateIndex() {
	req := esapi.IndicesCreateRequest{
		Index: "logstash-bimap-test01",
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		log_record.Logrecord("ERROR ","create index error" + err.Error())
		// panic(err)
	}
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
		log_record.Logrecord("ERROR ","delete index error" + err.Error())
		// panic(err)
	}
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
		log_record.Logrecord("ERROR ","cat index status error" + err.Error())
		// panic(err)
	}

	defer res.Body.Close()
	log.Println(res)
	fmt.Println(res)
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
		log_record.Logrecord("ERROR ","forcemerge error" + err.Error())
		// panic(err)
	}

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
		log_record.Logrecord("ERROR ","allocation error" + err.Error())
		// panic(err)
	}

	defer res.Body.Close()
	log.Println(res)
}
