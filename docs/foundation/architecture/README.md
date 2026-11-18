# 現行架構與決策

先以[需求導航](layout.md#從需求定位修改範圍)找規格與實作；架構文件按問題查閱，不作為每個任務必讀清單。

## 現行結構與規則

| 文件 | 回答的問題 |
| --- | --- |
| [架構總覽](overview.md) | 目前服務如何執行、有哪些入口？ |
| [需求導航與目錄責任](layout.md) | 需求應修改哪個模組？ |
| [後端功能邊界](backend-boundaries.md) | 核心、adapter 與組裝如何分工？ |
| [依賴與契約歸屬](dependencies.md) | 技術版本、共用規則與 import 邊界在哪裡？ |
| [開發原則](principles.md) | 如何選擇簡單且符合現行邊界的做法？ |
| [業務操作模型](business-operation.md) | 如何描述授權、規則與操作結果？ |
| [ADR](decisions/README.md) | 哪些重大決策已接受或被取代？ |

## 任務準備與按需參考

- [依任務選讀](../../README.md#按任務開始)；完成條件見[驗收矩陣](../../agents/acceptance.md#變更驗收矩陣)，原始交付順序見[架構基線](decisions/0002-baseline.md)。缺產品決策時先查當前功能規格。
- [實作契約與待決策](implementation-contract.md)：保留早期切片與決策閘門；個別功能以最新功能契約為準，不將舊待決策表當作目前完成狀態。
- [未來規劃與提案](../../planning/README.md)：平台擴充、目標業務模型與舊交付階段已移出現行架構目錄。

## 接手時的證據與用語

- 具體功能 happy path／拒絕範例與 JSON fixtures，放測試 testdata，避免在多頁重複維護。
- 一張需求 ID → Use Case → 檔案／函數 → 測試的追溯表；功能確定後填真實路徑，不預填不存在證據。
- 用詞表：actor=經驗證操作者；target=發送目標；accepted=API 接受；completed=本地處理完成，均不等於已讀。
- 各錯誤的重試／HTTP／觀測映射；未知不得默認可重試。
- 版本與來源：工具在 mise、SDK 在 go.mod；文件記查核來源，不另設版本真相。
- 圖要標 execution/import/data 關係；一圖一問題；規劃圖不可混進 Graphify 現況圖冒充已實作。

## 歷史查核來源（2026-11-07）

Context7：`/websites/go_dev_doc`（MixedCaps／建構函數）、`/line/line-openapi`（事件 source／mode、retry key、Reply）、`/websites/cloud_google_firestore_native`（交易重試、資料刪除），2026-11-07。
Context7 回傳片段不是完整官方 reference；未回傳的 token 期限、資格與 client-specific 行為必須再核對 endpoint 文件與實測。
Graphify AST 只描述可擷取程式關係，Markdown 規劃用 Mermaid 與連結；Repomix 用來提供可裁剪上下文，不證明應用驗收。

## 相鄰契約

[功能規格](../../features/README.md) · [HTTP](../api/README.md) · [Session](../identity/README.md) · [資料](../persistence/README.md) · [前端](../frontend/README.md) · [驗收矩陣](../../agents/acceptance.md)

[本目錄指引](AGENTS.md) · [父層指引](../../AGENTS.md) · [文件總入口](../../README.md)

[現行實作圖索引](../../diagrams.md#現行實作) · [未來設計圖](../../diagrams.md#未來設計)。
