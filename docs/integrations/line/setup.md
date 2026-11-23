# LINE 接入設定

保存 LINE 平台與 MINI App 的設定步驟。

此處說明外部平台設定；平台能力、設定契約與呈現方式見 [LINE 整合索引](./)。

## Go LINE 接入：啟動

- 官方 Go SDK；版本見 backend/go.mod。已接入 Webhook、MINI App Session、Firestore 出勤／請假及 Cloud Tasks 摘要。群組角色管理、Service Message 與常駐 Worker 不在此範圍。

- bash scripts/tooling/tool-exec.sh mise run dev

- PORT 預設 8080，不自動讀 .env。保留 /location定位示範，沒有健康檢查路由。 Messaging API 三項全缺時停用、部分缺失時啟動失敗。MINI App 以 LINE_MINI_APP_ENV 選環境，預設 developing；未知值或選定三元組缺失／非法時啟動失敗。 設定名稱與 Channel 分離規則見 [LINE 環境變數](configuration.md)。

<a id="mini-app-setup"></a>

## MINI App 設定與登入

1. 在部署環境依[公開環境設定](configuration.md#公開環境設定)選擇階段，注入該組 Channel ID、LIFF
   ID 與 Endpoint URL。
2. 在對應的 LINE Console Channel 啟用 `openid` scope；需要顯示名稱時再啟用 `profile`。
3. 將該階段的 MINI App Endpoint 設為 `https://<service-host>/miniapp`，並以永久連結
   `https://miniapp.line.me/<LIFF ID>` 開啟。

- Go 提供 Next.js 靜態匯出的 /miniapp 與 React 程式。前端載入官方LIFF v2
  SDK，由 /api/v1/miniapp/config 取得公開 LIFFID 並執行 liff.init。外部瀏覽器由使用者點擊登入，redirectUri保留本站目前頁面與查詢參數。

- 驗證與資源授權的時序集中於[MINI App信任流程](../../contracts/http.md#mini-app-信任流程)。瀏覽器以liff.getIDToken() 的 Bearer
  token 與空物件呼叫
  `POST /api/v1/session`。Go 驗證成功才建立 FirestoreSession，回傳安全 Cookie、本人資料、CSRF
  token 與有效期。後續本人及出勤API 經 SessionProvider.request 使用Cookie／CSRF，近期有活動且接近到期才續期。

- 不接受前端 profile、decoded token 或任意 userId作身分。Session ID 透過 HttpOnly Cookie 傳遞，CSRF
  token僅保留於記憶體；原始 Session ID、CSRF 與 LINE
  token不存入資料庫或日誌。 Session 並不授予任意群組或他人資料的存取權。

- 登出呼叫 `DELETE /api/v1/session`
  撤銷伺服器狀態並清除Cookie；失敗時保留重試入口， 不能只清畫面就宣稱已登出。撤銷應用 Session不等於撤銷 LINE 帳號授權。 期限、續期、Cookie 優先順序、既有 Bearer相容性及錯誤處理以
  [Session契約](../../contracts/identity.md)為單一來源；路徑與請求格式見[API 契約](../../../api/)。

- 程式接入不代表 MINI
  App 已完成 Channel 建立、資格確認或發布審查。本機可看頁面與未設定提示；真實登入需要正確的 HTTPS
  Endpoint 與 Channel設定。

<a id="firestore-唯讀連線"></a>
<a id="閱讀與修改入口"></a>
<a id="相關規格"></a>

## 接入後驗證

- 驗證 [Webhook 接線](webhook.md)、[聊天指令](commands.md)及 [Rich Menu](../../operations/rich-menu.md)。
- 執行 [Firestore 唯讀連線檢查](../../development/tooling/backend/firestore.md#唯讀連線檢查)。
- 核對[平台來源與環境限制](sources.md)、[相關規格](../../)及[文件指引](../../AGENTS.md)。
