package job

import (
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"log"
	"strings"
	// "time"
	"context"
	"es-curator/global"
	"io"
)

func Catnodes() {
	req := esapi.NodesInfoRequest{
		NodeID: []string{"Q4kU"},
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		// log_record.Logrecord("ERROR ", "cat nodes error"+err.Error())
		global.Logger.Error(err.Error())
		// panic(err)
	}

	defer res.Body.Close()
	log.Println(res)
	fmt.Println("res",res)
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
		// log_record.Logrecord("ERROR ", "cat nodes error"+err.Error())
		global.Logger.Error(err.Error())
		// panic(err)
	}
	defer res.Body.Close()
	// log.Println(res)
	// fmt.Println(res)
	resString, _ := io.ReadAll(res.Body)
	var s CatNode
	json.Unmarshal(resString, &s)
	defer res.Body.Close()
	// fmt.Println("s",s)
	fmt.Println(s)
	return s
}

func CatNodesWithNodeName(nodeName string) CatNode {
	nodesinfo := CatNodes()

	var nodesInfoWithName CatNode

	for _, node := range nodesinfo {

		if node.Name == nodeName {
			nodesInfoWithName = append(nodesInfoWithName, node)
		}

	}

	return nodesInfoWithName
}

// 判定 nodeRole 相對應的 nodeName
func NodeRoleDetermination(role string) (nodeName []string) {
	nodeinfo := CatNodes()
	for _, data := range nodeinfo {
		if strings.Contains(data.NodeRole, role) {
			// fmt.Println("nodeName",data.Name,"nodeRole",data.NodeRole)
			nodeName = append(nodeName, data.Name)

		}

	}
	// fmt.Println("yes", nodeName)
	return nodeName
}

func Nodetest() {
	nodeinfo := CatNodes()
	for data := range nodeinfo {
		fmt.Println(nodeinfo[data].DiskTotal, nodeinfo[data].NodeRole)

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
		// log_record.Logrecord("ERROR ", "node status error"+err.Error())
		global.Logger.Error(err.Error())
	}

	defer res.Body.Close()
	// log.Println(res)
	// fmt.Println(res)
}
