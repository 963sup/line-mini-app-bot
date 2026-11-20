# backend/internal/business/attendance 工作指引

## 目錄用途

管理打卡、出勤紀錄、補卡、請假、審核與出勤看板。工時計算與事件狀態放在 `worktime/`。

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：定位前預檢、狀態與時序](../../../../docs/domains/attendance/README.md#定位前預檢狀態與時序) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
