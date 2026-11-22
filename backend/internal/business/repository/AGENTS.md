# backend/internal/business/repository 工作指引

## 閱讀與修改入口

先查閱[numbering.go](domain/entity/numbering.go)、[repository.go](domain/entity/repository.go)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 工作方式

[現行實作圖：空間載入時序](../../../../docs/domains/repository/README.md#空間載入時序)

Project 和 Repository 各自保存一個 `ownerId`。LINE UserID（U）以登入本人核驗，GroupID（C）以當下群組成員資格核驗。Project 可連結多個 Repository，每筆來源引用各自授權。Space 隸屬 Repository。

## 寫入庫內受限資源的維護與驗收

即使庫 ID 相符，只要擁有者已變更或狀態不再是 ACTIVE，就不能沿用舊授權完成受限操作。

修改此規則時，核對[判定流程](../../../../docs/domains/repository/README.md#decision-flow)與[判定實作](domain/entity/repository.go)。
