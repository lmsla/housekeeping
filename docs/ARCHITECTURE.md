# BiMAP Housekeeping Service Architecture

## Overview

BiMAP Housekeeping 是一個專注於 Elasticsearch 索引生命週期管理的輕量級服務。經過架構簡化後，移除了複雜的 Universal Monitoring System，專注於核心的索引管理功能。

## Current Architecture (Simplified)

### Core Modules

```
es-curator/
├── main.go                 # 服務入口點，初始化和協調
├── global/                 # 全域變數和配置
│   └── global.go
├── structs/                # 資料結構定義
│   └── env.go
├── utils/                  # 工具函數
│   ├── utils.go           # 配置載入和環境設置
│   └── crontab.go         # Cron 排程管理
├── log_record/            # 日誌記錄系統
├── metrics/               # 輕量級本地指標收集
│   └── metrics.go
└── job/                   # 核心業務邏輯
    ├── job.go             # ES 客戶端初始化
    ├── actions.go         # Action 控制和執行
    ├── tools.go           # ES 操作工具
    ├── filters.go         # 索引過濾邏輯
    ├── indices.go         # 索引管理操作
    └── [other job files]
```

### Key Components

#### 1. Service Entry (`main.go`)
- 環境配置載入
- 日誌系統初始化
- 本地指標收集器啟動
- ES 客戶端連線
- Cron 排程或單次執行模式
- 優雅關閉處理

#### 2. Configuration Management (`utils/`)
- YAML 配置檔案解析
- 環境變數支援
- Cron 表達式處理
- 配置驗證

#### 3. Action Engine (`job/actions.go`)
- Action 控制流程 (`Action_controll`)
- 過濾器處理和日誌記錄
- 統一的 Action 執行器 (`ActionExecutor`)
- 測試模式支援

#### 4. ES Operations (`job/`)
- **Index Management**: 刪除、關閉、開啟索引
- **Allocation**: 索引分配到指定節點角色
- **Force Merge**: 索引段合併
- **Filtering**: 基於年齡、模式、空間、節點角色的過濾
- **Health Monitoring**: 叢集健康狀態檢查

#### 5. Local Metrics (`metrics/`)
- 操作統計 (成功/失敗計數)
- 性能指標 (響應時間)
- 資源使用監控 (記憶體、Goroutine)
- ES 健康狀態追蹤

#### 6. Logging System (`log_record/`)
- 結構化日誌 (JSON 格式)
- 多層級日誌 (Info, Debug, Error)
- ES 日誌回寫支援
- 詳細操作記錄

## Supported Actions

| Action | Description | Key Features |
|--------|-------------|--------------|
| `delete_indices` | 刪除符合條件的索引 | 支援測試模式，安全檢查 |
| `allocation` | 索引分配到指定節點角色 | 支援 hot/warm/cold 層級 |
| `forcemerge` | 強制合併索引段 | 可配置合併段數量 |
| `close` | 關閉索引 | 節省資源，保留資料 |
| `open` | 開啟索引 | 恢復索引使用 |
| `rollover` | 索引別名滾動更新 | 基於大小/文檔數/時間自動創建新索引 |

## Filter Types

| Filter | Description | Usage |
|--------|-------------|--------|
| `age` | 基於索引年齡 | 清理舊索引 |
| `pattern` | 基於索引名稱模式 | 正規表達式匹配 |
| `space` | 基於索引大小 | 管理儲存空間 |
| `node_role` | 基於節點角色 | 配合 allocation |
| `water_level` | 基於磁碟水位 | 儲存容量管理 |

## Safety Features

### 1. Test Mode
- 所有 Action 支援測試模式
- 預覽操作結果，不實際執行
- 安全的配置驗證

### 2. Configuration Validation
- 啟動時配置完整性檢查
- 關鍵配置缺失時拒絕執行
- 測試模式狀態明確顯示

### 3. Error Handling
- Panic 恢復機制
- 詳細錯誤日誌記錄
- 操作失敗統計和報告

## Deployment Modes

### 1. Cron Mode (`ExecuteCron: true`)
- 背景服務模式
- 定期執行排程任務
- 支援優雅關閉

### 2. Single Run Mode (`ExecuteCron: false`)  
- 單次執行模式
- 立即執行所有 Action
- 執行完畢後輸出指標摘要

## Configuration

主要配置檔案：
- `config.yml`: ES 連線和 Action 定義
- `setting.yml`: 服務設定和日誌配置

## Recent Improvements

### Universal Monitoring Removal (v1.1.5)
- ✅ 移除 Universal Monitoring System 複雜依賴
- ✅ 簡化 go.mod，降低外部依賴
- ✅ 保留輕量級本地指標收集
- ✅ 專注核心 Housekeeping 功能
- ✅ 提升服務啟動速度和穩定性

### Code Quality Improvements
- ✅ 函數重構，降低複雜度
- ✅ 變數命名規範化
- ✅ 清理註解程式碼
- ✅ 統一錯誤處理機制

---

*架構文件版本: 1.1.5*  
*最後更新: 2025-01-11*  
*維護者: BiMAP 開發團隊*