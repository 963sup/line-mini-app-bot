# line-環境變數規格

<a id="line-環境變數規格"></a>

```mermaid
flowchart LR
  h["LINE 環境變數規格"]
  h --- n0["所有名稱使用全大寫與底線。依產品區分前綴，不<br/>共用模糊的 LINE_CHANNEL_* 名稱。 &gt; Messaging<br/> API 三項與 MINI App Channel／LIFF ID 已供 <br/>接入能力 讀取與檢查。 &gt; Service Message <br/>token 尚未使用。"]
  click n0 "../../guides/line-setup/README.md"
```

<a id="1-命名與來源"></a>

```mermaid
flowchart LR
  h["1. 命名與來源"]
  h --- n0["前綴：LINE_MESSAGING_API_<br/>所屬產品：LINE Messaging API<br/>規則：官方帳號、Webhook、Reply／Push 與 Rich<br/> Menu"]
  h --- n1["前綴：LINE_MINI_APP_<br/>所屬產品：LINE MINI App<br/>規則：LIFF、身分驗證、MINI App Channel 與 <br/>Service Message"]
```

```mermaid
flowchart LR
  h["1. 命名與來源"]
  h --- n0["Go 欄位可對應為 MessagingAPI 與 MiniApp 兩組<br/>設定，禁止兩組憑證互相回退。 本機、測試與正<br/>式部署沿用相同變數名稱，由不同部署環境注入各<br/>自的值。"]
```

<a id="2-messaging-api"></a>

```mermaid
flowchart LR
  h["2. Messaging API"]
  h --- n0["環境變數：LINE_MESSAGING_API_CHANNEL_ID<br/>用途：明確識別 Messaging API Channel<br/>機密：否<br/>必要條件：啟用 Messaging API 時"]
  h --- n1["環境變數：LINE_MESSAGING_API_CHANNEL_SECRET<br/>用途：驗證 Webhook 簽章<br/>機密：是<br/>必要條件：接收 Webhook 時"]
  h --- n2["環境變數：<br/>LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN<br/>用途：呼叫 Reply／Push、Rich Menu 等 API<br/>機密：是<br/>必要條件：使用預先簽發 token 發送時"]
```

```mermaid
flowchart LR
  h["2. Messaging API"]
  h --- n0["Channel access token 須依官方選定的 token 類<br/>型管理有效期與輪替；若改用即時簽發， 另行設<br/>計金鑰與簽發設定，不把 client assertion 或私<br/>鑰塞入 access token 欄位。"]
```

<a id="3-mini-app"></a>

```mermaid
flowchart LR
  h["3. MINI App"]
  h --- n0["環境變數：LINE_MINI_APP_CHANNEL_ID<br/>用途：後端驗證 LINE ID token 的預期 Channel<br/>／audience<br/>機密：否<br/>必要條件：使用 MINI App 身分驗證時"]
  h --- n1["環境變數：LINE_MINI_APP_LIFF_ID<br/>用途：前端 liff.init 的 LIFF ID<br/>機密：否<br/>必要條件：啟動 MINI App 前端時"]
  h --- n2["環境變數：LINE_MINI_APP_WEB_URL<br/>用途：Flex URI action的桌面替代入口，精確<br/>HTTPS網頁Endpoint<br/>機密：否<br/>必要條件：啟用桌面開啟卡時；不從請求Host推導"]
  h --- n3["環境變數：LINE_MINI_APP_CHANNEL_ACCESS_TOKEN<br/>用途：呼叫 MINI App Service Message API<br/>機密：是<br/>必要條件：啟用 Service Message 且採預先簽發 <br/>token 時"]
```

