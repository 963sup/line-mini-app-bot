# Attendance entity

業務物件、識別、狀態與不變條件。

## 實作與規則

[出勤紀錄](record.go)、[班次](session.go)、[狀態](state.go)、[補正](correction.go) 與 [請假](leave.go) 保存身分及生命週期資料。變更前由領域運算或政策核對合法性。

## 驗證與入口

繼承[模組工作指引](../../AGENTS.md)，依[模組驗證入口](../../README.md#規格與驗證)執行相關檢查。[上層導覽](../README.md)。
