# 空間統計

提供 Repository 的統計查詢與結果模型，並透過 `http/` 提供 API。

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[statistics.go](../domain/entity/statistics/statistics.go)。

## 相關規格

[基礎規格](../../../../../docs/domains/repository/insights.md)

## 閱讀入口

操作、權限、資料與驗收依 [基礎規格](../../../../../docs/domains/repository/insights.md)。不擴充未授權功能；核心、HTTP 與 Firestore 各自保留責任，前端組合既有元件。

[適用工作指引](../AGENTS.md) · [上層導覽](../README.md)
