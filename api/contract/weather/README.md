# Weather / Forecast 既有入口契約

本文件描述既有 LINE 入口，不加入公開 OpenAPI、不產生瀏覽器 client。
完整產品規則見 [群組位置天氣](../../../docs/features/weather.md)。

- Actor／Context：共用 webhook 驗證簽章後，群組來源的位置訊息；沿用 active、reply token 與非重送條件。
- Resource／Action：以該訊息的經緯度、地址查詢目前有效的區域預報；不保存位置。
- Authority／Policy：私訊、一般文字與地圖網址不觸發；沿用 CWA 縣市與距離、有效時段、逾時及回應大小限制。
- Result：由 LINE 回覆官方區域預報；未設定、超出支援範圍或上游失敗時，回覆既有提示；不產生出勤或位置紀錄。
- 重送：位置事件不重查；回覆失敗不保證送達，使用者可重新分享。

後端 core、CWA adapter、LINE adapter 分別位於 business/weather/forecast、其 cwa/、line/。
組裝在 internal/server/line.go，與出勤、摘要資料層無依賴。
沒有既有公開 Weather HTTP API 或 MINI App 天氣畫面。
