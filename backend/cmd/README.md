# backend/cmd 目錄導覽

## 用途與範圍

保存 Go 可執行指令的入口。

## 架構與責任邊界

指令負責解析輸入、組裝依賴與執行指定工作；共用服務與業務規則由 internal 套件維護。

## 閱讀與修改入口

先查閱[firestorecheck](firestorecheck)、[projectmigrate](projectmigrate)、[richmenu](richmenu)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：格式與資料相容](../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)


從 [firestorecheck](firestorecheck)、[projectmigrate](projectmigrate)、[richmenu](richmenu)、[server](server) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：格式與資料相容](../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../docs/diagrams.md#未來設計)
