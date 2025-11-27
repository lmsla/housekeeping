package global

import (
	"es-curator/structs"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

var (
	EnvConfig        *structs.EnviromentModel
	Elasticsearch    *elasticsearch.Client
	// Action        *structs.Action
	ActionStruct     *structs.ActionStruct
	Logger           *zap.SugaredLogger
	Detail_Logger    *zap.SugaredLogger
	Stderr_logger    *logrus.Logger
	// MaxShardsPerNode 存儲集群的 cluster.max_shards_per_node 設定值
	// 預設 1000，程式啟動時從 ES API 動態查詢
	MaxShardsPerNode int = 1000
)
