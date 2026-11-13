# 架構延續：固定診斷期限與完整取消責任

基準為 `4e55e5456e0e1386dfbde92990d72890ec3a2604`。上一輪八檔已建立此提交；GitHub push 被自動審核拒絕，具體目的地授權仍待確認，不能稱為遠端已同步。本輪兩項新修正保持本機可審查，未 commit、未推送。環境 bootstrap 由另一執行者負責，本文只記錄架構變更與對應證據。

## 基準、排序與方法

已核對前輪完整來源的 529 個 SHA-256 均未變，重用仍適用的 Graphify、Repomix、Go／TS 指標及已通過檢查；對新問題先查圖，再核對目前來源與套件文件。本輪重新檢視 tooling／analysis、release／API scripts、typed client、Session 與元件請求生命週期，沒有重做前輪已否決的通用 HTTP／CRUD 框架。

| 優先 | 根因／影響 | 最小決策 |
| --- | --- | --- |
| P1 | MCP doctor 每收到通知或錯誤 ID 都重設等待期限，故障 server 可使診斷永不結束 | 同一 request 使用固定 monotonic deadline，不增加 MCP 框架 |
| P2 | 已取消的 untilAbort 提前返回，遺漏已啟動 Promise 的 rejection observer | 兩條取消分支均附上原有 Promise handlers，不改外層錯誤契約 |
| P3／延後 | doctor.sh 的 gcloud 版本字串與 mise pin 重複 | 當前值一致，沒有故障；不為一個字串建立設定框架 |

先處理可重現且影響操作期限／錯誤責任的狹小缺口；保留 typed transport、Session 認證與元件結果世代的不同責任，不為減行數合併三層。

## 第一輪：診斷 request 只有一個期限

[mcp-doctor.py](../../scripts/tooling/mcp-doctor.py) 的 stdio `send` 原先在循環內反覆呼叫 `Queue.get(timeout=...)`。持續通知或不匹配 ID 都會重新得到完整等待時間，而不是消耗同一 request 的額度。

只新增 `time.monotonic()` 固定 deadline，每次等待使用剩餘時間；到期仍走既有 `queue.Empty`、具名 JSON-RPC timeout 及 subprocess 清理流程。合法通知、matching response、提前退出與 HTTP transport 保持原有分工。正式腳本 162→167 行，只增加一個標準函式庫 import，無新外部依賴或工具契約。

同一隔離 fake server、設定 0.3 秒，持續通知與錯誤 ID 的修前版本均到 1.5 秒安全上限仍未結束；修後分別約 0.3021／0.3049 秒回原具名 timeout。正常握手、無回應、提前退出亦符合原語義，五種情境的 child 全部清理。這是本機樣本，非嚴格 real-time 上限或整體 MCP 效能評測。

## 第二輪：取消也要處理已啟動工作

[untilAbort](../../frontend/miniapp/src/shared/http/lifecycle.ts) 接受的是已啟動 Promise。[Session.request](../../frontend/miniapp/src/identity/Session.tsx) 可先把 signal 設為已取消，再建立 operation 並交給該 helper。舊 helper 在預先取消時立即 return，外層雖然回 AbortError，原 transport 的拒絕卻沒有 handler，另產生全域未處理拒絕。

最小修正移除提前 return：已取消則立即 reject 外層，未取消才加 listener；兩者都繼續附加既有成功／失敗 handlers。沒有 catch 吞掉正常錯誤，也沒有改 Session／client 或新增取消框架。檔案仍為 64 行、imports 不變。

六場景覆蓋預先取消且原 Promise 已拒絕／稍後拒絕／稍後成功、開始後取消再拒絕、正常成功與正常拒絕。修前兩個預先取消的拒絕各產生 1 次 unhandled rejection，修後全部為 0；外層取消仍是 AbortError，正常 transport failure 仍回原錯誤，listener 註冊與清理通過核對。

## 各輪比較與驗證

