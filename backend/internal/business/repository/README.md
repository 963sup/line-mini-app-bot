# backend/internal/business/repository 目錄導覽

[領域文件](../../../../docs/domains/repository/README.md) · [全域索引](../../../../docs/domains/README.md#domain-matrix)

管理 Repository 的歸屬、查詢與編號，並承載 Space、標籤、里程碑、Wiki 與統計等子功能。

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

從 [numbering.go](domain/entity/numbering.go)、[repository.go](domain/entity/repository.go)、[firestore](adapter/firestore)、[http](adapter/http)、[labels](labels)、[line](adapter/line)、[milestones](milestones) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 閱讀入口

[現行實作圖：空間載入時序](../../../../docs/domains/repository/README.md#空間載入時序) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

## 領域追蹤入口

[領域規格](../../../../docs/domains/repository/README.md) · [判定流程](../../../../docs/domains/repository/README.md#decision-flow) · [判定實作](domain/entity/repository.go)

[上層導覽](../README.md)

## 下層文件入口

- [statistics](statistics/README.md)
