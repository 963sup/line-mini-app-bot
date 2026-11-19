# 儲存庫洞察（Insights）

<a id="儲存庫統計"></a>

```mermaid
flowchart LR
  h["洞察：現況總覽"]
  h --- n0["成員可讀取所屬儲存庫的事項狀態、討論、里程碑<br/>與專案數。專案數表示在該庫所屬協作空間建立、<br/>以該庫為初始來源的專案；後來才連結此庫的跨庫<br/>專案不計入。"]
```

## 現況總覽與 Pulse 規劃

Mini App 的 Repository 分析入口命名為「洞察（Insights）」。選取「洞察」時才載入統計資料。既有 statistics 是其中的「現況總覽」：回答現在有多少問題、討論與里程碑，保留既有 API 與計數定義。「近期動態（Pulse）」是 Insights 下的期間報告。Owner Project 的跨庫看板進度由 [Project](../project.md) 定義。

| 視圖 | 問題 | 資料來源 |
|---|---|---|
| 現況總覽（已實作統計） | 現在有幾個已關閉問題？ | Go 查詢本庫 Issue 的當前狀態，回傳觀察時間。 |
| Pulse（規劃） | 過去七天有幾個問題新建、關閉或重開？ | Go 依 Issue 建立時間與有前後狀態的變更紀錄統計。 |

Pulse 首版預設最近 7 天，可選最近 30 天。Go 固定一次查詢的 UTC 截止時間，採左含右不含的滾動期間，回傳起訖時間供前端顯示。每項指標各按 Issue ID 去重：同一問題在期間內關閉兩次又重開，在「關閉」與「重開」各計一個。首次建立只計入「新建」。修改正文、分類或 Project item 進度不計為關閉或重開。歷史覆蓋不足時顯示可用期間或「資料不足」；無活動且覆蓋完整才顯示零。

權限沿用本庫讀取資格。驗收涵蓋期間邊界、同一 Issue 多次轉換、跨庫排除與失權。實作前先驗證歷史紀錄完整性、查詢範圍與成本；前端不逐筆載入 Issue 歷史來計數。參照 [GitHub Pulse](https://docs.github.com/en/repositories/viewing-activity-and-data-for-your-repository/using-pulse-to-view-a-summary-of-repository-activity)。

<a id="操作契約"></a>

```mermaid
flowchart LR
  h["操作契約"]
  h --- n0["項目：Actor／Resource／Action<br/>規則：Session／Bearer 核驗本人，讀取所屬儲存<br/>庫的完整統計"]
  h --- n1["項目：Context／Authority<br/>規則：啟用儲存庫與當下群組成員資格；只可讀取<br/>有權範圍"]
  h --- n2["項目：Policy／Decision<br/>規則：統計完整集合，不以單頁筆數作總數；沿用<br/>各功能自己的狀態"]
  h --- n3["項目：Result<br/>規則：真實數量與觀察時間；讀取失敗顯示錯誤，<br/>不偽裝為 0 或沿用過期成功數字"]
  h --- n4["項目：併發／一致性<br/>規則：即時查詢，無快取或背景工作；不提供跨請<br/>求統計快照保證"]
  h --- n5["項目：Event／Audit<br/>規則：唯讀，不建立額外寫入或外部通知"]
```

<a id="實作與驗收入口"></a>

```mermaid
flowchart LR
  h["實作與驗收入口"]
  h --- n0["用例位於 <br/>backend/internal/business/repository/statistics<br/>，不建立通用統計引擎。資料量增長時依實際查詢<br/>成本另行評估聚合，不自動增加計數器、排程或趨<br/>勢圖。"]
  h --- n1["OpenAPI 來源 · Project 規格 · 驗收矩陣。真實<br/> LINE 身分、群組與正式資料庫須有獨立驗收證據<br/>，不沿用舊報告的通過結論。"]
  n1 --- n1l0["OpenAPI 來源"]
  click n1l0 "../../../api/contract/repository/statistics.yaml"
  n1 --- n1l1["Project 規格"]
  click n1l1 "../project.md"
  n1 --- n1l2["驗收矩陣"]
  click n1l2 "../../development/acceptance.md"
```
