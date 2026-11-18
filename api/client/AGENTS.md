# 共用 API 客戶端

適用於本目錄及其子目錄；繼承 [上層指引](../AGENTS.md)，並遵循其引用的根目錄規則。此處只補充本範圍責任；使用者與主機指令優先，不增加外部寫入授權。

## 職責與邊界

- 維護公開 HTTP 的共用 TypeScript client；操作型別來自契約生成，不手改 generated 內容或手寫另一套 DTO。
- 新業務沿用現有 createAPIClient 與前端 Session 認證請求，不按能力複製 client、登入或重試機制。
- 內部 Summary worker 不加入瀏覽器 client；不把 token、服務帳戶憑證或服務端設定打包到前端。
- 修改契約先執行上層生成與相容性流程；修改 client 行為另跑相關既有型別與消費端檢查。
