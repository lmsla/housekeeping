package job

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"

	// "log"
	"strings"
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

// CatAllocation 返回每個節點的 shard 分配資訊
type CatAllocation []struct {
	Node        string `json:"node"`         // 節點名稱
	Shards      string `json:"shards"`       // Shard 總數（關鍵指標）
	DiskIndices string `json:"disk.indices"` // 索引佔用磁碟
	DiskUsed    string `json:"disk.used"`    // 已使用磁碟
	DiskAvail   string `json:"disk.avail"`   // 可用磁碟
	DiskTotal   string `json:"disk.total"`   // 總磁碟
	DiskPercent string `json:"disk.percent"` // 磁碟使用率
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, global.Elasticsearch)
	if err != nil {
		global.Logger.Error("CatNodes request failed: ", err.Error())
		return CatNode{}
	}

	ResponseStatusCheck(res, "CatNodes")

	defer res.Body.Close()

	resString, err := io.ReadAll(res.Body)
	if err != nil {
		global.Logger.Error("Failed to read response body in CatNodes", "error", err)
		return CatNode{}
	}

	var s CatNode
	if err := json.Unmarshal(resString, &s); err != nil {
		global.Logger.Error("Failed to unmarshal JSON in CatNodes", "error", err)
		return CatNode{}
	}
	return s
}

func NodeStatus() {
	req := esapi.NodesStatsRequest{
		NodeID: []string{"Q4kUx91HR6mSOt0I6oydug"},
		// Metric: []string{"indices"},
		// IndexMetric: []string{"docs"},
		Pretty: true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, global.Elasticsearch)
	if err != nil {
		global.Logger.Error("NodeStatus request failed: ", err.Error())
		return
	}
	ResponseStatusCheck(res, "NodeStatus")
	defer res.Body.Close()
}

// 判定 nodeRole 相對應的 nodeName
func NodeRoleDetermination(role string) (nodeName []string) {
	nodeinfo := CatNodes()
	for _, data := range nodeinfo {
		if strings.Contains(data.NodeRole, role) {
			nodeName = append(nodeName, data.Name)

		}

	}
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := req.Do(ctx, global.Elasticsearch)
	if err != nil {
		global.Logger.Error("Catnodes request failed: ", err.Error())
		return
	}
	ResponseStatusCheck(res, "CatNodes")
	defer res.Body.Close()
	fmt.Println("res", res)
}

// CatAllocationAPI 獲取每個節點的 shard 分配統計資訊
func CatAllocationAPI() CatAllocation {
	req := esapi.CatAllocationRequest{
		H:      []string{"node", "shards", "disk.indices", "disk.used", "disk.avail", "disk.total", "disk.percent"},
		Format: "json",
		Bytes:  "gb",
		Pretty: true,
		V:      newTrue(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	res, err := req.Do(ctx, global.Elasticsearch)
	if err != nil {
		global.Logger.Error("CatAllocation request failed: ", err.Error())
		return CatAllocation{}
	}
	defer res.Body.Close()

	ResponseStatusCheck(res, "CatAllocation")

	resString, err := io.ReadAll(res.Body)
	if err != nil {
		global.Logger.Error("Failed to read response body in CatAllocation", "error", err)
		return CatAllocation{}
	}

	var s CatAllocation
	if err := json.Unmarshal(resString, &s); err != nil {
		global.Logger.Error("Failed to unmarshal JSON in CatAllocation", "error", err)
		return CatAllocation{}
	}

	return s
}
