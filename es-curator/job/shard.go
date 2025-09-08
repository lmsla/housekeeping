package job

import (
	"encoding/json"
	// "fmt"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	// "log"
	// "strings"
	// "time"
	"context"
	"es-curator/global"
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
	res, err := req.Do(context.Background(), es)
	if err != nil {
		global.Logger.Error("CatShards request failed: ", err.Error())
	}
	ResponseStatusCheck(res, "CatShards")
	defer res.Body.Close()

	// fmt.Println(res)
	resString, _ := io.ReadAll(res.Body)
	var s CatShard
	json.Unmarshal(resString, &s)
	defer res.Body.Close()
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
	res, err := req.Do(context.Background(), es)
	if err != nil {
		global.Logger.Error("CatShardsbyNodeName request failed: ", err.Error())
	}

	ResponseStatusCheck(res, "CatShardsbyNodeName")
	defer res.Body.Close()

	resString, _ := io.ReadAll(res.Body)
	var s CatShard
	json.Unmarshal(resString, &s)
	defer res.Body.Close()

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
	res, err := req.Do(context.Background(), es)
	if err != nil {
		global.Logger.Error("CatIndicesbyNodeName request failed: ", err.Error())
	}
	defer res.Body.Close()
	// log.Println(res)
	// fmt.Println(res)
	resString, _ := io.ReadAll(res.Body)
	var s CatShard
	json.Unmarshal(resString, &s)

	ResponseStatusCheck(res, "CatIndicesbyNodeName")
	defer res.Body.Close()

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
