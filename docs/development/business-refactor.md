# Business 程式碼重構

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

身分／Session 公開契約保留在 openapi.yaml。摘要、天氣不因分類而新增公開 HTTP API。
請假仍是登記，不新增核准、駁回、扣假或工時連動。
Actor、Resource、Authority、Policy 與結果依各既有功能規格及 [操作模型](../architecture/business-operation.md)；本次不另定產品規則。

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
│   └── firestore/
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
