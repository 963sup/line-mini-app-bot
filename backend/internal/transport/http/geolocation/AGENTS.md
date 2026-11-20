# backend/internal/transport/http/geolocation 工作指引

## 用途與範圍

提供瀏覽器定位診斷所需的後端處理，與出勤打卡的定位和授權規則分開。

## 架構與責任邊界

依目錄用途提供後端入口或共用能力；服務接線由 bootstrap 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

先查閱[geolocation.go](geolocation.go)、[本目錄導覽](README.md)、[上層工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：現有出勤整合檔案與責任](../../../../../docs/architecture/integrations/web-geolocation.md#現有出勤整合檔案與責任)

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：現有出勤整合檔案與責任](../../../../../docs/architecture/integrations/web-geolocation.md#現有出勤整合檔案與責任) · [未來設計圖](../../../../../docs/diagrams.md#未來設計)
