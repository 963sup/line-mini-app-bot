# 發布流程

提交、推送、CI 與正式部署是四個獨立階段，每一階段都需要對應授權與證據。

## 開發與驗證

- `main` 是唯一長期整合分支。每次從最新的 `origin/main` 開 `task/<任務名稱>`，沿用該分支提交到任務完成。既有的 `work` 供查閱舊工作。
- 發布前核對遠端、任務分支、工作樹、暫存區與未追蹤檔。先提交本次發布範圍內的變更，再確認與最新 `main` 的差異。若有其他進行中的工作，使用隔離的 worktree 或 checkout。
- 依範圍執行[本機檢查](verification.md#選擇檢查)。純文件變更執行 `pnpm docs:check`。

## 用任務分支送 main

「走 GitHub MCP 推送流程」涵蓋下列步驟。若使用者指定只推分支或只開 PR，以指定階段為終點。

1. 抓取最新的 `origin/main`，確認任務分支的差異和檢查結果。直接推送 `task/<任務名稱>`。若要求送 PR 前只有一筆提交，可在首次推送前壓縮本機提交。已推送的任務分支以新提交修正。
2. 用 GitHub MCP 查詢 `task/<任務名稱> → main` 的 PR，沿用符合的 PR 或建立一個。若寫入顯示確認表單，等使用者處理後用 MCP 重查。只有 MCP 寫入確定失敗、重查確認尚未生效且既有授權仍適用時，才用同一個已驗證 GitHub 身分的 `gh` 補做，完成後再用 MCP 核對。
3. 用 MCP 核對 PR head 與對應的 CI run。可用 `gh run watch <run-id> --exit-status --interval 15` 持續監看。CI 失敗時依[驗證流程](verification.md#用-github-mcp-處理-pr-紅燈)排查。若 PR head 或 `main` 前進，重新核對合併差異與對應 CI。
4. 等最新版本的必要 CI 通過後，用 MCP 的 `merge_pull_request` 指定 `merge_method: "squash"`，並以 `expected_head_sha` 鎖定通過檢查的 head。用 MCP 確認 PR 已合併，再核對 `main` 的提交與發布內容。
5. 確認 GitHub 已清理來源分支。若儲存庫的合併後自動刪除設定關閉，先核對來源分支仍指向已合併的 PR head，並確認無其他開啟中的 PR 使用該分支，再依授權清理。下一輪從最新 `main` 開新任務分支。回報 PR、CI 與發布 SHA。正式部署另依部署授權處理。

## 候選部署

1. 取得目標 project、region、service 與目前 traffic。
2. 建立不可變映像與候選 revision。
3. 核對 runtime env、Secret references、service account 與資料相容。
4. 在不影響正式流量的條件下做健康、API 與核心流程 smoke。
5. 確認回退 revision。

## 切換與觀察

經發布授權後切換流量，觀察錯誤率、延遲、授權拒絕、工作佇列與 LINE 核心流程。達停止條件時立即回復既有流量配置。部署後才做的資料清理必須等回退窗口結束並另有授權。
