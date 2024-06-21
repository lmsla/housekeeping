package global

import (
	"es-curator/structs"
	"github.com/elastic/go-elasticsearch/v8"
	"go.uber.org/zap"
	// "go.uber.org/zap/zapcore"
)

var (
	EnvConfig     *structs.EnviromentModel
	Elasticsearch *elasticsearch.Client
	// Action        *structs.Action
	ActionStruct  *structs.ActionStruct
	Logger        *zap.SugaredLogger
	Detail_Logger *zap.SugaredLogger
)
