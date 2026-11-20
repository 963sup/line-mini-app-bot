# 工程群組

## 用途與範圍

保存從已驗簽 LINE 事件取得的群組資料與成員候選資訊，提供群組查詢。

## 架構與責任邊界

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[group.go](group.go)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[專案設計](../../../../docs/domains/project/project.md) · [現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 閱讀入口

Group 直接以 LINE GroupID 為文件 ID，保存已驗簽的群組事件與成員候選清單。Project 歸屬與授權見[專案設計](../../../../docs/domains/project/project.md)。

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
