# backend/internal/business/weather 工作指引

## 閱讀與修改入口

先查閱[forecast](domain/entity/forecast)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

## 收到可回覆的群組位置的維護與驗收

以群組位置事件測試預報入口；非群組或無法回覆的事件不得觸發同一群組回覆流程。

修改此規則時，核對[判定流程](../../../../docs/domains/weather/README.md#decision-flow)與[代表測試](adapter/line/forecast/forecast_test.go)。
