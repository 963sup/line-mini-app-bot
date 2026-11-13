# 架構迭代：來源錯誤邊界與留言生命週期

本次基準為乾淨的 `2bf4722599b6ac30435ebe1bc29f070b97d96b42`。依使用者要求，只保留本機可審查變更，不 commit、不推送、不部署。先量測全庫來源結構，再按風險與收益排序；不以節點數、行數或重複外觀作為重構理由。

## 基準與排序

Graphify 使用前次剛重建的圖，Repomix 使用同提交的完整 XML，兩者交叉核對 333 個圖來源均存在且包含於 528 檔打包。原始證據在 `/workspace/tooling/reports/remote-graphify-repomix-20261113-2bf4722/`；本次各輪與臨時驗證在 `/workspace/tooling/reports/architecture-iteration-20261113/`。

| 優先 | 候選與根因 | 最小決策 |
| --- | --- | --- |
| P1／第一輪 | Project 同一用例的來源授權階段吞掉取消／逾時，來源 hydration 階段卻保留；失敗分類不一致 | 保留標準取消錯誤，只改既有 guard，不建立共用錯誤框架 |
| P1／第二輪 | Discussion 父層可卸載正在寫入／結果不明的 Comments，丟失子層原防重鍵 | 只協調會破壞子樹的操作，留言狀態及請求仍由 Comments 擁有 |
| 延後 | HTTP Register 外觀重複，但有認證、CSRF、Project items 例外及能力訊息差異 | 沒有已確認收益足以抵銷共用框架與整合驗證成本 |
| 保留 | server fan-out 高、identity／HTTP 共用 fan-in 高 | 現有 DI 組裝與穩定平台服務，不因 degree 拆模組 |
| 保留 | Project legacy resolver 與 marker | 有明確遷移／相容契約，沒有移除窗口與使用證據 |

第一性原理是保住操作、狀態及錯誤的擁有邊界；80/20 用於先處理有可重現影響的兩個狹小缺口；奧卡姆剃刀用於拒絕沒有必要的 controller、provider、通用 CRUD 或錯誤框架。

## 量測定義

Go 圖收錄 `backend/internal` 中有手寫 production Go 檔的 package，排除測試、cmd、生成 API 及外部依賴；直接 import 邊由來源抽取，已與同來源 `go list -json ./...` 精確核對。舊報告的 65 packages 含兩個僅有測試的 package，本輪為 63；這不是重構收益。

TS 圖收錄 `frontend/miniapp/src` 的字面靜態 import／export，含 type imports；排除 `.d.ts`、官方 `shared/ui` 與 `sidebar-07` 模板、外部及無法解析的 import。Module 為 `business/<domain>`，其餘以第一層目錄分組。Cycle 是有向圖大小大於 1 的強連通分量，不由 Graphify 無向關係推論。

模組內邊比＝兩端同 module 的 import 邊／全部 import 邊，只是結構代理；共用 React、Session 與呈現元件的正常依賴會降低比例，不能當成內聚品質分數。沒有 runtime、效能或未來交付修改模組數的量測。

同一 `metrics.py` 用於各輪，保存來源 SHA-256、精確邊及統計；不可將本次數字與排除定義不同的歷史報告直接相減。

## 第一輪：統一 Project 取消語義

[Project.progress](../../backend/internal/business/project/project.go) 原先對 `Sources.Authorize` 的所有錯誤回成功降級；同函式的 `Store.Progress` 則排除 `context.Canceled`／`DeadlineExceeded`。Repository 讀取錯誤可沿 `projectSources.Authorize` 原樣傳回，這是實際可達的用例邊界，不只是理論上的 port。

