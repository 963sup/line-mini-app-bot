# 發布流程

提交、推送、CI 與正式部署是四個獨立階段，每一階段都需要對應授權與證據。

## 提交

- 先完成適用驗證並審閱 staged diff。
- 一筆提交應完整、可獨立審查與回退，包含相關契約、程式、設定、測試與文件。
- 提交訊息描述具體結果；未完成或未驗證事項不要包裝成完成。

## 推送

推送前核對分支、remote、待推送提交與遠端是否前進。遠端有新提交時，整合後重新執行受影響檢查。只有使用者明確授權時才推送；不得重寫已發布歷史，除非取得單次明確授權與回復方案。

## 從 work 合併到 main

使用者指定「走 GitHub MCP 推送流程」即啟動從推送到分支同步的完整流程。若使用者指定「只推到 work」或「只開 PR」，以指定階段為終點。正式部署另依部署授權處理。

1. 推送後，用 GitHub MCP 查詢來源為 `work`、目標為 `main` 的開啟中 PR。有則沿用，沒有就建立。核對變更檔案與預期範圍，持續追蹤最新 head SHA 的必要檢查直到通過。若 CI 失敗，依[驗證流程](verification.md)修正並追蹤新一輪檢查。
2. 使用 **Create a merge commit**。呼叫 `merge_pull_request` 時指定 `merge_method: "merge"`，並將 `expected_head_sha` 設為已驗證的提交。若 head 已更新，重新確認檢查結果。這樣 `main` 會保留 `work` 原有的提交與 SHA。
3. 確認 PR 已合併後，從儲存庫執行下列命令同步長期使用的 `work`。先確認遠端 `work` 可以快轉至 `main`，再以一般 push 更新遠端分支。

```bash
git fetch origin refs/heads/work:refs/remotes/origin/work refs/heads/main:refs/remotes/origin/main &&
  git merge-base --is-ancestor origin/work origin/main &&
  git push origin refs/remotes/origin/main:refs/heads/work
```

4. 再次以上述明確的兩個 refspec 抓取分支 與 `git rev-list --left-right --count origin/work...origin/main`。兩個計數皆為零表示已對齊。若祖先檢查失敗、推送被拒或遠端又有新提交，保留分支歷史，重新取得遠端狀態並回報差異。不得用 force push 覆蓋他人的提交。

完整流程的完成回報須包含 PR 連結、該 PR 最新提交的 CI 結果、合併提交 SHA 與遠端兩支的 ahead／behind 計數。CI 執行中須繼續追蹤；遇到權限或分支保護阻擋時，回報阻擋原因與尚未完成的階段。

## 候選部署

1. 取得目標 project、region、service 與目前 traffic。
2. 建立不可變映像與候選 revision。
3. 核對 runtime env、Secret references、service account 與資料相容。
4. 在不影響正式流量的條件下做健康、API 與核心流程 smoke。
5. 確認回退 revision。

## 切換與觀察

經發布授權後切換流量，觀察錯誤率、延遲、授權拒絕、工作佇列與 LINE 核心流程。達停止條件時立即回復既有流量配置。部署後才做的資料清理必須等回退窗口結束並另有授權。
