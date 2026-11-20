# backend/internal/diagnostics/geolocation 工作指引

## 目錄用途

提供瀏覽器定位診斷所需的後端處理，與出勤打卡的定位和授權規則分開。

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：現有出勤整合檔案與責任](../../../../docs/architecture/integrations/web-geolocation.md#現有出勤整合檔案與責任) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
