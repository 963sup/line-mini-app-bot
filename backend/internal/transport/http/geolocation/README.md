# backend/internal/transport/http/geolocation 目錄導覽

提供瀏覽器定位診斷所需的後端處理，與出勤打卡的定位和授權規則分開。

依目錄用途提供後端入口或共用能力；服務接線由 bootstrap 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

從 [geolocation.go](geolocation.go) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 閱讀入口

[現行實作圖：現有出勤整合檔案與責任](../../../../../docs/architecture/integrations/web-geolocation.md#現有出勤整合檔案與責任)

[上層導覽](../README.md)