```mermaid
flowchart LR
  h["3. MINI App"]
  h --- n0["MINI App 的 Developing／Review／Published 有<br/>不同內部 Channel；每個部署只能注入 該環境匹<br/>配的 Channel ID、LIFF ID 與 token，不得混用<br/>。Messaging API 與 MINI App 即使位於相同 <br/>Provider，也使用各自的 Channel 憑證。"]
  h --- n1["目前沒有使用 MINI App Channel secret 的流程<br/>，因此不預設加入該變數。 未來若選定的授權流<br/>程需要，再明確增加 <br/>LINE_MINI_APP_CHANNEL_SECRET。"]
  h --- n2["Quick Reply的「MINI App」先送既有Webhook <br/>postback action=miniapp，回覆一張Flex開啟卡<br/>；手機再點卡片後使用原MINI App URI，比原先多<br/>一次點擊。桌面LINE使用altUri.desktop開啟已設<br/>定的HTTPS網頁入口；Quick Reply本身不支援此屬<br/>性，Rich Menu與其他quick chips不變。功能說明<br/>卡的MINI App按鈕共用同一桌面分流規則。"]
  h --- n3["缺少或無效的LINE_MINI_APP_WEB_URL不推測網址<br/>，卡片明示目前供手機LINE使用。範例部署的<br/>_MINI_APP_WEB_URL採已知服務的/miniapp作公開<br/>候選，不代表已核LINE Console Endpoint或完成<br/>桌面SSO。Managed trigger只換image時不會自動<br/>注入新變數，必須另外經授權更新服務runtime設<br/>定／新revision後驗收；本機.env.example不自動<br/>載入。"]
  h --- n4["依據：官方URI action與外部瀏覽器，官方來源庫<br/>revision <br/>259d1fa7698d597f95a7f044f2592379a63defbe。桌<br/>面需要手動LINE登入，不能把網頁入口當授權或繞<br/>過目前成員核驗。"]
  n4 --- n4l0["官方URI action"]
  click n4l0 "https://developers.line.biz/en/reference/messaging-api/#uri-action"
  n4 --- n4l1["外部瀏覽器"]
  click n4l1 "https://developers.line.biz/en/docs/line-mini-app/develop/external-browser/"
```

<a id="4-不放進固定環境變數的資料"></a>

```mermaid
flowchart LR
  h["4. 不放進固定環境變數的資料"]
  h --- n0["資料：使用者 LINE ID token／LIFF access <br/>token<br/>正確位置：請求或安全的使用者 session；按流程<br/>驗證，不記錄於日誌"]
  h --- n1["資料：Service notification token<br/>正確位置：按使用者操作／session 安全保存，發<br/>送後更新最新值"]
  h --- n2["資料：Webhook reply token<br/>正確位置：短效事件處理資料，不作全域設定或長<br/>期排程憑證"]
  h --- n3["資料：user ID、group ID<br/>正確位置：具存取控制的業務資料，不作服務共用<br/>認證"]
```

<a id="5-注入與驗證"></a>

```mermaid
flowchart LR
  h["5. 注入與驗證"]
  h --- n0["mise.toml 管理工具與非機密共用預設；真實 <br/>Channel／LIFF ID 由部署環境提供，避免提交某<br/>個環境的固定值。"]
  h --- n1["Cloud Run 正式環境機密由 Secret Manager 綁定<br/>。現行 GitHub managed trigger 只更新 image <br/>並沿用 service template；它不從 Dockerfile、<br/>backend/.env.example 或 repository <br/>deploy/cloudbuild.yaml 匯入 runtime 變數。變<br/>更設定必須建立新的 Cloud Run revision。 本機<br/>由執行環境安全注入，不提交至 mise、範例檔或<br/>版本庫。"]
  h --- n2["前端只公開必要的 LIFF ID 等非機密資訊；所有 <br/>Channel token／secret 限後端使用。"]
  h --- n3["啟動時依啟用功能檢查必填值，錯誤只列變數名稱<br/>，不輸出值；不讀取舊 LINE_CHANNEL_* 別名。"]
  h --- n4["Channel ID 與 LIFF ID 使用字串，不能互換；群<br/>組來源 ID 也不能拿來作使用者授權。"]
  h --- n5["單靠變數名稱不能證明 token 屬於正確 Channel<br/>；部署驗收需檢查環境對應並測試實際授權。"]
  h --- n6["OPENAI_API_KEY、GEMINI_API_KEY、CWA_API_KEY <br/>也必須使用 Secret reference，不能以 literal <br/>env 寫入。Secret 沒有 enabled version 時不能<br/>綁定；目前 Gemini Secret 尚無 version。"]
  h --- n7["整合行為與資格限制見 LINE 能力矩陣。"]
  click n7 "line.md"
```
