package main

import (
	"housekeeping/internal/global"
	"housekeeping/internal/job"
	"housekeeping/internal/log_record"
	"housekeeping/internal/utils"
	"testing"
)

func TestMain(t *testing.T) {
	if err := utils.LoadEnvironment(); err != nil {
		t.Skipf("skip startup integration test: %v", err)
	}

	if global.EnvConfig == nil {
		t.Skip("skip startup integration test: environment config is not initialized")
	}

	//// init logger
	log_record.InitLogger()
	log_record.InitDetailLogger()
	log_record.InitStderrLogger()

	if err := utils.ValidateConfig(); err != nil {
		t.Skipf("skip startup integration test: invalid config: %v", err)
	}

	// 初始化 Elasticsearch 客戶端
	if err := job.SetElkClient(); err != nil {
		t.Skipf("skip startup integration test: elasticsearch unavailable: %v", err)
	}

	nodes := job.CatNodes()
	if len(nodes) == 0 {
		t.Fatal("expected CatNodes to return at least one node")
	}
}
