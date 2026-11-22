# 瀏覽器定位

[Web Geolocation API 整合契約](browser-api.md)說明定位權限、使用者同意、資料驗證與本機示範。打卡授權、儲存與業務規則由[出勤規格](../../domains/attendance/README.md)維護。

`/location` 與 `POST /api/location/validate` 提供公開、無狀態的定位診斷。診斷不辨識本人、不寫入出勤資料，也不授予出勤權限。需要辨識使用者的流程依[瀏覽器 Session 契約](../../architecture/identity/browser-session.md#session-套用範圍)及所屬領域的授權規則。

[外部整合索引](../README.md)
