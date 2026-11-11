# 檔案、套件、函數與契約命名

> 狀態：本專案技術規範；不是所有規則都由 Go 官方強制。新增程式依此命名，既有程式按修改範圍調整。

## Go 命名

| 對象 | 規則 | 範例 |
| --- | --- | --- |
| package | 小寫短名，不用底線與 common/utils；避免輸出 API 重複 package 名 | `group`、`jobs`；`group.Service` |
| 檔案 | 小寫 snake_case，依主要責任命名 | `handle_postback.go`、`claim_outbox.go` |
| 測試 | 對應檔案加 `_test.go`；fixture 放鄰近 `testdata/` | `claim_outbox_test.go` |
| 公開型別／函數 | MixedCaps，名稱表達業務動作 | `UserTarget`、`HandlePostback` |
| 私有函數／欄位 | mixedCaps | `validateTarget`、`leaseVersion` |
| 縮寫 | Go 識別字保持一致 | `HTTPClient`、`channelID`、`userID`、`requestURL` |
| 建構函數 | 多型別時 NewType；單一主要型別可 New | `NewService`、`NewHandler`；初始化可能失敗回傳 error |
| 方法 | 動詞＋必要受詞；不重複 receiver 語意 | `Claim`、`RenewLease`、`MarkAccepted` |
| bool | 正向判斷 | `CanManage`、`IsExpired`；避免雙重否定 |
| interface | 消費端定義、能力命名；不加 I 前綴 | `EventStore`、`ReplySender`、`Clock` |
| context | I/O 方法第一參數、具名 ctx | `Handle(ctx context.Context, cmd Command) error` |
| error | 最後回傳；errors.Is/As 可識別，包裝保留 %w | `ErrForbidden`、`ErrLeaseLost` |
| 時間 | Go time.Time／time.Duration；時間點 UTC | `NextRunAt`、`LeaseUntil`、`PollInterval` |

不要為所有操作取名 Execute/Process/Manager，除非 receiver 已明確表達該使用案例。
不要為每個 struct 強制建立 interface。小 package 可用 `ports.go`，介面多時按能力拆檔。

出勤內部命名以實際行為為準：`ResolvePunchAttempt` 可能保存拒絕收據，`SubmitPunch` 提交定位操作；`verifyAndRebuildRecords` 驗證證據並重建回傳資料，`RebuildSessions` 重播事件產生班次，`RecordsOverlappingMonth` 明確包含跨月重疊。HTTP 路徑和既有資料欄位不隨內部重新命名而變更。
複雜交易用一個具名命令表達身分、時間單位與模式；短方法不強制包成命令。資料層直接在公開方法轉譯技術錯誤，不為每個方法另外建立同名 `Operation` 轉呼叫。


## 目前檔案命名

Go 檔使用 snake_case；文件與 Shell／Python 腳本使用 kebab-case；Web JS／CSS 使用 kebab-case。
不要求不同語言共用同一分隔符，也不依檔案行數強制拆分。

- business/attendance/records/types.go：出勤資料型別；ports.go：使用端的儲存能力契約。
- worktime/event_stream.go：班次事件序列驗證。
- backend/internal/business/attendance/firestore/attendance_launch.go：打卡啟動憑證；business/summary/firestore/：摘要工作持久化。
- backend/internal/business/summary/summarization/summarizer/client.go、gemini.go、openai.go：模型選擇與供應商整合。
- frontend/miniapp/src/shell/Shell.tsx：共同版型；identity/Session.tsx：身分生命週期；shared/http/lifecycle.ts：取消與過期回應；business/attendance/：功能畫面與 hooks。
- cmd 的 attendancemigrate、firestorecheck、richmenu 保留既有命令名，避免無必要改變使用方式。

## 責任對照：規劃檔案，不是已存在 API

| 路徑 | 主要符號／責任 |
| --- | --- |
| `backend/internal/httpapi/webhook/handler.go` | `Handler.ServeHTTP`；驗簽、轉換、可靠接收 |
| `internal/direct/application/handle_postback.go` | `Service.HandlePostback`；個人操作與授權 |
| `internal/group/application/handle_postback.go` | `Service.HandlePostback`；群組操作與授權 |
| `internal/group/application/authorize.go` | `Authorizer.CanManage`；明確 actor 與 group |
| `internal/jobs/domain/outbox.go` | `OutboxJob`、`UserTarget`、`GroupTarget`、工作狀態 |
| `internal/jobs/application/claim_outbox.go` | `Worker.Run` 協調、store.ClaimDue 領取；租約資料型別以 domain 定義 |
| `internal/adapters/firestore/jobs/outbox.go` | `OutboxStore.ClaimDue`、`RenewLease`、`MarkAccepted` |
| `internal/adapters/line/messaging/reply.go` | `Client.Reply`；SDK 型別轉換 |
| `internal/adapters/line/messaging/push.go` | `Client.PushUser`、`Client.PushGroup` |
| `internal/adapters/line/rendering/flex.go` | `RenderFlex`；業務結果轉訊息，不授權 |
| `internal/adapters/garr/execution/pool.go` | 程序內執行與取消，不保存工作 |

具體 signature 先在任務契約確認，不為填滿目錄創建空檔。避免讓 application 同時依賴未定義的工作型別與 adapter DTO。

## 邊界欄位命名

Go `ChannelID` → JSON `channelId` → Firestore `channelId`；schemaVersion 作版本欄位。
LINE 原始 payload 保留官方欄位名，不改寫 webhookEventId 等名稱。
內部 `EventID` 對應 LINE webhookEventId，mapping 明寫在 adapter。
JSON 統一 lowerCamelCase，環境變數依 [LINE 設定](../integrations/line-environment.md)，不可把機密放公開 DTO。

測試函數 `TestService_HandlePostback`，table case 使用行為如 `rejects missing actor`；benchmark `BenchmarkRenderFlex`。
文件名小寫 kebab-case，標題使用具體責任；不使用 final/new/v2-copy。API 版本與 schema 版本各自管理。

依據：Context7 `/websites/go_dev_doc`，2026-11-07 核對 [Effective Go](https://go.dev/doc/effective_go) 的 MixedCaps、package 與建構函數慣例；snake_case 檔案與 JSON 規則是專案選擇。
