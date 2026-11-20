# backend/internal/business/attendance 目錄導覽

## 用途與範圍

管理打卡、出勤紀錄、補卡、請假、審核與出勤看板。工時計算與事件狀態放在 `domain/worktime/`。

## 架構與責任邊界

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[errors.go](application/errors.go)、[ports.go](application/ports.go)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：定位前預檢、狀態與時序](../../../../docs/domains/attendance/README.md#定位前預檢狀態與時序)

## 閱讀入口

[domain](domain) 維護工時、補正、審核與出勤紀錄規則。[application](application) 定義使用端需要的能力、補正審核與月報用例。[adapter](adapter) 負責 HTTP、LINE 與 Firestore 的協定和交易。瀏覽器依預檢結果取得定位，依後端回傳的紀錄與工時呈現畫面。


[現行實作圖：定位前預檢、狀態與時序](../../../../docs/domains/attendance/README.md#定位前預檢狀態與時序) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
