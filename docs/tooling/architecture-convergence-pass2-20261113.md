# 架構收斂第二輪：Project 來源故障隔離

基準為 `ded8538885b669ddea5c5f6296d3b6cc0589e599`，2026-11-13 UTC。開始fetch後，work／遠端work／main相同且工作樹乾淨，無待推送內容。本輪不重做[上一輪](architecture-convergence-20261113.md)的Membership搬移與事件decoder。

## 範圍、方法與基準

先完成全量Repomix與Graphify隔離快照，再以同一統計腳本產生有向imports、static direct calls、SCC及fan-in／out。排除、module、literal TS imports、非多型callgraph及變更半徑代理的定義，完全沿用上一輪；Graphify無向EXTRACTED／INFERRED邊不是runtime依賴。Community採預設且未固定seed，僅記錄觀測值，不據此宣稱品質改善。

基準audit：`/workspace/tooling/tmp/analysis-audit-4gnc3891`。950個分類檔中，Repomix納入524檔、850,114 tokens；其餘依analysis-scope契約排除。缺漏／多收檔案、重複node IDs、dangling endpoints均0。Repomix 1.18.1、Graphify 0.9.103、callgraph 0.52.0；腳本`/workspace/tooling/tmp/architecture-metrics.py` SHA為`59f02de3c39dd1d1152f8fe799348b9a128475285e1b54c688fc78a6a1961cd4`。輸入及hash保存於`architecture-pass2-before-metrics.json`。

## 先排序，再修正

| 優先 | 判定／根因 | 決定 |
| --- | --- | --- |
| P1，確認 | canonical Project已通過Space授權並讀取／更新自有metadata，仍必須成功hydrate每個來源Issue才回應。來源資料消失或故障會使Project讀取失敗；metadata更新已提交也可能回錯誤，讓來源技術故障重新成為Project可用性的父容器 | 精確分類來源hydration錯誤，canonical回自有metadata及`sourceAvailable=false`；自有資料錯誤、取消／逾時仍失敗，legacy保留原錯誤語義 |
| P2，延後 | 五份HTTP Register約287行相似，但有CSRF、認證先後、Project items防重例外及領域錯誤差異；本輪未確認可觀察故障 | 不為DRY建立通用HTTP框架 |
| 不修改 | 前端Project使用Space API、來源資格另外核驗；選取／Session取消隔離與sourceAvailable呈現符合既有契約 | 無已確認前端權威錯置，不新增跨元件狀態機、不改官方元件 |
| 不修改 | 高degree主要為server組裝、identity及傳輸工具；有向來源圖未見非平凡SCC | 不按圖形大小拆模組或合併不同業務 |

依據為[Project操作契約](../features/projects.md)：Space授權決定Project自有資料，來源Repository僅供事項讀取。來源資料不完整時，不揭露來源內容、不用部分進度冒充成功。這是既有邊界的修正，沒有新增Project狀態或權限政策。

## 實際變更與驗證

- [Project用例](../../backend/internal/business/project/project.go)只處理明確的`ErrSourceUnavailable`；canonical回自有metadata並清除部分衍生count，`sourceAvailable=false`時現有前端隱藏進度。沒有宣稱未知總數是零。Legacy及來源項目讀寫不套用此降級。
- [Firestore items adapter](../../backend/internal/business/project/firestore/items.go)分離ProjectItem自有結構decode與來源Issue hydration；只有後者在canonical進度計算時包裝來源錯誤。Unwrap只公開Project來源分類及Unavailable，技術cause保留adapter私有欄位，遵守[錯誤邊界](../architecture/backend-boundaries.md)；Project container／marker／item損壞，以及hydration取消／逾時維持失敗。
- [Service回歸](../../backend/internal/business/project/access_test.go)與[本機emulator回歸](../../backend/internal/business/project/firestore/access_test.go)先實際重現失敗：Get／List／Update被來源hydrate阻擋，metadata版本2已提交卻回失敗；刪除被引用Issue後GET回invalid input。修正後兩組focused tests通過，另驗證legacy原始ErrInput、來源項目拒絕、只公開領域錯誤分類、自有item損壞及取消／逾時不被吞。

只改上述4個Go檔案與本報告／功能契約／Backend README／兩層文件索引。Production淨增31行、維護測試增144行。前端、OpenAPI／生成碼、依賴、資料結構及遷移CLI均未修改；沒有新框架或雙寫。實際功能修改只涉及Project一個business module，不能據此宣稱全庫功能的修改模組數下降。

