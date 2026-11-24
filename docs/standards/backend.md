# 後端規範

後端使用 Go，負責可信身分、授權、業務規則、狀態轉換、資料一致性與外部整合。

## 套件責任

- `internal/business/<context>/domain`：實體、值、政策與純規則。
- `internal/business/<context>/application`：用例協調與使用端 port。
- `internal/business/<context>/adapter`：HTTP、LINE、Firestore 或其他技術轉譯。
- `internal/platform/identity` 提供可信使用者與驗證錯誤；其 `session` 套件管理建立、輪替、到期、續期與撤銷。Firestore 轉接器在交易內核對通道與 CSRF 後撤銷 Session。
- `internal/platform/line` 驗證 LINE token，`lineid` 統一群組及使用者 ID 格式。`internal/platform/google` 驗證服務呼叫的 Google token。資源所有權與授權由業務領域判定。
- `internal/transport/http/identity` 處理 Bearer、Cookie、Origin、CSRF 標頭與 HTTP 回應。`internal/bootstrap` 驗證部署設定並組裝驗證器與 Session 服務。
- `internal/platform`：不擁有業務語意的共用技術能力。
- `internal/transport`：共用傳輸入口。
- `internal/bootstrap`：依賴組裝與程序生命週期。

## Go 慣例

- 套件名稱使用 `lowercase`，例如 `navigation`、`richmenu`。
- 匯出識別字使用 `PascalCase`，例如 `AttendanceRichMenus`。
- 非匯出識別字使用 `camelCase`，例如 `loadAttendanceMenu`。
- 優先使用標準庫與小介面；介面由使用端定義。
- 傳遞 request `context.Context`，外部呼叫必須有期限並正確釋放資源。
- 錯誤保留操作上下文，但不得記錄 secret、access token、原始身分 token 或不必要個資。
- 公開資料格式使用契約產生型別；外部輸入在邊界驗證。
- 依賴使用建構函式明確注入，避免全域可變狀態與不必要的 DI framework。

## 寫入與外部副作用

交易 callback 可能重跑，因此不得直接在 callback 內傳送訊息或呼叫不可重複的外部動作。重要寫入需定義版本衝突、重送鍵、回應遺失與結果未知的處理。外部副作用要有明確的提交點與恢復策略。

## 設定

專案自定義環境變數使用 `SCREAMING_SNAKE_CASE`，例如 `LINE_MINI_APP_ENV`。環境變數只在 bootstrap 或專用設定載入器解析。必填值、型別、允許值與安全預設必須在載入邊界驗證；診斷只能列出設定名稱，不能輸出秘密值。

## 驗證

修改 Go 行為時，至少執行受影響套件測試，並依風險補上 `go test`、`go vet`、建置、race、漏洞或 Firestore emulator 驗證。格式化範圍只限本次擁有的檔案。

## 資料一致性

- MUST define transaction and consistency requirements for state-changing operations.
- MUST protect consistency-sensitive reads and writes against concurrent modification.
- MUST define idempotency, retry, and recovery behavior where applicable.
- MUST preserve persisted data compatibility or provide an explicit migration.
- MUST publish business events only after the corresponding state change succeeds.
- MUST bound reads over growing collections by resource scope and an explicit result limit; paginated reads MUST use stable ordering.
- MUST assign a stable operation or delivery identity to retryable external side effects.

- MUST NOT bypass required transaction or consistency boundaries.
- MUST NOT separate consistency-sensitive checks from their protected writes.
- MUST NOT execute external side effects inside retried transaction callbacks.
- MUST NOT retry non-idempotent writes without duplicate-effect protection.
- MUST NOT silently discard persistence failures.
- MUST NOT modify persisted identifiers or encodings without compatibility handling.
- MUST NOT interpret an unknown operation outcome as confirmed success or failure.
- MUST NOT load an unbounded Firestore collection into memory; growing result sets MUST use pagination or bounded batch processing.
- MUST NOT emit duplicate business notifications for the same committed business event and audience.

- SHOULD NOT introduce repositories or services that merely duplicate lower-level interfaces.
- SHOULD NOT use generic repositories when domain-specific access semantics are required.
- SHOULD NOT introduce caching before identifying a measurable performance bottleneck.
- SHOULD NOT introduce denormalized data without demonstrated query or consistency requirements.
- SHOULD NOT introduce asynchronous messaging for operations adequately served by synchronous execution.
- SHOULD NOT introduce distributed transactions when simpler consistency boundaries are sufficient.

## Go 實作與介面

- SHOULD prefer idiomatic Go and the standard library.
- SHOULD prefer concrete types until an actual abstraction boundary exists.
- SHOULD define interfaces from the consumer's perspective.
- SHOULD use explicit constructors and dependency injection.
- SHOULD keep transport adapters thin and focused on translation.
- SHOULD preserve explicit error semantics and cancellation behavior.
- SHOULD use Firestore transactions or version conditions according to consistency requirements.

## 整合與背景工作

- SHOULD isolate provider-specific behavior behind explicit integration boundaries.
- SHOULD translate external events into domain-specific commands or messages.
- SHOULD prefer existing Google Cloud managed services when they satisfy operational requirements.
- SHOULD use asynchronous processing when justified by latency, reliability, or delivery requirements.

- SHOULD NOT assume events or background jobs are delivered exactly once or in order unless the infrastructure provides an enforceable guarantee.
- SHOULD NOT regenerate an operation identity when retrying the same business operation.
- SHOULD NOT acknowledge durable work before its required persistence or processing guarantee is satisfied.
- SHOULD NOT introduce unbounded retries, polling, or recursive execution.
- SHOULD NOT rely on process-local memory or instance-local locks as the sole authority for durable distributed state or invariants.
- SHOULD NOT rely on a continuously running service instance for required durable execution.
- SHOULD NOT allow stale workers to overwrite results produced by newer operation versions.
- SHOULD NOT remove recovery records while an operation outcome remains unresolved.
- SHOULD NOT silently discard permanently failed background operations.

- MAY introduce caching when measurable performance requirements justify it.
- MAY introduce denormalized projections for validated query requirements.
- MAY introduce asynchronous processing for operations with distinct latency or delivery requirements.
- MAY use Cloud Tasks, Pub/Sub, or Cloud Scheduler according to the operation's execution semantics.
- MAY introduce outbox or inbox patterns when reliable event delivery requires them.
- MAY use batch processing where it preserves required consistency and recovery behavior.
