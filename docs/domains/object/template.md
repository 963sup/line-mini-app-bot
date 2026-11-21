# ObjectTemplate 最小閉環

ObjectTemplate 是 Object 領域內的獨立聚合，歸屬單一 Repository。現行流程為「建立不可變 v1 → 選用 → 原子建立獨立 Object 樹 → 開啟結果」。範本核心不依賴 Object 實體、HTTP 或 Firestore；應用層協調兩種模型。

## 操作與權限

Actor 取自既有 Session／Bearer 驗證後的 request context，前端不能指定建立者。所有操作沿用 Object 的 Repository 成員存取規則，由 `Repositories.Get` 確認目前成員與有效範圍；不是只有範本建立者能套用。

| 操作 | 輸入與結果 |
| --- | --- |
| 建立範本 | 名稱與節點樹，立即保存為不可修改的 v1，回傳完整快照。 |
| 列出／讀取 | 同 Repository 的範本，列表每頁20筆，使用綁定操作者與範圍的 opaque cursor。 |
| 套用範本 | 範本 ID、重試鍵及選填父 Object ID，回傳新根 ID、所有 Object ID 與已建立的節點關係。第一版只接受 v1。 |

公開格式與錯誤回應以 [Object 契約](../../../api/contract/object/objects.yaml) 為準。路徑為 `/api/v1/repositories/{repositoryId}/object-templates`、`/{id}` 與 `/{id}/instances`。

## 輸入與領域規則

- 範本名稱最多120字。節點 ID 在版本內唯一，為1至64字的英數、底線或連字號。每版只有一個根節點，其餘父節點必須存在，所有節點都可從根抵達，禁止循環。節點清單順序保留在版本與套用結果中；一般 Object 列表仍沿用既有排序。
- 每棵樹最多100個節點，正規化後的建立內容 JSON 最多48,000 UTF-8 bytes，HTTP 本體上限沿用64 KiB。限制確保快照、收據與整棵樹能在單次交易內處理。
- 節點保存名稱、說明、類型、預設數量與單位。應用層沿用 Object 的精確數量與內容驗證；帶數量的施作需有單位，套用時完成數由後端設為0。範本不保存實際完成數、施工日誌或驗收結果。
- 範本與目標限定同一 Repository。選填父 Object 必須已存在於該範圍。套用時重新配置所有 ID 與父連結，根連至選填父節點；未提供即建立新根。
- 每個 Object 保存不可修改的 `templateId`、`templateVersion`、`templateNodeId`。後續修改進度或內容保留來源，不會改動範本或另一棵實例樹。

## 一致性、重送與失敗

建立範本及套用各有獨立的操作者／Repository／`Idempotency-Key` 收據。相同鍵與正規化內容重送回原結果，內容不同回409 `idempotency_conflict`。套用收據包含原始節點 ID、名稱與父連結；結果回放代表建立時狀態，目前內容由既有 Object 讀取 API 取得。

Firestore adapter 在交易內重新檢查 Repository 的有效容器、範本快照及選填父節點，全部讀取先於寫入。Object 文件與套用收據一起提交；任何錯誤均不留下部分 Object 或成功收據。Repository 成員判定沿用目前服務的交易前存取查詢，未新增另一套授權機制。操作不呼叫 LINE，也沒有雲端部署副作用。

儲存位置均位於 `channels/{channelId}/repositories/{repositoryId}` 下：`objectTemplates` 保存不可變版本；`objectTemplateCreationReceipts` 與 `objectTemplateInstanceReceipts` 保存重送結果；實際節點仍在既有 `objects` 集合。未提供來源欄位的既有 Object 維持可讀，不搬移舊 `subobjects`。

錯誤包括400 `invalid_input`、404 `not_found`、409 `idempotency_conflict`／`version_conflict`，以及結果未確認時的503 `unavailable`。拒絕發生在寫入前或交易中；未知結果須保留原鍵與原內容重送，不能依掃描物件列表猜測成功與否。

## 分層與入口

