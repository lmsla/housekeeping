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
)
