package job

import (
	// "log"
	"strconv"
	"time"
	"es-curator/global"
)

func CatCluster() {
	// 定時獲取 ES node 資訊並寫入 ES
	go func() {
		ticker := time.NewTicker(time.Duration(global.EnvConfig.Log.Health_check_interval) * time.Second)
		defer ticker.Stop()
		// []string{"ip", "nodeRole", "name", "diskTotal", "diskUsedPercent", "diskUsed", "uptime", "version"},
		for range ticker.C {
			nodes := CatNodes()
			for _, node := range nodes {
				diskUsedPercent, _ := strconv.ParseFloat(node.DiskUsedPercent, 64)

				data := map[string]interface{}{
					"@timestamp":      time.Now().Format(time.RFC3339),
					"ip":              node.IP,
					"nodeRole":        node.NodeRole,
					"name":            node.Name,
					"diskTotal":       node.DiskTotal,
					"diskUsedPercent": diskUsedPercent,
					"diskUsed":        node.DiskUsed,
					"uptime":          node.Uptime,
					"diskAvailable":   node.DiskAvailable,
				}
				// 寫入 ES
				ClusterLogToES(data)
			}
		}
	}()
}
