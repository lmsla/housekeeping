package job

import (
	"encoding/json"
	"time"
	// "fmt"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	// "log"
	// "strings"
	"context"
	"housekeeping/internal/global"
	"io"
)

type CatShard []struct {
	Index        string `json:"index"`
	Shard        string `json:"shard"`
	Prirep       string `json:"prirep"`
	State        string `json:"state"`
	Ip           string `json:"ip"`
	Node         string `json:"node"`
	Store        string `json:"store"`
	Docs         string `json:"docs"`
	CreationDate string `json:"creation.date"`
}

func CatShards() CatShard {
	req := esapi.CatShardsRequest{
		// i:ip,r:nodeRole,
		// Index:  []string{"logstash-l7_network-20240515"},
		H:      []string{"index", "shard", "prirep", "state", "ip", "node", "store"},
		Format: "json",
		// Bytes:  "kb",
		// FullID: true,
		Pretty: true,
		V:      newTrue(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, global.Elasticsearch)
	if err != nil {
		global.Logger.Error("CatShards request failed: ", err.Error())
		return CatShard{}
	}
	ResponseStatusCheck(res, "CatShards")
	defer res.Body.Close()

	// fmt.Println(res)
	resString, err := io.ReadAll(res.Body)
	if err != nil {
		global.Logger.Error("Failed to read response body in CatShards", "error", err)
		return CatShard{}
	}

	var s CatShard
	if err := json.Unmarshal(resString, &s); err != nil {
		global.Logger.Error("Failed to unmarshal JSON in CatShards", "error", err)
		return CatShard{}
	}
	// fmt.Println("s",s)

	// fmt.Println(s)
	return s
}

func CatShardsbyNodeName(nodeName string) CatShard {
	req := esapi.CatShardsRequest{
		// i:ip,r:nodeRole,
		// Index:  []string{"logstash-l7_network-20240515"},
		H:      []string{"index", "shard", "prirep", "state", "ip", "node", "dataset.size","store"},
		Format: "json",
		Bytes:  "kb",
		// FullID: true,
		Pretty: true,
		V:      newTrue(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, global.Elasticsearch)
	if err != nil {
		global.Logger.Error("CatShardsbyNodeName request failed: ", err.Error())
		return CatShard{}
	}

	ResponseStatusCheck(res, "CatShardsbyNodeName")
	defer res.Body.Close()

	resString, err := io.ReadAll(res.Body)
	if err != nil {
		global.Logger.Error("Failed to read response body in CatShardsbyNodeName", "error", err)
		return CatShard{}
	}

	var s CatShard
	if err := json.Unmarshal(resString, &s); err != nil {
		global.Logger.Error("Failed to unmarshal JSON in CatShardsbyNodeName", "error", err)
		return CatShard{}
	}

	var nodeSelected CatShard
	for _, data := range s {
		if data.Node == nodeName {
			nodeSelected = append(nodeSelected, data)
		}

	}
	// fmt.Println("nodeSelected",nodeSelected)
	return nodeSelected
}

func CatIndicesbyNodeName(nodeName string) []string {
	var indices []string
	req := esapi.CatShardsRequest{
		// i:ip,r:nodeRole,
		// Index:  []string{"logstash-l7_network-20240515"},
		H:      []string{"index", "shard", "prirep", "state", "ip", "node", "store", "docs", "creation.date"},
		Format: "json",
		Bytes:  "kb",
		// FullID: true,
		Pretty: true,
		V:      newTrue(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, global.Elasticsearch)
	if err != nil {
		global.Logger.Error("CatIndicesbyNodeName request failed: ", err.Error())
		return []string{}
	}
	defer res.Body.Close()

	ResponseStatusCheck(res, "CatIndicesbyNodeName")

	// log.Println(res)
	// fmt.Println(res)
	resString, err := io.ReadAll(res.Body)
	if err != nil {
		global.Logger.Error("Failed to read response body in CatIndicesbyNodeName", "error", err)
		return []string{}
	}

	var s CatShard
	if err := json.Unmarshal(resString, &s); err != nil {
		global.Logger.Error("Failed to unmarshal JSON in CatIndicesbyNodeName", "error", err)
		return []string{}
	}

	var nodeSelected CatShard
	for _, data := range s {
		if data.Node == nodeName {
			nodeSelected = append(nodeSelected, data)
		}
	}
	for _, data := range nodeSelected {
		indices = append(indices, data.Index)

	}
	indices = RemoveDuplicates(indices)
	// fmt.Println("nodeSelected", nodeSelected)
	return indices
}
