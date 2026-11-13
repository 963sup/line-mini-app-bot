# Graphify／Repomix 儲存庫分析與方案

## 已發布基礎與最小收斂（2026-11-13）

先依使用者要求將原先83個待提交檔案驗證後發布：`d26f350b64096cc35f807dbb29ca6371ab8071c7`，work／main均核對一致。包括Issue、Discussion、Labels、Milestones、Project／items與statistics；僅清理兩個TSX檔尾端空白，未把分析／重構混入該提交。因此下方兩輪的「未提交／Discussion缺adapter」是歷史快照，不能當作發布後現況。

再對此乾淨基礎做完整CLI分析：920個分類檔案、494 managed全部由Repomix收錄，774,852 tokens；Graphify 4,378節點、11,029關係、309個source_file來源，missing／extra、重複ID、懸空端點均0，`--verify`成功。固定來源及manifest／hash位於 `/workspace/tooling/tmp/analysis-audit-chayas4s/`。使用Graphify定位、完整XML與原碼交叉核對，分別做後端及前端唯讀檢查；沒有以高度數或重複外觀直接判定缺陷。

### 已確認根因與實際修正

| 優先 | 事實／根因 | 最小修正與證據 |
| --- | --- | --- |
| P1 | Issue與Discussion的Firestore adapter對已轉譯錯誤再次呼叫translate時，`errors.Is`命中後回傳sentinel，丟棄原storageFailure及其私有cause。Create→Get的巢狀邊界會觸發，與[錯誤邊界規範](../architecture/backend-boundaries.md)不符 | 兩處改為回傳原err。隔離in-package overlay原先兩包皆失敗，修後race smoke通過；並檢查公開領域分類仍可匹配、原SDK cause不經Unwrap公開 |
| P1 | Repository根核心未列入既有core map，守門忽略其禁止實作依賴與serialization tags的規則 | 只補 `business/repository: {}`。隔離副本注入httpapi依賴與JSON tag，修前既有檢查誤通過，修後正確拒絕兩項；正式原碼沒有注入檔，完整既有suite通過 |

修正只涉及 [Issue轉譯](../../backend/internal/business/issue/firestore/store.go)、[Discussion轉譯](../../backend/internal/business/discussion/firestore/store.go) 與 [架構守門](../../backend/internal/architecture/boundaries_test.go)。沒有改公開API、授權、交易、狀態、資料格式、依賴或官方UI元件；不建立共用workflow引擎或為消除兩段相似程式新增抽象。正式程式為兩行等量替換，既有測試守門增一行；報告文字另計。

### 查證、驗證與剩餘限制

Context7先resolve Go，再查 `/golang/go/go1.27.3` 的AST／error tree；本機Go1.27.4 `go doc errors.Is`、`errors.As`核對已安裝版行為。後端review另查 `/googleapis/google-cloud-go` 官方transaction語意；Google Developer Knowledge／Google skills無可用入口，未做Google資源操作，也未改Firestore交易。版本及外部查證不代替實際重現。

發布基礎實跑：Go格式、lint、module verify、vet、完整既有race suite、build、vuln；API能力／生成同步／相容性、設定、前端型別、23個既有前端測試及Next靜態build均通過。收斂後重跑Go格式／lint／完整check（含race）／build／vuln及兩包overlay race；API／前端未變，沿用基礎驗證。govulncheck無應用可達漏洞，仍有3筆未被應用呼叫的required-module項目，不稱所有依賴安全。

臨時負向注入及overlay保留在 `/workspace/tooling/tmp/convergence-boundary-probe/` 與 `/workspace/tooling/tmp/backend-review-translation-l4lzfw9t/`，未新增正式測試平台、重現檔或產物至Git。程式修後再執行既有Graphify更新與全量audit，最終CLI輸出／hash另保留於該次新audit目錄，不覆寫前述基準。

前端review重現：留言POST尚未完成時，父層重新讀取討論會卸載留言畫面並丟失該畫面的防重鍵；[討論規格](../features/discussions.md)已明列切題／離頁前須先確認結果。本輪作為殘餘限制，未升格為授權／狀態漏洞或加入新的跨頁持久化。API／HTTP adapter及領域scope有相似程式，但目前責任與錯誤分類不同，沒有足夠證據支持為減行數抽取通用層；補卡分檔亦繼續延後。這是一次有界收斂，不是全庫無缺陷證明，真實LINE／Firestore／手機與部署驗收仍未執行。

