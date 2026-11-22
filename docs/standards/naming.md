# 業務實體、集合與 API 命名

新增或修改模型、OpenAPI、資料集合與查詢欄位前，先確定名稱指的是實體、集合、操作還是能力。保留已發布契約與已儲存資料的名稱；必要更名須有遷移方案及相容性驗證。

| 對象 | 規則 | 範例 |
| --- | --- | --- |
| 實體、資料模型 | 型別用單數，依語言用 PascalCase 或 MixedCaps | `Issue`、`Discussion`、`Repository` |
| 集合 | 使用複數，大小寫依語言或儲存契約 | `issues`、`discussions`、`repositories` |
| REST 集合與成員 | 集合用複數小寫，多字用 kebab-case；成員接在集合路徑下 | `/issues`、`/repositories/{id}` |
| 操作與 operationId | 用動詞表達行為，受詞依單筆或集合命名 | `getRepository`、`listRepositories`、`approveCorrection` |
| DTO 與能力 | 依分頁、命令、結果或業務責任命名 | `RepositoryPage`、`CorrectionSubmission`、`identity`、`attendance`、`worktime`、`summarization` |

既有單例 `/api/v1/session`、operationId、產生型別、collection 與欄位屬相容性契約。第三方 payload 沿用官方名稱，不規則複數依語意判斷。API 以 [OpenAPI 來源及驗證流程](../../api/README.md)為準；GitHub 資源與[跨層命名提案](../proposals/platform/package-layout.md)僅供設計參考。驗收時核對實體／集合與路徑、operationId 的對應，記錄既有例外。

## Go 命名

| 對象 | 專案慣例 |
| --- | --- |
| package 與檔案 | 套件名短、小寫，避開底線及 common、utils 等無法辨認責任的名稱。匯出 API 避免重複套件名，如 `group.Service`。檔名按責任用 snake_case，如 `handle_postback.go`；測試加 `_test.go`，fixture 放鄰近 testdata/。 |
| 型別、函數與方法 | 匯出用 MixedCaps，私有用 mixedCaps。方法用動詞加必要受詞，如 `Claim`、`RenewLease`；只有 receiver 已明確時才用 Execute 等通稱。縮寫保持一致，如 `HTTPClient`、`channelID`、`requestURL`。 |
| 建構與介面 | 多個型別時用 `NewType`，單一主要型別可用 `New`；初始化可能失敗時回傳 error。介面由使用端依能力命名，如 `EventStore` 或 `ReplySender`，不加 I 前綴，也不用為每個 struct 建介面。小套件可用 ports.go，介面多時按能力分檔。 |
| 狀態、時間與錯誤 | bool 用正向名稱，如 `CanManage`、`IsExpired`。I/O 方法首參數為 `ctx context.Context`。error 最後回傳，以 `%w` 包裝並保留 `errors.Is/As` 語意。時間點用 UTC 與 `time.Time`，時長用 `time.Duration`。 |

出勤的 `ResolvePunchAttempt` 可能保存拒絕收據，`SubmitPunch` 提交定位操作；`verifyAndRebuildRecords` 驗證證據並重建紀錄，`RebuildSessions` 重播事件產生班次，`RecordsOverlappingMonth` 含跨月重疊。複雜交易可用具名命令表達身分、時間單位與模式；短方法不必包裝。資料層在公開方法轉譯技術錯誤，避免為每個方法新增轉呼叫的同名 Operation。[出勤套件](../../backend/internal/business/attendance/README.md)

## 目前檔案命名

Go 檔用 snake_case，文件、Shell／Python 腳本及 Web JS／CSS 依專案慣例用 kebab-case。不同語言依各自慣例；不依檔案行數拆分。下列是現行入口：

- [出勤實體](../../backend/internal/business/attendance/domain/entity/record.go)、[儲存 Port](../../backend/internal/business/attendance/application/port/store.go)、[事件序列驗證](../../backend/internal/business/attendance/domain/service/event_stream.go)、[打卡啟動憑證](../../backend/internal/business/attendance/adapter/firestore/attendance_launch.go)。
- [摘要持久化](../../backend/internal/business/summary/adapter/firestore/summary.go)與[模型整合](../../backend/internal/business/summary/adapter/summarizer/summarization/client.go)分開；後者按 client、Gemini 與 OpenAI 的責任分檔。
- 前端的[共同版型](../../frontend/miniapp/src/shell/Shell.tsx)、[身分生命週期](../../frontend/miniapp/src/identity/Session.tsx)、[請求取消與過期回應](../../frontend/miniapp/src/shared/http/lifecycle.ts)及[出勤畫面](../../frontend/miniapp/src/business/attendance/README.md)分別由其模組管理。
- [Go 執行入口](../../backend/main.go)辨識 `check-firestore` 與 `sync-line-navigation` 兩個子命令；變更名稱時需一併核對[部署設定](../operations/container-build.md#容器建置契約)及呼叫端。
