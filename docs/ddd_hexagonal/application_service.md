# 應用服務（Application Service）

Application Service 執行一個明確用例，安排取得事實、領域決策、保存及結果回傳。它可以是函數、具名命令處理器或服務方法，不要求特定框架。

## 責任與流程

1. 接收 Adapter 轉換的命令及可信 Actor，驗證必要的用例輸入。
2. 透過 Port 取得資源、資格及版本，協調授權與業務政策。
3. 呼叫模型、Factory 或 Domain Service 執行領域決策。
4. 依一致性需求保存結果、冪等收據與稽核。
5. 明確區分資料提交與外部通知的結果。

具體角色政策與合法狀態轉換應有自己的模型擁有者，不能因為用例需要協調，就將所有業務判斷改寫進 Application。HTTP status、JSON 格式與資料庫 document 也不屬於用例。

## 命令與查詢

命令定義預期改變、重送與併發語意；查詢定義可見範圍、投影、分頁與錯誤。分開命名不代表必須建立 CQRS 系統、兩套資料庫或訊息匯流排。

## 設計例子

接受成員邀請的用例，須協調登入者、邀請、目前管理資格、上層資格、合法轉換與提交。HTTP 與 Webhook 可以驅動同一用例，各自保留認證來源與呈現格式；新增入口不能繞過相同政策。

## 設計產出與核對

列明 Actor、Resource、Action、Authority、前置條件、讀寫集合、版本、防重與結果。

[Domain Service](domain_service.md) · [一致性](consistency.md) · [六邊形架構](hexagonal_architecture.md) · [DDD 設計索引](design_guide.md)