## 第二輪：最新快照與前後差異

2026-11-13（Asia/Taipei）重新實際執行兩個 CLI。下列結果優先於後方第一輪的歷史觀察；只更新報告，不實施方案。基準為 HEAD `d3803b68359b46ab50adddf0cdd2224e4c18623a` 的 dirty working tree，並非該提交的乾淨內容。

### 本輪範圍與實測

先執行既有 `analysis-scope.py --audit`：871個分類檔案，445 managed、426 excluded；tracked audit 的 Repomix 無缺檔／多檔，Graphify 3,317節點、7,866關係，完成時 stale=false。因為本輪已有多個新模組尚未追蹤，再明列10個來源加入**獨立分析快照**，不修改正式 scope contract、不 stage 他人的檔案：

- `backend/internal/business/issue/{README.md,issue.go,http/routes.go,firestore/store.go}`
- `backend/internal/business/discussion/discussion.go`
- `frontend/miniapp/src/business/issue/{README.md,View.tsx}`
- `frontend/miniapp/src/business/repository/navigation.ts`
- `docs/features/{issues.md,discussions.md}`

本輪產物於 `/workspace/tooling/tmp/analysis-audit-qofjl2jn/`。`current-source/` 是445 tracked＋10明列新檔的固定來源；`current-manifest.json` 記錄每檔hash，`current-repomix.xml` 及 `current-analysis/graphify-out/graph.json` 是針對同一455檔快照實際重跑的兩個CLI產物，`current-evidence-summary.json` 保存結果與artifact hashes。原始tracked audit保留於同目錄，數值不混用。

| 指標 | 第一輪主分析 | 第二輪455檔快照 | 解讀 |
| --- | --- | --- | --- |
| Repomix檔案 | 444 | 455 | 增11，其中包含舊報告與10個新檔；不是11個新功能 |
| Repomix tokens | 616,606 | 670,034 | 增53,428（約8.66%）；非執行效能指標 |
| 字元 | 1,956,241 | 2,142,423 | XML保留完整來源，不是壓縮摘要 |
| Graphify節點／關係 | 2,883／7,204 | 3,412／8,319 | 增529／1,115；範圍與程式均有改變 |
| 有source_file的來源／communities | 277／206 | 283／209 | 代表可觀測圖譜覆蓋，不能證明動態路徑完整 |
| EXTRACTED／INFERRED | 6,768／436 | 7,817／502 | 本輪約6.03%推測關係，仍需原碼核對 |
| missing／extra、duplicate IDs、dangling endpoints | 全0 | 全0 | XML與455個明列來源精確相符，圖結構自洽 |

工具仍為 Graphify `0.9.103`、Repomix `1.18.1`；Graphify code-only模型token為0 input／0 output。Repomix安全掃描回報No suspicious files detected，不能作無機密或安全性證明。本輪未重新執行應用、效能或渠道測試。

### 已失效的舊結論與仍存在的事項

1. **Issue已超出scaffold，舊說法失效。** [contract.go](../../backend/internal/httpapi/contract.go) 快照L13、L33–38已嵌入IssueServer與四個操作；[routes.go](../../backend/internal/httpapi/routes.go) L47–55註冊list/create/get/update；[server.go](../../backend/internal/server/server.go) L85接入Issue服務。[正式空間View](../../frontend/miniapp/src/business/repository/View.tsx) L10、L64引用正式Issue畫面。因此不再建議「從零開始Issue」或宣稱正式空間只有基本資料。這是本機未提交實作存在的證據，不能直接推論已發布或正式驗收。
2. **Issue具備明確的核心與adapter分工。** [issue.go](../../backend/internal/business/issue/issue.go) L58–82處理建立者、版本及三狀態轉換，L102–110透過Repository服務取範圍；[store.go](../../backend/internal/business/issue/firestore/store.go) L168–184使用共用編號建立callback，L201–245於交易中讀取並套用更新／紀錄。[功能文件](../features/issues.md) 已記載先前本機及emulator驗證，但本輪僅核對文字與原碼，未將他人報告的測試結果當作本輪重跑結果。
3. **Discussion處於進行中的跨檔不一致快照。** [capabilities.json](../../api/capabilities.json) 的discussion與discussion.comments已指向 `backend/internal/business/discussion/http/routes.go`，455檔快照沒有該檔；只有核心 [discussion.go](../../backend/internal/business/discussion/discussion.go) 與功能／契約文件，contractServer也尚未嵌入DiscussionServer。兩筆能力指向同一缺少adapter是已確認的快照差異；不是兩個獨立漏洞，也不代表後續工作樹仍缺檔。應先由該任務完成／核對接線，不能據此把Discussion列為完整已實作。
4. **架構守門有進展，Repository缺口仍未補齊。** [boundaries_test.go](../../backend/internal/architecture/boundaries_test.go) 本輪L20新增 `business/issue → business/repository`；core map仍沒有 `business/repository`，也尚未包含新的discussion核心。原有一般HTTP／transport限制仍存在。這是覆蓋範圍問題，未因此發現既有非法import或序列化tag。

