# 發布流程

提交、推送、CI 與正式部署是四個獨立階段，每一階段都需要對應授權與證據。

## 提交

- 先完成適用驗證並審閱 staged diff。
- 一筆提交應完整、可獨立審查與回退，包含相關契約、程式、設定、測試與文件。
- 提交訊息描述具體結果；未完成或未驗證事項不要包裝成完成。

## 推送

推送前核對分支、remote、待推送提交與遠端是否前進。遠端有新提交時，整合後重新執行受影響檢查。只有使用者明確授權時才推送；不得重寫已發布歷史，除非取得單次明確授權與回復方案。

## 從 work 合併到 main

1. 收到合併指示後，用 GitHub MCP 確認 PR 的來源為 `work`、目標為 `main`，並確認最新 head SHA 的必要檢查已通過。
2. 使用 **Create a merge commit**。呼叫 `merge_pull_request` 時指定 `merge_method: "merge"`，並將 `expected_head_sha` 設為已驗證的提交。若 head 已更新，重新確認檢查結果。這樣 `main` 會保留 `work` 原有的提交與 SHA。
3. 確認 PR 已合併後，從儲存庫執行下列命令同步長期使用的 `work`。先確認遠端 `work` 可以快轉至 `main`，再以一般 push 更新遠端分支。

```bash
git fetch origin &&
  git merge-base --is-ancestor origin/work origin/main &&
  git push origin refs/remotes/origin/main:refs/heads/work
```

4. 再次執行 `git fetch origin` 與 `git rev-list --left-right --count origin/work...origin/main`。兩個計數皆為零表示已對齊。若祖先檢查失敗、推送被拒或遠端又有新提交，保留分支歷史，重新取得遠端狀態並回報差異。不得用 force push 覆蓋他人的提交。

## 候選部署

1. 取得目標 project、region、service 與目前 traffic。
2. 建立不可變映像與候選 revision。
3. 核對 runtime env、Secret references、service account 與資料相容。
4. 在不影響正式流量的條件下做健康、API 與核心流程 smoke。
5. 確認回退 revision。

## 切換與觀察

經發布授權後切換流量，觀察錯誤率、延遲、授權拒絕、工作佇列與 LINE 核心流程。達停止條件時立即回復既有流量配置。部署後才做的資料清理必須等回退窗口結束並另有授權。
