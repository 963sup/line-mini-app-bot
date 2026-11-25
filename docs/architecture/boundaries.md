# 架構邊界

後端以業務能力為第一層邊界，每個能力可包含 domain、application 與 adapter。邊界用責任與依賴方向判定，不以目錄數量或是否同一程序判定。

## 後端分層

| 層          | 責任                                | 不得擁有                       |
| ----------- | ----------------------------------- | ------------------------------ |
| Domain      | 業務狀態、值、政策、不變條件        | HTTP、Firestore、LINE SDK 型別 |
| Application | 用例協調、輸入／輸出 port、交易意圖 | 傳輸格式、供應端細節           |
| Adapter     | DTO、持久化、外部 API、錯誤轉譯     | 跨領域業務規則                 |
| Bootstrap   | client 建立、依賴注入、路由、啟停   | 業務判斷                       |
| Transport   | 共用 HTTP／LINE 接入                | 功能內部狀態                   |

## 跨領域合作

消費端定義所需的最小介面，由 composition root 注入提供端 adapter。不得匯入另一領域的 Firestore document、HTTP handler 或內部 entity 以繞過公開語意。跨領域 DTO 必須轉譯，避免形成 shared mutable model。

`platform/identity` 提供可信 actor。Repository 判定儲存庫所有者，Project 判定個人、群組與組織的專案存取資格。Repository 核心驗證擁有者類型與識別碼，Project 在既有的 LINE 範圍使用 `platform/line/lineid` 驗證格式。HTTP 處理器經 Session 服務執行建立、讀取、續期與撤銷，Firestore 交易由持久化轉接器負責。[身分、平台與傳輸的套件位置](../standards/backend.md)

Repository、Organization、Team、Enterprise 與 Membership 透過具名查詢或操作契約合作，資源擁有者負責寫入。企業連結組織時，Enterprise 驗證企業擁有者，Organization 在同一筆交易驗證組織擁有者、現有歸屬並保存連結。企業目錄透過 Organization、Team 查詢可見項目，LINE 成員檢查在 Firestore 讀取交易結束後執行。

Repository 統計在同一份交易快照中向 Issue、Discussion 及 Project 取得各自的計數；Project 負責選擇擁有者的儲存路徑。

## 前端邊界

前端以 `business/<capability>` 組織功能。Shell 與 shared UI 不呼叫具體業務 API；Identity 擁有 Session 生命週期；平台 SDK 封裝於 `platform/`。頁面只組合功能，不重新實作授權或業務計算。

## 契約邊界

OpenAPI 定義跨 HTTP 邊界的資料與錯誤。Adapter 可轉換契約型別與領域型別，但 Domain 不依賴產生的傳輸 DTO。資料庫 schema 是 adapter 的保存契約，不是公開 API。

## 邊界檢查

[邊界測試](../../backend/architecture_test.go)掃描 `backend/internal/` 的 Go 生產程式碼，依賴白名單以該目錄為基準。[回歸測試](../../backend/architecture_regression_test.go)在暫存 Go module 執行同一支掃描器，驗證允許與拒絕案例。執行方式見[驗證流程](../workflows/verification.md#後端架構測試)。

新增 import、共用型別或跨模組呼叫時，先回答：

1. 規則由哪個 bounded context 擁有？
2. 消費端真正需要的最小能力是什麼？
3. 失敗、期限與一致性由哪一層處理？
4. 是否繞過了授權、版本或交易檢查？