本輪查詢已先經Graphify定位IssueServer、ApplyUpdate、Discussion及Numbering，再由XML／快照原文核對。圖中的 `cn()` degree170、react67仍是正常共用hub；Attendance52、Record50未變。未因新節點增加而要求拆微服務、改寫UI primitives或重構補卡交易。

### 本輪建議的四個方案

| 優先 | 方案與具體邊界 | 取捨 | 驗收證據 |
| --- | --- | --- | --- |
| P0 | **先收斂Discussion進行中接線**：由擁有者對照能力清單、契約入口、核心、adapter、server與正式UI；保留尚未實作的狀態。報告不替其實作 | 小至中；避免先擴充labels／projects等其他功能，期間須承認snapshot不完整 | 每個宣告的adapter／core路徑都存在且納入來源清單；適用能力／生成碼／接線檢查成功；缺項明列而非宣稱完成 |
| P1 | **先封存Issue可交付證據與新檔**：沿用既有功能文件，讓該任務擁有者核對四API、正式UI、資格、重送鍵及版本衝突，再只提交其task-owned檔案 | 小；重用仍適用的既有驗證，變更後才補相應檢查，不盲目重跑所有測試 | 清楚區分程式存在、檢查實跑、Git發布與LINE／Firestore實機驗收；所有import的新檔納入已驗證提交，原碼hash可追溯 |
| P1 | **補齊既有core map的領域覆蓋**：先確認目前Repository與Discussion所有同package檔，再按現行純核心規則維護既有map；保留已新增的Issue規則 | 小；不新增架構平台／萬用規則引擎，也不放寬SDK白名單 | Repository／Discussion與其他核心同受import及serialization tags規則檢查，適用既有架構檢查成功；未決依賴先說明，不直接例外放行 |
| P2 | **將455檔基準與問題小包分工**：保留全量manifest／bundle，日常以Graphify定位後打包本次能力的契約、接線、核心、adapter、UI與文件 | 小；670k tokens不宜每次全讀，但小包不能自稱全庫完整 | 每次列出含入／排除／未追蹤來源與實際tokens；來源hash、XML路徑、圖端點檢查可重現，視覺化只寫副本 |

建議下一步優先核對進行中Discussion的能力接線與Issue可交付證據，再處理核心守門；補卡分檔維持第一輪的條件式後續方案，沒有相關需求就不實作。

### 本輪檢查與時效

已核對455個來源hash、XML完整涵蓋及圖ID／端點；在報告撰寫前的來源重查，455個快照檔與現行來源hash沒有差異。這不鎖定其他任務後續新增的檔案；完成／發布時仍以凍結manifest為範圍。tracked audit的 `--verify` 可用於初始未變動產物，455檔擴充快照另以current manifest核對，不冒稱原script已自動涵蓋untracked。

使用原始graph副本 `current-visualization/graphify-out/` 產生原生GRAPH_REPORT.md及graph.html，避免重寫原始artifact hash。文件本機連結與whitespace另檢查；不將本報告的提交視為Issue／Discussion的發布，也不提交XML、graph或其他任務程式碼。Context7／官方文件中未變的工具查證沿用第一輪，並以本輪實際CLI及原碼驗證差異，未新增外部技術選型。

## 第一輪：歷史快照（保留原始證據）

