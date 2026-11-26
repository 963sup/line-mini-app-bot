# 發布流程

提交、推送、CI 與正式部署是四個獨立階段，每一階段都需要對應授權與證據。

## 提交

- 先完成適用驗證並審閱 staged diff。
- 一筆提交應完整、可獨立審查與回退，包含相關契約、程式、設定、測試與文件。
- 提交訊息描述具體結果；未完成或未驗證事項不要包裝成完成。
- 接到發布請求時，一次核對工作樹、暫存區、未追蹤檔案及遠端 `main`、`work` 的提交差異。「全部本地」涵蓋版控檔案的變更及待納入版控的新檔。PR 同時包含遠端 `work` 已有且尚未進入 `main` 的提交。
- 依變更範圍選擇[本機檢查](verification.md#選擇檢查)，並行執行互不依賴的項目。相同內容與環境已有通過結果時可以沿用。遇到唯讀家目錄，第一次執行 Go 測試或含 Go 產生器的 API 檢查前，確認 `GOCACHE`、`GOMODCACHE` 與 `GOPATH` 下的 checksum 資料目錄可寫，並沿用可寫的既有快取。

## 推送

推送前核對分支、remote、待推送提交與遠端是否前進。遠端有新提交時，整合後重新執行受影響檢查。只有使用者明確授權時才推送；不得重寫已發布歷史，除非取得單次明確授權與回復方案。

用 GitHub MCP 發布工作樹變更時，可直接使用已驗證的 `create_tree` → `create_commit` → `update_ref` 路徑。以遠端 `work` head 的 tree 為基底，並以同一個 head 當 commit parent 和 `update_ref` 的 `expected_sha`，指定 `force: false`。更新後核對遠端 SHA 與提交檔案範圍。若 `push_files` 在建立 tree 時回傳權限 403，確認遠端 head 後改走上述 GitHub MCP 路徑。

## 從 work 壓縮後送 main

收到完整發布流程的要求時，先完成 `work` 的提交與推送，再從遠端 `main` 建立獨立發布分支。若使用者指定只推 `work` 或只開 PR，以指定階段為終點。正式部署另依部署授權處理。

1. 抓取 `origin/main` 與 `origin/work`，記下 SHA。使用乾淨的工作區，從 `origin/main` 建立 `release/work-<work SHA 短碼>`，執行 `git merge --squash origin/work`。若有衝突，先釐清並解決，再檢查暫存區。若沒有差異，結束發布流程。
2. 檢查暫存區與測試結果，提交一次。用 `git rev-list --count origin/main..HEAD` 確認分支只有一筆提交，再推送發布分支並開啟以 `main` 為目標的 PR。重跑時可沿用來源、目標與內容都相同的現有 PR。
3. 核對 PR 的 head SHA 與 CI run，等所有必要檢查通過。如果 `main` 或發布分支前進，先重新驗證合併結果。以 GitHub 的 **Squash and merge** 合併並指定已驗證的 head SHA，讓 `main` 新增一筆發布提交。
4. 重新抓取遠端分支。若 `work` 仍是本次發布的 SHA，且 `git diff --quiet origin/work origin/main` 確認兩邊內容相同，可在隔離副本從 `origin/work` 合併 `origin/main`，再一般推送到 `work`。這會保留原始工作提交並讓最新 `main` 成為下次發布的基準。如果 `work` 已前進或內容不同，保留遠端分支並回報差異，避免覆蓋後續工作。
5. 用 `git merge-base --is-ancestor origin/main origin/work` 檢查祖先關係，並用 `git diff --exit-code origin/main origin/work` 確認已同步時的內容一致。回報 PR、CI 結果、`main` 發布 SHA 與 `work` 同步結果。

## 候選部署

1. 取得目標 project、region、service 與目前 traffic。
2. 建立不可變映像與候選 revision。
3. 核對 runtime env、Secret references、service account 與資料相容。
4. 在不影響正式流量的條件下做健康、API 與核心流程 smoke。
5. 確認回退 revision。

## 切換與觀察

經發布授權後切換流量，觀察錯誤率、延遲、授權拒絕、工作佇列與 LINE 核心流程。達停止條件時立即回復既有流量配置。部署後才做的資料清理必須等回退窗口結束並另有授權。
