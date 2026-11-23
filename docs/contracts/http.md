# HTTP 與信任邊界

<a id="http-與身分契約"></a>

本頁維護請求、回應、Webhook 信任邊界及模組間的驗證流程。Cookie、CSRF、Session 期限、續期和撤銷以[身分與 Session 契約](identity.md)為準。

<a id="基礎入口"></a>

## 基礎入口

- [POST /webhooks/line/{channelID} 使用該渠道的 secret 驗證原始請求。未配置的路徑不會註冊。Go 限制 body 為 1 MiB，超限回 413；簽章錯誤回 401，其他解析錯誤回 400。整批事件共用 5 秒處理期限。Webhook 實作](../../backend/internal/transport/line/webhook.go)

- 現行實作

```mermaid
flowchart TD
  V["已驗簽的事件批次"] --> N{"還有事件？"}
  N -->|否| OK["HTTP 200"]
  N -->|是| H["依序執行 handler，直到事件被處理"]
  H -->|業務處理錯誤| ERR["HTTP 503"]
  H -->|無需回覆或未處理| N
  H -->|需要回覆| R["以 reply token 回覆 LINE"]
  R -->|成功| N
  R -->|失敗| T{"handler 要求重試？"}
  T -->|否| N
  T -->|是| S{"重送的 message 或 postback，且 LINE 回 400 Invalid reply token？"}
  S -->|是| N
  S -->|否| ERR
```

[HTTP 200 表示這次事件迴圈完成，空批次也回 200。持久接收後才回應的流程屬於平台提案。規劃中的 Inbox 時序](../proposals/webhook-delivery/design.md#webhook-接收處理與發送)

## 回應與送達

整合結果若使用 `accepted`，表示外部 API 接受請求；`completed` 表示本地處理完成。LINE 訊息是否送達或已讀須依該流程取得相應證據。未知錯誤依各操作契約判定是否可重試。

<a id="api-一致性"></a>

## API 一致性

- 公開 API 以 /api/v1 為前綴。共用 JSON 錯誤包含 error.code 與 error.message，例如 {"error<br>":{"code":"forbidden","message":"operation not allowed"}}。各端點的媒體類型與 HTTP 狀態碼以 OpenAPI 為準。共用契約 · 錯誤回應實作

  - [共用契約](../../api/contract/common/transport.yaml)

  - [錯誤回應實作](../../backend/internal/transport/http/json.go)

- [Cookie Session 的建立、續期、撤銷及 CSRF 檢查見 Session 契約。](identity.md)

<a id="mini-app-信任流程"></a>

## MINI App 信任流程

現行實作

```mermaid
sequenceDiagram
  participant B as MINI App / LIFF JS
  participant H as Go HTTP adapter
  participant L as LINE token verifier
  participant U as Use Case
  B->>H: LINE token（非 profile）
  H->>L: 驗證 token、預期 channel 與有效性
  L-->>H: 已驗證身分或拒絕
  H-->>B: Cookie Session + CSRF token
  B->>H: 資源操作 + session + CSRF
  H->>U: 可信 actor + 不可信資源輸入
  U->>U: 依操作檢查本人、成員或資源權限
  U-->>H: 結果或領域錯誤
  H-->>B: 安全 DTO / 一致錯誤
```

- LINE token 不能當成任意群組授權證明；網址、postback data 都不是授權。 依據：LIFF 後端驗證、LINE 整合。

  - [LIFF 後端驗證](https://developers.line.biz/en/docs/liff/using-user-profile/)

  - [LINE 整合](../integrations/line/capabilities.md)
