# 天氣

## 領域設計

**子域定位**　外部預報屬通用能力；群組查詢是產品支援能力。

**通用語言**　位置是查詢輸入，預報由氣象來源提供，群組回覆是一次查詢的呈現結果。

**上下文與契約**　LINE 入口辨識能回覆的群組事件，天氣適配器取得中央氣象署預報。出勤定位的目的與授權不因兩者都使用位置而共用。

**驗收案例**　以群組位置事件測試預報入口；非群組或無法回覆的事件不得觸發同一群組回覆流程。

[對照實作或測試](../../../backend/internal/business/weather/adapter/line/forecast/forecast.go) · [戰略設計與分類準則](../../architecture/ddd/README.md)

天氣功能的查詢與 LINE 互動見天氣規格。

[相關業務規格](weather.md) · [後端weather 實作說明](../../../backend/internal/business/weather/README.md)
