package job

import (
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"log"
	"fmt"
	// "net/http"
	// "time"
	"context"
	"io"
)



func Catnodes() {
	req := esapi.NodesInfoRequest{
		NodeID: []string{"Q4kU"},
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	log.Println(res)
	fmt.Println(res)
}

type CatNode []struct {
	IP              string `json:"ip"`
	NodeRole        string `json:"nodeRole"`
	Name            string `json:"name"`
	DiskTotal       string `json:"diskTotal"`
	DiskUsedPercent string `json:"diskUsedPercent"`
	Uptime          string `json:"uptime"`
	Version         string `json:"version"`
}

func CatNodes() CatNode {
	req := esapi.CatNodesRequest{
		// i:ip,r:nodeRole,
		H:      []string{"ip", "nodeRole", "name", "diskTotal", "diskUsedPercent", "uptime", "version"},
		Format: "json",
		// Bytes:  "kb",
		// FullID: true,
		Pretty: true,
		V:      newTrue(),
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	log.Println(res)
	// fmt.Println(res)
	resString, err := io.ReadAll(res.Body)
	var s CatNode
	json.Unmarshal(resString, &s)
	defer res.Body.Close()
	// fmt.Println(s)
	return s

}


func Nodetest() {
	nodeinfo := CatNodes()
	for data := range nodeinfo{
		fmt.Println(nodeinfo[data].DiskTotal,nodeinfo[data].DiskUsedPercent)

	}
}

func NodeStatus() {
	req := esapi.NodesStatsRequest{
		NodeID: []string{"Q4kUx91HR6mSOt0I6oydug"},
		// Metric: []string{"indices"},
		// IndexMetric: []string{"docs"},
		Pretty: true,
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	log.Println(res)
	fmt.Println(res)
}