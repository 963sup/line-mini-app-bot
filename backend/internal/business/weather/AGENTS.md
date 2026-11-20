# backend/internal/business/weather 工作指引

## 用途與範圍

提供天氣預報功能。預報模型位於 `domain/entity/forecast/`，中央氣象署與 LINE 整合分別位於 `adapter/cwa/forecast/` 和 `adapter/line/forecast/`。

## 架構與責任邊界

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[forecast](domain/entity/forecast)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

## 收到可回覆的群組位置的維護與驗收

以群組位置事件測試預報入口；非群組或無法回覆的事件不得觸發同一群組回覆流程。

修改此規則時，核對[判定流程](../../../../docs/domains/weather/README.md#decision-flow)與[代表測試](adapter/line/forecast/forecast_test.go)。
