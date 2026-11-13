# Business 模組、契約與實作分佈

## 協作能力的本機分佈（2026-11-13）

依未提交差異重新核對 API、後端、文件與前端的對應。觀察基準為 `work` 的 HEAD `01631be9de257e9b9c32e86736a9f4a8d22a0914` 加當時工作樹；下表包含尚未提交的新檔，描述本機接線，不表示這些功能已隨本文件發布。功能擁有者提交或程式接線改變後，須同步核對本表。

API 路徑相對於 `api/contract/`，後端相對於 `backend/internal/business/`，前端相對於 `frontend/miniapp/src/business/`。來源契約仍由 `api/contract/openapi.yaml` 組合，能力歸屬以 `api/capabilities.json` 為準；本表只提供跨層導覽，不複製操作規則。

| 能力 | API 來源 | 後端核心／HTTP／資料 adapter | 正式畫面／功能規格 |
| --- | --- | --- | --- |
| Issue | `issue/issues.yaml` | `issue/issue.go`、`issue/http/routes.go`、`issue/firestore/store.go` | `issue/View.tsx`；功能規格 `docs/features/issues.md` |
| Discussion／單層留言 | `discussion/{discussions,comments}.yaml` | `discussion/discussion.go`、`discussion/http/routes.go`、`discussion/firestore/{store,comments}.go` | `discussion/{View,Comments}.tsx`；`docs/features/discussions.md` |
| Labels | `repository/labels.yaml` | `repository/labels/{labels.go,http/routes.go}`、`repository/firestore/labels.go` | `repository/labels/{View,Picker}.tsx`；`docs/features/labels.md` |
| Milestones | `repository/milestones.yaml` | `repository/milestones/{milestone.go,http/routes.go}`、`repository/firestore/milestones.go` | `repository/milestones/{View,Picker}.tsx`；`docs/features/planning.md` |
| Project／items | `project/{projects,items}.yaml` | `project/project.go`、`project/http/routes.go`、`project/firestore/{store,items}.go` | `project/{View,Items}.tsx`；`docs/features/planning.md` |
| Repository statistics | `repository/statistics.yaml` | `repository/statistics/{statistics.go,http/routes.go}`、`repository/firestore/statistics.go` | `repository/statistics/View.tsx`，放在正式空間總覽；`docs/features/planning.md` |

### 接線與完成狀態

- **本輪靜態核對完成**：上述能力的 core／adapter 路徑存在；`backend/internal/httpapi/contract.go`、`routes.go` 有相應能力介面與路由，`backend/internal/server/server.go` 註冊各 adapter，`server/repository.go` 組裝服務。`frontend/miniapp/src/business/repository/View.tsx` 在正式空間依頁籤組合上述畫面，非僅存在獨立示範檔。
- **仍為雛形**：`project/fields.yaml`、`project/views.yaml` 在能力清單中為 scaffold；不能由 Project／items 已接線推論自訂欄位與視圖已實作。Labels 僅有 list/create，statistics 僅有 GET；不因此補造改名、刪除或統計寫入能力。
- **文件中的既有驗證**：各功能文件可能記載其任務執行過的 API／Go／前端／emulator 檢查。這次文件整理沒有重跑，不能把文件敘述當作本輪測試通過；程式有變更時由功能擁有者依實際差異補檢查及證據。
- **Git 與正式驗收另列**：這批能力含未提交程式與功能文件。本次僅提交文件導覽；它們是否已納入可交付提交、推送、CI、Cloud Run、真實 LINE／Firestore 與手機驗收，須逐項以實際紀錄確認，不由本表推論。

原有 `repository/preview/Workspace.tsx` 保留為示範，正式入口為 `repository/View.tsx`；共用 `shell/Shell.tsx` 與身分生命週期不因分佈整理而改寫。既有出勤、摘要與天氣維持原來入口；後續功能仍依產品契約及 [分階段落地](foundation.md) 決定，不由目錄數量擴張範圍。

