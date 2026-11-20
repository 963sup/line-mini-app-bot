# Attendance port

儲存、成員核驗與交易介面。

## 實作與規則

[一般儲存能力](store.go)、[打卡能力](punch.go)、[審核交易](review.go) 與 [月報讀取](monthly.go) 定義用例需要的輸出能力。Firestore 或 LINE 適配器實作介面。

## 驗證與入口

繼承[模組工作指引](../../AGENTS.md)，依[模組驗證入口](../../README.md#規格與驗證)執行相關檢查。[上層導覽](../README.md)。