日期：2026-11-13（Asia/Taipei）。本次實際執行兩個 CLI，產出靜態結構、來源交叉核對與五個優先方案；不實作方案、不修改業務程式或官方 UI 元件，也不部署。

## 結論與建議順序

建議先完成來源快照與能力狀態的對照，再針對具體變更補齊既有架構守門。現有程式已具備模組化單體、生成 API 與能力分工，沒有證據支持全庫重構、拆微服務或建立通用業務引擎。補卡交易與前端預覽是後續變更應特別定位的區域；節點度數及檔案長度本身不是缺陷。

| 順序 | 方案 | 目的 | 規模／取捨 |
| --- | --- | --- | --- |
| P0／A | 以固定來源清單建立可重現分析快照 | 避免未追蹤依賴與並行修改造成假完整 | 小；維持全量基準，日常另用小範圍包 |
| P1／B | 將能力契約、正式入口、預覽及驗收狀態對照 | 避免把介面預覽、scaffold 或規格視為正式功能 | 小至中；先整理現有證據，再決定下一個產品切片 |
| P1／C | 使既有核心依賴守門涵蓋 Repository | 防止新領域繞過現有核心 import／serialization 規則 | 小；只補既有機制，不新增測試平台或抽象層 |
| P2／D | 補卡有實際變更時才按責任拆分 | 降低交易敏感區域的修改負擔 | 中；必須保留原子性、授權及重送語義，暫不動程式 |
| P2／E | 日常按問題打包，保留全量稽核基準 | 降低閱讀成本及模型 context 壓力 | 小；小包不能宣稱全庫涵蓋，依賴問題仍需 lockfile |

## 分析基準與可重現性

- HEAD：`f0ad5713fe31d2573f49d67482729d47383ee1a6`；分支 `work`；輸入為當時的 **dirty working tree**，不是該提交的乾淨 checkout。逐檔 SHA-256、排除理由與工具版本保存在原始 `report.json`。
- 主分析使用 [analysis-scope.toml](../../analysis-scope.toml) 的 Git universe：tracked 檔案減工作樹刪除項，加明列 additions（本次為空）。由 [analysis-scope.py](../../scripts/analysis/analysis-scope.py) 複製來源到獨立目錄，實際呼叫 Repomix 與 Graphify；不將本機工具、憑證或 node_modules 掃入。
- 原始產物目錄：`/workspace/tooling/tmp/analysis-audit-1led4vab/`，包含 `source/`、`report.json`、`evidence-summary.json`、`repomix.xml`、兩工具 log，以及 `graphify-out/{graph.json,GRAPH_REPORT.md,graph.html}`。這是本機暫存證據，不提交原始碼包，也不是永久下載網址。
- `report.json` 是初始 audit 紀錄；後續 `cluster-only` 會重寫圖譜標籤與 graph hash。最終圖、HTML、原生報告及 bundle 的 hashes 以 `evidence-summary.json` 為準；來源快照的逐檔 hash 再核對全部相符。
- 額外明列兩個當時未追蹤檔：`frontend/miniapp/src/business/repository/navigation.ts`、`docs/features/issues.md`，另放 `supplement-source/`、`supplement.xml`、`supplement-manifest.json` 及獨立 supplemental graph。它們不混入主分析的 444 檔完整涵蓋率。
- 分析期間其他工作持續修改 Issue 契約、API 接線與架構檢查。`--verify` 已實際拒絕把舊快照當作當前來源；不為追趕並行工作無限重跑。報告中的程式行號均指 `source/`／XML 的固定快照，導向儲存庫的連結只便於後續定位，現行行號與結論須重新核對。

重跑與核對方式（從儲存庫根目錄執行）：

```bash
bash scripts/tooling/tool-exec.sh python3 scripts/analysis/analysis-scope.py --audit
bash scripts/tooling/tool-exec.sh graphify query "Register AttendanceServer repositoryServer Authenticator" --graph /absolute/new-audit/graphify-out/graph.json
bash scripts/tooling/tool-exec.sh graphify query "Workspace RepositoryProvider route preview" --graph /absolute/new-audit/graphify-out/graph.json
# 需保留初始 graph hash 時，先複製 audit 產物，再在副本產生可視化。
bash scripts/tooling/tool-exec.sh graphify cluster-only /absolute/audit-copy
bash scripts/tooling/tool-exec.sh python3 scripts/analysis/analysis-scope.py --verify /absolute/new-audit/report.json
```

