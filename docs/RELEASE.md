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