本輪先用 Graphify 定位能力及接線，再以 Repomix 限定包核對87個檔案（197,849 tokens），逐一檢查能力清單、HTTP／server 接線與正式畫面組合。包為問題範圍證據，不宣稱全庫涵蓋；不新增外部技術判斷、產品規則或應用測試。下方保留先前重構的日期、基準及驗證，不能當作這批新增功能的驗證結果。

## 2026-11-11 既有重構紀錄

日期：2026-11-11。基準提交：`fbe7ed30e4fa94b9aa06687902cc478fddbbe9a8`。
本次只重組既有功能，不新增端點、產品規則、資料遷移、套件依賴或雲端資源。
既有未提交的 Business Operation 指引與其他文件修改另外保留。

## 先固定契約

`api/contract/openapi.yaml` 保持唯一入口。出勤原本的單一檔案依既有操作拆為
`attendance/{punch,records,worktime,corrections,approvals,leave}.yaml`。
路徑、operationId、型別名稱、授權、錯誤、nullable／空陣列及 deprecated 語義保持不變。
相容性 baseline 不更新，Go／TypeScript 仍由同一份 Redocly bundle 生成。

拆分前後使用 Redocly `bundle --dereferenced` 解析，完整 JSON 結構必須相等；
再執行生成同步、baseline 及基準提交的歷史相容性檢查。

## 業務、契約、實作與驗證對照

下表後端路徑相對於 `backend/internal/business/`，前端相對於 `frontend/miniapp/src/business/`。
共同身分驗證、Session、傳輸、server 組裝及靜態頁路由保留原本責任。

| 業務 | 既有入口／契約 | 後端核心與規則 | 前端／既有驗證 |
| --- | --- | --- | --- |
| Punch | launch、punches、preflight；punch.yaml | attendance/punch；交易在 attendance/firestore | attendance/punch；HTTP／LINE／交易既有測試 |
| Records | groups、records、monthly、summary；records.yaml | attendance/records，包含既有月報組裝 | attendance/records；分頁、月報與 legacy JSON fixture |
| Worktime | 計算 schema；worktime.yaml，無新增端點 | attendance/worktime | 沿用紀錄畫面；事件重播、工時計算既有測試 |
| Corrections | 申請、查詢、本人取消；corrections.yaml | attendance/corrections；交易在 attendance/firestore | attendance/corrections；版本、防重與更正測試 |
| Approvals | review、approve、reject、reviewer-groups；approvals.yaml | attendance/approvals 的審核資格、核准證據驗證 | attendance/approvals；禁止自審、證據及交易測試 |
| Leave | leaves；leave.yaml | attendance/leave；交易在 attendance/firestore | attendance/leave；既有 HTTP／工作流程檢查 |
| Board | LINE 指令；沒有公開 HTTP 契約 | attendance/board；查詢在 attendance/firestore | 沒有既有獨立畫面；沿用 LINE 檢查 |
| MessageCollection | LINE 文字、編輯、收回事件 | summary/messagecollection，LINE adapter 在其 line/ | 沒有 MINI App 畫面；文字收集、排除指令既有測試 |
| Summarization | LINE 摘要指令、POST /internal/summary（Google 工作身分） | summary/summarization；http、cloudtasks、line、summarizer 在其下 | 沒有 MINI App 畫面；模型、引用、重試與 HTTP 既有測試 |
| Forecast | LINE 群組位置事件 | weather/forecast；cwa、line 在其下 | 沒有 MINI App 畫面；CWA 與 LINE 既有測試 |

