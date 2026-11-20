# backend/internal/business/project 工作指引

## 用途與範圍

管理工程專案、來源引用及檢查紀錄，負責專案歸屬與引用來源的授權規則。

## 架構與責任邊界

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[inspections.go](domain/entity/inspections.go)、[project.go](domain/entity/project.go)、[references.go](domain/entity/references.go)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：最小來源引用骨架](../../../../docs/domains/project/project.md#最小來源引用骨架)

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：最小來源引用骨架](../../../../docs/domains/project/project.md#最小來源引用骨架) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

Project 和 Repository 各自保存一個 `ownerId`。LINE UserID（U）以登入本人核驗，GroupID（C）以當下群組成員資格核驗。Project 可連結多個 Repository，每筆來源引用各自授權。Space 隸屬 Repository。
