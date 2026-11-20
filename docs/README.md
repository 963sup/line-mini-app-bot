# 專案文件入口

## 用途與範圍

集中維護架構、領域規格、開發流程、操作指南與提案。

## 架構與責任邊界

實作現況由程式與契約驗證，文件按主題維護。

## 閱讀與修改入口

先查閱[architecture](architecture)、[domains](domains)、[development](development)、[guides](guides)、[proposals](proposals)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[系統執行入口](architecture/runtime/overview.md#執行入口) · [頁面與身分生命週期](architecture/runtime/react-lifecycle.md#頁面與身分生命週期)


先依任務選一個入口，讀相關規格與局部指引後再查程式碼；不需要依序讀完所有分類。文件與設定存在不代表已完成部署或真實環境驗收。

## 日常開發看哪張圖

[全部圖表按狀態查閱](diagrams.md)。初次接手先看執行入口；修改功能時，依手上的問題選讀一張。

| 開發問題 | 現行實作圖 |
| --- | --- |
| 瀏覽器、LINE 與 Cloud Tasks 的請求在哪裡匯合？ | [系統執行入口](architecture/runtime/overview.md#執行入口) |
| 前端從背景回到頁面，何時重新核驗身分？ | [頁面與身分生命週期](architecture/runtime/react-lifecycle.md#頁面與身分生命週期) |
| MINI App 的 LINE token 如何換成 Session，資源操作在哪裡授權？ | [MINI App 信任流程](architecture/contracts/http.md#mini-app-信任流程) |
| LINE Webhook 何時回成功、何時要求重送？ | [Webhook 事件處理](architecture/contracts/http.md#基礎入口) |
| 改 API 或 Firestore 欄位時，格式在哪一層轉換？ | [後端資料邊界](architecture/runtime/backend-boundaries.md#格式與資料相容) |
| 專案如何收錄多個 Repository的來源項目？ | [Project 與來源引用](domains/project.md#最小來源引用骨架) |

出勤、請假、送件等流程從[功能索引](domains/README.md)進入。評估未來設計時才讀[平台規劃](proposals/README.md)；[Inbox、Outbox 與 Webhook 時序](proposals/webhook-delivery/design.md)描述規劃中的持久接收機制。

## 按任務開始

| 要做的事 | 先讀 | 再按需要查閱 |
| --- | --- | --- |
| 找修改位置 | [需求到實作導航](architecture/runtime/code-layout.md#從需求定位修改範圍) | 所屬功能規格與目錄 AGENTS.md |
| 修改業務或授權 | [功能索引](domains/README.md) | [操作模型](architecture/contracts/business-operations.md)、[契約歸屬](architecture/runtime/dependencies.md#契約與不變條件的維護位置) |
| 修改公開 API | [API 契約與生成流程](../api/README.md) | 對應功能與 [HTTP 契約](architecture/README.md#http-契約) |
| 準備開發環境 | [開發入口](guides/README.md) | [工具與 MCP](development/tooling/README.md) |
| 判定修改是否完成 | [驗收矩陣](development/acceptance.md#變更驗收矩陣) | 對應功能及外部環境驗收 |
| 部署、發布或排錯 | [營運入口](architecture/README.md#部署與營運) | 相關設定與真實環境證據 |
| 評估未來能力 | [規劃與提案](proposals/README.md) | 已接受 ADR 與現行功能規格 |

## 目錄責任

| 任務 | 文件入口 |
| --- | --- |
| 操作、驗收與回復 | [開發操作](guides/README.md) · [發布與回復](guides/release/release-acceptance.md) |
| 業務規則與授權 | [功能索引](domains/README.md) |
| 跨功能機制與契約 | [系統架構](architecture/README.md) |
| 修改程式與文件 | [命名](standards/naming.md) · [Go](standards/go.md) · [文件規範](standards/documentation.md) |
| 查已採用決策或未來設計 | [ADR](architecture/README.md#architecture-decision-records) · [提案](proposals/README.md) |
| 工具與代理協作 | [開發工作流程與工具](development/README.md) |

## 按模組閱讀

[Backend](../backend/README.md) · [API](../api/README.md) · [Frontend](../frontend/README.md) · [Scripts](../scripts/README.md) · [Deploy](../deploy/README.md)

## 文件歸屬與維護

功能規則維護於 domains，共用機制維護於 architecture；已接受的重大架構決策維護於 [ADR](architecture/README.md#architecture-decision-records)。規劃不覆蓋現行契約；舊報告不再保存在 docs，歷史變更透過 Git 查閱。其他入口以連結引用完整規則，不複製政策或版本。

新增或搬移頁面時，同步更新所屬索引、引用、相對路徑與錨點；依[文件規範](standards/documentation.md)驗證。工具版本與命令以 [mise.toml](../mise.toml)及工作區設定為準。
