# Rollover Action 使用指南

## 概述

Rollover 功能允許 BiMAP Housekeeping 自動管理索引別名的滾動更新。當索引達到指定條件時（大小、文檔數或時間），系統會自動創建新的索引並更新別名指向。

## 使用場景

- **時間序列數據管理**: 日誌索引自動滾動 (如每日、每週滾動)
- **大小控制**: 防止索引過大影響性能
- **文檔數限制**: 控制單一索引的文檔數量
- **索引生命週期管理**: 配合其他 action 實現完整的 ILM 策略

## 配置參數

### 必需參數

- `rollover_alias`: 要執行 rollover 的別名名稱

### 觸發條件 (至少需要一個)

- `max_size`: 最大索引大小 (如 "50gb", "1tb")
- `max_docs`: 最大文檔數 (整數)
- `max_age`: 最大索引年齡 (如 "30d", "1w", "24h")

### 可選參數

- `new_index_name`: 新索引名稱 (留空使用預設命名規則)

## 配置範例

### 基本 rollover 配置

```yaml
actions:
  - action: rollover
    description: "Daily log rollover"
    options:
      rollover_alias: "logs-active"
      max_age: "1d"
      max_size: "50gb"
      max_docs: 100000000
```

### 僅基於大小的 rollover

```yaml
actions:
  - action: rollover
    description: "Size-based rollover"
    options:
      rollover_alias: "metrics-write"
      max_size: "10gb"
```

### 僅基於時間的 rollover

```yaml
actions:
  - action: rollover
    description: "Weekly rollover"
    options:
      rollover_alias: "weekly-reports"
      max_age: "7d"
```

### 完整的 ILM 策略範例

```yaml
actions:
  # 1. 執行 rollover
  - action: rollover
    description: "Daily logs rollover"
    options:
      rollover_alias: "logs-write"
      max_age: "1d"
      max_size: "5gb"
      delay: 300  # 等待 5 分鐘

  # 2. 將舊索引分配到 warm 節點
  - action: allocation
    description: "Move old indices to warm tier"
    filters:
      - filtertype: age
        source: creation_date
        direction: older
        unit: days
        unit_count: 1
    options:
      allocation_type: require
      key: _tier_preference
      value: data_warm
      delay: 60

  # 3. 刪除超過 30 天的索引
  - action: delete_indices
    description: "Delete indices older than 30 days"
    filters:
      - filtertype: age
        source: creation_date
        direction: older
        unit: days
        unit_count: 30
```

## 前置條件

### 1. 別名設置

在使用 rollover 之前，需要確保：

1. **別名已存在**: 目標別名必須已經存在於 ES 中
2. **初始索引**: 別名至少指向一個可寫入的索引
3. **索引命名規則**: 索引名稱應該遵循 rollover 命名慣例 (如 logs-000001, logs-000002)

### 2. 別名創建範例

```bash
# 創建初始索引
PUT /logs-000001
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 1
  }
}

# 創建別名並設定為可寫
POST /_aliases
{
  "actions": [
    {
      "add": {
        "index": "logs-000001",
        "alias": "logs-write",
        "is_write_index": true
      }
    }
  ]
}
```

## 工作原理

1. **條件檢查**: 系統檢查當前寫入索引是否滿足任何觸發條件
2. **索引創建**: 如果條件滿足，創建新索引 (遵循命名規則遞增)
3. **別名更新**: 自動更新別名指向新索引，並設定為可寫入
4. **結果記錄**: 記錄 rollover 操作結果和新舊索引信息

## 監控與日誌

### 成功 rollover 日誌範例

```json
{
  "level": "info",
  "msg": "Rollover successful",
  "alias": "logs-write",
  "old_index": "logs-000001",
  "new_index": "logs-000002",
  "logType": "Procedures"
}
```

### 條件未滿足日誌範例

```json
{
  "level": "info", 
  "msg": "Rollover conditions not met",
  "alias": "logs-write",
  "logType": "Procedures"
}
```

### 指標收集

Rollover 操作會自動記錄到本地指標系統：

- 總操作次數
- 成功/失敗統計
- 平均執行時間
- 每次操作的詳細日誌

## 最佳實踐

### 1. 合理設置觸發條件

```yaml
# ✅ 好的做法：多條件保護
options:
  rollover_alias: "logs-write"
  max_age: "1d"        # 確保至少每天滾動
  max_size: "5gb"      # 防止索引過大
  max_docs: 50000000   # 限制文檔數量
```

```yaml
# ❌ 避免：僅單一條件
options:
  rollover_alias: "logs-write"
  max_size: "50gb"     # 可能導致索引很久才滾動
```

### 2. 配合延遲使用

```yaml
actions:
  - action: rollover
    options:
      rollover_alias: "logs-write"
      max_age: "1d"
      delay: 300  # rollover 後等待 5 分鐘，確保索引穩定
```

### 3. 測試模式驗證

在生產環境使用前，建議：

1. 設定 `test_mode: true`
2. 觀察日誌輸出
3. 確認觸發條件符合預期
4. 驗證別名和索引狀態

## 常見問題

### Q1: Rollover 失敗，顯示 "alias not found"

**解決方案**: 確認別名已存在且正確配置

```bash
# 檢查別名狀態
GET /_aliases/your-alias-name
```

### Q2: 新索引沒有創建

**解決方案**: 檢查觸發條件是否過於嚴格

```bash
# 檢查當前索引統計
GET /your-current-index/_stats
```

### Q3: 測試模式下看不到效果

**解決方案**: 測試模式只會記錄日誌，不會實際執行。查看測試模式專用日誌：

```json
{
  "mode": "TestMode",
  "msg": "Rollover alias 'logs-write' processed in test mode"
}
```

## 版本兼容性

- ✅ Elasticsearch 7.x+
- ✅ 支援 ILM 策略
- ✅ 相容現有 Curator 配置遷移

---

*文件版本: 1.0*  
*最後更新: 2025-01-11*  
*適用版本: BiMAP Housekeeping v1.1.5+*