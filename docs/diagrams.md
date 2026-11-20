# Mermaid 圖表索引

初次接手可從[六張核心圖](README.md#日常開發看哪張圖)開始。以下按圖的用途列出原文章節；實作狀態依儲存庫內容判定，部署與裝置驗收依各頁紀錄。

## 現行實作

- [基礎出勤 · 定位前預檢、狀態與時序](domains/attendance/README.md#定位前預檢狀態與時序)
- [群組出勤看板與本人漏下班檢查 · 計數語意](domains/attendance/board.md#計數語意)
- [群組請假登記 · 最小流程](domains/attendance/leave.md#最小流程)
- [出勤查詢與更正 · 更正提交重試](domains/attendance/queries.md#更正提交重試)
- [出勤查詢與更正 · 審核畫面狀態同步](domains/attendance/queries.md#審核畫面狀態同步)
- [工時計算第一階段 · 共用狀態規則與事件重建](domains/attendance/worktime.md#共用狀態規則與事件重建)
- [工時計算第一階段 · 補卡與人工審核](domains/attendance/worktime.md#補卡與人工審核)
- [工時計算第一階段 · 休息未確認與暫估工時](domains/attendance/worktime.md#休息未確認與暫估工時)
- [個人或工程群組擁有的 Project · 最小來源引用骨架](domains/project/project.md#最小來源引用骨架)
- [個人或工程群組擁有的 Project · 工程驗收送件](domains/project/project.md#工程驗收送件)
- [空間討論與單層留言 · 最小實作邊界與驗證](domains/repository/discussions.md#最小實作邊界與驗證)
- [里程碑](domains/repository/milestones.md#里程碑)
- [Repository 共用資源編號基礎 · 內部接入方式](domains/repository/numbering.md#內部接入方式)
- [物件自動匯入 · 自動匯入操作契約](domains/repository/object-imports.md#自動匯入操作契約)
- [儲存庫物件 · 物件與子物件歸屬](domains/repository/objects.md#物件與子物件歸屬)
- [個人與群組 Repository · 空間載入時序](domains/repository/README.md#空間載入時序)
- [基礎群聊摘要 · 資料與可靠性](domains/summary/summary.md#資料與可靠性)
- [HTTP 與身分契約 · 基礎入口](architecture/contracts/http.md#基礎入口)
- [HTTP 與身分契約 · MINI App 信任流程](architecture/contracts/http.md#mini-app-信任流程)
- [後端功能邊界 · 格式與資料相容](architecture/runtime/backend-boundaries.md#格式與資料相容)
- [LINE 平台架構總覽 · 執行入口](architecture/runtime/overview.md#執行入口)
- [React 前端生命週期 · 頁面與身分生命週期](architecture/runtime/react-lifecycle.md#頁面與身分生命週期)
- [Go browser Session 與 Firestore · 續期交易與前端觸發](architecture/identity/browser-session.md#續期交易與前端觸發)
- [Web Geolocation API 整合契約 · 現有出勤整合檔案與責任](architecture/integrations/web-geolocation.md#現有出勤整合檔案與責任)

## 未來設計
- [Organization 所有權與 Team 歸屬](domains/repository/organization.md)

- [套件依賴與圖譜 · 允許方向](architecture/runtime/dependencies.md#允許方向)
- [LINE 整合：能力與邊界 · 3. Webhook 事件與可靠性](architecture/integrations/line.md#3-webhook-事件與可靠性)
- [Firestore 資料契約 · 邏輯關係（非資料庫外鍵）](proposals/firestore-data-contracts.md#邏輯關係非資料庫外鍵)
- [Webhook 接收、Inbox 與 Outbox 投遞](proposals/webhook-delivery/design.md#webhook-接收處理與發送)
- [未來平台目錄規劃 · 後續平台擴充參考流程](proposals/platform/package-layout.md#後續平台擴充參考流程)
- [LINE 工程管理平台目標架構（提案） · 一、完整系統總架構](proposals/line-engineering/README.md#一完整系統總架構)
- [LINE 工程管理平台目標架構（提案） · 二、LINE Developers 的正確設定結構](proposals/line-engineering/README.md#二line-developers-的正確設定結構)
- [LINE 工程管理平台目標架構（提案） · 四、業務資源的完整階層](proposals/line-engineering/README.md#四業務資源的完整階層)
- [LINE 工程管理平台目標架構（提案） · 流程 A：群組首次加入 Bot，自動建立 Repository](proposals/line-engineering/README.md#流程-a群組首次加入-bot自動建立-repository)
- [LINE 工程管理平台目標架構（提案） · 流程 B：使用者從群組開啟 MINI App](proposals/line-engineering/README.md#流程-b使用者從群組開啟-mini-app)
- [LINE 工程管理平台目標架構（提案） · 流程 C：1:1 私訊查詢個人工程資訊](proposals/line-engineering/README.md#流程-c11-私訊查詢個人工程資訊)
- [LINE 工程管理平台目標架構（提案） · 流程 D：工程操作、附件上傳與 LINE 通知](proposals/line-engineering/README.md#流程-d工程操作附件上傳與-line-通知)
- [LINE 工程管理平台目標架構（提案） · 八、通知與背景工作的完整閉環](proposals/line-engineering/README.md#八通知與背景工作的完整閉環)

## 開發與維運指引

- [HTTP API 契約 · 契約生成流程](../api/README.md#契約生成流程)

- [專案文件入口 · 目錄責任](README.md#目錄責任)
- [Acceptance and evidence · 變更驗收矩陣](development/acceptance.md#變更驗收矩陣)
- [Systematic debugging and root cause analysis · Diagnostic loop](development/debugging.md#diagnostic-loop)
- [Agent delivery workflow · Minimal handoff and evidence reuse](development/workflow.md#minimal-handoff-and-evidence-reuse)
- [Agent delivery workflow · Repository publication](development/workflow.md#repository-publication)
- [Agent delivery workflow · Business Operation compliance](development/workflow.md#business-operation-compliance)
- [Codex Cloud: mise + MCP · Acceptance](guides/local-development/setup.md)
- [Markdown 文件規範 · 編寫與修正流程](standards/documentation.md#編寫與修正流程)
- [基礎設施與分階段落地 · 第一階段：執行與部署基礎（目前優先）](guides/local-development/environment-verification.md)
- [資源導向的通用業務操作模型 · 模型與語意](architecture/contracts/business-operations.md#模型與語意)
- [開發原則與決策 · 3. 每次功能交付順序](standards/engineering.md#3-每次功能交付順序)
- [Cloud Run 部署 · 部署後確認與回滾](guides/release/cloud-run.md#部署後確認與回滾)
- [Monorepo 整合、發布與回復 · 切換、觀察、回復](guides/release/release-acceptance.md#切換觀察回復)
- [Monorepo 整合、發布與回復 · 舊版本與映像清理](guides/release/release-acceptance.md#舊版本與映像清理)
- [Troubleshooting](guides/release/troubleshooting.md#troubleshooting)
- [Troubleshooting · Cloud Build、Cloud Run 與部署後同步](guides/release/troubleshooting.md#cloud-buildcloud-run-與部署後同步)
- [context7 · Anonymous fallback](development/tooling/mcp/context7.md#anonymous-fallback)

## 現況與提案對照

- [Cloud Run 部署 · 建置入口](guides/release/cloud-run.md#建置入口)