本次初始重構時，身分／Session 公開契約保留在 openapi.yaml；後續能力歸屬整理見下節「API 契約能力歸屬」。摘要、天氣不因分類而新增公開 HTTP API。
請假仍是登記，不新增核准、駁回、扣假或工時連動。
操作者身分、資源範圍、授權、業務規則與結果依上表各既有功能規格，並沿用
[既有架構原則](../architecture/principles.md#3-每次功能交付順序) 的使用案例與邊界檢查；本次不另定產品規則。

## 依賴與交易邊界

```text
business/
├── attendance/
│   ├── punch/                 → records
│   ├── records/               → worktime
│   ├── worktime/
│   ├── corrections/           → records、worktime
│   ├── approvals/             → corrections、records
│   ├── leave/
│   ├── board/                 → records
│   └── http/、line/、firestore/、presentation/
├── summary/
│   ├── messagecollection/
│   │   └── line/
│   ├── summarization/         → messagecollection 的訊息模型
│   │   └── http/、line/、cloudtasks/、summarizer/
│   └── firestore/、line/（跨功能組合入口）
└── weather/
    └── forecast/
        └── cwa/、line/
```

attendance 根套件僅保留共享錯誤及原有入口所需的組合儲存契約，不保留模型 alias 或轉呼叫 facade。
Corrections 與 Approvals 為同層套件；審核資格與證據驗證在 approvals，出勤投影規則在 corrections。
核准的權限、版本、出勤更新與 audit 仍在同一 Firestore transaction 執行，禁止拆成多次提交。
因此跨子功能的 HTTP／LINE 組合入口與 Firestore 文件 codec／交易保留於所屬領域的 adapter。
摘要的收集與生成共用版本／撤回狀態，Firestore adapter 同樣保留共同交易邊界。
不為搬目錄建立通用引擎、額外 service 或重複 repository。

前端只移動既有畫面為 `business/attendance/{punch,records,corrections,approvals,leave}`。
既有 `/miniapp/history`、`/miniapp/correction`、`/miniapp/review` 等 URL 不改名。
沒有既有獨立畫面的功能不建立空目錄或新畫面。

## 查證與驗證

- Context7：Go `/golang/go/go1.27.3` 的套件依賴規則；直接或間接自我 import 均非法。
- Context7：Redocly `/redocly/redocly-cli` 的 v2 多檔 `$ref` 與 bundle 流程；本機 CLI 版本以 package.json 固定值為準。
- Next.js：安裝版本 16.4.5 的 `node_modules/next/dist/docs/02-pages/01-getting-started/02-project-structure.md`；保留 Pages Router 與公開路由。
- LINE 官方文件來源預設分支 main，2026-11-11 查得 revision `dea09dc6e63fc5c6a3754abfbb304fc6fce4cc54`；本次只搬移 adapter，不變更 LINE API／事件語義。
- Graphify 用乾淨 AST 重建排除搬移殘留，再以 Repomix 核對來源覆蓋；編譯與既有架構測試驗證依賴。

## 本次實際驗證結果

- 解析後的 OpenAPI 完整結構相等；生成的 Go／TypeScript 檔案完全相同，baseline 與基準提交相容性通過。
- 前端型別檢查、12 項既有測試及完整 pnpm build 通過，既有公開頁面路徑保留。
- Go module verify、vet、完整 race suite、lint（0 issues）、fmt-check 與 build 通過。
- 搬移前後 295 個既有函數（含測試）經 token 正規化比對均保留；忽略套件 qualifier、格式，將抽出的 CanReview 原條件還原比較。ValidLeaveKey 使用內容相同的原 UUID pattern，只有私有變數歸屬改變。
- 既有 architecture 測試改為涵蓋各業務子套件的明確依賴白名單；無通用引擎或跨領域依賴放行。
- govulncheck 在重構前後均回報既有 Go 標準庫弱點 `GO-2026-6859`（Go 1.27.3，修補版 1.27.4）；本次不升級 toolchain／依賴，不宣稱弱點檢查通過。
- 本機已有 Firestore emulator，但其 class version 69（Java 25）與現有 Java 21 不相容，既有 emulator 案例未執行；不以單元測試取代資料庫整合驗收。

本機驗證與產物記錄位於 `/tmp/business-refactor-20261111/`，不納入 Git。
測試通過與外部整合驗收分開記錄：未設定 Firestore emulator 的案例會跳過，
本次不宣稱真實 LINE、模型、CWA 或正式 Cloud Run 已驗收，也不執行部署。

## Summary／Weather 拆分補齊（2026-11-11）

後續核對發現 MessageCollection 的 LINE adapter 仍處理摘要指令及佇列排程，
且依賴包含收集與生成的 Store。本次將既有指令處理移入 summarization/line/commands.go，
messagecollection/line 只負責收集、編輯與收回；summary/line 為兩者的組合入口。
收集使用 messagecollection.Store，指令使用 summarization.CommandStore，worker 使用 summarization.Store。
共用 Firestore 儲存仍實作這些小介面，保留原本交易與版本一致性。
既有 LINE 測試移至組合入口，既有架構檢查增加收集不得依賴生成的限制。

API 分類補上 [Summary 入口契約](../../api/contract/summary/README.md) 與
[Weather 入口契約](../../api/contract/weather/README.md)。公開 OpenAPI、產生碼、
HTTP 路徑、LINE 回覆與前端 URL 均未更動。前端範圍見 [MINI App 說明](../../frontend/README.md#business-畫面範圍)。
Weather 原有 Forecast、CWA、LINE 責任已分離，本次核對其既有測試，不做無必要搬移。

查證：Context7 Go go1.27.3 語言規格的介面嵌入與方法集合；本機 Go 為 go.mod 的 1.27.4。
LINE 官方文件來源 main HEAD 仍為 dea09dc6e63fc5c6a3754abfbb304fc6fce4cc54；只拆分既有事件邏輯，不改 SDK 或事件語意。

本次 Go fmt-check、lint、module verify、vet、完整 race suite、build 通過；
API 產生碼同步與相容性、前端型別、設定檢查通過。govulncheck 無可達弱點，
另回報 required modules 中三項未呼叫弱點；不把此結果描述為所有相依模組無弱點。
未執行真實 LINE、Firestore、Cloud Tasks、模型或 CWA 整合驗收。
本機圖譜、打包與驗證紀錄保存在 /tmp/business-followup-20261111/，不納入 Git。

## API 契約能力歸屬（2026-11-12）

原結構只拆分 Attendance，Identity／Session 與 Repository 仍內嵌根契約；出勤來源反向引用根入口的 APIError，Corrections 與 Approvals 互相引用。這次把來源歸屬與依賴方向整理一致，完整索引及維護規則見 [契約能力索引](../../api/contract/README.md)。

- Identity 擁有設定、本人與 Session 的操作及 DTO；Repository 擁有容器列表、詳情與 DTO。
- common 僅保存共用錯誤與認證方式；openapi.yaml 保留 metadata、路徑與 components 引用索引。能力檔不再反向引用根入口。
- 補卡列表與審核共用的 CorrectionReview、撤回與審核共用的 CorrectionDecisionResult 移到 Corrections，保留型別名稱與內容；Approvals 單向引用 Corrections。
- Summary／Weather 維持既有事件／內部工作契約；共用 client、codegen、compatibility 不按能力複製。

本次僅整理契約來源及文件，不改 HTTP handler、權限、產品規則或資料交易。handler 參數傳遞與 runtime 驗證的統一是另一項行為調整，不能由契約搬移推定已完成。

查證沿用本次 Context7 查得的 Redocly v2 多檔 `$ref`／bundle 文件，本機固定版本以 package.json 為準。Graphify 用於程式關係定位與更新，不能證明所有契約引用；Repomix 限定打包契約、生成腳本、HTTP 接線與相關文件，另以完整解析結構比較確認公開契約語義。

實際驗證：拆分前後完整 dereferenced JSON 相等；Go／TypeScript 重新生成後與原檔完全相同；`pnpm api:check`、`pnpm api:compat --base d395e0023cb719ee5622ecf1b8ed54c9d696ec0f`、設定檢查、變更 Markdown 的本機檔案連結與 `git diff --check` 均通過。程式與文件圖譜均已重建；文件圖譜未解析本機連結為 0。未修改應用行為，因此未重跑完整 Go／前端測試或外部整合驗收。
