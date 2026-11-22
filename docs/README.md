# 專案文件入口

先依任務選一個權威入口，再查該頁所指的契約、程式與代表測試。

## 按任務開始

| 要做的事 | 先讀 | 再按需要查閱 |
| --- | --- | --- |
| 找修改位置 | [需求到實作導航](architecture/runtime/code-layout.md#從需求定位修改範圍) | 所屬功能規格與目錄 AGENTS.md |
| 修改業務或授權 | [功能索引](domains/README.md) | [操作模型](contracts/business-operations.md)、[契約歸屬](architecture/runtime/dependencies.md#契約與不變條件的維護位置) |
| 修改公開 API | [API 契約與生成流程](../api/README.md) | 對應功能與 [HTTP 契約](architecture/README.md#http-契約) |
| 劃分模型、聚合或交易邊界 | [DDD 問題索引](references/ddd/README.md#按問題查閱) | [現行後端邊界](architecture/runtime/backend-boundaries.md)與目標領域測試 |
| 修改自有文件或巢狀指引 | [文件歸屬](#文件歸屬與維護) | [文件維護與驗證](AGENTS.md#驗證) |
| 準備開發環境 | [初始化指南](development/environment/setup.md) | [工具與 MCP](development/tooling/README.md) |
| 判定修改是否完成 | [驗收矩陣](development/acceptance.md#變更驗收矩陣) | 對應功能及外部環境驗收 |
| 部署、發布或排錯 | [營運入口](architecture/README.md#部署與營運) | 相關設定與真實環境證據 |
| 查工具在 Codex 沒有曝光 | [MCP 載入檢查](development/tooling/mcp/README.md#project-loading-check) | [專案 Codex 設定](../.codex/README.md) |
| 評估未來能力 | [規劃與提案](proposals/README.md) | 已接受 ADR 與現行功能規格 |

## 日常開發看哪張圖

初次接手先看執行入口；修改功能時，依手上的問題選讀原文章節。圖表所述的實作、部署和裝置驗收狀態以各頁證據為準。

| 開發問題 | 現行實作圖 |
| --- | --- |
| 瀏覽器、LINE 與 Cloud Tasks 的請求在哪裡匯合？ | [系統執行入口](architecture/runtime/overview.md#執行入口) |
| 前端從背景回到頁面，何時重新核驗身分？ | [頁面與身分生命週期](architecture/runtime/react-lifecycle.md#頁面與身分生命週期) |
| MINI App 的 LINE token 如何換成 Session，資源操作在哪裡授權？ | [MINI App 信任流程](contracts/http.md#mini-app-信任流程) |
| LINE Webhook 何時回成功、何時要求重送？ | [Webhook 事件處理](contracts/http.md#基礎入口) |
| 改 API 或 Firestore 欄位時，格式在哪一層轉換？ | [後端資料邊界](architecture/runtime/backend-boundaries.md#格式與資料相容) |
| 專案如何收錄多個 Repository的來源項目？ | [Project 與來源引用](domains/project/project.md#最小來源引用骨架) |

其他圖依目的查閱：

| 目的 | 閱讀入口 |
| --- | --- |
| 查現行業務流程與狀態 | [領域索引](domains/README.md#domain-matrix) · [定位打卡時序](domains/attendance/README.md#定位前預檢狀態與時序) · [補卡與人工審核](domains/attendance/worktime.md#補卡與人工審核) |
| 推導模型、聚合、交易與用例 | [DDD 設計圖](references/ddd/design-diagrams.md) · [聚合邊界](references/ddd/aggregate.md#邊界與存取入口) · [提交與發布](references/ddd/transaction.md#提交與發布) |
| 開發、驗收或排錯 | [開發流程](development/README.md) · [驗收矩陣](development/verification/commands.md#變更驗收矩陣) · [部署與回復](operations/release-acceptance.md#切換觀察回復) |
| 評估尚未採用的設計 | [提案索引](proposals/README.md) · [跨領域協作](proposals/domain-collaboration.md) · [Webhook、Inbox 與 Outbox](proposals/webhook-delivery/design.md#webhook-接收處理與發送) |

## 目錄責任

| 文件分類 | 單一維護位置與讀法 |
|---|---|
| [架構](architecture/README.md) | 系統邊界、信任模型、儲存責任與執行架構。 |
| [契約](contracts/README.md) | 查跨邊界的身分、授權、資料與操作契約。 |
| [參考](references/README.md) | 查 DDD 概念、設計方法與外部來源。 |
| [外部整合](integrations/README.md) | 查 LINE、瀏覽器定位、Firestore 與 Cloud Storage 的整合契約，並前往 Google 服務索引。 |
| [領域](domains/README.md) | 從 18 個領域查規則、核心流程與驗收依據。 |
| [決策](decisions/README.md) | 查已採用 ADR 的背景與理由，依編號引用。 |
| [開發](development/README.md) | 查變更流程、工具、除錯與驗收。 |
| [維運](operations/README.md) | 查正式部署、發布驗收、故障排查與復原。 |
| [提案](proposals/README.md) | 查待評估的設計與落地條件。 |
| [規範](standards/README.md) | 查命名、程式、文件及交付準則。 |

## 按模組閱讀

[Backend](../backend/README.md) · [API](../api/README.md) · [Frontend](../frontend/README.md) · [Scripts](../scripts/README.md) · [Deploy](../deploy/README.md)

## 文件歸屬與維護

功能規則維護於 domains，共用機制維護於 architecture；已接受的重大架構決策維護於 [ADR](decisions/README.md)。規劃不覆蓋現行契約；舊報告不再保存在 docs，歷史變更透過 Git 查閱。其他入口以連結引用完整規則，不複製政策或版本。

新增或搬移頁面時，同步更新所屬索引、引用、相對路徑與錨點；依[文件規範](standards/documentation.md)驗證。工具版本與命令以 [mise.toml](../mise.toml)及工作區設定為準。

[上層導覽](../README.md)
