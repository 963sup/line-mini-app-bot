# Google Maps Platform 簡要規格

**狀態：按需導入。**

## 用途與最小流程

在 Mini App 顯示工地／工區位置、標記及前往地點；地址查找與路線另按需求開啟。

取得有權地點 → 按需載入地圖 → 顯示座標與標記；使用者選定／查找地址時由 Go 驗證可保存資料。

## 資料、授權與失敗邊界

裝置座標仍由瀏覽器定位取得；地圖或地理編碼不證明人在現場，不改變既有出勤規則，也不預設半徑打卡或持續追蹤。

## 導入條件

Maps JavaScript API 負責互動顯示；Geocoding、Places、Routes 個別評估。瀏覽器 key 限制正式網域及 API，服務端憑證分離；啟用前確認帳務、配額、歸屬標示及資料保存條款。

## 驗收

手機／LINE WebView 載入、拒絕定位、錯誤地址、配額耗盡、載入失敗與未授權地點均有清楚結果；地圖失敗仍能閱讀必要文字。

## 來源與相關規格

[官方文件](https://developers.google.com/maps/documentation/javascript/overview)，查證範圍見[整合總覽](./)。[瀏覽器定位](../geolocation/browser-api.md) · [金鑰限制](https://developers.google.com/maps/api-security-best-practices)。

[Google 服務索引](./)
