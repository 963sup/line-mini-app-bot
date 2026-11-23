# 全庫標準化盤點與批次整改清單

本文件集中保存本輪找到的問題與缺口，作為後續批次實施入口。盤點後依使用者指示開始分批實施；卡片保留原始證據，另記修正及驗收狀態。 已完成的修正與測試歷史維持在[標準化驗收紀錄](standardization.md)，本頁不把它們重新列成待修。

## 基準與涵蓋範圍

- 日期：2026-11-23；來源基準：`3497f48ede824f1aca51295abac692f14c307a55`。
- 基準時共有 1,859 個已追蹤且存在的檔案；其中 Go 429、TSX 159、TS 22、Markdown 962。Markdown 數包含上游技能，並非自有文件數。
- 業務目錄 17 個；能力清單有 39 個 capability、109 個 operation、23 個 resource 登記、6 個事件入口與 11 個命名相容例外。這些是索引規模，不是已驗收比例，也不是獨立 bounded context 數。
- 以 Repomix 參考技能導航，對當前來源做索引、模式掃描與風險群組核對；檢查了業務／API／前端、身分與共用傳輸、建置／CI／工具、規格／提案。未逐條動態執行 109 個操作，也未查詢正式雲端、LINE、配額或真實裝置。
- 生成碼只核對來源與既有生成證據；二進位、上游技能／shadcn 快照不做人工重寫。`.mise/locks/` 是工具依賴鎖資料，不能僅因位於隱藏目錄就當快取刪除。
- DDD 參考目錄完全保留。開始盤點時既有未提交的 `attendance/domain/service/worktime.go`、`journal/WorkContext.tsx` 與 DDD 的兩個檔案不在盤點寫入範圍；涉及其行為的結論標為待驗證。

「已確認」只指卡片列明的本地證據；「待驗證」不是已證明的 bug；未交付的候選模式不代表現行 legacy 功能壞掉。全庫盤點是此版的可追蹤清單，不保證未知缺陷已窮盡。

## 先按批次排程

