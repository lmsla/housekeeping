# BiMAP Housekeeping Documentation

此資料夾包含 BiMAP Housekeeping 服務的所有技術文件。

## 📁 文件結構

### 核心文件
- [`ARCHITECTURE.md`](./ARCHITECTURE.md) - 服務架構文件，描述簡化後的系統結構
- [`ROLLOVER_GUIDE.md`](./ROLLOVER_GUIDE.md) - Rollover 功能使用指南，索引滾動更新完整說明
- [`OPTIMIZATION_TODO.md`](./OPTIMIZATION_TODO.md) - 優化待辦清單，追蹤改善項目進度

### 配置與部署
- [`ESRoleSetting.md`](./ESRoleSetting.md) - Elasticsearch 節點角色設定說明
- [`ubuntu Service 啟用步驟.md`](./ubuntu%20Service%20啟用步驟.md) - Ubuntu 系統服務部署指南

### 版本與發布
- [`RELEASE.md`](./RELEASE.md) - 版本發布說明
- [`UNIVERSAL_MONITORING_SPEC.md`](./UNIVERSAL_MONITORING_SPEC.md) - Universal Monitoring 系統規格 (已從主服務剝離)

## 🏗️ 架構概覽

BiMAP Housekeeping 是專注於 Elasticsearch 索引生命週期管理的輕量級服務：

```
功能模組:
├── 索引管理 (刪除、開啟、關閉)
├── 索引分配 (Hot/Warm/Cold 層級)
├── 強制合併 (Segment 優化)
├── 智能過濾 (年齡、模式、大小、節點角色)
└── 本地指標收集 (操作統計、性能監控)
```

## 🚀 快速開始

1. **查看架構**: 閱讀 [`ARCHITECTURE.md`](./ARCHITECTURE.md) 了解系統設計
2. **配置服務**: 參考 [`ESRoleSetting.md`](./ESRoleSetting.md) 設定 ES 環境
3. **部署服務**: 依照 [`ubuntu Service 啟用步驟.md`](./ubuntu%20Service%20啟用步驟.md) 部署
4. **優化追蹤**: 查看 [`OPTIMIZATION_TODO.md`](./OPTIMIZATION_TODO.md) 了解改善計畫

## 📋 當前狀態

- **版本**: v1.1.5
- **狀態**: 穩定運行
- **重要變更**: Universal Monitoring System 已剝離，專注核心功能
- **完成率**: 31.6% 優化項目已完成

## 🔧 維護與支援

如需技術支援或建議改善，請：
1. 查看相關文件
2. 檢查待辦清單是否已涵蓋
3. 聯繫 BiMAP 開發團隊

---
*文件庫最後更新: 2025-01-11*