# Attendance firestore

文件映射、查詢與交易。

## 實作與規則

[打卡交易](attendance.go)、[補正交易](attendance_correction.go)、[文件映射](documents.go) 與 [索引查詢](attendance_query.go) 實作儲存 port。保存路徑、欄位名稱、核准證據與交易邊界保持既有契約。

## 驗證與入口

繼承[模組工作指引](../../AGENTS.md)，依[模組驗證入口](../../README.md#規格與驗證)執行相關檢查。[上層導覽](../README.md)。