| 批次              | 目標／項目                                                      | 前置與可獨立驗收的交付                                                                              |
| ----------------- | --------------------------------------------------------------- | --------------------------------------------------------------------------------------------------- |
| 0：語意與驗收基準 | [A07](#a07)、[A08](#a08)、[B02](#b02)、[B04](#b04)、[B07](#b07) | 先確認資源／投影、角色、保存與適用範圍。輸出權威決策及逐操作驗收矩陣，不直接改名稱或資料。          |
| 1：交付閘門       | [A01](#a01)、[A02](#a02)、[A04](#a04)                           | 不依賴新產品規則。完整交付來源閉包、emulator CI 覆蓋與對應測試導覽；先修可阻擋實際交付的缺口。      |
| 2：現行規則稽核   | [A03](#a03)、[C01–C05](#c01)                                    | 按同一規則掃全部消費端，再分資源提交完整修正；每筆含 Backend、API、必要 Frontend 與拒絕／重送證據。 |
| 3：獨立資格與遷移 | [B01](#b01)、[B03](#b03)、[B05](#b05)、[B06](#b06)              | 分別依賴對應產品決策；以 Organization／Team／Enterprise 的正常入口閉環交付，不先切模式再補功能。    |
| 4：互動與證據     | [C06](#c06)、[C07](#c07)、[A05](#a05)                           | 按相同錯誤／切換情境覆蓋全部相關入口；只補能捕捉具體反例的門檻，沿用未失效的共用檢查。              |
| 5：目標環境驗收   | [D01–D04](#d01)                                                 | 所需環境身分、測試渠道及操作授權就緒後，分開驗收部署、資料庫、裝置及外部服務；不以本地綠燈代替。    |
| 6：重建與效能     | [A06](#a06)、[E01](#e01)、[E02](#e02)                           | 先定義可重現性目標與量測需求，再選最小改動；不預先增加 broker、cache、outbox 或通用框架。           |

同批次可先共用盤點和驗證證據，但不代表要合成一筆巨大 commit。修復順序由依賴、風險及所有權決定；修改前重查卡片來源版本，處理相同原因的所有入口，完成後再選下一批，避免每次重新全庫探索。

## 問題卡片

<a id="a01"></a>

### A01｜Cloud Build 上傳來源缺少欄位政策腳本

**P0 · 已修復；本地驗收通過 · 標準 16、31、32 · 責任：部署／工具**

- 證據：[cloudbuild.yaml](../../../cloudbuild.yaml) 在盤點基準 3497f48e 的 firestore-field-policies step 執行 scripts/db/ensure-field-policies.py；當時 [.gcloudignore](../../../.gcloudignore) 只放行 ensure-attendance-indexes.py。當時實跑 gcloud meta list-files-for-upload，882 個上傳檔案中漏掉欄位政策腳本。
- 調整：同步來源 allowlist 與建置 step，建立所有 Cloud Build 本地腳本及其本地依賴的上傳閉包檢查；先修這一批，不只補單一檔名。
- 關閉條件：離線／唯讀上傳清單包含每個建置入口及其依賴；加入漏檔反例，移除任一必要檔案時檢查失敗；以實際上傳內容重建步驟。此缺陷影響 repository cloudbuild.yaml 的手動／自訂 trigger 路徑，不代表 main 的 managed inline trigger 一定使用該檔。
- 前置：無；先維持現行相容行為。

- 本輪實施：`.gcloudignore` 已放行必要腳本（已整合並行索引提交 `40f1f98a`），並把 Cloud Build 靜態 script entry 與 sibling Python helper 的來源閉包納入 check-config。四個離線反例／正常案例通過；實際 gcloud 上傳清單同時包含欄位政策與索引 helper。檢查器對尚未支援的 SDK include 指令明確失敗；動態路徑仍需另行發布驗證。
- 實施提交：`40f1f98a（上傳基線）／76ef1291（驗收閘門）`；遠端驗收見本輪共用 CI 證據。

<a id="a02"></a>

### A02｜CI 的 emulator 驗收只涵蓋兩個 Project 案例

**P0 · 已修復；遠端 CI 驗收通過 · 標準 17、19、20、21、22、23、30、34 · 責任：CI／儲存驗收**

- 證據：[.github/workflows/check.yml](../../../.github/workflows/check.yml) 先執行 mise run check，之後才啟動 emulator；後段只選 TestPersonalRepositoryAndProject 與 TestProjectSettingsBrowserRoundTrip。[backend/internal/business/enterprise/adapter/firestore/provisioning_test.go](../../../backend/internal/business/enterprise/adapter/firestore/provisioning_test.go) 與 [backend/internal/business/team/adapter/firestore/assignment_lifecycle_test.go](../../../backend/internal/business/team/adapter/firestore/assignment_lifecycle_test.go) 在缺少 FIRESTORE_EMULATOR_HOST 時會 SKIP。
- 調整：列出所有依賴 emulator 的測試與實際 CI 執行範圍，統一 emulator 啟動、清理及整合測試階段；保留瀏覽器個別前置條件。
- 關閉條件：在乾淨 runner 實跑所有納入範圍的交易／授權案例，保存逐案例 PASS 與 SKIP；必要案例不得以未選中或 SKIP 充當成功。不要停用檢查來縮短時間。
- 前置：無；先維持現行相容行為。

- 本輪實施：CI 改為 emulator 及 Playwright 就緒後跑完整 fresh race JSON suite，對照 go list 的預期套件，拒絕失敗、缺席與 skipped test cases。修正後首次本機實跑 168 套件、494 個測試／子案例全部通過，零案例跳過；新增結果檢查器的四項反例／正常測試。最終重新執行為 169 個套件、510 個測試／子案例通過，零案例跳過。
- 實施提交：`76ef1291`；遠端驗收見本輪共用 CI 證據。

<a id="a03"></a>

### A03｜Enterprise 核心錯誤依賴 Organization

**P1 · 已修復；本地驗收通過 · 標準 1、6、7、13 · 責任：Enterprise／HTTP 錯誤邊界**

- 證據：[backend/internal/business/enterprise/application/command/enterprise.go](../../../backend/internal/business/enterprise/application/command/enterprise.go) 引用 Organization 的 ErrNotFound／ErrConflict；[backend/internal/business/organization/domain/entity/errors.go](../../../backend/internal/business/organization/domain/entity/errors.go) 同時包含企業開通與團隊歸屬錯誤。架構清單已明確保留這個相容依賴。
- 調整：先盤點 errors.Is 的所有消費端、公共錯誤碼及 HTTP 映射，再決定領域自有錯誤與 adapter 翻譯位置；不以改 package 名稱掩蓋相同耦合。
- 關閉條件：同樣成功／拒絕輸入維持公共 status、code、recovery metadata；核心依賴清單能移除不再需要的跨領域例外，且舊錯誤相容案例通過。
- 前置：無；先維持現行相容行為。

- 本輪實施：Enterprise command 使用自有 domain 錯誤；HTTP adapter 翻譯到既有傳輸相容分類，保留 errors.Is 與公共 403／404／409。核心架構例外已移除。九個命令授權／副作用情境、四個 HTTP 相容情境及九個架構正反案例通過。共享 legacy adapter／儲存錯誤分類仍保留，不宣稱所有領域錯誤都已重構。
- 實施提交：`0924735e`；遠端驗收見本輪共用 CI 證據。

<a id="a04"></a>

### A04｜Team 導覽仍聲稱沒有獨立測試

**P1 · 已修復；本地驗收通過 · 標準 9、30、34 · 責任：Team 文件**

- 證據：[docs/domains/team/README.md](../../domains/team/design.md) 仍有 “No independent Go tests found here”；實際已有 [backend/internal/business/team/adapter/firestore/activation_test.go](../../../backend/internal/business/team/adapter/firestore/activation_test.go)、[backend/internal/business/team/adapter/firestore/group_catalog_test.go](../../../backend/internal/business/team/adapter/firestore/group_catalog_test.go) 及生命週期案例。
- 調整：批次核對各 domain README 的現行實作、測試入口與已執行證據，歷史 smoke 記錄保留日期與限制，不誤當當前覆蓋。
- 關閉條件：每個現行領域都有正確代表測試入口及環境前提；測試檔存在與案例實跑分開描述；本機連結／錨點通過。
- 前置：無；先維持現行相容行為。

- 本輪實施：已移除 Team 導覽的錯誤聲明，連到生命週期、開通及資格／觀察三個代表測試檔，並標示 loopback emulator 前提與「檔案存在不等於已執行」。全域搜尋已無此錯誤聲明。
- 實施提交：`76ef1291`；遠端驗收見本輪共用 CI 證據。

<a id="a05"></a>

### A05｜文件檢查的零失效不涵蓋所有引用語法

**P2 · 已修復；本地驗收通過 · 標準 9、30 · 責任：文件工具**

- 證據：[scripts/check/check-config.py](../../../scripts/check/check-config.py) 只解析 inline Markdown links 並略過 fenced code；[scripts/generate/docs-graph.py](../../../scripts/generate/docs-graph.py) 也跳過 Mermaid 內文。本輪臨時 fixture 中，失效 reference-link、HTML href 及 Mermaid click 均未被拒絕；自有文件仍有 6 個 Mermaid click 語法出現處。
- 調整：先定義實際支援的本地引用語法，再補對應解析與反例；統一文件圖及檢查器的 coverage 說明，外部網址另列網路驗收。
- 關閉條件：失效 inline／reference／HTML／Mermaid 本地目的地各有拒絕證據，合法 fenced 範例與上游快照保持排除。未新增支援前只能宣稱「已涵蓋語法零失效」，不宣稱全引用皆有效。
- 前置：無；先維持現行相容行為。

- 本輪實施：文件檢查新增 reference definitions、HTML href、fenced Mermaid click 與獨立 .mermaid 目的地；文件圖共用同一解析器。新增目的地缺失、錨點缺失、獨立圖及 literal 範例驗收，六項文件測試通過。Setext 與遠端網址可達性仍明列範圍外。
- 實施提交：`76ef1291`；遠端驗收見本輪共用 CI 證據。

<a id="a06"></a>

### A06｜工具 bootstrap 與 apt 輸入未完全固定

**P2 · 已確認可重現性限制 · 標準 16、31、32、33 · 責任：工具鏈／建置**

- 證據：[.github/workflows/check.yml](../../../.github/workflows/check.yml) 對 mise CDN 使用 minimum_release_age: 0s，註解為 latest；[Dockerfile](../../../Dockerfile) 已固定三個 base image digest，但 apt-get update/install 未固定 Debian 套件快照或版本。[mise.toml](../../../mise.toml) 固定應用工具版本，不能等同所有 bootstrap 輸入不可變。
- 調整：先定義要求的是功能可重建還是位元完全一致，再固定／記錄 bootstrap 與作業系統套件解析結果，設置受控安全更新節奏。
- 關閉條件：兩個乾淨環境的來源、工具、OS 套件、依賴與 image digest 可對照；不因追求一致而永久凍結安全更新。保留目前已固定的 base digests 及鎖檔。
- 前置：無；先維持現行相容行為。

<a id="a07"></a>

### A07｜HTTP 資源登記與領域實體需分層說明

**P0 · 已修復；本地驗收通過 · 標準 1、2、3、4、6、10、11、12 · 責任：API／領域規格**

- 證據：[api/contract/capabilities.json](../../../api/contract/capabilities.json) 的 resources 有 23 筆，entity 欄位同時包含 Repository、Project 與 GroupItem、PunchResponse、PopplerText；也有 Project 的兩種路由。這份清單不能直接當全部 Aggregate 或 Firestore collection 清單。
- 調整：逐條標記領域資源、命令結果、查詢投影及相容路由，對應唯一責任領域與持久化權威；只補語意分類，名稱更改另走相容性決策。
- 關閉條件：23 筆皆可追到負責領域、URI 語意、資料權威及必要相容例外；Group／Team、Project 的來源與所有權、Organization／Enterprise 的 Owner 政策不混用。
- 前置：無；先維持現行相容行為。

- 本輪實施：23 個資源條目均加入 owner／kind，以 format version 2 明確區分 resource、projection、command-result；owner 必須匹配實際操作。補三項正常／未知分類／錯誤歸屬／重複路徑回歸。保留所有公開 URI、schema 名稱、儲存識別及 HTTP API 版本；解釋 entity 是既有公開 schema 名稱而非 Aggregate 登記。
- 實施提交：`7498c7bd`；遠端驗收見本輪共用 CI 證據。

<a id="a08"></a>

### A08｜36 項里程碑沒有逐操作驗收分母

**P0 · 待驗收範圍定義 · 標準 1、9、26、30、34 · 責任：驗收規格**

- 證據：[docs/development/verification/standardization.md](standardization.md) 明確採四個 25% 里程碑，沒有「每項全部適用操作及已驗收數」；75% 不是全庫合規率，修一個案例不能推導整類完成。
- 調整：為每一項列適用資源／操作、必要案例、環境與非適用條件；把已完成、未驗證、阻塞分開計數。舊里程碑保留為歷史，後續另報有分母的驗收進度。
- 關閉條件：36 類皆有可追溯驗收集合，分母變更留下原因；新問題不被藏在固定百分比裡；未知覆蓋不得算通過。
- 前置：無；先維持現行相容行為。

<a id="b01"></a>

### B01｜Organization 獨立資格的正常入口

**P1 · 已知未完整交付 · 標準 1、2、18、19、20、21、22、23、25、26 · 責任：Organization／Membership**

- 證據：[docs/domains/organization/design.md](../../domains/organization/design.md) 的獨立資格交付邊界要求建立、邀請、接受、People、Owner 移轉、退出／移除的完整入口；目前產品主線仍為 legacy。[backend/internal/business/membership/application/invitations.go](../../../backend/internal/business/membership/application/invitations.go) 的既有切片不等於完整模式切換。
- 調整：按一個資源的完整操作閉環交付 Backend→API→Frontend，包含錯誤、重送與失去資格；不要先手改模式欄位。
- 關閉條件：從正常入口完成全部步驟，並驗證唯一有效 Owner、重送不重新開權、移轉與退出競爭、撤權後導覽與下層資格失效。
- 前置：A07／A08；依既定 Organization 規格，不套用 Enterprise 多 Owner。

<a id="b02"></a>

### B02｜Team 群組開通委任的產品決策

**P0 · 待決策 · 標準 1、17、18、19、20、23 · 責任：Team／渠道治理**

- 證據：[docs/domains/team/design.md](../../domains/team/design.md) 的群組開通授權仍需渠道審核人員、可核驗獨立委任證據與效期；[docs/proposals/membership.md](../../proposals/membership.md) 的閉環清單列為啟用阻塞。
- 調整：定案審核人員來源、證據、效期、撤銷及自審限制，落在同一權威規格。
- 關閉條件：可判定誰能核發、何時失效、如何撤銷及何種證據有效；Bot 看過群組或申請人發過訊息不得自行升權。
- 前置：無；先維持現行相容行為。

<a id="b03"></a>

### B03｜Team 開通的申請、審核與消耗閉環

**P1 · 已知未完整交付 · 標準 2、12、17、19、20、21、22、23、25 · 責任：Team／Membership／Bootstrap**

- 證據：[docs/domains/team/design.md](../../domains/team/design.md) 描述申請、核發與交易消耗流程；[frontend/miniapp/src/business/team/TeamActions.tsx](../../../frontend/miniapp/src/business/team/TeamActions.tsx) 現行表單接線並未完成該治理閉環。
- 調整：在既有 Team／Membership 邊界內實作契約與正常入口，不把管理授權、成員角色及 LINE 觀察混成一項。
- 關閉條件：覆蓋自審拒絕、跨群組／帳號重放、審核前開通拒絕、撤銷與消耗競爭、Owner 換手、首任人員退出及原命令重送。
- 前置：B02；適用上層 Organization／Enterprise 資格可用。

<a id="b04"></a>

### B04｜Enterprise 多 Owner／manager 的獨立權威

**P0 · 待決策 · 標準 1、2、17、19、20、21 · 責任：Enterprise**

- 證據：[docs/domains/enterprise/design.md](../../domains/enterprise/design.md) 明確列出多 Owner 權威、manager 可執行動作與最後 Owner 保護尚待定案。
- 調整：先定義角色可執行的具體操作、Owner 新增／撤銷／退出與版本競爭，再固定狀態與 API 契約。
- 關閉條件：每個角色與狀態都有允許／拒絕結果；不複製 Organization 單一 Owner 指標，manager 不能自行升 Owner。
- 前置：無；先維持現行相容行為。

<a id="b05"></a>

### B05｜Enterprise 獨立資格的邀請與成員入口

**P1 · 已知未完整交付 · 標準 12、18、19、20、21、22、23、25、26 · 責任：Enterprise／Membership**

- 證據：[docs/domains/enterprise/design.md](../../domains/enterprise/design.md) 尚未完成直接成員加入、名單、退出、移除及重新加入入口；旗下組織可見性不代表直接企業資格。
- 調整：依 B04 交付直接資格與正常入口，再接 Enterprise Team 消費端，保持直接與衍生資格可區分。
- 關閉條件：建立、加入、退出、重加入、撤權及 Team 消費端均有正常流程、拒絕／重送與 generation 失效證據。
- 前置：B04；不以旗下組織成員名單代替企業資格。

<a id="b06"></a>

### B06｜資料準備、切換與回復工具

**P1 · 已知未交付 · 標準 4、11、16、19、21、22、26、32 · 責任：資料遷移／各資源**

- 證據：[docs/domains/organization/design.md](../../domains/organization/design.md) 列出資料準備、切換及回復未交付；[scripts/db/README.md](../../../scripts/db/) 的現有索引／欄位政策工具不是 Membership 遷移工具。[docs/proposals/firestore-data-contracts.md](../../proposals/firestore-data-contracts.md) 的 schemaVersion=1 是未實作提案。
- 調整：為每個切換建立舊／新 reader-writer 組合、試跑、可重跑準備、版本條件、切換檢查與回復界線，不在請求路徑無界回填。
- 關閉條件：混合版本及部分失敗可重跑；回復不恢復已撤權資格、不覆寫新資料；資料與權限不變量經 staging 驗收。
- 前置：B01／B05 對應正常入口完成；保存與刪除範圍依 B07。

<a id="b07"></a>

### B07｜保存、刪除與稽核保留的跨資源政策

**P1 · 待決策 · 標準 2、16、17、21、23、26、28、35、36 · 責任：資料責任領域／營運**

- 證據：[docs/proposals/firestore-data-contracts.md](../../proposals/firestore-data-contracts.md) 的保存期與刪除 SLA 依 P05 待定；[docs/proposals/membership.md](../../proposals/membership.md) 的獨立邀請、收據窗口與稽核保存待定。Summary 已有自身 24／48 小時規則，不代表全部資源採同一 TTL。
- 調整：按資料目的列出邀請、重送收據、稽核、訊息、附件與衍生資料的保留、清除、法律／產品限制及操作責任。
- 關閉條件：每類資料有保存及刪除完成條件；刪除與未完成工作、重送、追溯及子集合相容；不得默認永久保存或任意套用共同 TTL。
- 前置：無；先維持現行相容行為。

<a id="c01"></a>

### C01｜全寫入操作的生命週期與授權矩陣

**P1 · 待驗證 · 標準 2、17、19、20、21 · 責任：各資源領域**

- 證據：[api/contract/capabilities.json](../../../api/contract/capabilities.json) 列 109 個操作；現有 Team／Enterprise 回歸只能證明各自切片。代表來源為 [backend/internal/business/team/adapter/firestore/store.go](../../../backend/internal/business/team/adapter/firestore/store.go)、[backend/internal/business/enterprise/adapter/firestore/store.go](../../../backend/internal/business/enterprise/adapter/firestore/store.go)、[backend/internal/business/repository/adapter/firestore/labels.go](../../../backend/internal/business/repository/adapter/firestore/labels.go)。
- 調整：一次列出全部寫入操作的可信 Actor、作用域、授權模式、生命週期、版本及交易讀取點，按共同判斷批次找漏檢；先重現再改。
- 關閉條件：矩陣逐操作核對正常、停用、未知模式、失去資格與競爭；拒絕無部分寫入。未核對欄位明列未知，不因相似程式假設同樣正確。
- 前置：無；先維持現行相容行為。

<a id="c02"></a>

### C02｜識別與游標的作用域、格式及上限

**P1 · 待驗證 · 標準 3、4、10、11、14、17 · 責任：API／各查詢 adapter**

- 證據：[backend/internal/platform/firestore/key/cursor.go](../../../backend/internal/platform/firestore/key/cursor.go) 與 [backend/internal/business/attendance/adapter/firestore/attendance_query.go](../../../backend/internal/business/attendance/adapter/firestore/attendance_query.go)、[backend/internal/business/project/adapter/firestore/store.go](../../../backend/internal/business/project/adapter/firestore/store.go) 有各自識別／游標規則。
- 調整：對照 OpenAPI 與儲存語意，核對 UUID／LINE ID／不透明引用、游標長度、型別、scope、排序 tie-breaker 與過期政策；只整合相同規則。
- 關閉條件：空值、非法 ID、超長、跨 actor／channel／resource／filter 游標及同時間排序有拒絕或穩定結果；不為統一名稱重寫既有持久 ID。
- 前置：無；先維持現行相容行為。

<a id="c03"></a>

### C03｜時間與衍生數值的權威及邊界

**P1 · 待驗證 · 標準 5、10、11、20、25 · 責任：Attendance／Journal／Project／前端呈現**

- 證據：[backend/internal/business/attendance/domain/service/worktime.go](../../../backend/internal/business/attendance/domain/service/worktime.go) 與 [docs/domains/attendance/worktime.md](../../domains/attendance/worktime.md) 定義工時；[frontend/miniapp/src/business/journal/Workspace.tsx](../../../frontend/miniapp/src/business/journal/Workspace.tsx) 使用業務日，[frontend/miniapp/src/business/project/Inspections.tsx](../../../frontend/miniapp/src/business/project/Inspections.tsx) 使用本地日期顯示。
- 調整：分清 instant、業務日期、月界線與顯示時區；核對已確認／估計工時是否被相加，以及交易重試的時間採收到或提交時點。不能全域機械改成 UTC 顯示。
- 關閉條件：跨午夜／跨月、進行中班次、端點時間、過期邊界與估計值均符合契約；前端不另做權威工時計算。兩個正在編輯的相關檔案須先完成所有權／版本確認。
- 前置：無；先維持現行相容行為。

<a id="c04"></a>

### C04｜跨入口錯誤語意與復原 metadata

**P1 · 待驗證 · 標準 12、13、17、27、28 · 責任：HTTP／LINE／內部 worker**

- 證據：[backend/internal/transport/http/json.go](../../../backend/internal/transport/http/json.go)、[backend/internal/business/internal/adapter/http/response.go](../../../backend/internal/business/internal/adapter/http/response.go) 與 [backend/internal/transport/http/geolocation/geolocation.go](../../../backend/internal/transport/http/geolocation/geolocation.go) 有不同入口用途；A03 是已確認的領域耦合。
- 調整：列出每個入口的錯誤來源、穩定碼、HTTP status、可重試／原請求重試提示及敏感資料限制，區分既有相容例外。
- 關閉條件：同業務拒絕在其契約內可辨識；權限／無效輸入不變成無界重試；原始供應端錯誤、token 與私人內容不洩漏。不為格式一致破壞既有範例端點。
- 前置：無；先維持現行相容行為。

<a id="c05"></a>

### C05｜重送收據、回應遺失與提交／外部副作用窗口

**P1 · 待驗證 · 標準 21、22、23、27、35 · 責任：各寫入使用案例／Summary**

- 證據：[docs/domains/repository/numbering.md](../../domains/repository/numbering.md) 記錄共用編號／收據；[backend/internal/business/summary/adapter/firestore/summary.go](../../../backend/internal/business/summary/adapter/firestore/summary.go) 與 [backend/internal/business/summary/adapter/messaging/cloudtasks/summarization/queue.go](../../../backend/internal/business/summary/adapter/messaging/cloudtasks/summarization/queue.go) 分別處理提交與入列。Summary 文件明定入列失敗由 LINE 重送恢復。
- 調整：對各收據核對 actor/resource/type/content scope、回放時權限、保留窗口、版本與外部投遞；Summary 注入提交後入列失敗及回應遺失。先核對既有恢復路徑，再決定是否需要新 outbox。
- 關閉條件：同鍵同內容只產生一次效果、不同內容衝突、撤權後不洩漏結果；已提交但未入列能依既定契約恢復，不能把固定 task ID 當原子跨服務交易。
- 前置：無；先維持現行相容行為。

<a id="c06"></a>

### C06｜表單與 Backend 決策的分工及取消／切換

**P1 · 待驗證 · 標準 10、12、18、19、24、25、36 · 責任：前端各功能／API**

- 證據：[frontend/miniapp/src/business/object/quantity.ts](../../../frontend/miniapp/src/business/object/quantity.ts) 是未提交輸入提示；[frontend/miniapp/src/shared/http/lifecycle.ts](../../../frontend/miniapp/src/shared/http/lifecycle.ts) 與 [frontend/miniapp/src/identity/Session.tsx](../../../frontend/miniapp/src/identity/Session.tsx) 維護取消及身分。存在前端驗證不等於非法重複業務決策。
- 調整：按表單、列表、詳情及選擇器建立成功／永久拒絕／暫時錯誤／未知結果／切 actor/resource 測試矩陣；區分立即輸入提示與授權、業務計算。
- 關閉條件：後端拒絕仍是權威，舊回應不覆寫新作用域；失去可見性清空私人資料，暫時分頁錯誤保留可重試游標，未知寫入保留原 payload/key。
- 前置：無；先維持現行相容行為。

<a id="c07"></a>

### C07｜工程證據的適用版本與覆蓋範圍

**P1 · 待驗證 · 標準 7、9、15、30、31、34 · 責任：驗收／架構／生成工具**

- 證據：[backend/internal/architecture/boundaries_test.go](../../../backend/internal/architecture/boundaries_test.go) 是靜態 Go import 檢查；[scripts/contract/check-compatibility.mts](../../../scripts/contract/check-compatibility.mts) 是保守結構比較；[scripts/generate/docs-graph.py](../../../scripts/generate/docs-graph.py) 不是語意驗證。[docs/development/verification/standardization.md](standardization.md) 有先前快照測試證據。
- 調整：將已通過的檢查對應 commit、dirty paths、工具版本、案例及 SKIP；審查 frontend 邊界是否需要現有檢查能捕捉的具體反例，再增加門檻。
- 關閉條件：每個結論有確切版本與涵蓋／不涵蓋範圍；生成同步不等於公開行為相容，圖譜不等於動態授權。未變更來源沿用有效證據，不每批重跑所有檢查。
- 前置：無；先維持現行相容行為。

<a id="d01"></a>

### D01｜目標 Cloud Run 設定、身分及發布／回復

**P1 · 環境待驗收 · 標準 16、17、18、27、31、32、33 · 責任：部署營運**

- 證據：[docs/operations/cloud-run.md](../../operations/cloud-run.md) 區分 managed inline trigger 與 repository cloudbuild.yaml；[scripts/ops/cloud-run-snapshot.py](../../../scripts/ops/cloud-run-snapshot.py) 提供服務／revision metadata。此輪沒有讀取目標雲端實際狀態。
- 調整：先唯讀核對 project/region/service、來源 SHA/image digest、runtime identity、IAM、Secret 數字版本與執行期設定，再依已授權發布範圍演練。
- 關閉條件：目標來源、映像、Ready revision、traffic、必要路由與回復版本一致；回復不破壞資料格式。不得把修改 repository YAML 當 managed service 設定已改變。
- 前置：需可用雲端唯讀身分；部署／切流量／回復依當次授權。

<a id="d02"></a>

### D02｜Firestore 索引、TTL 與刪除政策的實際部署

**P1 · 環境待驗收 · 標準 2、11、17、21、26、32、36 · 責任：Firestore／資料營運**

- 證據：[firestore.indexes.json](../../../firestore.indexes.json) 與 `scripts/db/ensure-field-policies.py`（盤點當時的腳本） 僅代表宣告及工具；A01 證實來源上傳還有缺口。
- 調整：核對目標 database 的 composite indexes、TTL、單欄豁免與 IAM；在授權 staging 驗證等待就緒、錯誤／逾時及過期資料讀取行為。
- 關閉條件：必要查詢實跑、索引／TTL 狀態符合宣告，過期資料即使未刪也不被讀出；不以 emulator 證明正式 IAM 或已部署 TTL。
- 前置：先修 A01；套用政策與刪除資料另需對應環境授權。

<a id="d03"></a>

### D03｜真實 LINE、定位與工作階段

**P1 · 環境待驗收 · 標準 5、17、18、19、24、25、27 · 責任：LINE／身分／裝置**

- 證據：[docs/operations/release-acceptance.md](../../operations/release-acceptance.md) 列出 iOS、Android、外部瀏覽器驗收；[docs/integrations/geolocation/browser-api.md](../../integrations/geolocation/browser-api.md) 有定位限制；本機 fixtures 不提供真實裝置證據。
- 調整：依正常 LINE 入口測 Channel/Provider、liff.state、登入返回、過期／撤銷、定位拒絕／逾時／背景切換及私人資料清理。
- 關閉條件：每個裝置／OS／LINE 版本分別記錄結果；不以 mock token、GPS 或獨立瀏覽器成功推論所有 LINE 裝置可用。
- 前置：實際測試渠道、已配置 LIFF 與測試帳號／裝置。

<a id="d04"></a>

### D04｜Cloud Tasks／模型／通知與遙測落地

**P1 · 環境待驗收 · 標準 17、23、27、28、29、35 · 責任：Summary／外部依賴／營運**

- 證據：[backend/internal/transport/http/telemetry.go](../../../backend/internal/transport/http/telemetry.go) 把 allowlisted OTel 資料輸出到結構化日誌，並非已配置 Cloud Trace／Monitoring；[docs/domains/summary/summary.md](../../domains/summary/summary.md) 列 OIDC、模型及真實 LINE 投遞仍需驗收。
- 調整：驗證服務身分／audience、模型截斷與錯誤、429／timeout／未知 Push 結果、trace 串接、日誌查詢與告警；明確選定哪些 log exports 需轉成原生指標／trace。
- 關閉條件：失敗分類、重送與期限符合契約；必要背景工作可追蹤，告警能被觸發與解除，敏感內容不進日誌；不把 OTel SDK 存在當告警已可用。
- 前置：外部測試服務與授權；供應端憑證不寫入清單或日誌。

<a id="e01"></a>

### E01｜組織／企業清單的容量與查詢成本

**P2 · 量測後決定 · 標準 14、27、29、36 · 責任：Organization／Enterprise／Team**

- 證據：[backend/internal/business/enterprise/adapter/firestore/store.go](../../../backend/internal/business/enterprise/adapter/firestore/store.go) 的導覽／旗下組織／團隊與 [backend/internal/business/team/adapter/firestore/store.go](../../../backend/internal/business/team/adapter/firestore/store.go) 的候選／指派使用 GetAll，部分在迴圈內查詢。尚無本輪量測證明容量問題。
- 調整：先取得每帳號組織／Team 的實際上限、延遲及讀取成本，再決定分頁、projection、查詢整併或是否需要快取。
- 關閉條件：以同資料量與授權作用域比較前後成本；排序、空頁、游標與撤權一致。不為假設的規模先建通用查詢或快取框架。
- 前置：無；先維持現行相容行為。

<a id="e02"></a>

### E02｜進階事件／Outbox／搜尋／代理功能的採用

**P2 · 需求與量測後決定 · 標準 1、6、7、27、29、35、36 · 責任：Workflow／Knowledge／Agent／外部整合**

- 證據：[跨領域協作提案](../../proposals/domain-collaboration.md)、[Webhook 投遞提案](../../proposals/webhook-delivery/design.md) 與 [Knowledge 領域設計](../../domains/knowledge/design.md) 包含候選設計／部分元件，不是全部當前產品缺陷。
- 調整：逐能力寫清真實使用案例、現有同步／重送方案限制、可靠投遞或效能目標；有證據後才選 CloudEvents、outbox、worker、向量索引或快取。
- 關閉條件：採用決策能對應具體需求、責任領域、版本及失敗／恢復驗收；未採用的候選能力不強制湊成微服務或通用平台。
- 前置：無；先維持現行相容行為。

## 36 項覆蓋與進度

既有里程碑仍為 **1–25、27–36 各 75%，26 為 50%，平均 74.3%**，權威與歷史見[標準化驗收紀錄](standardization.md)。以下列的是每類的盤點入口與下一份必要證據，並不把未發現問題當成 100% 合規。A08 完成後再提供有適用範圍及分母的驗收比例。

以下分母是各標準關聯的整改卡，不是全部業務操作／環境的測試數；「本地完成」的六張卡亦已通過本輪遠端 CI。沒有關聯卡片填「—」，不推算成 100%。全清單本地完成 **6／28（21.4%）**；A08 的業務驗收分母仍未建立。

| 標準           | 本地完成／關聯卡 | 整改卡完成率 | 仍待處理／驗收                                                                                                                                |
| -------------- | ---------------: | -----------: | --------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 領域         |             2／7 |        28.6% | [A08](#a08)、[B01](#b01)、[B02](#b02)、[B04](#b04)、[E02](#e02)                                                                               |
| 2 資源生命週期 |             1／7 |        14.3% | [B01](#b01)、[B03](#b03)、[B04](#b04)、[B07](#b07)、[C01](#c01)、[D02](#d02)                                                                  |
| 3 命名         |             1／2 |        50.0% | [C02](#c02)                                                                                                                                   |
| 4 識別         |             1／3 |        33.3% | [B06](#b06)、[C02](#c02)                                                                                                                      |
| 5 時間         |             0／2 |         0.0% | [C03](#c03)、[D03](#d03)                                                                                                                      |
| 6 結構         |             2／3 |        66.7% | [E02](#e02)                                                                                                                                   |
| 7 依賴         |             1／3 |        33.3% | [C07](#c07)、[E02](#e02)                                                                                                                      |
| 8 格式         |             0／0 |            — | 本輪無未完成關聯卡；不代表全庫驗收完成                                                                                                        |
| 9 文件         |             2／4 |        50.0% | [A08](#a08)、[C07](#c07)                                                                                                                      |
| 10 型別        |             1／4 |        25.0% | [C02](#c02)、[C03](#c03)、[C06](#c06)                                                                                                         |
| 11 資料契約    |             1／5 |        20.0% | [B06](#b06)、[C02](#c02)、[C03](#c03)、[D02](#d02)                                                                                            |
| 12 API         |             1／5 |        20.0% | [B03](#b03)、[B05](#b05)、[C04](#c04)、[C06](#c06)                                                                                            |
| 13 錯誤        |             1／2 |        50.0% | [C04](#c04)                                                                                                                                   |
| 14 查詢        |             0／2 |         0.0% | [C02](#c02)、[E01](#e01)                                                                                                                      |
| 15 生成        |             0／1 |         0.0% | [C07](#c07)                                                                                                                                   |
| 16 設定        |             1／5 |        20.0% | [A06](#a06)、[B06](#b06)、[B07](#b07)、[D01](#d01)                                                                                            |
| 17 安全        |            1／12 |         8.3% | [B02](#b02)、[B03](#b03)、[B04](#b04)、[B07](#b07)、[C01](#c01)、[C02](#c02)、[C04](#c04)、[D01](#d01)、[D02](#d02)、[D03](#d03)、[D04](#d04) |
| 18 身分        |             0／6 |         0.0% | [B01](#b01)、[B02](#b02)、[B05](#b05)、[C06](#c06)、[D01](#d01)、[D03](#d03)                                                                  |
| 19 授權        |            1／10 |        10.0% | [B01](#b01)、[B02](#b02)、[B03](#b03)、[B04](#b04)、[B05](#b05)、[B06](#b06)、[C01](#c01)、[C06](#c06)、[D03](#d03)                           |
| 20 狀態        |             1／8 |        12.5% | [B01](#b01)、[B02](#b02)、[B03](#b03)、[B04](#b04)、[B05](#b05)、[C01](#c01)、[C03](#c03)                                                     |
| 21 交易        |            1／10 |        10.0% | [B01](#b01)、[B03](#b03)、[B04](#b04)、[B05](#b05)、[B06](#b06)、[B07](#b07)、[C01](#c01)、[C05](#c05)、[D02](#d02)                           |
| 22 併發        |             1／6 |        16.7% | [B01](#b01)、[B03](#b03)、[B05](#b05)、[B06](#b06)、[C05](#c05)                                                                               |
| 23 冪等        |             1／8 |        12.5% | [B01](#b01)、[B02](#b02)、[B03](#b03)、[B05](#b05)、[B07](#b07)、[C05](#c05)、[D04](#d04)                                                     |
| 24 UI          |             0／2 |         0.0% | [C06](#c06)、[D03](#d03)                                                                                                                      |
| 25 表單        |             0／6 |         0.0% | [B01](#b01)、[B03](#b03)、[B05](#b05)、[C03](#c03)、[C06](#c06)、[D03](#d03)                                                                  |
| 26 資料遷移    |             0／6 |         0.0% | [A08](#a08)、[B01](#b01)、[B05](#b05)、[B06](#b06)、[B07](#b07)、[D02](#d02)                                                                  |
| 27 可靠性      |             0／7 |         0.0% | [C04](#c04)、[C05](#c05)、[D01](#d01)、[D03](#d03)、[D04](#d04)、[E01](#e01)、[E02](#e02)                                                     |
| 28 日誌        |             0／3 |         0.0% | [B07](#b07)、[C04](#c04)、[D04](#d04)                                                                                                         |
| 29 可觀測性    |             0／3 |         0.0% | [D04](#d04)、[E01](#e01)、[E02](#e02)                                                                                                         |
| 30 測試        |             3／5 |        60.0% | [A08](#a08)、[C07](#c07)                                                                                                                      |
| 31 建置        |             1／4 |        25.0% | [A06](#a06)、[C07](#c07)、[D01](#d01)                                                                                                         |
| 32 部署        |             1／5 |        20.0% | [A06](#a06)、[B06](#b06)、[D01](#d01)、[D02](#d02)                                                                                            |
| 33 版本        |             0／2 |         0.0% | [A06](#a06)、[D01](#d01)                                                                                                                      |
| 34 開發流程    |             2／4 |        50.0% | [A08](#a08)、[C07](#c07)                                                                                                                      |
| 35 事件        |             0／4 |         0.0% | [B07](#b07)、[C05](#c05)、[D04](#d04)、[E02](#e02)                                                                                            |
| 36 快取        |             0／5 |         0.0% | [B07](#b07)、[C06](#c06)、[D02](#d02)、[E01](#e01)、[E02](#e02)                                                                               |

## 業務目錄覆蓋

下表是盤點導航，不重新定義領域規則；詳情以[領域矩陣](../../domains/)、各規格及能力清單為準。Go 檔案數含測試，測試檔數不等於案例數或已執行數。

| 目錄                                                            | Go／測試檔數 | 本輪追蹤範圍                                                                       |
| --------------------------------------------------------------- | -----------: | ---------------------------------------------------------------------------------- |
| [agent](../../../backend/internal/business/agent)               |         5／2 | 候選能力、已有限定元件；[E02](#e02)，不自動升格正式業務入口。                      |
| [attendance](../../../backend/internal/business/attendance)     |       72／15 | 時間、可信身分、改正與查詢；[C01–C05](#c01)、[D03](#d03)。                         |
| [discussion](../../../backend/internal/business/discussion)     |        10／1 | Repository 內的內容／留言、編號、授權與游標；[C01–C06](#c01)。                     |
| [enterprise](../../../backend/internal/business/enterprise)     |         9／1 | 開通、Owner、關聯與獨立治理；[A03](#a03)、[B04–B05](#b04)、[C01](#c01)。           |
| [issue](../../../backend/internal/business/issue)               |        11／2 | Repository 範圍、狀態、標籤及編號；[C01–C06](#c01)。                               |
| [journal](../../../backend/internal/business/journal)           |        11／4 | 業務日期、提交／草稿、作用域及來源；[C03](#c03)、[C05](#c05)、[C06](#c06)。        |
| [knowledge](../../../backend/internal/business/knowledge)       |         5／1 | 候選檢索與向量索引接線；[E02](#e02)。                                              |
| [membership](../../../backend/internal/business/membership)     |        17／2 | 邀請、關係、世代及模式；[B01–B07](#b01)、[C05](#c05)。                             |
| [object](../../../backend/internal/business/object)             |       30／15 | 父子物件、數量、候選、模板與歷史；[C01–C06](#c01)。                                |
| [operations](../../../backend/internal/business/operations)     |         9／3 | PDF／模型處理介面與候選能力；[C04](#c04)、[D04](#d04)、[E02](#e02)。               |
| [organization](../../../backend/internal/business/organization) |        18／6 | 單一 Owner、名單與獨立資格切換；[B01](#b01)、[B06](#b06)、[C01](#c01)。            |
| [project](../../../backend/internal/business/project)           |        28／9 | 獨立所有權、來源引用、版本與表單；[A07](#a07)、[C01–C06](#c01)。                   |
| [repository](../../../backend/internal/business/repository)     |       51／13 | 容器資格、能力、共用編號與相容性；[C01](#c01)、[C02](#c02)、[C05](#c05)。          |
| [summary](../../../backend/internal/business/summary)           |       30／10 | 收集、生成、佇列、保存與通知；[B07](#b07)、[C05](#c05)、[D02](#d02)、[D04](#d04)。 |
| [team](../../../backend/internal/business/team)                 |        12／3 | LINE 觀察、管理歸屬、上層關係與指派；[A04](#a04)、[B02–B03](#b02)、[C01](#c01)。   |
| [weather](../../../backend/internal/business/weather)           |         6／2 | 位置／時間範圍與外部服務失敗；[C03](#c03)、[C04](#c04)、[D04](#d04)。              |
| [workflow](../../../backend/internal/business/workflow)         |         4／2 | 候選協調與確認元件；[E02](#e02)。                                                  |

## 非業務層覆蓋

| 區域                                    | 盤點與後續責任                                                                                                              |
| --------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| API／生成碼                             | capability、operation、resources、events、命名例外與生成來源；[A07](#a07)、[C02](#c02)、[C04](#c04)、[C07](#c07)。          |
| identity／transport／linegroup          | 驗證、Session、錯誤、取消、資料限制及群組事實；[C01](#c01)、[C04](#c04)、[C06](#c06)、[D03](#d03)。                         |
| bootstrap／platform                     | 依賴組裝、runtime 設定、關閉、外部期限及供應端能力；[A03](#a03)、[D01](#d01)、[D04](#d04)。                                 |
| frontend／共享元件                      | 表單、身分切換、取消、查詢、靜態路由；[C03](#c03)、[C06](#c06)。上游 UI 不作格式重寫。                                      |
| scripts／mise／CI                       | 來源閉包、emulator 覆蓋、解析器、工具版本及可重建性；[A01](#a01)、[A02](#a02)、[A05](#a05)、[A06](#a06)。                   |
| Docker／Cloud Build／Firestore manifest | 宣告、上傳來源、映像及實際目標環境分開核對；[A01](#a01)、[A06](#a06)、[D01](#d01)、[D02](#d02)。                            |
| docs／技能／參考                        | 規格、歷史與提案狀態、連結支援範圍及證據版本；[A04](#a04)、[A05](#a05)、[A08](#a08)、[C07](#c07)。DDD、上游技能與快照只讀。 |

## 此輪實際檢查與關閉方式

1. Git 檔案索引、狀態與副檔名／業務目錄統計；結構化讀取能力清單；搜尋狀態、授權、交易、重送、時間、I/O 邊界、未決規格及整合測試 SKIP。模式命中只產生候選，再核對相鄰原碼，不直接當成漏洞。
2. `gcloud meta list-files-for-upload --verbosity=error` 實際確認 A01。此命令只列本地上傳來源，沒有送出建置或修改雲端。
3. 臨時 Git fixture 實際確認 A05 的 reference／HTML／Mermaid 引用限制；既有文件的 6 個 Mermaid click 是語法出現數，未據此宣稱六個目標皆壞掉。
4. 核對 CI 的 emulator 啟動位置、篩選案例與測試的 SKIP 條件，確認 A02 是 CI 設定覆蓋缺口；本輪未重跑應用全套或真實服務，前輪通過結果仍屬其原本版本／環境證據。
5. 已執行設定／本機連結／錨點檢查（518 個 Markdown）、文件圖（532 個文件，失效本地連結 0）及 `git diff --check`，均通過。文件驗證不等於卡片已修復。

每項開始前核對來源是否已被其他工作修改；關閉卡片必須填實際提交、驗收命令／結果與剩餘限制，更新狀態而非刪除失敗歷史。若只完成部分子案例，保留未完成範圍。實施批次、驗收歷史與產品決策各保留自己的權威，不複製一整套新規則到其他 AGENTS.md。

## 本輪實施進度

已實施 A01、A02、A03、A04、A05、A07 六張卡；A02 的本機與遠端 runner 驗收均已通過。其餘卡片維持原狀態，不因規則或測試存在就標成完成。B02／B04／B07 的產品決策仍未提供，使用者指示「能處理的先處理」，因此維持 legacy 行為，先完成不依賴這些決策的工作。

A08 尚未關閉：目前有 28 張整改卡的有限分母，但仍沒有全部業務操作／環境的已完成驗收分母。後續回報將另列卡片進度；既有 36 項 75%／50% 為歷史里程碑，不能用本輪六項修正推算成全庫完成率。

本輪共用驗收：38 項 Python 回歸通過（含已整合的索引工具基線）；API 生成／路由／型別檢查與 14 項 dispatch／resource semantics 測試通過；相對基準 `e087d96d` 的公開及內部 API 結構相容性比較通過；Go 格式、lint、模組驗證、vet、race、build、vuln 通過，漏洞檢查未發現可達漏洞。靜態前端已建置；全套 JSON 驗收包含真實 loopback emulator 與已鎖定 Playwright 的 browser 案例。文件與來源上傳驗證沒有送出雲端建置或改動正式資料。

隔離快照以 `40f1f98a` 為基準重新通過設定／引用與 38 項 Python 回歸；實際 gcloud 上傳清單包含現行 ensure-firestore-indexes.py 及 ensure-field-policies.py。另有並行前端變更維持工作樹，未納入本輪提交或發布驗收。

遠端共用驗收：[GitHub Actions 45131718514](https://github.com/963sup/line-mini-app-bot/actions/runs/45131718514) 對應 `30d1152c318b68cb738fd848e31dff4088f8e4ef`，configuration、container、go 均成功；完整 Firestore／browser race 閘門回報 169 個套件、510 個測試／子案例通過、零案例跳過。此結果驗證上述六張卡的提交內容，不涵蓋其他並行工作樹修改或正式環境驗收。
