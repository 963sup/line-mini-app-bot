# Firestore Session 持久化

瀏覽器 Cookie、CSRF 及公開 API 以[應用 Session 契約](../../architecture/identity/browser-session.md)為準。

## 文件與相容性

Firestore sessions/{sha256(sessionID)} 保存 userID、displayName、channelID、createdAt、expiresAt、csrfTokenHash、absoluteExpiresAt、lastRenewedAt、revision、schemaVersion=2。原始 Session ID、CSRF token 與 LINE token 不寫入資料庫或日誌。核心不依賴 Firestore 型別，adapter 負責文件編解碼；以 transaction 原子建立新文件並刪除請求帶入的舊文件。每個實例擁有 Session Firestore client，使用既有 project/database 設定並在結束時關閉，不要求 Messaging channel 才能建立登入。

schemaVersion=1 缺少原始 LINE exp，只允許原固定期限內使用；refresh 不升級、不延長。重新驗證 LINE 後才建立 schemaVersion=2。未知或不完整的新版本文件拒絕使用，不能從舊 expiresAt 推算出更長的 provider 有效期。

## 續期與競爭

交易核驗身分、期限與 CSRF。剩餘期限不超過 10 分鐘、距上次續期至少 5 分鐘且新期限增加時，更新 expiresAt=min(now+30 分鐘, absoluteExpiresAt)、lastRenewedAt=now，並將 revision 加一。

每次交易嘗試都在讀取文件後讀時鐘，避免鎖定等待或重試沿用已過期的時間。交易只使用 Update，不使用 upsert；遇到登出或 TTL 刪除會衝突重試或找不到文件，不能重新建立已撤銷的 Session。交易 callback 不呼叫 LINE、不發 Cookie；僅在提交成功後回應 Cookie 與到期時間。並行續期不縮短伺服器期限。

每個業務請求至少一筆 Session 讀取；符合前端條件的 refresh 額外執行交易讀取，只有符合伺服器條件才增加寫入。尚未加入會延遲撤銷生效的本機快取。

## 部署與 TTL

沿用 GOOGLE_CLOUD_PROJECT 與 FIRESTORE_DATABASE_ID；缺少 project 時 Session 回 503；MINI App Channel 由 LINE_MINI_APP_ENV 選定的環境變數提供。runtime service account 須可讀寫指定資料庫；Go server SDK 使用 IAM，不以 Security Rules 取代應用授權。

firestore.indexes.json 宣告 sessions.expiresAt TTL 與索引豁免，其他 Session 欄位不建索引。提交設定不代表雲端已套用。TTL 通常在過期後 24 小時內清理，過期文件仍可能被查到；清理失效只影響保留時間與費用，不延長登入效力。

[索引宣告](../../../firestore.indexes.json) · [Firestore 整合索引](README.md)
