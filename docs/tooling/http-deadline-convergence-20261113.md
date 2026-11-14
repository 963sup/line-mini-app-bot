# HTTP 診斷期限與回歸驗證收斂

以乾淨的 `db65fb69ce844a730a72007af8e5d724c1e64335` 為基準，531 個來源 hash 與前輪已驗證內容一致。本輪只做本機可審查修改，未 commit、推送或部署；沒有重測未變的 Go／TS 應用。目標是實際解決前輪已確認的 HTTP SSE 診斷無界等待，不再只重述限制。

## 根因、排序與選擇

[mcp-doctor.py](../../scripts/tooling/mcp-doctor.py) 原先只有 urllib 的單次 blocking-operation timeout；SSE 通知、錯誤 ID 或持續未完成的資料可讓整次 request 永不結束。本機真實 fake server 曾超過 25 秒，26.591 秒才被外層安全界中止。新增正式回歸後，舊版本五種卡住情境全部超過四秒外層安全界，合計 21.114 秒、五項失敗。

選擇用一個小型標準庫子程序隔離完整 HTTP exchange。這會增加每次交換的程序建立成本，但可由父程序終止阻塞網路工作並回收 socket。沒有新增 HTTP 套件、client framework、全域 signal handler、private socket 存取或 multiprocessing fork。Threads＋close 無法可靠終止阻塞工作；全域 alarm 有主線程與巢狀 timer 的額外責任，未採用。

## 最小實作與獨立審查修正

- 父程序仍擁有 RPC、session／protocol headers、固定 monotonic deadline及子程序清理；`Popen.communicate` 只取得剩餘額度，逾時或中斷均 kill／reap。
- 新增 [mcp-http.py](../../scripts/tooling/mcp-http.py)，把原 urllib open／SSE／JSON 邏輯搬成唯一 HTTP 交換來源。URL、headers、message及session經 stdin JSON 傳入；argv與錯誤輸出不含憑證，worker只輸出一個結果或安全錯誤分類。
- 保留正常 JSON／SSE、202空通知、無新session header時的原session及後續protocol header。JSON錯誤ID的舊處理行為沒有順手改動。
- 獨立審查發現兩個診斷缺口並最小修正：SSE EOF仍回固定安全的 `no JSON-RPC response`；HTTP通知逾時顯示實際phase，不再只顯示 `Empty`。其他exception仍只傳安全類型，不廣泛序列化exception文字或response body。

HTTP採用既有 `startup_timeout_sec` 作每次交換額度，未配置時維持原HTTP預設；stdio預設不變。期限計入回到父程序後可觀測的startup耗時，但OS初始process creation可能不可中斷，不宣稱硬即時保證。

## 實際驗證與 CI 整合

正式 unittest涵蓋原stdio及HTTP：持續通知／錯ID、partial SSE line、JSON body及headers停住、正常JSON／SSE、202通知與session／protocol、HTTPError／EOF、initialized通知phase逾時，以及真實SIGINT中斷。每個HTTP結果均檢查child已reap、server觀察連線關閉，並檢查secret不出現在argv或stdout／stderr。

第一個CI discover曾有兩項fixture失敗：所有情境共用0.5秒額度，使HTTPError先被取消，中斷測試也未進request-ready。沒有以重試掩蓋；改為stall、正常分類與中斷分組budget，保留共同外層安全界及全部診斷／清理斷言。唯一最終CI相同命令實跑 **8/8 tests通過，29.409秒**；設定檢查、Python compile與whitespace通過。

[既有 configuration CI job](../../.github/workflows/check.yml) 新增標準 `unittest discover` step，只用標準庫與本機fake server，不增加測試平台、外部服務或Go gate依賴。本機同命令已通過，本輪未推送，未宣稱遠端CI已執行。

代表性真實HTTP smoke只執行已配置的OpenAI公開文件MCP initialize／tools-list，**PASS，5 tools**，驗證proxy／CA與inherited environment可用；沒有tools/call，也沒有重跑14個可選服務或ADC登入。

## 規模、依賴與比較

| 指標 | 基準 | 修改後 |
| --- | ---: | ---: |
| doctor production行數 | 167 | 177 |
| HTTP worker行數 | 0 | 33 |
| 本次production總行數 | 167 | 210（＋43） |
| 正式Python回歸行數 | 97 | 267（＋170） |
| 正式Python top-level tests | 3 | 8 |
| 新外部依賴／framework | 0 | 0 |
| HTTP實作來源位置 | 1 | 1（搬入worker） |

父程序import數仍為11，worker只用3個標準庫import；新增一條父程序執行worker的明確依賴，worker不反向import父程序。沒有把增加程序邊界說成減少所有耦合，收益是逾時與中斷的生命週期可完整終止。

Go production仍為63 packages／170 imports，TS仍為51 files／148 imports／4,067行；跨business imports仍為9／5、非平凡SCC仍為0／0。模組內邊比沿用44.71%／12.16%，只是結構代理，不把未變指標當內聚改善。Graphify更新、原始metrics與來源hash另存證據目錄；圖包含新增測試，不把節點增加當品質結論。

## 再分析與停止理由

獨立唯讀審查追蹤Session核心、Cookie／CSRF／Bearer邊界、Firestore renew與撤銷、出勤cursor。現有責任與契約一致，未確認另一個值得最小改動的缺陷；不為nilguard、不同領域錯誤分類或語法重複建立helper／介面。typed client、Session與元件世代隔離也保留原分工。

已完成本輪可證明收益的期限修正、診斷保留與CI收錄。剩餘改動若無可觀察根因，成本高於已證明收益，因此停止，不以迭代次數要求製造工作；這不是全庫無缺陷證明。沒有真實LINE、Firestore、手機、雲端資源或部署驗收。

## 來源與重現

Graphify先定位，再以Repomix及目前來源核對。Context7使用 `/python/cpython/v3.12.16` 的 urllib／subprocess文件，與本機Python3.12.14核對；官方文件明訂timeout後kill／wait及OS process creation例外。相關API來源：[subprocess](https://github.com/python/cpython/blob/v3.12.16/Doc/library/subprocess.rst)、[urllib](https://github.com/python/cpython/blob/v3.12.16/Doc/library/urllib.request.rst)。

證據目錄：`/workspace/tooling/reports/http-deadline-convergence-20261113/`，包含baseline531個hash、red／green／最終CI logs、source-size、Graphify、metrics、代表性live smoke、最終來源bundle與交叉核對。唯讀Session審查來源另在 `/workspace/tooling/reports/session-boundary-review-20261113/`。

```bash
bash scripts/tooling/tool-exec.sh python3 -m unittest discover -s scripts/tooling -p 'test_*.py' -v
bash scripts/tooling/tool-exec.sh python3 scripts/tooling/check-config.py
```

[工具索引](README.md) · [文件索引](../README.md)
