# backend/internal/business/weather 目錄導覽

[領域文件](../../../../docs/domains/weather/README.md) · [全域索引](../../../../docs/domains/README.md#domain-matrix)

提供天氣預報功能。預報模型位於 `domain/entity/forecast/`，中央氣象署與 LINE 整合分別位於 `adapter/cwa/forecast/` 和 `adapter/line/forecast/`。

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

從 [forecast](domain/entity/forecast) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 閱讀入口

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 領域追蹤入口

[領域規格](../../../../docs/domains/weather/README.md) · [判定流程](../../../../docs/domains/weather/README.md#decision-flow) · [代表測試](adapter/line/forecast/forecast_test.go)

[上層導覽](../README.md)
