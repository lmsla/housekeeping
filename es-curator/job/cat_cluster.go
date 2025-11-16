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
		// Panic Recovery: 保護監控 goroutine 不會因為 runtime panic 而終止
		defer func() {
			if r := recover(); r != nil {
				global.Logger.Errorw("❌ CatCluster goroutine panic recovered",
					"panic", r,
					"action", "goroutine_restarted")
				// 注意：這裡不 return，讓函數正常結束
				// 如果需要自動重啟，應在外層函數中實現
			}
		}()

		// 確保健康檢查間隔為正數，預設 60 秒
		interval := global.EnvConfig.Log.HealthCheckInterval
		if interval <= 0 {
			interval = 60
			global.Logger.Warn("Health check interval is invalid, using default 60 seconds")
		}

		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()

		global.Logger.Info("CatCluster monitoring started with shard tracking")

		for {
			select {
			case <-ctx.Done():
				global.Logger.Info("CatCluster monitoring stopped gracefully")
				return
			case <-ticker.C:
				// 獲取集群健康狀態
				clusterHealth := ClusterHealth()

				// 獲取每個節點的 shard 分配資訊
				allocations := CatAllocationAPI()

				// 獲取節點詳細資訊
				nodes := CatNodes()

				// 建立節點名稱到詳細資訊的映射
				nodeMap := make(map[string]struct {
					IP              string
					NodeRole        string
					DiskUsedPercent string
					Uptime          string
					Version         string
				})

				for _, node := range nodes {
					nodeMap[node.Name] = struct {
						IP              string
						NodeRole        string
						DiskUsedPercent string
						Uptime          string
						Version         string
					}{
						IP:              node.IP,
						NodeRole:        node.NodeRole,
						DiskUsedPercent: node.DiskUsedPercent,
						Uptime:          node.Uptime,
						Version:         node.Version,
					}
				}

				// 遍歷每個節點的 allocation 資訊
				for _, allocation := range allocations {
					// 處理 UNASSIGNED（未分配的分片統計）
					if allocation.Node == "UNASSIGNED" {
						// 解析未分配的 shard 數量
						unassignedCount, err := strconv.Atoi(allocation.Shards)
						if err != nil {
							global.Logger.Error("Failed to parse unassigned shard count", "error", err.Error())
							continue
						}

						// 如果有未分配的分片，記錄警告
						if unassignedCount > 0 {
							global.Logger.Warnw("⚠️  Cluster has UNASSIGNED shards",
								"unassigned_count", unassignedCount,
								"cluster_status", clusterHealth.Status)

							// 寫入 ES 記錄未分配分片資訊
							unassignedData := map[string]interface{}{
								"@timestamp":                time.Now().Format(time.RFC3339),
								"type":                      "unassigned_shards",
								"unassigned_shard_count":    unassignedCount,
								"cluster_status":            clusterHealth.Status,
								"cluster_unassigned_shards": clusterHealth.UnassignedShards,
							}
							ClusterLogToES(unassignedData)
						}
						continue
					}

					nodeInfo, exists := nodeMap[allocation.Node]
					if !exists {
						global.Logger.Warn("Node not found in CatNodes", "node", allocation.Node)
						continue
					}

					// 解析 shard 數量
					shardCount, err := strconv.Atoi(allocation.Shards)
					if err != nil {
						global.Logger.Error("Failed to parse shard count", "node", allocation.Node, "error", err.Error())
						continue
					}

					// 解析磁碟使用率
					diskUsedPercent, err := strconv.ParseFloat(nodeInfo.DiskUsedPercent, 64)
					if err != nil {
						global.Logger.Error("Failed to parse disk usage percent", "node", allocation.Node, "error", err.Error())
						continue
					}

					// 解析磁碟容量數據（從 string 轉為 float64）
					diskTotalGB, err := strconv.ParseFloat(allocation.DiskTotal, 64)
					if err != nil {
						global.Logger.Error("Failed to parse disk total", "node", allocation.Node, "value", allocation.DiskTotal, "error", err.Error())
						diskTotalGB = 0
					}

					diskUsedGB, err := strconv.ParseFloat(allocation.DiskUsed, 64)
					if err != nil {
						global.Logger.Error("Failed to parse disk used", "node", allocation.Node, "value", allocation.DiskUsed, "error", err.Error())
						diskUsedGB = 0
					}

					diskAvailableGB, err := strconv.ParseFloat(allocation.DiskAvail, 64)
					if err != nil {
						global.Logger.Error("Failed to parse disk available", "node", allocation.Node, "value", allocation.DiskAvail, "error", err.Error())
						diskAvailableGB = 0
					}

					diskIndicesGB, err := strconv.ParseFloat(allocation.DiskIndices, 64)
					if err != nil {
						global.Logger.Error("Failed to parse disk indices", "node", allocation.Node, "value", allocation.DiskIndices, "error", err.Error())
						diskIndicesGB = 0
					}

					// 計算 shard 使用率（預設上限 1000）
					maxShardsPerNode := 1000
					shardUsagePercent := float64(shardCount) / float64(maxShardsPerNode) * 100
					shardRemaining := maxShardsPerNode - shardCount

					// 判斷告警級別
					shardAlertLevel := "normal"
					if shardCount >= 950 {
						shardAlertLevel = "critical"
					} else if shardCount >= 800 {
						shardAlertLevel = "warning"
					}

					data := map[string]interface{}{
						"@timestamp": time.Now().Format(time.RFC3339),

						// 集群級別指標
						"cluster_status":                clusterHealth.Status,
						"cluster_unassigned_shards":     clusterHealth.UnassignedShards,
						"cluster_relocating_shards":     clusterHealth.RelocatingShards,
						"cluster_initializing_shards":   clusterHealth.InitializingShards,
						"cluster_active_shards_percent": clusterHealth.ActiveShardsPercentAsNumber,
						"cluster_number_of_nodes":       clusterHealth.NumberOfNodes,
						"cluster_number_of_data_nodes":  clusterHealth.NumberOfDataNodes,

						// 節點基本資訊
						"node_name":    allocation.Node,
						"node_ip":      nodeInfo.IP,
						"node_role":    nodeInfo.NodeRole,
						"node_uptime":  nodeInfo.Uptime,
						"node_version": nodeInfo.Version,

						// Shard 相關指標（核心監控）- 全部使用數值型別
						"shard_count":         shardCount,         // int
						"shard_max_per_node":  maxShardsPerNode,  // int
						"shard_usage_percent": shardUsagePercent, // float64
						"shard_remaining":     shardRemaining,    // int
						"shard_alert_level":   shardAlertLevel,   // string (normal/warning/critical)

						// 磁碟相關（使用數值型別，單位 GB）
						"disk_used_percent": diskUsedPercent, // float64
						"disk_total_gb":     diskTotalGB,     // float64
						"disk_used_gb":      diskUsedGB,      // float64
						"disk_available_gb": diskAvailableGB, // float64
						"disk_indices_gb":   diskIndicesGB,   // float64
					}

					// 寫入 ES
					ClusterLogToES(data)

					// 記錄告警日誌
					if shardAlertLevel == "critical" {
						global.Logger.Errorw("🔴 Node shard usage CRITICAL",
							"node", allocation.Node,
							"shard_count", shardCount,
							"usage_percent", shardUsagePercent,
							"remaining", shardRemaining)
					} else if shardAlertLevel == "warning" {
						global.Logger.Warnw("⚠️  Node shard usage WARNING",
							"node", allocation.Node,
							"shard_count", shardCount,
							"usage_percent", shardUsagePercent,
							"remaining", shardRemaining)
					}
				}
			}
		}
	}()
}