[Project 契約](../features/projects.md#操作契約) 要求取消／逾時仍失敗。修正只在第一個 guard 以 `errors.Is` 保留原錯誤；撤權或來源不可用的 canonical metadata 降級、legacy 拒絕、外部 API 與資料均不變。正式程式一行替換，淨行數 0、新依賴 0。

外部 Go overlay 對 Get／List／Update × cancel／deadline × 原生／包裝錯誤共 12 項，修前 0/12、修後 12/12 通過；既有 Project 套件與撤權／hydration 測試通過。`gopls references` 找到 5 個呼叫點，CLI `gopls check` 無診斷。Graphify 仍為 4,801 節點、12,374 關係、285 communities，來源拓撲未變；此輪收益是已確認的錯誤邊界不一致由 1→0。

## 第二輪：保住 Comments 的寫入狀態

[Discussion View](../../frontend/miniapp/src/business/discussion/View.tsx) 的 `open`／`create` 原先只看主題自己的寫鎖，會清除 `selected` 並卸載 [Comments](../../frontend/miniapp/src/business/discussion/Comments.tsx)。留言 POST 的 `attempt` 防重鍵卻只存在子元件；父層重新讀取、切題或新建時，後端可能已寫入而原鍵已遺失。[討論契約](../features/discussions.md) 要求結果不明時保留本次畫面的原鍵重試。

只新增父子同步握手：留言開始寫入前，父層確認沒有正在卸載子樹的導覽；接受後以 ref 立即阻擋父層 `open`／`create`，以 state 停用三類對應按鈕。POST 成功或明確拒絕解除，結果不明維持原鍵與阻擋；PATCH 完成解除，但失敗後仍須重新讀取。留言草稿、請求及 attempt 仍由 Comments 擁有，不建立 provider、全域鎖或共用 controller。

初版 DOM probe 重現父層按鈕未受阻；反向同批事件另重現「父導覽先開始、舊留言表單稍後仍送出 1 POST」。最終同步 boolean 握手使後者為 0 POST。七個場景全部通過：pending、結果不明原鍵重試、成功、明確拒絕、子先／父後同批事件、PATCH pending／失敗後 refresh、父先／子後同批事件。主題 metadata 更新不卸載同 ID 的留言，仍可與留言合法並行，沒有宣稱全子樹單寫鎖。

正式兩檔淨增 15 行，無新 import、package、provider 或公開 API。會破壞留言子樹的兩個父層操作函式由未協調 2→0；這是特定修改位置的代理，不是全庫耦合度下降。既有前端型別檢查、29/29 測試及 Next 靜態 build（14 個路由）通過，官方 UI primitives、use-mobile、sidebar-07 的 diff 為空。

## 比較與停止條件

| 指標 | 基準 | 第一輪 | 第二輪 |
| --- | ---: | ---: | ---: |
| 手寫 production Go／TS 實體行數 | 13,610／4,664 | 13,610／4,664 | 13,610／4,679 |
| Go packages／import edges | 63／170 | 63／170 | 63／170 |
| Go 跨 module／跨 business 邊 | 94／9 | 94／9 | 94／9 |
| Go 模組內邊比 | 44.71% | 44.71% | 44.71% |
| TS files／import edges | 53／151 | 53／151 | 53／151 |
| TS 跨 module／跨 business 邊 | 132／5 | 132／5 | 132／5 |
| TS 模組內邊比 | 12.58% | 12.58% | 12.58% |
| Go／TS 非平凡 SCC | 0／0 | 0／0 | 0／0 |
| Graphify nodes／edges／communities | 4,801／12,374／285 | 4,801／12,374／285 | 4,801／12,374／285 |

最高 Go fan-out 是 server 46、Attendance HTTP 10、Attendance Firestore 9；最高 fan-in 是 identity 12、Repository 與 HTTP transport 各 11。TS 最高 fan-in 為 lifecycle 22、Status 與 Session 各 20。這些結構熱點不是實際呼叫頻率，也沒有直接證明循環或職責錯置。

### 第三輪再分析：停止增加抽象

修後重新查圖、核對來源及量測，沒有發現需要處理的有向 import cycle。再對候選作以下判斷，因此本次停在兩輪實改，而非為達迭代次數繼續搬動程式：

- server 內部入邊為 0，現有組裝根沒有反向依賴；拆成更多 helper 不減實際耦合。
- Project HTTP body 上限 65,536 且 items POST 有防重鍵例外，Labels 上限 1,024 且所有 POST 要求鍵；共用 Register 至少增加容量、路由與安全政策參數，未有可觀察故障支持成本。
- Issue／Discussion translator 仍映射各自領域錯誤與 Repository 分類；Project 還有來源故障隔離。共用 callback 或映射表沒有消除領域責任，反而新增間接層。
- Project 交易外的 legacy resolver 與交易內 marker／lock 檢查責任不同，移除會影響遷移競態防護；正式遷移及舊 caller 窗口尚無移除證據。
- 根核心可達 module 集合未增加：Repository／Space／Attendance／Summary 各 1；Project 為 Project＋Space，Issue／Discussion 各為自身＋Repository，均為 2。沒有以新增共用層換取表面減行。

仍保留兩個明確限制：本次導航協調只涵蓋 Discussion 父層，切換 Repository tab、身份失效或整頁離開仍會卸載；LINE Membership 將 HTTP client failure 統一分類為 unavailable，本次只保留 port 已提供的標準取消，未擴張平台錯誤契約。前者若要跨頁保障須有持久化／導航契約，後者須先確認平台分類與領域公開錯誤邊界，不能在局部收斂中猜測改動。

## 最終整合驗證

| 實際檢查 | 結果 |
| --- | --- |
| Go fmt-check／lint | 通過，lint 0 issues |
| Go check | module verify、vet、完整既有 `go test -race ./...` 均通過 |
| Go build／vuln | 通過；0 個應用可達漏洞，另有 3 個 required-module 項目未被應用呼叫 |
| Project overlay／gopls | 12/12 回歸通過，既有 Project tests 通過，gopls 無診斷 |
| 前端 check／test／build | 通過；29/29 tests、14 個靜態路由 |
| Discussion 暫存 DOM probe | 7/7 通過，含雙向同批事件 |
| 文件本機路徑／錨點 | 本批五份 Markdown 共 172 個本機引用通過 |
| git diff --check／官方 UI diff | 通過／保護元件無變更 |

Go suite 未啟動 Firestore emulator，需要 emulator 的條件式案例不作本輪資料庫驗收；本次只改純用例取消分類與前端導航協調，未改儲存交易。API、生成碼、依賴版本及 lockfile 無變更。未為相同內容重跑無關 API 生成或正式雲端驗收。

已執行 `graphify update .`。因舊本機圖增量保留額外節點，再依既有指南以 `extract . --code-only --force --allow-dedup-shrink --out .` 重建，使本機圖與獨立圖均為 4,801 節點／12,374 關係。Root 重建當次 communities 為 274，兩輪獨立抽取為 285；分群沒有固定 seed，不把這個差異算成改善。逐輪比較仍使用同種獨立抽取的結果。

## 查證、重現與限制

Context7 已 resolve Go，再 query `/golang/go/go1.27.3`（最接近已安裝固定 Go 1.27.4 的可用版本）；回覆對 errors.Is 細節不足，另以 Go 1.27.4 `go doc errors.Is`、`context.Canceled` 補足包裝錯誤及取消語義。React 已 resolve `/reactjs/react.dev`，查官方 [狀態保存與重設](https://github.com/reactjs/react.dev/blob/main/src/content/learn/preserving-and-resetting-state.md) 及 [共同父層協調](https://github.com/reactjs/react.dev/blob/main/src/content/learn/sharing-state-between-components.md)；文件為 main，實際依賴 React 19.3.1，不宣稱文件固定同版。Next 16.4.5 另讀安裝版 StrictMode／static-export 指引，未改 router、樣式或框架。

shadcn MCP 未曝光且固定 CLI 未安裝，沒有把設定存在當作成功；本次未新增或修改官方元件，只沿用既有 Button 的公開 `disabled`。以本機 `React.ComponentProps<"button">`／props spread 與 Context7 `/shadcn-ui/ui` 的官方 Radix disabled 範例補核對，未新增安裝或改動 primitives。

重現命令（在儲存庫根目錄執行，證據目錄代入上述本機絕對路徑）：

```bash
bash scripts/tooling/tool-exec.sh go -C backend test \
  -overlay /workspace/tooling/reports/architecture-iteration-20261113/source-cancel-overlay.json \
  ./internal/business/project -count=1 -v
python3 /workspace/tooling/reports/architecture-iteration-20261113/metrics.py \
  /workspace/line-mini-app-bot /tmp/current-metrics.json
bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp check
bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp test
bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp build
TSX_TSCONFIG_PATH=/workspace/line-mini-app-bot/frontend/miniapp/tsconfig.json \
bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp exec node \
  --import tsx \
  --import /workspace/line-mini-app-bot/frontend/miniapp/tests/dom-setup.ts \
  /workspace/tooling/reports/architecture-iteration-20261113/discussion-lifecycle-probe.tsx
```

DOM probe 的實際 node／tsx 命令、兩次修前失敗與七場景成功紀錄，連同 source-cancel overlay、gopls、各輪 metrics、原始圖、Repomix 與檢查 logs 均保留於本次證據目錄。Graphify／Repomix 不涵蓋完整動態行為，圖的大小不變不能證明語義未變；安全掃描成功也不是無機密證明。

本輪遵循 [驗收規範](../agents/acceptance.md) 的「do not add test plans, cases or tools」及 [Go 開發規範](../development/go.md) 的測試開發暫緩要求；未新增正式測試或平台，使用外部 overlay／DOM probe 作臨時回歸並保留重現檔，既有檢查照常執行。這些本機證據不能代替真實 LINE、Firestore、手機、正式遷移或部署驗收。

[工具索引](README.md) · [文件索引](../README.md)
