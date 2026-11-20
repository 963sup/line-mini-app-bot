# backend/internal/business/repository 目錄導覽

## 用途與範圍

管理 Repository 的歸屬、查詢與編號，並承載 Space、標籤、里程碑、Wiki 與統計等子功能。

## 架構與責任邊界

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[numbering.go](domain/entity/numbering.go)、[repository.go](domain/entity/repository.go)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：空間載入時序](../../../../docs/domains/repository/README.md#空間載入時序)

## 閱讀入口

從 [firestore](firestore)、[http](http)、[labels](labels)、[line](line)、[milestones](milestones)、[numbering.go](domain/entity/numbering.go) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：空間載入時序](../../../../docs/domains/repository/README.md#空間載入時序) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
