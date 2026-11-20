# backend/internal/business/weather 目錄導覽

## 用途與範圍

提供天氣預報功能。預報模型位於 `domain/entity/forecast/`，中央氣象署與 LINE 整合分別位於 `adapter/cwa/forecast/` 和 `adapter/line/forecast/`。

## 架構與責任邊界

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[forecast](domain/entity/forecast)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 閱讀入口

從 [forecast](domain/entity/forecast) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：格式與資料相容](../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
