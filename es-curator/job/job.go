package job

import (
	"es-curator/global"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"

	// "github.com/elastic/go-elasticsearch/v8/esapi"
	"crypto/tls"
	"net/http"
)


var es *elasticsearch.Client

func SetElkClient() error {
    cfg := elasticsearch.Config{
        Addresses: global.EnvConfig.ES.URL,
        Username:  global.EnvConfig.ES.SourceAccount,
        Password:  global.EnvConfig.ES.SourcePassword,
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        },
    }

    var err error
    es, err = elasticsearch.NewClient(cfg)
    if err != nil {
        // fmt.Println("ES 客戶端初始化失敗: ", err)
        return fmt.Errorf("ES 客戶端初始化失敗: %w", err)
    }

    res, err := es.Info()
    if err != nil {
        // fmt.Println("無法連線到 Elasticsearch: ", err)
        return fmt.Errorf("無法連線到 Elasticsearch: %w", err)
    }
    defer res.Body.Close()

    if res.IsError() {
        // fmt.Println("elasticsearch 返回錯誤狀態: ", res.String())
        return fmt.Errorf("elasticsearch 返回錯誤狀態: %s", res.String())
    }

	global.Stderr_logger.Info("成功連線到 Elasticsearch")
    return nil
}



// var es *elasticsearch.Client



