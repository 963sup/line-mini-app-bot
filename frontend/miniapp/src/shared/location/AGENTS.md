# frontend/miniapp/src/shared/location 工作指引

## 閱讀與修改入口

先查閱[position.ts](position.ts)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：現有出勤整合檔案與責任](../../../../../docs/architecture/integrations/web-geolocation.md#現有出勤整合檔案與責任) · [瀏覽器定位規格](../../../../../docs/architecture/integrations/web-geolocation.md)

[現行實作圖：現有出勤整合檔案與責任](../../../../../docs/architecture/integrations/web-geolocation.md#現有出勤整合檔案與責任)

定位由瀏覽器 JavaScript 取得，Go 驗證不受信任的座標與操作權限。新增位置歷史或在場驗證須有產品契約。修改前先讀[瀏覽器定位規格](../../../../../docs/architecture/integrations/web-geolocation.md)。
