# backend/internal/business/issue 工作指引

## 閱讀與修改入口

先查閱[issue.go](domain/entity/issue.go)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 更新議題的維護與驗收

建立者可關閉議題後重新開啟；只有分類權限的成員不能藉由分類操作修改內容。

修改此規則時，核對[判定流程](../../../../docs/domains/repository/issue/README.md#decision-flow)與[代表測試](domain/entity/issue_test.go)。
