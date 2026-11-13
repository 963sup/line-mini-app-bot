# 移除空間預覽、補回歸與 MCP 優先發布

本次依使用者要求移除「空間 UI 預覽」，將 GitHub 操作規則改為 MCP 優先、CLI 補足能力缺口，並納入先前已驗證的 stdio deadline／Promise 取消修正。基準為 `4e55e5456e0e1386dfbde92990d72890ec3a2604`，另有七個既有本機變更，已先記錄 hash，未丟棄或重置。使用者本次明確授權補足關鍵正式回歸及驗證後發布至既有 work／main；不授權手動部署或資料操作。

## 逐輪更動與驗證

| 輪次 | 根因／目標與最小變更 | 實際驗證 |
| --- | --- | --- |
| 發布規則 | 遠端操作先用當前 host GitHub MCP；本機 commit 與必要原始 Git objects 上傳由 CLI 補充，保留精確 SHA | MCP get_repo／compare_commits 已成功；文件引用／whitespace 核對 |
| 移除預覽 | 刪除 577 行 Workspace 與一行 Pages 入口，移除 AppFrame／ServiceSidebar 的專用 props、兩條 Go 路由與必要 artifact；正式共用 navigation 保留 | 舊路由回歸修前兩處 HTTP 200≠404，修後通過；桌機／手機正式選單無預覽；完整前端 suite 與乾淨靜態 build 通過 |
| 關鍵回歸 | 將先前暫存的 stdio deadline 與 untilAbort 取消回歸納入既有 unittest／node:test，沒有新平台或依賴 | Python 3 tests／5 fake-server 情境、Node 6 tests 通過；隔離舊 HEAD 正確在原缺陷情境失敗 |
| 單一版本來源 | doctor 原先硬編碼 gcloud 版本，active 配置缺版本時仍誤報舊版本已安裝；改用 `mise where gcloud` | installed／active missing 的前後 metadata-only probe 通過，沒有執行登入或安裝 |
| 再分析 | 核對 HTTP SSE deadline、API dispatch 與既有領域邊界 | 已確認 SSE 殘餘問題，但嚴格期限與 response 清理需完整 transport 驗證，停止順手擴框架 |

三種回歸都驗證可觀察行為，沒有僅以程式文字或圖度數判定成功。Python fake server 有外層安全上限及就緒同步，舊版本通知／錯誤 ID 不會讓正式測試永遠卡住。Node 使用專案原有 `tsx --tsconfig ... --import ... --test` runner，確實展開測試；直接 node loader 只報 file wrapper 的結果沒有計為通過。

## 移除範圍與保留邊界

- 移除 `/miniapp/workspace-preview`、其 html 形式、頁面 artifact 要求及正式選單入口；目前靜態輸出沒有此頁。
- 保留正式 Repository／Project、定位診斷、共用 `navigation.ts`、既有退休資產相容 URL，沒有改公開 API、保存資料或遷移流程。
- 官方 primitives、use-mobile 及其他 sidebar 模板不改。`app-sidebar.tsx` 僅刪明確授權的預覽資料項與退休的 isActive 條件；保留未被正式入口使用的模板可選 API，不藉刪功能重寫外觀或互動。
- 活躍文件同步目前用途。舊分析報告保留歷史快照，只把指向已刪檔案的連結改為 code path 並註明已移除，不倒改歷史結論。

## 前後量測

