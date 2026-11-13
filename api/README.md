# HTTP API 契約

`contract/openapi.yaml` 是公開 HTTP 契約的唯一組合入口。Identity、Repository、Issue、Discussion 與 Attendance 各自擁有操作、參數與資料格式；共用錯誤與認證方式位於 `contract/common/`。完整歸屬與引用規則見 [契約能力索引](contract/README.md)。兩端產生器先使用 Redocly bundle 解析外部參照，再由同一份 bundle 產生 Go 與 TypeScript；不手改產生檔。

```bash
bash scripts/tooling/tool-exec.sh mise run install-web
bash scripts/tooling/tool-exec.sh pnpm api:generate
bash scripts/tooling/tool-exec.sh pnpm api:check
bash scripts/tooling/tool-exec.sh pnpm api:compat --base origin/main
```

Go 使用生成路由與 HTTP DTO，業務核心不 import 生成套件。TS 端以 api/client 的 createAPIClient 呼叫，paths/components/operations 型別都來自契約。瀏覽器查詢與互動控制器已使用 React hooks，共用 SessionProvider 的認證請求；見 [React 前端生命週期](../docs/architecture/react-frontend.md)。

## 相容性與 CI

api:check 同時檢查產生碼同步與 oasdiff 破壞性變更。compatibility/baseline 保存初次完整契約，避免尚未提交的初次遷移漏掉既有路徑；後續不應隨意更新它。CI 另外使用 pull request base SHA 或 push before SHA，從該提交讀取契約及參照檔再比較，不依賴可被本次修改的 baseline 取代歷史。缺少指定 base revision 時明確失敗；base 尚無 OpenAPI 時仍檢查初次 baseline。

設定／身分、Session、Repository 與出勤操作（含請假登記）以 `contract/openapi.yaml` 及其參照檔為準，不在本頁重複維護數量。範圍是自家 MINI App HTTP API；LINE webhook、內部摘要工作與獨立定位示範不屬於此公開契約，不複製 LINE 官方 API。

## 保持的資料語意

- 工時新增出勤、暫估合計與確認狀態；既有 Minutes 仍為已確認淨工時。EstimatedMinutes 已含 Minutes，不可重複相加，見 [工時計算](../docs/features/worktime.md)。
- 原有大小寫保留：班次回應為 RecordID/ClockInAt 等；更正輸入為 recordId/start 等。
- null 與空陣列分開描述並保留。時間為 RFC3339；月份依 Asia/Taipei 解讀，跨月計算與完整摘要仍由出勤核心決定。
- 座標使用 double，避免 Go 生成 float32 導致精度改變。工時 Duration 保留既有奈秒整數格式，Minutes 是已計算分鐘。
- 明細使用 opaque cursor，預設 20 筆，上限 100；月摘要獨立，不因分頁縮減。既有空 pageSize/cursor 仍代表預設值。
- 新客戶端應為同一更正草稿保存 Idempotency-Key，重試沿用；相同使用者／鍵／內容回原結果，內容不同回 409。舊客戶端可不帶鍵，但沒有跨請求防重保證。
- 既有 groups 與 monthly 非分頁端點保留並標為 deprecated，不新增相容路由。

## 驗證邊界

出勤請求先通過 CSRF 與身分驗證（Cookie Session 或既有無 Cookie Bearer），再做 OpenAPI 結構驗證，最後進入既有授權與用例。已驗證身分僅快取於該請求 context，避免一個請求重複呼叫身分服務；Session 的跨請求狀態另存 Firestore，不等同於 request context 快取。未設定出勤時保留 503。HEAD 沿用 GET 語意。

OpenAPI 驗證不取代班次版本、群組成員、審核權限、位置時效與原子交易。位置／申請內容與 token 不會出現在 validator 錯誤訊息。保留 512／2048／16384 bytes 的既有請求上限。

執行期驗證請求；回應以測試驗證契約與新舊 DTO JSON 一致，沒有在每次正式回應上增加緩衝驗證。標準 CSRF 拒絕維持 text/plain 403，已列在寫入端點契約。

## 後端實作位置

生成碼與通用接線保留於 backend/internal/httpapi；出勤 handlers／DTO 映射與 runtime 驗證位於 business/attendance/http，身分入口位於 identity/http。核心型別不含 JSON／Firestore tags，資料層負責文件編碼與錯誤轉譯。

Redocly lint／bundle 設定位於 `api/redocly.yaml`；根目錄的生成腳本明確帶入 `--config`，不依賴工作目錄自動尋找。

## 目錄導覽與維護

contract/ 依能力維護來源契約，openapi.yaml 組合公開 HTTP API；client/、codegen/、compatibility/ 分別保存共用客戶端、生成設定及相容性基準，不按業務複製這些工具。

先修改契約，再由根目錄 pnpm api:generate 更新 Go/TS 產生碼；不手改 generated 檔或任意刷新 baseline。業務核心不依賴 HTTP DTO。

從根目錄執行 bash scripts/tooling/tool-exec.sh pnpm api:check。歷史比較依 README.md 指定 base revision；缺少 base 不宣稱通過。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)

## 巢狀指引

- [契約來源](contract/AGENTS.md)：能力巢狀指引見 [契約索引](contract/README.md#巢狀指引)。
- [共用客戶端](client/AGENTS.md)。
- [生成設定](codegen/AGENTS.md)。

Issue 基礎操作、Discussion 基礎主題／留言及 Repository 標籤目錄已接入；Project／items 與 Repository milestones／statistics 已依基礎規格接入；Project fields／views 仍為雛形；既有 compatibility/ 及 baseline 保留。

## 相關文件

- [文件索引：API](../docs/README.md#按模組閱讀)
- [Session 與滑動續期](../docs/architecture/browser-session.md)
- [HTTP 契約設計](../docs/architecture/http-contracts.md)
- [資料契約](../docs/architecture/data-contracts.md)
- [出勤查詢與更正](../docs/features/attendance-queries.md)

- [群組請假登記](../docs/features/leave.md)

- [Business 模組與契約對照](../docs/development/business-refactor.md)

## Business 入口分類

公開 HTTP 能力為 Identity、Repository、Issue、Discussion、Attendance；Summary、Weather 描述既有事件／內部工作入口，不加入公開 codegen，也不新增 HTTP 端點。各能力來源與依賴見 [契約能力索引](contract/README.md)。

- [群組 Repository](../docs/features/repository.md)
- [空間標籤目錄](../docs/features/labels.md)：目錄列出／建立，分類沿用 Issue／Discussion 更新；不含改名或刪除。
- [空間事項第一階段](../docs/features/issues.md)：已接入公開 Issue 契約與生成型別；正式環境待驗收。
- [空間討論與留言](../docs/features/discussions.md)：主題及單層留言契約，作者編輯。

- [Repository、Issue 與 Discussion 模型及 Connection 分頁規劃](../docs/architecture/domain-model.md)：未來設計參考，現行 OpenAPI 契約維持原狀。

- [里程碑、專案與統計](../docs/features/planning.md)：正式契約與三層接線。

## 協作能力分佈與交付狀態

Issue、Discussion／留言、Labels、Milestones、Project／items 與 Repository statistics 的本機契約及跨層對照見 [能力分佈](../docs/development/business-refactor.md#協作能力的本機分佈2026-11-13)。其中包含尚未提交來源；Project fields／views 仍為 scaffold。API 生成同步、相容性與正式環境驗收須有各自證據，不能由契約或本文件存在推論。
