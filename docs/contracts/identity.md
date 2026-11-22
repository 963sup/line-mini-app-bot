# 身分與應用 Session 契約

<a id="go-browser-session-與-firestore"></a>

本頁維護應用的 Cookie、CSRF、有效期及前端生命週期。[LINE 身分交換](../integrations/line/mini-app-identity.md)定義登入來源，[Firestore Session 持久化](../integrations/firestore/browser-session.md)定義文件、交易與 TTL。[定位診斷](../integrations/geolocation/README.md)另有公開入口。

<a id="已實作登入查詢撤銷與滑動續期"></a>

```mermaid
flowchart LR
  h["已實作：登入、查詢、撤銷與滑動續期"]
  h --- n0["登入交換由 LINE MINI App 身分契約維護。"]
  h --- n1["Cookie 為 __Host-session：32 bytes <br/>crypto/rand 隨機值，以無 padding base64url <br/>編碼，Secure、HttpOnly、Path=/、SameSite=Lax<br/>，不設定 Domain。必須經 HTTPS 使用；本機 <br/>HTTP 不降級 Cookie。Cloud Run 終止 TLS 後，<br/>Go 以 https:// 加請求 Host 檢查 Origin，不信<br/>任 X-Forwarded-Host。"]
  h --- n2["新 Session 的初始有效期為 min(now+30 分鐘, <br/>absoluteExpiresAt)，絕對期限為 min(建立時間<br/>+8 小時, 已驗證 LINE token exp)。每次請求由 <br/>Go 判斷到期，不依賴 瀏覽器 Cookie 或 <br/>Firestore TTL 的清理時間。到期後不能透過續期<br/>復活，必須重新 驗證 LINE token；不自動重送未<br/>完成的業務寫入。"]
  h --- n3["文件欄位、編碼與原子持久化見 Firestore Session 契約。"]
  h --- n4["GET /api/v1/me 與全部出勤 API 支援 Cookie。<br/>Cookie 存在時優先驗證 Session； 遺失、過期、<br/>超過絕對期限或 channel 不符回 401；Firestore<br/> 故障回 503，不退回 Bearer。已驗證使用者經既<br/>有流程放入 request context，供業務 handlers <br/>取得。"]
  h --- n5["Cookie 認證的修改操作檢查相同 Origin、<br/>CrossOriginProtection 與 Session 綁定的 <br/>X-CSRF-Token。前端在讀取請求也帶 CSRF token<br/>，伺服器收到時同樣檢查，避免 其他頁籤替換 <br/>Cookie 後舊畫面顯示另一個帳號的資料。純 <br/>Cookie GET /me 仍允許。 Session 建立入口要求<br/> Origin 與 JSON 空物件，防止登入 CSRF；缺少<br/>契約必要 header 可能由生成路由先回 400。<br/>Session 不代替群組、資料所有權或審核授權。"]
  h --- n6["DELETE /api/v1/session 先驗證 Origin 與 CSRF<br/>，再刪除資料、清 Cookie、回 204。 不存在文件<br/>或無 Cookie 可重試回 204；過期但尚未清理的文<br/>件仍可帶原 CSRF 撤銷。 刪除失敗回 503 且不清<br/> Cookie。前端立即清除私有畫面，保留登出重試<br/>按鈕，不 顯示伺服器登出成功。已通過驗證的執<br/>行中操作仍可能完成，登出不是交易取消。"]
```

<a id="續期交易與前端觸發"></a>

```mermaid
flowchart LR
  h["續期交易與前端觸發"]
  h --- n0["POST /api/v1/session/refresh 要求 Cookie、相<br/>同 Origin、X-CSRF-Token 與 JSON 空物件。續期<br/>保留 Session ID 與 CSRF，不因讀取或續期而輪<br/>替識別碼。"]
  h --- n1["現行實作"]
```

```mermaid
sequenceDiagram
  participant B as 瀏覽器
  participant G as Go Session API
  participant F as Firestore
  B->>G: POST refresh
  G->>F: 交易讀取現存 Session
  alt 文件不存在或驗證失敗
    G-->>B: 拒絕續期
  else Session 有效且符合續期條件
    G->>F: Update 並提交
    F-->>G: 提交完成
    G-->>B: Cookie 與新到期時間
  else Session 有效但未達續期條件
    G-->>B: 目前有效期
  end
```

```mermaid
flowchart LR
  h["續期交易與前端觸發"]
  h --- n0["續期的交易條件與版本更新依 Firestore Session 契約。"]
  h --- n1["原子續期不得復活已撤銷 Session，提交後才回應 Cookie。"]
  h --- n2["前端每頁只允許一筆續期進行中，最多每分鐘嘗試<br/>一次；僅在頁面可見、兩分鐘內 有使用者活動、<br/>距到期不超過十分鐘時觸發。登入完成視為初始活<br/>動；之後只由可信 pointerdown、keydown、<br/>touchstart、wheel 事件記錄活動。每分鐘的檢查<br/>器與 visibilitychange 本身不算使用者活動，背<br/>景或閒置頁面不會無限保活。 業務請求前也檢查<br/>續期；續期失敗時尚未送出的業務操作不會被偷偷<br/>重送。"]
  h --- n3["CSRF 僅存記憶體。續期 401／403 清除已驗證狀<br/>態，要求重新驗證。網路或服務 錯誤保留原期限<br/>，後續檢查可重試；不把失敗當作已續期。重新登<br/>入、登出與 pagehide 會取消舊請求、拒收舊 <br/>generation 的前端結果。"]
  h --- n4["多頁籤共用 Cookie。重新開頁面、重新驗證與 <br/>BFCache 恢復仍交換新 Session， 可能使其他頁<br/>籤的舊 CSRF 失效，需要重新驗證。並行 Cookie <br/>回應仍可能亂序： 續期不換 ID，舊回應最多使瀏<br/>覽器提早失效，伺服器仍執行現存文件與絕對期限<br/> 檢查；與重新登入或登出交錯時，舊 Cookie 可<br/>能被重新帶回但不會使已刪文件復活。 不宣稱跨<br/>頁籤無縫登入或 Cookie 回應完全有序。"]
  h --- n5["舊文件相容性依 Firestore Session 契約。"]
```

