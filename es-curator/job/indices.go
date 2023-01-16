package job

import (
	// "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"log"

	// "strings"
	"fmt"
	"net/http"
	"time"
	// "crypto/tls"
	// "bytes"
	"context"
	"encoding/json"
	"io"
	// "time"
	// "strconv"
	// "sync"
)

func OpenIndices(Index []string) {
	req := esapi.IndicesOpenRequest{
		Index: Index,
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	log.Println(res)
}

func CloseIndices(Index []string) {
	req := esapi.IndicesCloseRequest{
		Index: Index,
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	log.Println(res)
}

func CreateIndex() {
	req := esapi.IndicesCreateRequest{
		Index: "prefixmore03-2023-01-05",
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	log.Println(res)
}

func DeleteIndex(Index []string) {
	req := esapi.IndicesDeleteRequest{
		// Index: []string{"test_index"},
		Index: Index,
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	log.Println(res)
}

func IndicesStatus() {
	req := esapi.IndicesStatsRequest{
		Index: []string{"logstash-imperva-20221208"},
		// Metric: []string{"_all"},
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

type CatIndicesRequest struct {
	Index []string

	Bytes                   string
	ExpandWildcards         string
	Format                  string
	H                       []string
	Health                  string
	Help                    *bool
	IncludeUnloadedSegments *bool
	Local                   *bool
	MasterTimeout           time.Duration
	Pri                     *bool
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

func newTrue() *bool {
	b := true
	return &b
}

type CatIndice []struct {
	Health       string `json:"health"`
	Status       string `json:"status"`
	Index        string `json:"index"`
	UUID         string `json:"uuid"`
	Pri          string `json:"pri"`
	Rep          string `json:"rep"`
	DocsCount    string `json:"docs.count"`
	DocsDeleted  string `json:"docs.deleted"`
	StoreSize    string `json:"store.size"`
	PriStoreSize string `json:"pri.store.size"`
	CreationDate string `json:"creation.date"`
	// CreationDate	time.Time
}

func CatIndices() CatIndice {
	req := esapi.CatIndicesRequest{
		ExpandWildcards: "open,closed",
		Format:          "json",
		Bytes:           "kb",
		H:               []string{"health", "status", "index", "uuid", "pri", "rep", "docs.count", "docs.deleted", "store.size", "pri.store.size", "creation.date"},
		V:               newTrue(),
		Pretty:          true,
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		panic(err)
	}
	// log.Println(res)
	resString, err := io.ReadAll(res.Body)
	var s CatIndice
	json.Unmarshal(resString, &s)
	defer res.Body.Close()

	return s
}