`--verify` 要求 HEAD、來源、分析契約及原始產物均一致；新增本報告、修改索引或提交後，舊 audit 不應再被宣稱為當前全庫驗證。重跑時使用新輸出目錄，勿覆寫他人的分析。

## 實際工具結果

| 項目 | 結果 | 限制 |
| --- | --- | --- |
| CLI | Graphify `0.9.103`；Repomix `1.18.1` | 是本次觀測值；版本配置仍以 mise.toml 為準 |
| 分類 | 870 檔：444 managed、426 excluded | upstream skills、安裝副本、產物與二進位依契約排除，不是空目錄 |
| Repomix | 444 檔、616,606 tokens、1,956,241 字元 | XML 實際路徑與 managed 清單完全相同；missing／extra 均為 0 |
| Graphify | 2,883 節點、7,204 關係、206 communities | 277 個有 source_file 的來源；不等於執行時呼叫覆蓋 |
| 圖完整性 | duplicate IDs＝0；dangling endpoints＝0 | 僅證明結構自洽 |
| 關係可信度 | EXTRACTED 6,768；INFERRED 436 | INFERRED 約 6.05%；不得當作確定呼叫鏈 |
| 成本 | Graphify extraction input/output tokens 均為 0 | `--code-only` 為本機 AST；不包含本次對話與報告撰寫成本 |
| 安全掃描 | Repomix 回報 No suspicious files detected | 不是安全稽核，也不能證明絕無機密 |

Graphify 辨識 290 個 code 輸入；129 個 Markdown 未做語意抽取，另有23個輸入未分類。最終167個 managed 檔沒有節點 source_file，主要為129個 Markdown、12個 TOML、8個 JSON、6個 YAML等；主分析沒有遺漏 `.go`／`.ts`／`.tsx`／`.py` 的來源檔節點。單檔有節點仍不代表所有符號及動態關係已抽取。

Repomix 各區塊檔數：backend 141、frontend 93、docs 82、api 53、scripts 23、根目錄20、.codex 13、.github 7、.vscode 5、deploy 4、assets 3。audit 在隔離來源中依明確清單處理忽略規則並保留安全掃描，與日常直接於工作樹執行 Repomix 的預設範圍不同，兩者數值不可直接比較。

## 結構觀察與來源核對

### 1. API 分工已存在，圖譜導覽不能取代能力實作狀態

Graphify 查詢定位 `Register()`、`RepositoryServer`、`AttendanceServer` 與 `Authenticator`；Repomix 逐段核對：

- [contract.go](../../backend/internal/httpapi/contract.go) L9–15 以 capability interfaces 組成生成契約，L27–29 的 RepositoryServer 只有 list/get；不是每個頁籤都有正式 API。
- [routes.go](../../backend/internal/httpapi/routes.go) L41–44 接 Repository 讀取路由；[Repository HTTP adapter](../../backend/internal/business/repository/http/routes.go) L46–86 先 Authenticate，再呼叫 service，沒有在 UI 取得資料後便省略後端檢查。
- [Repository Service](../../backend/internal/business/repository/repository.go) L57–104 以 ACTIVE 與 Membership 檢查存取；這是靜態存在的程式路徑，未做 LINE membership 實機驗收。
- [能力清單](../../api/capabilities.json) 快照 L76–129 明列 labels、milestones、issue、discussion、project 為 scaffold。Issue 正由其他工作修改，因此此狀態只能描述本次快照，不能用來否定後續成果。

結論：保留既有模組化單體與 generated API，不因圖上存在 Issue／Discussion 契約節點就宣稱功能已上線。

### 2. 正式空間與預覽已區分；下一步應補交付證據

Graphify 將 Workspace、RepositoryProvider、AppFrame 與 UI primitives 連起來；XML 核對：