| 指標 | 基準 | 第一輪 | 第二輪 |
| --- | ---: | ---: | ---: |
| MCP doctor 行數／標準庫 imports | 162／10 | 167／11 | 167／11 |
| untilAbort 所在檔案行數 | 64 | 64 | 64 |
| Go production packages／imports | 63／170 | 63／170 | 63／170 |
| TS production files／imports | 53／151 | 53／151 | 53／151 |
| Go 跨 module／business 邊 | 94／9 | 94／9 | 94／9 |
| TS 跨 module／business 邊 | 132／5 | 132／5 | 132／5 |
| Go／TS 模組內邊比 | 44.71%／12.58% | 相同 | 相同 |
| Go／TS 非平凡有向 SCC | 0／0 | 0／0 | 0／0 |
| Graphify nodes／edges | 4,801／12,374 | 4,802／12,375 | 4,802／12,375 |

指標定義沿用 [前輪報告](architecture-iteration-20261113.md#量測定義)，同一腳本用於本輪各快照。模組內邊比只是來源結構代理，沒有以比例不變證明內聚改善，也沒有宣稱 runtime 或跨服務依賴完整。增加標準庫時間依賴是有界等待的必要成本，不追求圖或行數單向下降。

設定檢查、Python compile、五場景 fake server、六場景取消 probe、前端型別及既有 29/29 測試均通過；實際結果與原始 logs 保存在證據目錄。`graphify update .` 已執行，新增的圖節點／關係對應標準庫時間依賴，第二輪沒有拓撲變化。未變的 Go 來源沿用前輪 fmt／lint／race／build／vuln 證據；本次沒有框架或靜態輸出變更，不再重跑已驗證的 Next full build。真實 MCP 工具的安裝／握手結果由環境 bootstrap 交接另列，fake server 通過不等於全部外部服務可用。

## 再分析與停止理由

- API client 僅建立 typed transport；Session 管認證世代、續期與期限；useLatestRequest 管元件最新結果。沒有確認這些責任應合併，通用 request framework 會增加參數及生命週期耦合。
- tooling scope contract、設定檢查與 release scripts 未確認另一個收益足夠的低風險結構缺陷；生成檔、退休資產相容 URL 及遷移檢查都有既有用途，沒有因看似冗餘刪除。
- gcloud pin 的重複字串是明確後續清理候選，現值一致，影響低於前兩項；本輪不新增跨腳本設定層。
- HTTP SSE doctor 目前只有 urllib 的 blocking-operation timeout，尚未實測持續通知下的總期限；本輪只驗證 stdio request deadline，不宣稱所有 transport 都有 wall-clock 上限。若要統一 HTTP streaming 期限，須先做有界重現並驗證 response 清理，不在此次環境收尾中推測改動。
- 前輪留下的跨頁留言持久化、LINE transport 取消分類仍需契約決定，不由局部錯誤修正擴張。

目前沒有另一項已確認且成本效益高於這些界線的最小改動，因此停止實作，不把有界來源審查稱為全庫無缺陷。環境 bootstrap 的工具缺失／憑證問題與原碼架構缺陷分開處理。

## 查證與重現

Context7 已 resolve `/python/cpython`，查 `/python/cpython/v3.12.16` 的 Queue.get：timeout 是每次呼叫的期限；同來源另查 urllib 的 timeout 是 blocking-operation timeout，未將它當作整個串流期限。本機固定 Python 為 3.12.14，同系列而非相同 patch。取消語義先 resolve MDN，再 query `/mdn/content` 的官方 [AbortSignal](https://github.com/mdn/content/blob/main/files/en-us/web/api/abortsignal/index.md) 與 [unhandledrejection](https://github.com/mdn/content/blob/main/files/en-us/web/api/window/unhandledrejection_event/index.md)；文件是 main，不代表真實瀏覽器版驗收。

原始來源、修前／修後 probes、每輪 metrics、Graphify 更新及 scoped Repomix 均放在 `/workspace/tooling/reports/architecture-continuation-20261113/`。沿用目前 [驗收階段限制](../agents/acceptance.md)，未新增正式測試或平台，暫存回歸檔保留可重現：

```bash
bash scripts/tooling/tool-exec.sh python3 /workspace/tooling/reports/architecture-continuation-20261113/probe-mcp-deadline.py
TSX_TSCONFIG_PATH=/workspace/line-mini-app-bot/frontend/miniapp/tsconfig.json \
bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp exec node --import tsx \
  /workspace/tooling/reports/architecture-continuation-20261113/abort-lifecycle-probe.ts
bash scripts/tooling/tool-exec.sh python3 scripts/tooling/check-config.py
bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp check
bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp test
```

[工具索引](README.md) · [文件索引](../README.md)
