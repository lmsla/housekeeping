package job

import (
	"es-curator/global"
	"es-curator/metrics"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"

	// "github.com/elastic/go-elasticsearch/v8/esapi"
	"crypto/tls"
	"net/http"
)


// 已廢棄：改用 global.Elasticsearch
// var es *elasticsearch.Client

func SetElkClient() error {
    startTime := time.Now()
    
    cfg := elasticsearch.Config{
        Addresses: global.EnvConfig.ES.URL,
        Username:  global.EnvConfig.ES.SourceAccount,
        Password:  global.EnvConfig.ES.SourcePassword,
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        },
    }

    var err error
    global.Elasticsearch, err = elasticsearch.NewClient(cfg)
    if err != nil {
        // 記錄連線失敗指標
        responseTime := time.Since(startTime)
        if metrics.GlobalMetrics != nil {
            metrics.GlobalMetrics.RecordESHealth(false, responseTime, "unknown")
        }
        return fmt.Errorf("ES 客戶端初始化失敗: %w", err)
    }

    res, err := global.Elasticsearch.Info()
    responseTime := time.Since(startTime)
    
    if err != nil {
        // 記錄連線失敗指標
        if metrics.GlobalMetrics != nil {
            metrics.GlobalMetrics.RecordESHealth(false, responseTime, "unreachable")
        }
        return fmt.Errorf("無法連線到 Elasticsearch: %w", err)
    }
    defer res.Body.Close()

    if res.IsError() {
        // 記錄連線失敗指標
        if metrics.GlobalMetrics != nil {
            metrics.GlobalMetrics.RecordESHealth(false, responseTime, "error")
        }
        return fmt.Errorf("elasticsearch 返回錯誤狀態: %s", res.String())
    }

    // 記錄連線成功指標
    if metrics.GlobalMetrics != nil {
        metrics.GlobalMetrics.RecordESHealth(true, responseTime, "green")
    }
    
    global.Stderr_logger.Info("成功連線到 Elasticsearch")
    return nil
}



// var es *elasticsearch.Client