同一 `metrics.py`、相同 module 與排除定義用於前後來源，定義見 [量測規範](architecture-iteration-20261113.md#量測定義)。Go／TS production 圖排除測試、生成碼與官方模板；Graphify 包含測試，不與 production 指標混用。

| 指標 | 基準 | 完成後 |
| --- | ---: | ---: |
| TS 手寫 production files／import edges | 53／151 | 51／148 |
| TS 跨 module／跨 business imports | 132／5 | 130／5 |
| TS production 實體行數 | 4,679 | 4,067（−612） |
| Go production packages／imports | 63／170 | 63／170 |
| Go 跨 module／跨 business imports | 94／9 | 94／9 |
| Go／TS 非平凡有向 SCC | 0／0 | 0／0 |
| Go／TS 模組內邊比 | 44.71%／12.58% | 44.71%／12.16% |
| Graphify nodes／edges | 4,802／12,375 | 4,812／12,325 |
| 正式前端 tests | 29 | 35（新增 6 個取消回歸） |

含授權模板資料，前端 production 淨刪 619 行；兩份新正式回歸共 163 行，分開計算。刪除示範內部邊會降低模組內邊比，不能解讀為內聚變差；圖新增測試節點也不是架構惡化。跨 business 邊未下降，沒有宣稱業務耦合已全面降低或量得 runtime 效能收益。

doctor.sh 的四個硬編碼版本出現位置移除，版本仍只由 mise.toml 決定；採用既有 active-version 解析，沒有新增設定服務。前輪 stdio 腳本淨增五行及一個標準庫 import、untilAbort 行數不變的證據沿用 [架構延續](architecture-continuation-20261113.md)。

## 最終檢查

Go fmt、lint（0 issues）、module verify、vet、完整既有 race suite及build均通過；前端 35/35 tests、型別及 Next 靜態 build 通過。乾淨獨占 build 的輸出為 13 個路由／匯出項，包含原正式頁與定位頁，沒有 workspace-preview。最初 sandbox bind EPERM 與兩個 build 共用 generated cache 的嘗試不作最終證據；清除 ignored `.next` 後，由唯一執行者用原命令完成，沒有修改 build flags 或 runner。

設定檢查、Python 正式回歸與 shell 語法／版本選取 smoke 通過。依賴與 Go 業務未變，沿用同來源仍適用的 vulnerability 證據（0 應用可達項目、3 個 required-module 未被應用呼叫項目），不為文件定稿重跑無關全庫檢查。沒有真實 LINE、手機、Firestore emulator、ADC 登入、正式資料或部署驗收。

## 停止理由與已確認殘餘

HTTP SSE doctor 的持續通知已在本機真實 HTTP fake server 重現：超過硬編碼 25 秒仍未返回，26.591 秒由外層安全 alarm 中止。stdio 修正不涵蓋此分支。僅每行比較時間不能限制阻塞 read；完整修正需同時覆蓋連線、SSE／JSON body、response 清理及整體期限，驗證成本高於本輪局部收斂，不新增不完整的 HTTP framework。證據在 `/workspace/tooling/reports/new-candidates-20261113/sse-deadline-probe.json`。

API dispatch 的受限 token reader 與 Go 編譯驗證各有用途，沒有確認另一個值得擴解析器的缺陷。既有 server 組裝根、跨業務錯誤分類及 Project legacy／marker 邊界仍有契約用途。跨頁留言持久化及平台取消分類需要額外契約決定，因此停止新增實作，不為達迭代輪數製造抽象或搬檔。

## 來源、產物與發布方式

Graphify 先定位，Repomix 核對目前來源；Context7 查證包含 Next 16.4.5 Pages／static export、Python `/python/cpython/v3.12.16` Queue／urllib、MDN Promise／AbortSignal、Mise `/jdx/mise/v2026.11.0` 的 active `where`，並與本機版本／help核對。GitHub refs 的 `force=false` fast-forward 語義經官方文件核對；MCP `expected_sha` 以當前 host schema 為準，不冒稱 REST body 欄位。

完整本機證據目錄為 `/workspace/tooling/reports/preview-removal-optimization-20261113/`；包含 baseline dirty hashes、前後 metrics、正式回歸負向證據、doctor pin probe、Graphify 更新、最終 Repomix 與逐檔內容／SHA 核對。預覽的修前／修後及最終乾淨 build logs 另在 `/workspace/tooling/reports/remove-preview-20261113/`。

發布遵循 [單一流程](../agents/workflow.md#repository-publication)：MCP 核對 repository 與分支，CLI 建立並在必要時上傳精確本機 commit，MCP 以 `force=false` 及 fresh `expected_sha` 快轉 main，再核對兩分支 SHA。若審核或保護拒絕則回報實際完成部分，不切換 transport 繞過。Git 推送與 CI／部署結果分開報告。

[工具索引](README.md) · [文件索引](../README.md)