<a id="session-套用範圍"></a>

```mermaid
flowchart LR
  h["Session 套用範圍"]
  h --- n0["範圍：MINI App 本人 API、出勤操作（含請假登<br/>記）<br/>驗證方式：Cookie Session；既有無 Cookie <br/>Bearer 仍相容"]
  h --- n1["範圍：前端打卡、歷史、申請更正、審核<br/>驗證方式：共用 SessionProvider.request，帶 <br/>Cookie／CSRF；LINE token 僅用於登入交換"]
  h --- n2["範圍：Session refresh／delete<br/>驗證方式：Cookie、Origin、CSRF，不以 Bearer <br/>代替"]
  h --- n3["範圍：Session create<br/>驗證方式：向 LINE 驗證 Bearer ID token，建立<br/>應用 Session"]
  h --- n4["範圍：MINI App 設定、靜態頁面、資產<br/>驗證方式：公開，不提供使用者授權"]
  h --- n5["範圍：/location、/api/location/validate<br/>驗證方式：公開無狀態診斷，不授予出勤或本人身<br/>分"]
  h --- n6["範圍：LINE webhook<br/>驗證方式：LINE 簽章，不能用瀏覽器 Session 取<br/>代"]
  h --- n7["範圍：/internal/summary<br/>驗證方式：Google ID token 的 audience／工作<br/>帳號驗證"]
```

```mermaid
flowchart LR
  h["Session 套用範圍"]
  h --- n0["應用瀏覽器業務已使用 Session，但後端尚未禁止<br/>舊 Bearer 客戶端。移除 Bearer 需要獨立契約與<br/>客戶端遷移；本次保留相容性，不改公開診斷或服<br/>務對服務驗證。"]
```

<a id="部署與驗收界線"></a>

```mermaid
flowchart LR
  h["部署與驗收界線"]
  h --- n0["LINE Channel 選擇見 MINI App 設定；執行身分與資料庫見 Firestore 整合。"]
  h --- n1["TTL 負責背景清除，應用仍在每次讀取時核對有效期。"]
  h --- n2["持久化讀寫成本見 Firestore Session 契約。"]
  h --- n3["本機編譯、mock 與契約驗證不證明 production <br/>IAM、TTL 已啟用或真機 Cookie 行為正常。需分<br/>別驗收 LINE iOS、Android、外部瀏覽器，以及真<br/>實 Firestore 續期／登出競爭、交易重試、背景<br/>頁面、Cookie 回應亂序與多頁籤帳號切換。"]
```

<a id="文件依據"></a>

```mermaid
flowchart LR
  h["文件依據"]
  h --- n0["Context7：Go go1.27.3 的 net/http <br/>CrossOriginProtection；它允許缺少 Origin 與 <br/>Sec-Fetch-Site 的請求，因此 Session 修改操作<br/>額外要求 Origin 與 CSRF。"]
  h --- n1["Context7：Firestore 交易與重試、 TTL、 IAM、<br/> 索引最佳實務。 SDK 方法另以 backend/go.mod <br/>指定版本的 go doc／編譯核對；Google <br/>Developer Knowledge 在此對話未提供。TTL JSON<br/> 核對 Firebase CLI FieldOverride。"]
  n1 --- n1l0["Firestore 交易與重試"]
  click n1l0 "https://docs.cloud.google.com/firestore/native/docs/manage-data/transactions"
  n1 --- n1l1["TTL"]
  click n1l1 "https://docs.cloud.google.com/firestore/native/docs/ttl"
  n1 --- n1l2["IAM"]
  click n1l2 "https://docs.cloud.google.com/firestore/native/docs/security/iam"
  n1 --- n1l3["索引最佳實務"]
  click n1l3 "https://docs.cloud.google.com/firestore/native/docs/best-practices"
  n1 --- n1l4["Firebase CLI FieldOverride"]
  click n1l4 "https://github.com/firebase/firebase-tools/blob/main/src/firestore/api-spec.ts"
  h --- n2["LINE 官方 LIFF 使用者資料， 先前查證 default<br/> branch main 的 revision 為 dea09dc；本次不<br/>變更 LINE 驗證流程。"]
  click n2 "https://github.com/line/line-developers-docs-source/blob/dea09dc6e63fc5c6a3754abfbb304fc6fce4cc54/docs/en/docs/liff/using-user-profile/index.html.md"
```
