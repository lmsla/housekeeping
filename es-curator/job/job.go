package job

import (
	// "github.com/elastic/go-elasticsearch/v7"
	"es-curator/global"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	// "encoding/json"
	"log"
	// "strings"
	"fmt"
	"net/http"
	"time"
	"crypto/tls"
	// "bytes"
	"context"
	// "io/ioutil"
	// "strconv"
	// "sync"
)

type IndicesCloseRequest struct {
	Index []string

	AllowNoIndices      *bool
	ExpandWildcards     string
	IgnoreUnavailable   *bool
	MasterTimeout       time.Duration
	Timeout             time.Duration
	WaitForActiveShards string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header
	// contains filtered or unexported fields
}

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

func Catnodes() {
	req := esapi.NodesInfoRequest{
		NodeID: []string{"Q4kU"},
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	log.Println(res)
	fmt.Println(res)
}

type CatNodesRequest struct {
	Bytes                   string
	Format                  string
	FullID                  *bool
	H                       []string
	Help                    *bool
	IncludeUnloadedSegments *bool
	Local                   *bool
	MasterTimeout           time.Duration
	S                       []string
	Time                    string
	V                       *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header
	// contains filtered or unexported fields
}

func CatNodes() {
	req := esapi.CatNodesRequest{
		H: []string{"r", "n", "du", "u"},
		// FullID: true,

	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	log.Println(res)
	fmt.Println(res)
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
		panic(err)
	}

	defer res.Body.Close()
	log.Println(res)
	fmt.Println(res)
}
