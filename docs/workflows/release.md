# 發布流程

提交、推送、CI 與正式部署是四個獨立階段，每一階段都需要對應授權與證據。

## 開發與驗證

- `main` 是唯一長期整合分支。每次從最新的 `origin/main` 開 `task/<任務名稱>`，沿用該分支提交到任務完成。既有的 `work` 供查閱舊工作。
- 發布前核對遠端、任務分支、工作樹、暫存區與未追蹤檔。先提交本次發布範圍內的變更，再確認與最新 `main` 的差異。若有其他進行中的工作，使用隔離的 worktree 或 checkout。
- 依範圍執行[本機檢查](verification.md#選擇檢查)。純文件變更執行 `pnpm docs:check`。

## 壓成一筆後送 main

「走 GitHub MCP 推送流程」涵蓋下列步驟。若使用者指定只推分支或只開 PR，以指定階段為終點。

1. 抓取最新的 `origin/main` 並記錄 SHA。從該提交開 `release/<任務名稱>`，執行 `git merge --squash task/<任務名稱>`。遇到衝突先解決並重跑受影響的檢查。暫存區沒有差異時結束流程。
2. 審閱暫存差異並提交一次。用 `git rev-list --count origin/main..HEAD` 確認只有一筆提交，並確認 parent 是上一項記錄的 `main` SHA。可用 GitHub MCP 的 `create_tree`、`create_commit`、`create_branch` 發布同一份已驗證內容。更新既有分支時先核對目前 head，使用 `update_ref` 的 `expected_sha` 與 `force: false`。若舊發布分支不能快轉，從最新 `main` 重建並使用新分支名稱。
3. 用 GitHub MCP 查詢 `release/<任務名稱> → main` 的 PR。有相同來源、目標與變更內容的 PR 就沿用，否則用 MCP 建立。寫入若出現確認表單，等待使用者處理後用 MCP 重查。只有 MCP 操作確定失敗、重查確認寫入尚未生效且既有授權仍適用時，才用同一個已驗證 GitHub 身分的 `gh` 補做，完成後再用 MCP 核對結果。
4. 用 MCP 核對 PR head 與對應的 CI run。可用 `gh run watch <run-id> --exit-status --interval 15` 持續監看。CI 失敗時依[驗證流程](verification.md#用-github-mcp-處理-pr-紅燈)排查。若 PR head 或 `main` 前進，重新核對發布差異與 CI。
5. 必要檢查全部通過後，用 MCP 的 `merge_pull_request` 指定 `merge_method: "squash"`，並以 `expected_head_sha` 鎖定已驗證的 head。用 MCP 確認 PR 已合併，核對 `main` 的發布 SHA 和檔案差異。
6. 發布完成即結束本輪短期分支。確認分支沒有發布後新增的內容，再依儲存庫設定自動刪除或依授權清理。下一輪直接從最新 `main` 開新分支。回報 PR、CI 結果及 `main` 的發布 SHA。正式部署另依部署授權處理。

## 候選部署

1. 取得目標 project、region、service 與目前 traffic。
2. 建立不可變映像與候選 revision。
3. 核對 runtime env、Secret references、service account 與資料相容。
4. 在不影響正式流量的條件下做健康、API 與核心流程 smoke。
5. 確認回退 revision。

## 切換與觀察

經發布授權後切換流量，觀察錯誤率、延遲、授權拒絕、工作佇列與 LINE 核心流程。達停止條件時立即回復既有流量配置。部署後才做的資料清理必須等回退窗口結束並另有授權。
