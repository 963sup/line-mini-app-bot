# 36 項標準化驗收紀錄

範圍為目前產品的本地與 loopback Firestore emulator 驗收。Organization 的既有產品主線維持 legacy 模式；Membership 已有邀請、資格與 Team 接線，但獨立模式的完整正常入口、資料切換與回復仍待交付，不以直接修改模式欄位代替交付。

待辦及實施順序統一見[全庫盤點與批次整改清單](standardization-backlog.md)。本頁保存已完成的驗收與歷史證據，不把待驗證項目算成通過。

## 計分與限制

沿用聊天中的證據里程碑：權威規範、實作、適用自動檢查、完整驗收各占 25%。這不是按檔案或操作統計的實測符合率。局部案例通過不代表全庫業務與目標環境均已驗收，因此不將所有項目調成 100%。

前一輪補齊 Team 寫入生命週期條件與自有 TypeScript 格式門檻。本輪補齊所有 business 核心套件的跨領域依賴閘門及 Membership 導覽，尚未取得完整產品／環境驗收，因此不調高進度。第 26 項維持 50%，其餘維持 75%，36 項等權平均為 74.3%。

| 項目           | 百分比 | 已有證據／尚缺範圍                                                                                                                    |
| -------------- | -----: | ------------------------------------------------------------------------------------------------------------------------------------- |
| 1 領域         |    75% | [領域矩陣](../../domains/)含 Membership 的部分實作狀態；逐操作語意驗收仍有限制                                                        |
| 2 資源生命週期 |    75% | Team 綁定及指派於交易內拒絕非 ACTIVE／停用狀態；全資源生命週期尚未逐項驗收                                                            |
| 3 命名         |    75% | 命名規範、API 能力檢查及 Go lint                                                                                                      |
| 4 識別         |    75% | 各資源契約與格式驗證；不任意更改公開或持久化 ID                                                                                       |
| 5 時間         |    75% | UTC instant、業務時區及邊界測試                                                                                                       |
| 6 結構         |    75% | 現有模組配置及架構測試                                                                                                                |
| 7 依賴         |    75% | [跨領域核心契約](../../architecture/runtime/dependencies.md#現行跨領域核心契約)逐套件核准；新套件同樣受檢查，不代替動態呼叫與授權驗收 |
| 8 格式         |    75% | Go 格式／lint；自有 TS／TSX／MTS 的固定版本 Prettier 與 CI 門檻                                                                       |
| 9 文件         |    75% | 規格同步、本機路徑及錨點檢查                                                                                                          |
| 10 型別        |    75% | Go／TypeScript 型別檢查及產生型別                                                                                                     |
| 11 資料契約    |    75% | OpenAPI、輸入驗證與儲存邊界測試                                                                                                       |
| 12 API         |    75% | 現行契約、路由及能力歸屬檢查                                                                                                          |
| 13 錯誤        |    75% | 共用錯誤與 recovery metadata；保留既定內部入口格式                                                                                    |
| 14 查詢        |    75% | 分頁、游標範圍及失敗重試案例                                                                                                          |
| 15 生成        |    75% | Go／TS 產生碼與來源一致性檢查                                                                                                         |
| 16 設定        |    75% | 載入、工具版本及 Secret reference 檢查；線上設定另驗收                                                                                |
| 17 安全        |    75% | Team 停用拒絕、身分與保護邊界測試；實際 IAM／Secret 狀態未驗收                                                                        |
| 18 身分        |    75% | Session／token 的拒絕、續期與撤銷案例；真實 LINE 裝置未驗收                                                                           |
| 19 授權        |    75% | 交易內權限與 Team 生命週期共同判斷；全角色組合未逐項驗收                                                                              |
| 20 狀態        |    75% | 領域狀態與 Team 非 ACTIVE 拒絕案例                                                                                                    |
| 21 交易        |    75% | Emulator 與 race 測試；Team 狀態由同一交易取得                                                                                        |
| 22 併發        |    75% | 版本條件與 emulator／race 案例                                                                                                        |
| 23 冪等        |    75% | 收據、固定鍵及重試案例；外部服務實際投遞另驗收                                                                                        |
| 24 UI          |    75% | 共用元件、171 項前端測試及 Chromium 專案案例；LINE 裝置另驗收                                                                         |
| 25 表單        |    75% | 生成型別、欄位錯誤與草稿／原請求保留案例                                                                                              |
| 26 資料遷移    |    50% | 既有格式相容處理；獨立 Membership 切換及回復不在本次範圍                                                                              |
| 27 可靠性      |    75% | 期限、取消、重試與本地故障案例；真實依賴另驗收                                                                                        |
| 28 日誌        |    75% | 結構化錯誤、關聯 ID 與敏感資料排除測試                                                                                                |
| 29 可觀測性    |    75% | OpenTelemetry log exports、背景 trace 傳遞及指標測試；雲端採集／告警另驗收                                                            |
| 30 測試        |    75% | 本次 unit／contract／emulator／browser 證據；不視為全部業務路徑覆蓋                                                                   |
| 31 建置        |    75% | 固定版本、依賴鎖檔及靜態建置；獨立乾淨環境重建另驗收                                                                                  |
| 32 部署        |    75% | 發布與回復流程及本地檢查；未執行線上部署或回復演練                                                                                    |
| 33 版本        |    75% | 指定 Git 基準的保守 API 相容性檢查                                                                                                    |
| 34 開發流程    |    75% | 自我審查、CI 與驗收矩陣；不宣稱獨立審查                                                                                               |
| 35 事件        |    75% | 既有事件與工作去重／排序案例；不為一致性強加 CloudEvents                                                                              |
| 36 快取        |    75% | no-store、範圍隔離及舊回應清理案例；未新增無量測需求的快取                                                                            |

## 既有功能驗收基線

程式與格式化驗證對應 `80d2804f`（Team 生命週期修正）及 `4de9baf1`（格式化與 CI）；後續變更後，依[證據重用條件](../workflow/handoff.md#minimal-handoff-and-evidence-reuse)決定哪些結果須重跑。

- `pnpm format:check`：自有 TS／TSX／MTS 格式通過。Prettier 版本由根 package.json／鎖檔固定，範圍由該命令與 `.prettierignore` 決定；生成碼與受保護上游元件排除。
- `pnpm api:check`：型別、11 項 dispatch 測試、109 個 operation、契約及產生碼一致性通過。
- 前端型別、171 項測試及靜態建置通過。
- `go mod verify`、`go vet -p=1 ./...` 及 `FIRESTORE_EMULATOR_HOST=127.0.0.1:8080 go test -race -p=1 -count=1 ./...` 通過；未提供瀏覽器 export 的案例由下一項單獨實跑。
- 設定 `MINIAPP_BROWSER_EXPORT` 並使用固定 Playwright 1.62.0／系統 Chromium 後，`TestPersonalRepositoryAndProject` 與 `TestProjectSettingsBrowserRoundTrip` 實際通過。LINE 身分建立使用 fixture，不能作為真實 LINE 證據。
- `TestTeamWritesRequireActiveLifecycle` 於 emulator 實際驗證 ACTIVE 成功，以及 INACTIVE、停用、缺少狀態拒絕；涵蓋管理綁定、指派與移除。
- Go lint、23 項工具測試及設定／本機連結檢查通過。

執行入口與環境前提見[驗收矩陣](../tooling/testing/commands.md)。本次沒有部署、套用線上 TTL、讀取 Secret payload 或改動正式資料。

## 本輪全庫盤點與優先序

盤點從 Repomix 參考技能定位，再核對目前原碼、API 能力清單、領域矩陣與架構檢查。Graphify MCP 索引停在 `7bec2eee`，落後本輪基準 `9d4edd5d`；只用於導航。靜態掃描覆蓋所有 business domain／application 原始碼，不等於逐一驗收所有業務路徑。上方 36 項是里程碑，不是以測試數量推算的合規覆蓋率。

| 優先序／標準 | 發現、處理及完成條件                                                                                                                                                                                    |
| ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| P0：1–9      | **已修正**：未列入清單的新核心套件可直接引用其他領域。現在跨領域目標須逐套件核准；原有必要引用及同領域分層保留。Membership 已補入領域矩陣，實作、部分接線與未完成切換分開標示。                         |
| P1：10–15    | 現有契約生成、型別、路由能力及指定基準相容性檢查持續作為證據入口。新增業務操作須先補契約及對應拒絕／重送結果；不能把生成通過當所有輸入、錯誤與分頁情境已覆蓋。                                          |
| P2：16–23    | Team 生命週期已由前輪交易及 emulator 案例驗證。本輪維持現有授權與角色語意。獨立 Membership 的正常入口、上下層資格及撤權競爭仍依[閉環與待定規則](../../proposals/membership.md#閉環判定與待定規則)交付。 |
| P3：24–26    | 既有前端回歸、型別與靜態輸出驗收；表單輸入提示不替代 Backend 權限判斷。資料遷移須補上正常操作、準備、切換與回復入口及跨版本驗收，不能僅修改模式欄位。                                                   |
| P4：27–34    | **已修正**：移除驗收矩陣的「不使用 emulator」與 Membership 提案的「限制新增測試案例」舊說明，改引用同一驗收矩陣。部署、IAM、Secret、真實 LINE 裝置與回復演練仍需各自的環境證據。                        |
| P5：35–36    | 維持既有事件、工作去重與快取隔離；新增事件協定、快取或可靠投遞機制前，先取得操作需求、失敗窗口或效能量測，不以標準化名義新增通用機制。                                                                  |

### 後續交付的具體邊界

1. **Organization 獨立資格。** 依[領域交付邊界](../../domains/organization/design.md#獨立資格的交付邊界)完成建立、邀請、People、Owner 移轉及退出／移除的正常入口，驗證唯一有效 Owner、重送、並行撤權與下層 generation 失效，再交付切換與回復。此範圍有既定規格，不能僅用模式切換宣告完成。
2. **Team 首次開通。** [Team 開通契約](../../domains/team/design.md#群組開通授權)已列出申請／審核／消耗方案；正式啟用仍須定案渠道審核人員、可採用的獨立委任證據與效期，並從正常入口驗收。自動觀察到 LINE 群組不等於申請人有管理權。
3. **Enterprise 獨立資格。** [多 Owner 與 manager 操作](../../domains/enterprise/design.md#獨立資格與治理的交付邊界)尚有產品決策；不能複製 Organization 的單一 Owner 政策。既有 Enterprise 使用 Organization 錯誤值的相容耦合已列於[依賴契約](../../architecture/runtime/dependencies.md#現行跨領域核心契約)，改動前須先核對 HTTP 錯誤映射及消費端。

### 本輪回歸證據

`TestUnregisteredBusinessContextBoundaries` 在臨時儲存庫執行真正的架構掃描器。修正前，未登記 domain、application 及 Membership 的三項跨領域拒絕案例失敗；修正後，八個正反案例及現行全庫 import 掃描通過。案例涵蓋新領域、既有未列核心、已核准套件擴大引用的拒絕，以及同領域、Repository 值型別、成員引用與既有錯誤相容性的保留。

來源：[掃描器](../../../backend/internal/architecture/boundaries_test.go) · [回歸案例](../../../backend/internal/architecture/boundaries_regression_test.go)。本輪變更未修改公開 API、持久資料或業務授權結果；靜態邊界檢查不代替業務端到端驗收。

本輪盤點涵蓋 17 個業務目錄（含候選領域及子能力）與 83 個 domain／application 套件；目錄數不是獨立 bounded context 的數量。2026-11-23 實際驗證：

- `gopls check`：兩個變更的架構檢查檔無診斷。
- `go test ./internal/architecture -count=1 -v`：現行依賴掃描與八個臨時儲存庫正反案例通過。
- `mise run fmt-check`、`lint`、`check`、`build`、`vuln`：通過；`check` 使用 loopback emulator 執行，包含模組驗證、vet 與 race suite，沿用仍有效的快取。漏洞檢查未發現可達漏洞。
- `pnpm format:check`、`pnpm check`：通過，包含契約／生成碼、型別、靜態輸出與設定檢查；前端 171 項及 Python 23 項回歸通過。
- 文件本機連結／錨點及 whitespace 檢查通過。DDD 參考資料及上游快照維持原狀；沒有執行雲端寫入或資料模式切換。

## 第二輪：企業開通重送的生命週期

基準為上一輪已發布的 `2f4b46d1`。現行 `POST /api/v1/enterprises` 在開通收據已有 enterpriseId 時，只核對舊 Owner 欄位；企業已停用、切至 `membership` 或未知模式仍回 200。這讓建立重送的授權與現行 legacy 企業生命週期不一致。

修正沿用 Enterprise 已有的 `legacyEnterpriseEnabled` 判斷，在原交易內同時核對目前開通資格、企業生命週期與 Owner，不新增授權來源、資源或資料遷移。OpenAPI 補上既有 403 的具體語意，重新生成 Go 與 TypeScript 契約；有效的空值／legacy 模式及舊單一 Owner 格式仍可回放。

[HTTP／Firestore 回歸案例](../../../backend/internal/business/enterprise/adapter/firestore/provisioning_test.go)涵蓋八種狀態。修正前，停用、獨立模式與未知模式三例錯回 200；修正後八例通過，並核對回應、企業筆數，以及企業與開通資格均未被重送改寫。既有 Bot 建立對話的 emulator／race 案例亦通過。

36 項里程碑維持：1–25、27–36 各 75%，26 為 50%，等權平均 74.3%。本輪增加第 2、17、19、20、21、23 項的具體證據，不將單一操作修復當成整項全庫驗收完成。尚未啟用獨立 Membership、修改正式資料或執行部署。

本輪發布驗證以獨立 Git index 快照為準，排除其他工作中的工時計算與 Journal 表單修改。`go test -race -p=2 ./...`（loopback emulator）、lint、vet、release build、漏洞檢查均通過；API 生成同步、對 `2f4b46d1` 的契約相容性比較、前端型別、文件連結／錨點及 whitespace 檢查亦通過。首次全套驗證因磁碟滿而中斷，清理已完成工作快照及可重建 Go 快取後重跑；快照的 browser 測試補齊既有靜態輸出後通過，未把環境失敗列為成功證據。DDD 的 69 個參考檔案維持原狀。

## 整改清單實施：交付門檻、錯誤邊界與資源語意

本輪實施 [A01／A02／A03／A04／A05／A07](standardization-backlog.md#本輪實施進度)：修正 Cloud Build 來源漏檔與 CI emulator 覆蓋；Enterprise 核心改用自有錯誤並維持公共 HTTP 語意；修正文檔測試入口及引用支援；23 個 HTTP 資源條目加入已檢查的 owner／representation kind，內部 inventory format 升為 2，HTTP API 版本及持久化格式維持相容。

最終本機驗收：169 套件、510 個 Go 測試／子案例，零跳過；38 項 Python 與 14 項 dispatch／resource semantics 測試通過。Go 格式／lint／vet／race／build／vuln、API 同步／型別／基準相容性、前端靜態建置及本地文件檢查通過。遠端 CI 結果在清單卡片另列。

目前卡片分母及未完成項目見[36 項進度表](standardization-backlog.md#36-項覆蓋與進度)。未定案的治理／保存政策維持待決策，不變更 legacy 行為。

實施提交：`76ef1291`（交付／驗收閘門）、`0924735e`（Enterprise 錯誤邊界）、`7498c7bd`（HTTP 資源語意）。另整合既有並行索引基線 `40f1f98a`；其他未完成的並行前端內容及 DDD 修改不納入。
