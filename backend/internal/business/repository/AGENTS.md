# backend/internal/business/repository 工作指引

## 用途與範圍

管理 Repository 的歸屬、查詢與編號，並承載 Space、標籤、里程碑、Wiki 與統計等子功能。

## 架構與責任邊界

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[numbering.go](domain/entity/numbering.go)、[repository.go](domain/entity/repository.go)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：空間載入時序](../../../../docs/domains/repository/README.md#空間載入時序)

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：空間載入時序](../../../../docs/domains/repository/README.md#空間載入時序) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

Project 和 Repository 各自保存一個 `ownerId`。LINE UserID（U）以登入本人核驗，GroupID（C）以當下群組成員資格核驗。Project 可連結多個 Repository，每筆來源引用各自授權。Space 隸屬 Repository。

## 寫入庫內受限資源的維護與驗收

即使庫 ID 相符，只要擁有者已變更或狀態不再是 ACTIVE，就不能沿用舊授權完成受限操作。

修改此規則時，核對[判定流程](../../../../docs/domains/repository/README.md#decision-flow)與[判定實作](domain/entity/repository.go)。