- [正式 View](../../frontend/miniapp/src/business/repository/View.tsx) 快照 L35–68 顯示真實空間基本資料、本人出勤入口及尚未開放的功能。
- [RepositoryContext](../../frontend/miniapp/src/business/repository/RepositoryContext.tsx) L42–68 透過 typed client 呼叫正式 list/get API；L74–83 取消舊請求並清空前一 owner 的資料。
- [預覽 Workspace](../../frontend/miniapp/src/business/repository/preview/Workspace.tsx) 共577行，L52 seed、L85–98 useState、L178 示範資料提示、L429 重整還原提示與 L443 setItems 都表明本地預覽。沒有證據支持「假成功冒充持久寫入」的缺陷判定。
- 正式 View L8 import `./navigation`；該檔當時未追蹤，不在主 audit，但已透過 supplement XML 核對是共用導覽資料。初始工作樹圖比隔離圖多2個節點、4條關係，source 差異正是此檔，不能將這種差異稱為 stale graph 或刪檔。

結論：這是功能就緒度與分析範圍的管理事項；沒有理由把預覽資料直接接成正式操作，更不改寫官方 shadcn primitives。

### 3. Repository 核心尚未列入既有依賴守門的 core map

[依賴規範](../architecture/dependencies.md) 要求核心只依賴標準庫與明確核准的純核心子套件。[boundaries_test.go](../../backend/internal/architecture/boundaries_test.go) 快照 L19–33 的 core map 包含 attendance、summary、forecast、identity，缺少 `business/repository`；L53 的 `core[dir]` 使用精確目錄比對，L87–96 的 serialization tag 檢查也只作用於該 map。

因此 Repository HTTP adapter 仍受一般 HTTP 規則檢查，但 Repository 根核心未受到 map 限制的「禁止外部 import」與「禁止 serialization tags」兩項守門。已核對 [repository.go](../../backend/internal/business/repository/repository.go) 的 import／struct：當時只有標準庫且無此類 tags，**未發現現存違規**。這是守門覆蓋缺口；分析期間 boundaries_test.go 已由其他工作修改，實施前先確認是否已補齊。

### 4. 補卡資料 adapter 是敏感變更區域，長度不構成重構理由

Graphify hubs 包含 Attendance（degree 52）及 Record（50）；對應 [attendance_correction.go](../../backend/internal/business/attendance/firestore/attendance_correction.go) 共541行，處理提案、決定、查詢、分頁與稽核解碼。L232–334 的 `decideCorrection` 同一交易內處理權限、舊決定、核准重送、重疊／狀態檢查、有效紀錄與 audit。

[操作模型](../architecture/business-operation.md) 要求授權與重要寫入協調，並保留重送／併發語義。此處是後續改動需要先釐清的責任集中點；本次沒有驗證交易實際行為，也沒有發現必須立即拆分的缺陷。不得只為降低圖度數，把檢查搬出交易或引入跨服務流程。

### 5. 高度數多半反映正常共用；全量閱讀成本已很高

Graphify God Nodes 的前幾名為 `cn()` degree 170、react 66、Attendance 52、Record 50、Job 38、WriteAPIError 38；Workspace degree 31。`cn()` 是官方 UI 廣泛共用工具，高度數不表示應拆分或改寫。

原生分析 cohesion 是圖結構密度而非品質分數：community 0＝0.050520566649598905、10＝0.09397163120567376、14＝0.13063063063063063、19＝0.10752688172043011。不同規模及生成內容會影響數值，沒有設定合格門檻。

Surprising Connections 中 Session→generated components、SessionProvider→createAPIClient 已可由 import／呼叫核對，是預期的 typed client 使用。`TestMiniAppAttendanceGetsDedicatedPage→serveMiniApp` 被標為 INFERRED，不能據此宣稱已測試該路徑。可繼續追問：RepositoryProvider 的 typed client 呼叫如何經能力路由到 Service 的 membership 檢查；本次不把未完整 trace 的整段圖當作確定執行鏈。

Repomix 最大項為 lockfile 95,778 tokens（15.5%）、generated Go API 21,673、generated TS schema 13,995、API compatibility baseline 10,802。它們各有依賴及契約價值，應放在全量基準；日常針對單一問題不必將616k tokens 全部放入模型。

## 五個方案的實施條件與驗收

### A／P0：先固定證據再作決策（建議第一步）

使用現有 analysis-scope audit；對當次需要的未追蹤新檔明列 supplemental manifest，或由其擁有者先納入正式範圍。記錄 HEAD、dirty、每檔 SHA、工具版本、排除理由與 bundle／graph hash。只重用 hash 一致的觀察；並行變更時固定快照、標記過期，不覆寫他人的檔案或暫存成果。

