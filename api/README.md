# HTTP API 契約

管理 API 契約、產生設定與 TypeScript client。修改 API 輸入或輸出時，先更新契約，再重新產生 Go 與 TypeScript 檔案。

契約是介面來源；client 與產生設定供消費端使用，依契約重新產生型別。

## 閱讀與修改入口

先查閱[capabilities.json](contract/capabilities.json)、[redocly.yaml](codegen/redocly.yaml)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[獨立Project](../docs/domains/project/project.md) · [React 前端生命週期](../docs/architecture/runtime/react-lifecycle.md)

工程群組／Project新增獨立入口；既有Repository-scoped Project路由由相容adapter呼叫同一用例，來源識別不再代表專案擁有權。資料與授權／相容界線見 [獨立Project](../docs/domains/project/project.md)，契約及生成碼遵循下方連結的生成流程，不手改generated檔。

[contract/openapi.yaml](contract/openapi.yaml) 是公開 HTTP 契約的唯一組合入口。各能力維護自己的操作與格式，共用錯誤與認證位於 `contract/common/`；歸屬見[契約能力索引](contract/README.md)。不手改產生檔，業務核心不 import 生成套件。

契約生成與檢查命令見[契約生成流程](../docs/development/tooling/api/change-api.md#契約生成流程)。

瀏覽器控制器與 SessionProvider 的認證請求見 [React 前端生命週期](../docs/architecture/runtime/react-lifecycle.md)。

## 相容性與 CI

api:check 驗證目前的 OpenAPI 契約、路由與產生碼同步。

設定／身分、Session、Repository 與出勤操作（含請假登記）以 `contract/openapi.yaml` 及其參照檔為準，不在本頁重複維護數量。範圍是自家 MINI App HTTP API；LINE webhook、內部摘要工作與獨立定位示範不屬於此公開契約，不複製 LINE 官方 API。

## 保持的資料語意

- 工時新增出勤、暫估合計與確認狀態；既有 Minutes 仍為已確認淨工時。EstimatedMinutes 已含 Minutes，不可重複相加，見 [工時計算](../docs/domains/attendance/worktime.md)。
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

生成碼與通用接線保留於 backend/internal/transport/http/api；出勤 handlers／DTO 映射與 runtime 驗證位於 business/attendance/adapter/http，身分入口位於 identity/http。核心型別不含 JSON／Firestore tags，資料層負責文件編碼與錯誤轉譯。

Redocly lint／bundle 設定位於 `api/codegen/redocly.yaml`；根目錄的生成腳本明確帶入 `--config`，不依賴工作目錄自動尋找。

## 目錄導覽與維護

[契約來源](contract/README.md)按能力維護，[客戶端](client/README.md)與[生成設定](codegen/README.md)共用。契約變更依[生成流程](../docs/development/tooling/api/change-api.md#契約生成流程)驗證。若須對歷史版本做相容性比較，明列比較的 revision 與差異；目前的 `api:check` 驗證工作樹內的契約與生成結果。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)

## 巢狀指引

- [契約來源](contract/AGENTS.md)：能力巢狀指引見 [契約索引](contract/README.md#巢狀指引)。
- [共用客戶端](client/AGENTS.md)。
- [生成設定](codegen/AGENTS.md)。

Issue 基礎操作、Discussion 基礎主題／留言及 Repository 標籤目錄已接入；Project／items 與 Repository milestones／statistics 已依基礎規格接入；Project fields／views 仍為雛形。

## 相關文件

- [文件索引：API](../docs/README.md#按模組閱讀)
- [Session 與滑動續期](../docs/contracts/identity.md)
- [HTTP 契約設計](../docs/contracts/http.md)
- [資料契約](../docs/proposals/firestore-data-contracts.md)
- [出勤查詢與更正](../docs/domains/attendance/queries.md)

- [群組請假登記](../docs/domains/attendance/leave.md)

- [目前模組與責任](../docs/architecture/runtime/code-layout.md)

## Business 入口分類

公開 HTTP、事件與內部工作依[契約能力索引](contract/README.md)判定；Summary、Weather 的事件及內部工作維持原入口。

- [群組 Repository](../docs/domains/repository/README.md)
- [空間標籤目錄](../docs/domains/repository/labels.md)：目錄列出／建立，分類沿用 Issue／Discussion 更新；不含改名或刪除。
- [空間事項第一階段](../docs/domains/repository/issue/design.md)：已接入公開 Issue 契約與生成型別；正式環境待驗收。
- [空間討論與留言](../docs/domains/repository/discussion/design.md)：主題及單層留言契約，作者編輯。

- [Repository、Issue 與 Discussion 模型及 Connection 分頁規劃](../docs/proposals/platform/domain-model.md)：未來設計參考，現行 OpenAPI 契約維持原狀。

- [里程碑](../docs/domains/repository/milestones.md)、[Project](../docs/domains/project/project.md)與[洞察（Insights）](../docs/domains/repository/insights.md)：正式契約與三層接線。

## 協作能力分佈與交付狀態

已接線與規劃中的操作以[能力清單](contract/capabilities.json)為準，各操作的請求與回應見 [OpenAPI 契約](contract/openapi.yaml)。

[現行實作圖：基礎入口](../docs/contracts/http.md#基礎入口)
