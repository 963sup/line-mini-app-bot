# 規劃與提案

本目錄保存未來能力、目標模型與歷史平台交付設計。這些文件不表示功能已實作，也不構成目前任務的必要前置條件。

現行功能先讀[功能索引](../features/README.md)，實作位置見[需求導航](../foundation/architecture/layout.md#從需求定位修改範圍)，已接受決策見 [ADR](../foundation/architecture/decisions/README.md)。提案中已落地的部分以所連結的現行契約為準。

## 按問題查閱

| 文件 | 用途 |
| --- | --- |
| [未來平台目錄](future-platform.md) | 評估持久 Inbox、Outbox、Worker 與管理後台的責任配置 |
| [工程平台目標架構](line-engineering-target.md) | 查閱 LINE 平台與跨業務能力的提案模型 |
| [Repository、Issue 與 Discussion 模型](domain-model.md) | 比較後續資源模型與產品設計參考 |
| [平台交付階段](delivery.md) | 查閱舊平台切片與未來驗收方向；目前檢查依[驗收矩陣](../agents/acceptance.md) |

## 相鄰的設計參考

- [資料契約提案](../foundation/persistence/data-contracts.md)：尚未實作的技術基線。
- [可靠性設計參考](../foundation/reliability/README.md)：通用 Inbox／Outbox 與狀態機，與現行功能區分。

## 維護邊界

提案須標示狀態與現行契約來源。能力被接受後，將有效規則落在功能規格或 ADR，保留必要的提案背景與引用；不以提案直接修改公開契約、資料或部署。

[文件總入口](../README.md) · [文件維護指引](../AGENTS.md)
