# Attendance dto

分頁、月報、看板及打卡回傳資料。

## 實作與規則

[打卡結果](punch.go)、[月報](monthly.go)、[分頁](page.go)、[補正詳情](correction.go) 與 [看板](board.go) 為應用層輸入輸出資料，不承擔 HTTP 或 Firestore 序列化標籤。

## 驗證與入口

繼承[模組工作指引](../../AGENTS.md)，依[模組驗證入口](../../README.md#規格與驗證)執行相關檢查。[上層導覽](../README.md)。
