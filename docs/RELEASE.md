# v1.1.6
1.新增action 之 index roll over 功能
2.修復 ES 失聯程式中斷退出問題，若 ES 暫時失聯，程式會繼續按時排程執行，直到 ES 恢復為止。

**重大安全更新** - 四層防禦架構實作

### 安全修復

1. **P0 級別修復** - 系統 Index 保護機制
   - 實作系統 index 黑名單函數 `isSystemIndex()`
   - Pattern filter 自動排除系統 index (`.kibana`, `.security`, `ilm-history-*`, etc.)
   - Delete action 執行前最終檢查 (handleDeleteIndices)
   - 測試覆蓋：13/13 通過 ✅

2. **P1 級別修復** - 配置驗證增強 (Fail Fast)
   - delete_indices action 必須配置 pattern filter
   - space/water_level filter 必須配合 pattern 使用
   - 禁止互斥 filter 組合（age+space, age+water_level, space+water_level）
   - 啟動時配置驗證（os.Exit(1) Fail Fast）
   - 測試覆蓋：8/8 通過 ✅

### 測試與文檔

3. **測試覆蓋率**
   - 新增 21 個測試案例（P0/P1 安全驗證 + 系統 index 識別）
   - 單元測試通過率：21/21 (100%) ✅
   - Code Review 評分：4.8/5.0 (A+)

4. **文檔重組**
   - 按照 IEEE Std 1016-2009 標準重組 SDD 文檔
   - 新增完整的測試規格文檔
   - 四層防禦架構詳細說明

### 相關文檔
- [安全審計報告](../claude/安全審計報告_Filter流程_2025-11-26.md)
- [Code Review 報告](../claude/Code_Review_Report_2025-11-27.md)
- [SDD 規格文檔](../claude/specs/)

# v1.1.5
1.新增 cluster health 及 action 寫回 es 做記錄及 dashboard 功能
2.加入 cat cluster health 間隔的控制選項

# v1.0.5 (2025.03.27)
1.alloction 執行前應先確認 index 當前位在哪個 node 再動作

# v1.0.3 (2025.02.04)
1.排除掉執行 action 時，出現重複 index 的導致執行失敗的狀況

# v1.0.2 (2024.01.07)
1.修復同時使用 pattern & space 時 index 篩選不到的問題

(cat indices with pattern 帶入的 index 數量過多，api 太長無法執行)

# v1.0.1 (2024.12.16)
1.修復filter by node role 部分 index 刪選錯誤

2.移除不必要的 print log