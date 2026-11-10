# HTTP API 契約

`contract/openapi.yaml` 是唯一入口，出勤路徑與資料格式位於 `contract/attendance.yaml`。兩端產生器先使用 Redocly bundle 解析外部參照，再由同一份 bundle 產生 Go 與 TypeScript；不手改產生檔。

```bash
bash scripts/tool-exec.sh mise run install-web
bash scripts/tool-exec.sh pnpm api:generate
bash scripts/tool-exec.sh pnpm api:check
bash scripts/tool-exec.sh pnpm api:compat --base origin/main
```

Go 使用生成路由與 HTTP DTO，業務核心不 import 生成套件。TS 端以 api/client 的 createAPIClient 呼叫，paths/components/operations 型別都來自契約。瀏覽器查詢與互動控制器的 React 遷移是後續階段。

## 相容性與 CI

api:check 同時檢查產生碼同步與 oasdiff 破壞性變更。compatibility/baseline 保存初次完整契約，避免尚未提交的初次遷移漏掉既有路徑；後續不應隨意更新它。CI 另外使用 pull request base SHA 或 push before SHA，從該提交讀取契約及參照檔再比較，不依賴可被本次修改的 baseline 取代歷史。缺少指定 base revision 時明確失敗；base 尚無 OpenAPI 時仍檢查初次 baseline。

目前 14 個 HTTP operations：2 個設定／身分端點、12 個出勤端點。範圍是自家 MINI App HTTP API；LINE webhook、內部摘要工作與獨立定位示範不屬於此公開契約，不複製 LINE 官方 API。

## 保持的資料語意

- 原有大小寫保留：班次回應為 RecordID/ClockInAt 等；更正輸入為 recordId/start 等。
- null 與空陣列分開描述並保留。時間為 RFC3339；月份依 Asia/Taipei 解讀，跨月計算與完整摘要仍由出勤核心決定。
- 座標使用 double，避免 Go 生成 float32 導致精度改變。工時 Duration 保留既有奈秒整數格式，Minutes 是已計算分鐘。
- 明細使用 opaque cursor，預設 20 筆，上限 100；月摘要獨立，不因分頁縮減。既有空 pageSize/cursor 仍代表預設值。
- 新客戶端應為同一更正草稿保存 Idempotency-Key，重試沿用；相同使用者／鍵／內容回原結果，內容不同回 409。舊客戶端可不帶鍵，但沒有跨請求防重保證。
- 既有 groups 與 monthly 非分頁端點保留並標為 deprecated，不新增相容路由。

## 驗證邊界

出勤請求先通過既有 CSRF 與 LINE 身分驗證，再做 OpenAPI 結構驗證，最後進入既有授權與用例。已驗證身分僅快取於該請求 context，避免一個請求重複呼叫身分服務；不是跨請求登入快取。未設定出勤時保留 503。HEAD 沿用 GET 語意。

OpenAPI 驗證不取代班次版本、群組成員、審核權限、位置時效與原子交易。位置／申請內容與 token 不會出現在 validator 錯誤訊息。保留 512／2048／16384 bytes 的既有請求上限。

執行期驗證請求；回應以測試驗證契約與新舊 DTO JSON 一致，沒有在每次正式回應上增加緩衝驗證。標準 CSRF 拒絕維持 text/plain 403，已列在寫入端點契約。

## 後端實作位置

生成碼與通用接線保留於 backend/internal/httpapi；出勤 handlers／DTO 映射與 runtime 驗證位於 attendance/http，身分入口位於 identity/http。核心型別不含 JSON／Firestore tags，資料層負責文件編碼與錯誤轉譯。
