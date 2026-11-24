# 整合架構

外部系統都透過 adapter 接入，核心只看具名能力與穩定錯誤分類。

## LINE

修改 LINE API、Webhook、導覽或業務指令前，可從[後端責任與驗收](line.md)查到對應套件、相容性條件與測試。

LINE Webhook 先以 channel secret 驗證原始 body，再把事件轉成用例輸入。Reply token 只適用單次回覆，不可持久化為一般認證。MINI App 登入由後端驗證 LINE token，後續瀏覽器操作使用應用 Session；LINE 身分不自動等於群組或資源權限。

## Firestore

Firestore 保存各 bounded context 的文件與 Session。Adapter 擁有 document schema、查詢、索引需求與 transaction。交易 callback 可能重跑；外部副作用必須移到提交後。TTL 只負責延遲清理，應用仍需在讀取時驗證到期。

## Google Cloud

- Cloud Run 執行單一 Go service。
- Cloud Build／Docker 建立映像並可部署 revision。
- Cloud Tasks 以 OIDC 呼叫內部工作端點。
- Secret Manager 以明確版本注入 runtime。
- Cloud Storage 保存不適合放入 Firestore 的物件內容。
- Logging、Monitoring 與 OpenTelemetry 提供可觀測性。

## AI 與外部資料

OpenAI、Gemini、CWA 及文件處理能力由 adapter 封裝。輸入需最小化、設定 timeout，並將供應端錯誤轉成可判斷的暫時性或永久性結果。模型輸出是不可信候選，必須經 schema、來源與業務規則驗證後才能寫入。

## 整合契約

每個整合必須定義：

- 身分與授權來源；
- 請求期限、取消與重試；
- idempotency／deduplication 範圍；
- 錯誤分類及降級；
- 敏感資料與 log 邊界；
- 本機、測試環境與正式環境各自需要的證據。

外部系統清單與權威來源見[參考資料](../references/external-systems.md)。

## LINE 群組成員查核

LINE transport 驗簽並解析群組事件。Team 的 LINE 轉接器將事件轉為群組狀態觀測，由 Firestore 轉接器保存。群組與成員候選索引用於探索；舊版群組可見性透過 LINE 平台客戶端核實目前的群組成員身分，正式團隊資源則依產品的團隊成員資格授權。[公開 API 契約](../../api/contract/group/) · [Team 群組查詢](../../backend/internal/business/team/application/query/group.go)

## 瀏覽器定位

瀏覽器以 navigator.geolocation 取得裝置提供的位置，Go 無法直接存取使用者 GPS。 MINI App 用少量 JavaScript 呼叫 API，再將必要欄位提交 Go HTTPS API；不需要 Node.js 服務。 這不是 Google Maps API，不需要 Google Maps key，也不提供地圖、地址反查或可信到場證明。 LINE 訊息的 location／位置 action 與 Web Geolocation 是不同輸入，不混用同意或可信度。

## Object 文件擷取與引文補查

Object 的文件辨識與範本草稿共用[文件準備用例](../../backend/internal/business/object/application/document/preparation.go)。OpenAI 轉接器擷取候選。PDF 候選缺少位置時，Object 透過文字擷取埠讀取原文，最多核對 32 筆引文，只為唯一匹配的頁面補上位置。

### 文件辨識回應與重送

適配器透過 Responses API 在同一個請求中使用 `input_file` 傳送原檔，並以 `text.format` 指定嚴格的 JSON Schema。候選包含名稱、十進位數量字串、單位、來源位置與引文；數量、單位及位置允許 null。完成、拒答、截斷和無效輸出各有對應的結果，呼叫端可檢視模型回應。每次重送都會重新辨識且可能計費。[File inputs](https://developers.openai.com/api/docs/guides/file-inputs?api-mode=responses) · [Structured Outputs](https://developers.openai.com/api/docs/guides/structured-outputs#supported-schemas)

[候選擷取實作](../../backend/internal/business/object/adapter/openai/candidates.go)

## 公開定位診斷端點

- 按下「取得位置並驗證」後才請求瀏覽器權限。成功時顯示座標、精度與採樣時間， 並傳送 POST /api/location/validate；取消、離開頁面或伺服器逾時後忽略遲到 callback， 並中止尚在等待的 fetch。已送出的資料無法撤回。

- 此公開端點只檢查資料，不讀寫資料庫、不關聯使用者、不賦予出勤或資源權限。 不將座標寫入一般應用日誌，回應只包含 valid 與伺服器 receivedAt；設定 no-store。 若日後加入業務操作，必須另接可信身分、CSRF 與資源授權。

POST /api/location/validate，Content-Type 為 application/json，body 上限 2 KiB。 不啟用跨來源 CORS，使用 Go CrossOriginProtection 拒絕瀏覽器跨來源寫入請求。

[定位端點](../../backend/internal/transport/http/geolocation/geolocation.go)
