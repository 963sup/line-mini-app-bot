# frontend/miniapp/src/shared/location 工作指引

遵循[上層指引](../AGENTS.md)。修改互動時，核對使用此模組的頁面與 miniapp 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：現有出勤整合檔案與責任](../../../../../docs/foundation/integrations/web-geolocation.md#現有出勤整合檔案與責任) · [未來設計圖](../../../../../docs/diagrams.md#未來設計)

定位由瀏覽器 JavaScript 取得，Go 驗證不受信任的座標與操作權限。新增位置歷史或在場驗證須有產品契約。修改前先讀[瀏覽器定位規格](../../../../../docs/foundation/integrations/web-geolocation.md)。
