# Attendance line

LINE 事件與出勤訊息呈現。

## 實作與規則

[打卡入口](attendance.go) 與 [月報呈現](attendance_history.go) 隔離 LINE 協定；呈現用的群組名稱整理在 [共用呈現](../presentation/group.go)。

群組核驗由 bootstrap 注入 [platform/line](../../../../platform/line/README.md) 至既有 application Membership port；HTTP adapter 保留允許、拒絕及暫時無法核驗的回應判定。

## 驗證與入口

繼承[模組工作指引](../../AGENTS.md)，依[模組驗證入口](../../README.md#規格與驗證)執行相關檢查。[上層導覽](../README.md)。
