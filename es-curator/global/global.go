package global

import (
	"es-curator/structs"
	"github.com/elastic/go-elasticsearch/v8"
)

var (
	EnvConfig     *structs.EnviromentModel
	Elasticsearch *elasticsearch.Client
	Action        *structs.Action
	ActionStruct  *structs.ActionStruct
)
