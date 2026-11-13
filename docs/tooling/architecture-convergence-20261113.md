# 儲存庫架構收斂：基準、實改與對照

基準日期：2026-11-13（UTC；檔名沿用基準日，交付日期依Asia/Taipei）。開始先fetch並核對本機work與遠端work／main均為 `87e736320bc82148f8a0eef5dca41d2be46cd93c`，工作樹乾淨，沒有未推送提交。先建立完整來源基準、指標與排序，再修改本批已確認的問題；沒有為降低節點數合併業務或引入新框架。

## 固定範圍與方法

- 全量來源遵循 [analysis-scope.toml](../../analysis-scope.toml)，以既有audit建立隔離來源、逐檔SHA及完整Repomix XML。基準518個managed檔，程式修改後523檔；後者包含新增維護測試，本報告及索引之後才加入。排除技能副本、產物、binary等理由沿用契約，沒有關閉安全掃描來縮數字。
- Graphify `0.9.103`使用code-only完整抽取，節點／關係包含生成碼與測試，與手寫production指標分列。圖為**無向圖**，EXTRACTED／INFERRED不是實際執行紀錄。Community使用CLI預設，未另固定或查證cluster seed；數量只報該次結果，不用品質歸因。
- Go production有向依賴：`go list -json ./...`的直接Imports，只收`backend/internal`本機packages，排除`httpapi/generated`、測試、cmd及外部依賴。Module定義為`business/<domain>`，其他以internal第一層目錄分組。
- TS production有向依賴：`frontend/miniapp/src`中的.ts／.tsx字面靜態import／export，包含type imports；排除.d.ts、官方shared/ui及sidebar-07模板，外部package或無法解析的import不加入。這是來源結構圖，沒有宣稱涵蓋HTTP呼叫或動態載入。
- Direct calls另用固定`golang.org/x/tools/cmd/callgraph@0.52.0 -algo=static`；只收兩端均為手寫internal函數的邊，排除生成碼／測試／cmd／外部端點。它解析直接非多型呼叫，**不涵蓋interface dispatch、反射或runtime trace**，不能取代授權／交易測試。原binary未安裝，這次只安裝mise既有pin，未改依賴或設定。
- Cycle統計為上述有向圖的非平凡SCC（大小>1），不從Graphify無向或相似性邊推論循環。Fan-in／out為不同來源邊數，不是呼叫次數或效能。

相同統計腳本、相同排除與module定義用於前後來源。工作區證據：`/workspace/tooling/tmp/architecture-{before,after}-metrics.json`、`architecture-{before,after}-go-packages.json`、`architecture-{before,after}-static-calls.txt`及 `architecture-metrics.py`。兩輪腳本SHA相同：`59f02de3c39dd1d1152f8fe799348b9a128475285e1b54c688fc78a6a1961cd4`；各輸入hash記錄於metrics JSON。

## 基準／修改後矩陣

「修改後」指固定程式與新測試的快照，不把本報告自身新增的文字混為程式改善。

| 指標 | 基準 | 修改後 | 結論 |
| --- | ---: | ---: | --- |
| Graphify nodes／edges | 4,765／12,185 | 4,792／12,342 | 增加含新helper／測試，不宣稱圖越小越好 |
| Graphify communities | 276 | 286 | CLI當次觀測，不能據此判定模組品質 |
| Go直接import packages／edges | 65／172 | 65／172 | package拓撲不變 |
| Go跨module／跨business import edges | 96／9 | 96／9 | 沒有把正常port或組裝列為違規 |
| Go有向import SCC cycles | 0 | 0 | 沒有循環需要修，未製造循環 |
| Go直接靜態call nodes／edges | 644／1,143 | 646／1,144 | 共用helper增加正常呼叫，非runtime量測 |
| Go跨module／跨business direct-call edges | 228／26 | 230／26 | 與import數分開；不得混報為跨業務依賴減少 |
| Go直接靜態call SCC cycles | 0 | 0 | 不含動態dispatch，沒有無循環的全程式保證 |
| TS靜態import files／edges | 53／151 | 53／151 | 前端未修改 |
| TS跨module／跨business edges | 132／5 | 132／5 | 同排除定義 |
| TS有向import SCC cycles | 0 | 0 | 不推論為所有runtime路徑無cycle |
| 同一group協定decoder實作位置 | 2 | 1 | 已人工核對同一規則；不是兩個業務合成一個 |
| 已確認應合併的重複業務規則群 | 0 | 0 | 本輪未找到，不把呈現檢查或不同業務角色灌入此數 |
| 精確本批手寫production實體行數 | 350 | 348 | −2；測試另增351行，沒有以刪測試換減量 |
| 全量Repomix files／tokens | 518／841,743 | 523／846,629 | 新測試提高閱讀量，與production架構指標分開 |

