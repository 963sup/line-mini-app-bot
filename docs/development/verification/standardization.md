# 36 項標準化驗收紀錄

本次範圍為目前產品的本地與 loopback Firestore emulator 驗收；Organization／Membership 維持既定 legacy 模式。獨立 Membership 的完整正向流程、資料切換與回復列為後續功能，不以直接修改模式欄位代替交付。

## 計分與限制

沿用聊天中的證據里程碑：權威規範、實作、適用自動檢查、完整驗收各占 25%。這不是按檔案或操作統計的實測符合率。局部案例通過不代表全庫業務與目標環境均已驗收，因此不將所有項目調成 100%。

本輪補齊 Team 寫入生命週期條件與自有 TypeScript 格式門檻。第 2、8、17、19、20、21 項由 50% 更新為 75%；第 26 項維持 50%，其餘維持 75%，36 項等權平均約 74%。

| 項目 | 百分比 | 已有證據／尚缺範圍 |
| --- | ---: | --- |
| 1 領域 | 75% | 模組責任及架構檢查；逐操作的完整語意驗收仍有限制 |
| 2 資源生命週期 | 75% | Team 綁定及指派於交易內拒絕非 ACTIVE／停用狀態；全資源生命週期尚未逐項驗收 |
| 3 命名 | 75% | 命名規範、API 能力檢查及 Go lint |
| 4 識別 | 75% | 各資源契約與格式驗證；不任意更改公開或持久化 ID |
| 5 時間 | 75% | UTC instant、業務時區及邊界測試 |
| 6 結構 | 75% | 現有模組配置及架構測試 |
| 7 依賴 | 75% | Composition 注入與依賴邊界檢查 |
| 8 格式 | 75% | Go 格式／lint；自有 TS／TSX／MTS 的固定版本 Prettier 與 CI 門檻 |
| 9 文件 | 75% | 規格同步、本機路徑及錨點檢查 |
| 10 型別 | 75% | Go／TypeScript 型別檢查及產生型別 |
| 11 資料契約 | 75% | OpenAPI、輸入驗證與儲存邊界測試 |
| 12 API | 75% | 現行契約、路由及能力歸屬檢查 |
| 13 錯誤 | 75% | 共用錯誤與 recovery metadata；保留既定內部入口格式 |
| 14 查詢 | 75% | 分頁、游標範圍及失敗重試案例 |
| 15 生成 | 75% | Go／TS 產生碼與來源一致性檢查 |
| 16 設定 | 75% | 載入、工具版本及 Secret reference 檢查；線上設定另驗收 |
| 17 安全 | 75% | Team 停用拒絕、身分與保護邊界測試；實際 IAM／Secret 狀態未驗收 |
| 18 身分 | 75% | Session／token 的拒絕、續期與撤銷案例；真實 LINE 裝置未驗收 |
| 19 授權 | 75% | 交易內權限與 Team 生命週期共同判斷；全角色組合未逐項驗收 |
| 20 狀態 | 75% | 領域狀態與 Team 非 ACTIVE 拒絕案例 |
| 21 交易 | 75% | Emulator 與 race 測試；Team 狀態由同一交易取得 |
| 22 併發 | 75% | 版本條件與 emulator／race 案例 |
| 23 冪等 | 75% | 收據、固定鍵及重試案例；外部服務實際投遞另驗收 |
| 24 UI | 75% | 共用元件、171 項前端測試及 Chromium 專案案例；LINE 裝置另驗收 |
| 25 表單 | 75% | 生成型別、欄位錯誤與草稿／原請求保留案例 |
| 26 資料遷移 | 50% | 既有格式相容處理；獨立 Membership 切換及回復不在本次範圍 |
| 27 可靠性 | 75% | 期限、取消、重試與本地故障案例；真實依賴另驗收 |
| 28 日誌 | 75% | 結構化錯誤、關聯 ID 與敏感資料排除測試 |
| 29 可觀測性 | 75% | OpenTelemetry log exports、背景 trace 傳遞及指標測試；雲端採集／告警另驗收 |
| 30 測試 | 75% | 本次 unit／contract／emulator／browser 證據；不視為全部業務路徑覆蓋 |
| 31 建置 | 75% | 固定版本、依賴鎖檔及靜態建置；獨立乾淨環境重建另驗收 |
| 32 部署 | 75% | 發布與回復流程及本地檢查；未執行線上部署或回復演練 |
| 33 版本 | 75% | 指定 Git 基準的保守 API 相容性檢查 |
| 34 開發流程 | 75% | 自我審查、CI 與驗收矩陣；不宣稱獨立審查 |
| 35 事件 | 75% | 既有事件與工作去重／排序案例；不為一致性強加 CloudEvents |
| 36 快取 | 75% | no-store、範圍隔離及舊回應清理案例；未新增無量測需求的快取 |

## 本次實際驗證

程式與格式化驗證對應 `80d2804f`（Team 生命週期修正）及 `4de9baf1`（格式化與 CI）；後續變更後，依[證據重用條件](../workflow/handoff.md#minimal-handoff-and-evidence-reuse)決定哪些結果須重跑。

- `pnpm format:check`：自有 TS／TSX／MTS 格式通過。Prettier 版本由根 package.json／鎖檔固定，範圍由該命令與 `.prettierignore` 決定；生成碼與受保護上游元件排除。
- `pnpm api:check`：型別、11 項 dispatch 測試、109 個 operation、契約及產生碼一致性通過。
- 前端型別、171 項測試及靜態建置通過。
- `go mod verify`、`go vet -p=1 ./...` 及 `FIRESTORE_EMULATOR_HOST=127.0.0.1:8080 go test -race -p=1 -count=1 ./...` 通過；未提供瀏覽器 export 的案例由下一項單獨實跑。
- 設定 `MINIAPP_BROWSER_EXPORT` 並使用固定 Playwright 1.62.0／系統 Chromium 後，`TestPersonalRepositoryAndProject` 與 `TestProjectSettingsBrowserRoundTrip` 實際通過。LINE 身分建立使用 fixture，不能作為真實 LINE 證據。
- `TestTeamWritesRequireActiveLifecycle` 於 emulator 實際驗證 ACTIVE 成功，以及 INACTIVE、停用、缺少狀態拒絕；涵蓋管理綁定、指派與移除。
- Go lint、23 項工具測試及設定／本機連結檢查通過。

執行入口與環境前提見[驗收矩陣](../tooling/testing/commands.md)。本次沒有部署、套用線上 TTL、讀取 Secret payload 或改動正式資料。
