# line-平台架構總覽

<a id="line-平台架構總覽"></a>

## LINE 平台架構總覽

- Cloud Run 執行一個 Go HTTP 服務。Next.js 在建置階段匯出 MINI App，映像將靜態檔案放在 /app/frontend/，由 Go 提供頁面與 API。服務從 backend/main.go 啟動，功能接線集中在 backend/internal/bootstrap。建置流程 · 服務組裝 · 部署決策

  - [建置流程](../../../Dockerfile)

  - [服務組裝](../../../backend/internal/bootstrap/server.go)

  - [部署決策](../../decisions/0003-cloud-run-primary.md)

<a id="執行入口"></a>

## 執行入口

現行實作

```mermaid
architecture-beta
  group run(cloud)[Cloud Run]
  service line(internet)[LINE Webhook]
  service mini(internet)[MINI App Browser]
  service tasks(server)[Cloud Tasks]
  service api(server)[Go HTTP Service] in run
  service assets(disk)[Next.js Static Export] in run
  line:R --> L:api
  mini:B --> T:api
  tasks:R --> L:api
  api:R -- L:assets
```

- Go 驗簽後分派 LINE 事件，由 handler 決定回覆及重試。MINI App 由 Go 提供頁面與 API，後端分別核驗身分及資源存取權。Cloud Tasks 呼叫內部摘要工作入口。Webhook 與身分 · Session · 群聊摘要

  - [Webhook 與身分](../../contracts/http.md)

  - [Session](../../contracts/identity.md)

  - [群聊摘要](../../domains/summary/summary.md)

<a id="業務與共用系統基礎"></a>

## 業務與共用系統基礎

- 個人或群組擁有 Repository；個人或組織擁有 Project。Space 是庫內工區；Project 透過引用跨庫收錄物件、子物件、問題與討論，讀取時核驗各來源庫的權限。Project 規格 · 庫內工區

  - [Project 規格](../../domains/project/project.md)

  - [庫內工區](../../domains/repository/spaces.md)

- 出勤、日誌、群聊摘要與天氣各自維護操作規格與驗收條件。持久 Inbox、通用 Outbox、租約 worker 與管理後台的構想另存於平台規劃。功能索引 · 平台規劃圖

  - [功能索引](../../domains/)

  - [平台規劃圖](../../proposals/platform/package-layout.md#後續平台擴充參考流程)

<a id="文件路由"></a>

## 文件路由

| 任務               | 參考文件                                                                                                  |
| ------------------ | --------------------------------------------------------------------------------------------------------- |
| 修改模組與依賴     | [目前目錄](code-layout.md)、[後端邊界](backend-boundaries.md)                                             |
| 修改資料與交易     | [Firestore 整合](../../integrations/firestore/connection.md)                                              |
| 準備裝置或部署驗收 | [整合與發布驗收](../../operations/release-acceptance.md)、[Cloud Run 部署](../../operations/cloud-run.md) |
| 比較現況與目標     | [工程平台目標架構](../../proposals/line-engineering/)                                                     |

<a id="文件路由-1"></a>