Go最高package fan-in維持identity12、repository11、transport/http11；最高fan-out為server46、attendance/http10、attendance/firestore9。前者是共用身份／傳輸，server是明確DI組裝根；沒有額外行為證據，不能只憑degree把它們判成應拆分的God Module。Direct-call最高fan-in為WriteAPIError40、WriteJSON40、repository.ValidID28；最高fan-out為server.init46、server.Run31、Attendance.RecordsPage12，亦不是熱路徑效能證據。

跨business直接imports的9條來源關係已核對：Issue／Discussion核心及資料adapter依既有Repository／Labels契約；Project核心只依Space。Project沒有重新依賴Repository作擁有容器，前端Items只讀來源事項；migration marker及legacy adapter保留為已記錄過渡責任，不機械式刪除。

### 功能修改涉及模組：代理值與實際證據

固定代理值為「由business根核心沿有向imports可達的module集合」。前後均為：Repository、Space各1；Project為Project＋Space共2；Issue、Discussion各為自身＋Repository共2；Attendance、Summary各1。這個指標**完全未下降**，只描述靜態核心依賴，不預測一次功能開發會改多少模組。

本批實際變更涉及Repository observer、Space observer、LINE transport與server組裝4個module，不能把這4個交付scope藏起來。協定定義位置確實由兩個Domain adapter檔縮成一個platform decoder；這只是特定協定的修改位置代理，不是已觀察到未來功能可由2個模組降到1個模組。未量測未發生的功能交付成本。

## 排序、根因與實際變更

| 排序 | 判定與根因 | 本批處理 |
| --- | --- | --- |
| P1 | Membership只有平台HTTP核驗責任，卻放在Repository adapter並引用Repository validator／error；Space組裝因此要引用另一業務的技術實作 | 移至既有transport/line，保留使用端Membership port；不另建通用授權框架 |
| P1 | Repository與Space各72行observer，在替換Domain前綴後完全相同；同一事件型別、mode／group／時間／ID／users協定解析有兩份 | 共用DecodeGroupEvent；兩observer各自保留Store、名稱查詢、領域錯誤、生命週期及不消耗事件語義，不建立泛型Observer／event bus |
| P2 | 獨立比對時確認舊Membership對nil receiver／nil Client會panic，屬配置失敗沒有安全回傳 | 同一guard增加nil檢查，回中立Unavailable；明列為fail-closed配置防護，不假稱完全純搬移 |
| 延後 | 五處HTTP Register約287行相似，但牽涉CSRF／認證順序、Project items例外及各能力訊息 | 本批不抽取；需要專門整合驗收，不能為DRY一次移走安全邊界 |
| 不修改 | 前端mine／disabled／狀態標籤為呈現及防禦，後端仍執行資格、作者、版本與狀態規則 | 沒有確認越權或業務規則搬錯層；不引入大型共用控制器。A→B來源race候選違反目前單group唯一Repository模型，已撤回 |

實際production檔案：[group.go](../../backend/internal/transport/line/group.go)、[membership.go](../../backend/internal/transport/line/membership.go)、[Repository observer](../../backend/internal/business/repository/line/observer.go)、[Space observer](../../backend/internal/business/space/line/observer.go)、[Repository組裝](../../backend/internal/server/repository.go)、[Space組裝](../../backend/internal/server/space.go)。舊Repository-owned membership.go已移除，沒有留下純轉呼叫alias。

