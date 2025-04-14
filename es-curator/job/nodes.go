package job

import (
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/esapi"

	// "log"
	"strings"
	// "time"
	"context"
	"es-curator/global"
	"io"
)

type CatNode []struct {
	IP              string `json:"ip"`
	NodeRole        string `json:"nodeRole"`
	Name            string `json:"name"`
	DiskTotal       string `json:"diskTotal"`
	DiskUsedPercent string `json:"diskUsedPercent"`
	DiskUsed        string `json:"diskUsed"`
	Uptime          string `json:"uptime"`
	Version         string `json:"version"`
	DiskAvailable   string `json:"diskAvail"`
}

func CatNodes() CatNode {
	req := esapi.CatNodesRequest{
		// i:ip,r:nodeRole,
		H:      []string{"ip", "nodeRole", "name", "diskTotal", "diskUsedPercent", "diskUsed", "uptime", "version", "diskAvail"},
		Format: "json",
		// Bytes:  "kb",
		// FullID: true,
		Pretty: true,
		V:      newTrue(),
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		global.Logger.Error("CatNodes request failed: ", err.Error())
		// panic(err)
	}

	ResponseStatusCheck(res, "CatNodes")

	defer res.Body.Close()

	resString, _ := io.ReadAll(res.Body)
	var s CatNode
	json.Unmarshal(resString, &s)
	defer res.Body.Close()
	return s
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
		global.Logger.Error("NodeStatus request failed: ", err.Error())
	}
	ResponseStatusCheck(res, "NodeStatus")
	defer res.Body.Close()
	// log.Println(res)
	// fmt.Println(res)
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

// CatNodes by nodeName
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

func Nodetest() {
	nodeinfo := CatNodes()
	for data := range nodeinfo {
		fmt.Println(nodeinfo[data].DiskTotal, nodeinfo[data].NodeRole)

	}
}

// for test
func Catnodes() {
	req := esapi.NodesInfoRequest{
		NodeID: []string{"Q4kU"},
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		global.Logger.Error("Catnodes request failed: ", err.Error())
		// panic(err)
	}
	ResponseStatusCheck(res, "CatNodes")
	defer res.Body.Close()
	// log.Println(res)
	fmt.Println("res", res)
}
