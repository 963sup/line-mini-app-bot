# 專案基礎

## 用途與範圍

管理工程專案、來源引用及檢查紀錄，負責專案歸屬與引用來源的授權規則。

## 架構與責任邊界

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[inspections.go](inspections.go)、[project.go](project.go)、[references.go](references.go)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[基礎規格](../../../../docs/domains/project.md) · [現行實作圖：最小來源引用骨架](../../../../docs/domains/project.md#最小來源引用骨架)

## 閱讀入口

操作、權限、資料與驗收依 [基礎規格](../../../../docs/domains/project.md)。不擴充未授權功能；核心、HTTP 與 Firestore 各自保留責任，前端組合既有元件。

[現行實作圖：最小來源引用骨架](../../../../docs/domains/project.md#最小來源引用骨架) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