Membership的HTTP邏輯逐段對照基準：原本就要求200且profile.userId等於非空請求者，404＝deny且無error，403／429／無效profile／body／網路失敗＝Unavailable；本輪沒有新增接受條件。內部錯誤由Repository.ErrUnavailable改為平台ErrMembershipUnavailable，各Domain HTTP原本對此類失敗的503及訊息保持；token、1秒deadline與已配置的redirect拒絕策略不變。Group join／leave可合法沒有user；它不是空使用者通過membership核驗。

## 實際驗證與查證

- 新增4份維護測試共351行：platform decoder／membership、Repository observer、Space observer。五類事件、無user join／leave、unsupported／inactive／invalid group／時間／ID、member順序、名稱查詢、Store錯誤與事件不消耗均涵蓋。HTTP測試涵蓋matching／mismatched／empty profile、404／403／429、JSON／body大小、nil／空設定、cancel及redirect。
- Focused tests先通過，再執行Go格式、lint、module verify、vet、完整既有race suite、build與vuln；API能力／生成同步／相容檢查及設定檢查通過。無應用可達漏洞，仍有3筆未被應用呼叫的required-module項目。前端／契約／資料模型／migration未修改，未重跑不受影響的前端build或正式遷移。
- Context7實際resolve及query `/websites/pkg_go_dev_net_http`：[Client.Do](https://pkg.go.dev/net/http#Client.Do)區分非2xx與transport error、關閉Body；[ErrUseLastResponse](https://pkg.go.dev/net/http#ErrUseLastResponse)核對拒絕redirect。另對照本機Go1.27.4及現有SDK types。
- LINE官方來源預設main已核對HEAD `e39b8daf919dc7a9396e44b5fc55a2aad918d932`，查閱group-chats、receiving-messages與其Messaging API reference。這是文件／本機測試證據，沒有真實LINE群組或手機驗收。

## 原始證據與重現

基準完整audit：`/workspace/tooling/tmp/analysis-audit-knj9p32s/`；固定程式修改後audit：`/workspace/tooling/tmp/analysis-audit-okhecox3/`。兩者report.json保存逐檔來源hash、排除理由、版本及artifact hashes；missing／extra、duplicate IDs、dangling endpoints均為0。基準bundle SHA：`b423956407b0dbc65b9b36389a0892479429e63c8c0cdde83d37439241c5c04e`；修改後bundle SHA：`eb67844b9f75a770150870205d1658de8394881c662b053d9cbe7c873b68ecef`。

```bash
bash scripts/tooling/tool-exec.sh python3 scripts/analysis/analysis-scope.py --audit
bash scripts/tooling/tool-exec.sh go -C backend list -json ./... > /tmp/packages.json
bash scripts/tooling/tool-exec.sh bash -c 'cd backend && callgraph -algo=static ./...' > /tmp/calls.txt
python3 /workspace/tooling/tmp/architecture-metrics.py \
  --root /absolute/audit/source --packages /tmp/packages.json \
  --graph /absolute/audit/graphify-out/graph.json --calls /tmp/calls.txt --out /tmp/metrics.json
```

Go packages／calls必須由同一固定來源取得，並保留script／inputs hashes；基準是在任何應用修改前執行。報告／索引加入後另做最終full audit核對交付來源，原始固定程式metrics不因報告自身token變化改寫。本機暫存證據不是永久下載服務，不提交巨大原碼包。

剩餘限制：這是全來源結構定位後的一次有界收斂，不是全庫無缺陷或安全性證明。動態dispatch／跨服務授權與正式資料行為仍需對應執行證據；沒有正式LINE、雲端migration、資料刪除或部署。既有Space／Project的source權限與保守rollback界線保持，後續依 [Project規格](../features/projects.md)處理，不藉架構清理移除過渡碼。

[工具索引](README.md) · [文件索引](../README.md)