驗收：managed 與 XML 路徑完全相等，圖零重複ID／懸空端點；每個被引用的新檔都有明確納入或限制說明。來源變動時 `--verify` 失敗被正確揭露。若另產生視覺化，使用副本或記錄衍生 hash，不假稱原始 artifact 未變。

### B／P1：建立能力交付對照，再選一個已核准切片

沿用 capabilities.json 與既有功能文件，列出「已接線、僅 scaffold、UI preview、未做真實渠道驗收」的證據，避免另建一份版本真相。先與 Issue 進行中的工作核對；不把當前 scaffold、預覽或本報告本身視為新功能授權。

驗收：被選功能的 Actor、Resource、Action、Authority、Policy／Decision、Result 已在既有功能規格回答；契約、adapter、核心、UI 及驗收證據可定位。正式頁的未開放狀態與預覽的示範提示保持正確；六個問題或產品決策缺失時，依賴部分標 blocked。基線要求先完成 foundation／LINE／基礎出勤，不因看板預覽而跳過。

### C／P1：只補現有架構檢查的覆蓋

先讀最新 boundaries_test.go；若 Repository 尚未納入，將其純核心加入既有 core map，允許依賴以實際核心關係為準。更動前檢查 numbering.go 等同 package 檔案，不因新目錄存在自動加入 SDK 白名單。

驗收：Repository 與既有核心遵守相同標準庫／純核心依賴及 serialization tag 規則；HTTP、transport、httpapi 的既有守門維持。執行既有架構檢查及適用 Go checks。這是後續小範圍工作，不新增測試平台；若其他任務已完成則只核對證據，不重複實作。

### D／P2：變更補卡時再做責任拆分

採用「同 package 按提案、決定、查詢、稽核 codec 分檔」的最小方案，保持公開介面及同一交易；相較引入共用 workflow engine，成本與驗證範圍較小。沒有相關需求或缺陷時維持現況。

驗收：既有核准／拒絕／取消與重送行為不變；權限、版本／重疊檢查、有效紀錄及必要 audit 維持同一原子邊界。執行既有 focused tests 與規定 checks，未執行的 Firestore／LINE 驗收另列；不因純分檔宣稱安全性或效能提高。

### E／P2：Graphify 導覽、Repomix 按範圍核對

保留本次全量 XML 作基準；日常先用 Graphify query／path／explain 定位，再使用固定 Repomix CLI 明確 include 契約、直接呼叫者、port、adapter、相關文件與測試。只在依賴／生成契約問題中帶入完整 lockfile／generated；每個小包都列出實際檔案及未納入邊界。

驗收：本次問題需要的來源全數收錄，輸出 XML 可解析並保留安全掃描；以實際 token 數衡量閱讀成本。`--compress` 是結構摘要，若涉及業務判斷仍讀完整來源；不能用壓縮包代替全量涵蓋，也不修改正式分析契約來追求較小數字。

## 查證、檢查與限制

已使用專案 Graphify／Repomix-explorer skills、固定 CLI、分析契約與現有架構／操作文件。Context7 先 resolve `/yamadashy/repomix`，再查 XML parsable-style、include／ignore、安全掃描與 tokens，並核對本機 `repomix --help`；Context7 上游 main 文件不保證與固定1.18.1完全同版。Graphify 的既有套件與 CLI 實測是本次依據，未宣稱其上游文件全量查證。

已執行分類／audit、Graphify 查詢與 cluster-only、Repomix 完整與補充打包、來源快照 hashes、XML 檔案覆蓋、圖完整性及文件連結／whitespace 檢查。並行來源造成 `--verify` 拒絕是限制揭露，不能記為當前全庫 PASS。原生報告未設定 LLM backend，community labels 使用自動／placeholder，不以名稱推導業務分類。

本次未執行應用測試、完整安全掃描、效能評測、LINE 手機／membership 或 Firestore／Cloud Run 實機驗收。所有方案仍為後續工作；只交付此分析文件與索引，不修改模型、公開 API 或資料，不推論正式環境已就緒。

[工具與分析索引](README.md) · [分析範圍與證據契約](analysis-scope.md) · [專案文件索引](../README.md)