Focused tests先失敗再通過，之後Go fmt、lint（0 issues）、module verify、vet、完整`go test -race ./...`、build及vuln全部通過。Race suite帶`FIRESTORE_EMULATOR_HOST=127.0.0.1:8787`，使用既有隔離本機project及每測試唯一channel；未連正式資料。Vuln無應用可達問題，仍有3筆required-module項目未被應用呼叫。API完整生成／能力／相容檢查及對ded8538的顯式相容檢查、設定檢查均通過。前端／契約未修改，未重跑無關前端build；官方UI檔案diff為空。

## 前後指標

| 指標 | 基準 | 修改後 |
| --- | ---: | ---: |
| Graphify nodes／edges／communities | 4,792／12,342／286 | 4,801／12,374／285 |
| Go import packages／edges | 65／172 | 65／172 |
| Go跨module／跨business imports | 96／9 | 96／9 |
| Go import非平凡SCC | 0 | 0 |
| Go static direct-call nodes／edges | 646／1,144 | 647／1,145 |
| Go跨module／跨business direct calls | 230／26 | 230／26 |
| Go direct-call非平凡SCC | 0 | 0 |
| TS static import files／edges | 53／151 | 53／151 |
| TS跨module／跨business imports | 132／5 | 132／5 |
| TS非平凡SCC | 0 | 0 |
| 已確認應合併的重複business規則群 | 0 | 0 |
| group協定decoder位置（上一輪成果） | 1 | 1 |
| 全量Repomix收錄files | 524 | 525 |

本批修正可用性邊界，不減少package拓撲；增加helper／測試的Graphify節點不等於架構惡化。Import最高fan-in仍identity12、Repository11、HTTP transport11；最高fan-out server46、Attendance HTTP10、Firestore9。Direct-call最高仍WriteAPIError40、WriteJSON40、ValidID28；server.init46、Run31、RecordsPage12。沒有runtime或效能量測，不能把degree當呼叫頻率。

跨business的9條Go imports仍是Issue／Discussion對Repository及Labels既有契約，以及Project對Space；Project核心沒有Repository import。5條TS跨business邊是Repository組合Issue／Discussion及其Label／Milestone picker。靜態核心可達module代理仍為Repository／Space各1，Project＝Project＋Space2，Issue／Discussion各2，Attendance／Summary各1；未把正常port／adapter或過渡migration誤算成新的父子業務。

修改後metrics保存於`/workspace/tooling/tmp/architecture-pass2-after-metrics.json`，含同腳本inputs hashes。較早一次audit因README連結並行更新而被staleness檢查正確拒絕，未冒稱通過；獨立review另要求技術cause維持私有，已修正並重驗。報告及索引定稿、明確stage後，再用同一full audit驗證最終來源，結果路徑保存於`/workspace/tooling/tmp/architecture-pass2-final-audit.log`。該audit的report.json保留逐檔hash、bundle／graph hashes及實際tokens；不把報告文字增加當成程式改善，也不為自我hash循環改寫報告。

基準bundle SHA：`f7d78e0319940ed14cb10694619b2e34e6b63471b7f0a780c16aa5e2fa5ba296`；修改後完整hash依最終audit的report.json。原始包留於任務暫存區，不提交巨大bundle或個資；重現命令沿用上一輪，artifact前綴改為`architecture-pass2-*`。

## 查證與限制

實際Context7 resolve／query使用官方`/googleapis/google-cloud-go`，核對Firestore交易所有讀取使用同transaction ID與callback重試契約；本機依賴Firestore v1.29.2。Google專用skill／Developer Knowledge MCP未在本session曝光，此處採官方SDK來源；沒有雲端資料驗收。

Go實際resolve為`/golang/go`，查詢go1.26.7與最接近已安裝Go1.27.4的go1.27.3。Context7對errors.Is有文件，但對多重wrap回覆不足；另以已安裝工具鏈`go doc errors.Join`、`go doc fmt.Errorf`核對多原因可供Is／As檢查。沒有把不相關的Context7回覆當作API證明。

本輪是有界來源審查，不是全庫無缺陷證明。Static direct calls不涵蓋interface dispatch／反射，TS圖不涵蓋HTTP／動態載入，Graphify不保證收錄Markdown／TOML。未執行正式LINE操作、資料遷移或部署；既有migration marker及保守rollback保留。

[工具索引](README.md) · [文件索引](../README.md)