| 層級 | 責任與位置 |
| --- | --- |
| 前端 | [Templates.tsx](../../../frontend/miniapp/src/business/object/Templates.tsx) 收集名稱、節點名稱與父節點，選用並送出一次套用請求；未知結果凍結輸入、重試沿用原請求。 |
| API | [契約](../../../api/contract/object/objects.yaml) 生成 Go／TypeScript 型別；[HTTP adapter](../../../backend/internal/business/object/adapter/http/templates.go) 映射格式與可信 Actor。 |
| 應用層 | [templates.go](../../../backend/internal/business/object/application/templates.go) 驗證權限與輸入、準備 Object 建立內容，依[儲存介面](../../../backend/internal/business/object/application/port/template.go)寫入。 |
| 領域 | [template/version.go](../../../backend/internal/business/object/domain/template/version.go) 驗證單根樹與不可變版本，不匯入 Object 實體或基礎設施。 |
| Firestore | [templates.go](../../../backend/internal/business/object/adapter/firestore/templates.go) 實作快照、分頁、交易及收據。 |
| 組裝 | [repository.go](../../../backend/internal/bootstrap/repository.go) 注入既有 Repository 服務及 Object store，不由 HTTP adapter 建立資料庫 client。 |

服務方法均以 `context.Context` 為第一參數，沿用呼叫端取消與期限；HTTP 期限12秒。`actor` 來自可信驗證，其他字串與 DTO 是未信任輸入，由應用層驗證。

- `CreateTemplate(ctx context.Context, actor, rid, key string, in dto.TemplateCreation) (dto.Template, error)`：驗證並建立 v1。
- `ListTemplates(ctx context.Context, actor, rid, cursor string) (dto.TemplatePage, error)`：授權後分頁讀取。
- `GetTemplate(ctx context.Context, actor, rid, id string) (dto.Template, error)`：授權後讀取指定快照。
- `InstantiateTemplate(ctx context.Context, actor, rid, id, key, parentID string) (dto.TemplateInstance, error)`：授權、驗證、準備節點並原子套用。

前端以 API 回傳的新根 ID 開啟既有 Object 編輯畫面，並呈現回傳樹的節點連結。第一版 UI 只編輯範本與節點名稱、父子選項；類型、數量、單位可由 API 提供預設，或建立後沿用 Object 畫面補充。階層有效性、權限與身分配置由 Go 決定，前端不連 Firestore。

## 驗證與限制

[應用層測試](../../../backend/internal/business/object/application/templates_test.go)、[HTTP 測試](../../../backend/internal/business/object/adapter/http/templates_test.go)、[Firestore 交易測試](../../../backend/internal/business/object/adapter/firestore/templates_test.go)與[前端互動測試](../../../frontend/miniapp/tests/object-templates.test.tsx)涵蓋最小閉環、拒絕、重送及結果開啟。

```bash
bash scripts/tooling/tool-exec.sh pnpm api:check
bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp check
bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp test
bash scripts/tooling/tool-exec.sh --tool go bash -c 'cd backend && go test -race ./internal/business/object/... ./internal/architecture'
```

交易實測另需啟動本機 Firestore 模擬器，並設定 `FIRESTORE_EMULATOR_HOST` 為 loopback 位址。gcloud 592 的模擬器需要 Java 25+，系統 Java 21 無法執行；本次使用環境已有的 Java 25 執行，未更改專案工具版本。隔離測試各自建立隨機測試專案，不使用 ADC 或正式資料。

實測案例：同一範本套用兩次得到獨立樹；多個相同請求併發只建立一棵；修改一棵的完成數不改另一棵；回應遺失重送仍回原結果；第二筆已緩衝寫入前發生驗證錯誤時整筆回滾。未進行正式環境、LINE 裝置或實體瀏覽器驗收。

Firestore 交易順序依 [Google 官方文件](https://docs.cloud.google.com/firestore/native/docs/manage-data/transactions) 與本專案 `cloud.google.com/go/firestore v1.29.2` 的 `RunTransaction` 實作核對。Context7 本次達配額，未取得該服務的新查證結果。Graphify 用於入口探索，Repomix 打包 Object 與接線範圍，公開契約及文件另以原文核對。

## 後續範圍

草稿、發布審核、新版、退役、跨 Repository 共用及舊 `subobjects` 資料遷移尚未實作。未來若加入新版本，已建立的 Object 仍須保留原來源與獨立生命週期；不因目前領域核心能表達版本而宣稱相關產品流程已接通。
