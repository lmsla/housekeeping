# `es-curator` 技術分析文件

## 1. 總覽

`es-curator` 是一個使用 Go 語言開發的 Elasticsearch 索引管理工具。其核心功能類似於官方的 `Elastic Curator`，旨在根據使用者定義的規則，自動化地對 Elasticsearch (ES) 索引執行生命週期管理任務，例如刪除、搬遷 (allocation)、合併 (forcemerge) 和滾動 (rollover) 等。

此工具可以作為單次執行的腳本，也可以配置為常駐的 Cron Job，定期執行維護任務。

## 2. 核心組件與架構

應用程式主要由以下幾個部分組成：

| 檔案/目錄 | 主要職責 |
| :--- | :--- |
| `main.go` | **應用程式進入點**。負責初始化設定、日誌、ES 連線，並根據配置決定執行模式（單次或 Cron Job）。 |
| `setting.yml` | **應用程式設定檔**。定義 ES 連線資訊、執行模式（`test_mode`, `execute_cron`）和日誌設定。 |
| `config.yml` | **任務設定檔**。定義具體要執行的 `actions` (任務) 和 `filters` (過濾器)。 |
| `job/actions.go` | **任務調度核心**。包含主要業務邏輯 `Action_controll`，負責遍歷 `config.yml` 中的任務，應用過濾器，並執行相應操作。 |
| `job/filters.go` | **過濾器實作**。提供各種過濾索引的具體邏輯，例如按年齡 (`age`)、名稱 (`pattern`)、磁碟空間 (`space`) 等。 |
| `job/indices.go` | **ES API 封裝**。封裝了 `go-elasticsearch` 客戶端，提供與 ES 互動的底層函數，如 `CatIndices`, `DeleteIndex`, `Allocation` 等。 |
| `job/job.go` | **ES 客戶端建立**。負責初始化及建立與 ES 的連線。 |
| `utils/` | **工具函數**。包含載入設定檔、設定 Cron Job 等輔助功能。 |
| `metrics/` | **指標收集**。一個本地指標收集系統，用於追蹤操作計數、成功/失敗率、執行時間和資源使用情況。 |
| `log_record/` | **日誌管理**。初始化和設定應用程式的日誌系統。 |

## 3. 執行流程

應用程式的執行流程如下：

1.  **初始化 (`main.go`)**:
    *   載入 `setting.yml` 的環境配置。
    *   初始化 `zap` 日誌系統，可同時輸出到檔案和 ES。
    *   初始化本地指標收集器。
    *   使用 `job.SetElkClient()` 建立與 ES 的連線。

2.  **模式判斷 (`main.go`)**:
    *   檢查 `setting.yml` 中的 `execute_cron` 旗標。
    *   若為 `true`，則啟動 Cron Job，週期性地呼叫 `job.Action_controll()`。同時，設定訊號處理程序以實現優雅關閉。
    *   若為 `false`，則直接執行一次 `job.Action_controll()`，完成後即退出。

3.  **任務控制 (`job.Action_controll`)**:
    *   讀取 `config.yml` 中定義的 `actions` 列表。
    *   **循序**遍歷每一個 `action`。
    *   若 `action` 被禁用 (`disable_action: true`)，則跳過。
    *   為每個啟用的 `action` 產生一個唯一的執行 ID。

4.  **過濾與篩選 (`processAction`)**:
    *   處理 `action` 下定義的 `filters`。
    *   呼叫 `job/filters.go` 中的過濾函數（如 `FilterType_age`, `FilterType_pattern`），從 ES 取得符合條件的索引。
    *   將多個過濾器的結果進行交集或組合，產生最終要執行的索引列表。

5.  **執行操作 (`executeAction`)**:
    *   根據 `action` 的類型（如 `delete_indices`, `allocation`），呼叫對應的處理函數（如 `handleDeleteIndices`, `handleAllocation`）。
    *   **安全檢查**: 在執行前，會檢查 `setting.yml` 中的 `test_mode`。
        *   若為 `true`，僅記錄將要執行的操作和目標索引，**不會**進行任何實際變更。
        *   若為 `false`，則呼叫 `job/indices.go` 中封裝的 ES API 函數，執行實際操作。
    *   操作完成後，根據 `delay` 選項進行短暫休眠。

6.  **完成與指標 (`main.go`)**:
    *   單次任務執行完畢或收到關閉訊號時，呼叫 `printMetricsSummary()` 函數，將執行期間收集的各項指標（如操作成功/失敗次數、執行時間、資源消耗等）輸出到日誌中。

## 4. 設定詳解

### `setting.yml`

此檔案控制應用程式的整體行為。

-   `es`: ES 叢集的連線資訊。
-   `information.test_mode`: **極其重要的安全開關**。`true` 時，所有操作都只會模擬執行並產生日誌，不會對叢集造成任何實際影響。
-   `information.execute_cron`: `true` 會使程式作為守護行程執行，`false` 則為單次執行。
-   `log.toes`: `true` 會將程式自身的日誌寫入指定的 ES 叢集，方便集中管理。

### `config.yml`

此檔案定義具體的索引管理策略。

-   **actions**: 一個任務陣列，每個任務都是一個獨立的操作單元。
-   **filters**: 每個 `action` 下的過濾器，用於精確指定操作對象。
    -   `filtertype: pattern`: 透過前綴、後綴或正規表示式來匹配索引名稱。
    -   `filtertype: age`: 根據索引的 `creation_date` 進行篩選（比...舊、比...新、或在某個時間範圍內）。
    -   `filtertype: space`: 當索引總大小超過 `disk_space` 時，選出最舊的索引進行操作。
    -   `filtertype: node_role`: 根據索引分片所在的節點角色 (`data_hot`, `data_warm` 等) 進行篩選。
    -   `filtertype: water_level`: 根據叢集節點的平均磁碟使用率（水位）來觸發操作。

## 5. 技術觀察與建議

-   **優點**:
    -   **測試模式**: `test_mode` 是個非常出色的安全功能，讓使用者可以在部署新規則前充分驗證其行為。
    -   **日誌與指標**: 內建了詳細的結構化日誌和指標收集，極大地提高了系統的可觀測性和問題追蹤能力。
    -   **優雅關閉**: 支援 `SIGINT` 和 `SIGTERM` 訊號，能夠在關閉前完成正在執行的任務並輸出指標摘要。
    -   **程式碼結構**: 模組劃分清晰，職責分明，易於理解和擴充。

-   **可改進之處**:
    -   **併發執行**: 目前 `Action_controll` 是循序執行每個 `action`。如果多個 `action` 之間沒有依賴關係，可以考慮將它們改為併發執行，以提高在任務繁重時的處理效率。
    -   **啟動時設定驗證**: 程式目前僅對日誌路徑有簡單檢查。建議在啟動時增加對 `config.yml` 和 `setting.yml` 更全面的驗證邏輯，例如檢查 `filtertype` 是否合法、必要的選項是否存在等，以便及早發現設定錯誤。
    -   **擴充單元測試**: 雖然已有測試檔案，但建議為 `job/filters.go` 中的過濾器邏輯增加更多的單元測試。由於這些過濾器大多是處理資料的純函數，非常適合進行單元測試，可以確保其邏輯的正確性。
