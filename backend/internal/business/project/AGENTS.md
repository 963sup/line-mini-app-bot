# backend/internal/business/project 工作指引

## 閱讀與修改入口

先查閱[inspections.go](domain/entity/inspections.go)、[project.go](domain/entity/project.go)、[references.go](domain/entity/references.go)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 工作方式

[現行實作圖：最小來源引用骨架](../../../../docs/domains/project/project.md#最小來源引用骨架) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

Project 和 Repository 各自保存一個 `ownerId`。LINE UserID（U）以登入本人核驗，GroupID（C）以當下群組成員資格核驗。Project 可連結多個 Repository，每筆來源引用各自授權。Space 隸屬 Repository。

## 加入跨庫來源引用的維護與驗收

加入跨庫引用時應核對選定來源及其專案連結；持有專案權限不會自動授予其他儲存庫的來源讀權。

修改此規則時，核對[判定流程](../../../../docs/domains/project/README.md#decision-flow)與[代表測試](application/reference_source_test.go)。
