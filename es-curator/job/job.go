package job

import (
	"es-curator/global"
	"github.com/elastic/go-elasticsearch/v8"
	// "github.com/elastic/go-elasticsearch/v8/esapi"
	"crypto/tls"
	"net/http"

)

// type IndicesCloseRequest struct {
// 	Index []string

// 	AllowNoIndices      *bool
// 	ExpandWildcards     string
// 	IgnoreUnavailable   *bool
// 	MasterTimeout       time.Duration
// 	Timeout             time.Duration
// 	WaitForActiveShards string

// 	Pretty     bool
// 	Human      bool
// 	ErrorTrace bool
// 	FilterPath []string

// 	Header http.Header
// 	// contains filtered or unexported fields
// }

var es *elasticsearch.Client

func SetElkClient() {
	var err error
	cfg := elasticsearch.Config{
		Addresses: global.EnvConfig.ES.URL,
		Username:  global.EnvConfig.ES.SourceAccount,
		Password:  global.EnvConfig.ES.SourcePassword,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	es, err = elasticsearch.NewClient(cfg)
	if err != nil {
		panic(err) // 連線失敗
	}

	// log.SetFlags(0)

}


