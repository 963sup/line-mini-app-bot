# backend/internal/business/group 工作指引

## 閱讀與修改入口

先查閱[group.go](application/query/group.go)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

## 查閱受限群組資料的維護與驗收

曾在事件中出現的使用者即使被列為候選，仍須通過群組成員核驗才能取得受限資料。

修改此規則時，核對[判定流程](../../../../docs/domains/group/README.md#decision-flow)與[代表測試](application/query/group_test.go)。
