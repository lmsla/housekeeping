package job

import (
	"context"
	"strconv"
	"time"
	"es-curator/global"
)

func CatCluster(ctx context.Context) {
	// 定時獲取 ES node 資訊並寫入 ES
	go func() {
		// 確保健康檢查間隔為正數，預設 60 秒
		interval := global.EnvConfig.Log.HealthCheckInterval
		if interval <= 0 {
			interval = 60
			global.Logger.Warn("Health check interval is invalid, using default 60 seconds")
		}
		
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()
		
		global.Logger.Info("CatCluster monitoring started")
		
		for {
			select {
			case <-ctx.Done():
				global.Logger.Info("CatCluster monitoring stopped gracefully")
				return
			case <-ticker.C:
				nodes := CatNodes()
				for _, node := range nodes {
					diskUsedPercent, err := strconv.ParseFloat(node.DiskUsedPercent, 64)
					if err != nil {
						global.Logger.Error("Failed to parse disk usage percent: ", err.Error())
						continue
					}

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
		}
	}()
}
