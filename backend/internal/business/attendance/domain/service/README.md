# Attendance service

跨物件計算、補正與證據核對。

## 實作與規則

[工時計算](worktime.go)、[狀態轉移](transition.go)、[事件序列核對](event_stream.go)、[補正套用](correction.go)、[核准證據核對](approval.go) 與 [月報計算](monthly.go) 為純 Go 領域運算。回傳的計算結果不負責 API 分頁或呈現。

## 驗證與入口

繼承[模組工作指引](../../AGENTS.md)，依[模組驗證入口](../../README.md#規格與驗證)執行相關檢查。[上層導覽](../README.md)。
