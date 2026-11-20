# backend/internal/business/attendance 目錄導覽

## 用途與範圍

管理打卡、出勤紀錄、補卡、請假、審核與出勤看板。工時計算與事件狀態放在 `worktime/`。

## 架構與責任邊界

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[errors.go](errors.go)、[ports.go](ports.go)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：定位前預檢、狀態與時序](../../../../docs/domains/attendance/README.md#定位前預檢狀態與時序)

## 閱讀入口

從 [approvals](approvals)、[board](board)、[corrections](corrections)、[errors.go](errors.go)、[firestore](firestore)、[http](http) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：定位前預檢、狀態與時序](../../../../docs/domains/attendance/README.md#定位前預檢狀態與時序) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
