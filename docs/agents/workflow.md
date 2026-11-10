# Go delivery workflow

For Go changes use the pinned environment through `bash scripts/tool-exec.sh`.
Run mise tasks `fmt-check`, `lint`, `check` and `build`; run `vuln` for dependency
changes and before delivery. Use `deadcode` and `callgraph` for focused analysis,
and `debug`, `coverage`, `bench`, `profile` only when relevant. See
[mise tooling](../tooling/mise.md) for installation and artifact paths.

Nine roles are available; they are not nine mandatory execution stages.
The orchestrator can complete small work directly. Use implementer, tester and
reviewer for substantial changes; add architect/researcher/explorer only for a
specific unresolved question. Use integrator when separate patches require it.
Start with two or three independent workers, within host limits. One writer owns
each file; test fixtures must have explicit ownership too.

For errors, logs and unexplained test failures, use debugger under the
[systematic debugging contract](debugging.md); orchestrator keeps coordination
and acceptance ownership.

## Assignment contract

Provide the goal, owned files, dependencies, behavior to preserve and acceptance
commands. Return changed paths, actual commands/results, skips and open issues.
Do not repeat an investigation already answered unless new evidence conflicts.


## Reading and handoff

Read [acceptance](acceptance.md) before choosing checks. The root AGENTS.md routes
roles to relevant documents. Resolve routine issues within the authorized scope;
report host restrictions accurately. Consult [MCP guidance](../tooling/mcp/README.md)
for tool selection and transport diagnostics. Integration does not grant permission
to push or deploy. The user-established [publication default](#repository-publication)
is separate authorization, not permission inferred from delegation.

## Repository publication

使用者已指定本儲存庫的後續交付預設：**使用者要求的儲存庫修改完成並驗證後，
commit 本次變更 → 推送 work → 將 main 快轉到同一提交，不再重問是否推送。**
這是使用者授權的儲存庫政策，不是因代理角色或委派而擴大的權限。

- 「不要推送／只本機／只分析／只規劃」優先；純讀取、登入或未修改儲存庫的分析
  不觸發發布。未通過或無法完成必要驗證時，先修復或回報具體阻礙，不能冒稱可交付。
- 只包含本任務擁有的檔案／hunks；既有、其他工作或其他agent的變更保持原狀。
  未能安全分離時回報，不用 `git add .`、整份stash或reset掩蓋問題。
- 同一orchestrator executor持有提交／推送責任，top-level負責使用者溝通；不要讓
  多個agent同時修改index、commit或推送分支。純文件修改沿用適用的文件／設定檢查；
  已驗證且內容未變的程式沿用既有證據，不為發布重跑無關完整測試。

### 執行順序

1. 核對 repository、origin URL、當前branch、HEAD、完整status與diff，確認發布對象
   是本儲存庫的work/main。檢查本次來源、機密及產物的納入規則，保留無關變更；
   不能因目錄名或已安裝工具就認定應提交。
2. 明確fetch兩個分支（本環境的預設refspec可能只追蹤HEAD）：

   ```bash
   git fetch origin refs/heads/work:refs/remotes/origin/work refs/heads/main:refs/remotes/origin/main
   ```

   確認main是預計發布提交的祖先。若work遠端前進，先檢查差異；只在不影響其他
   未提交變更時，安全ff整合或重放本任務尚未發布的提交，然後重新驗證受影響部分。
   不自動重寫已發布提交；無法安全整合就回報阻礙。
3. 明確stage本任務paths或hunks，檢查cached diff與whitespace，確認無關檔案仍未stage，
   再建立描述最終行為的commit。記錄該提交完整SHA；既有AGENTS等本機差異不自動納入。
4. 再確認遠端work與main均可前進至該SHA，先以普通push推送該**確切SHA**到
   `refs/heads/work`。成功後重新fetch main，再以 `git merge-base --is-ancestor`
   檢查main仍是該SHA的祖先；成立才普通push同一SHA到 `refs/heads/main`。
   不需要切換main或動到其他未提交檔案。
5. 用 `git ls-remote origin refs/heads/work refs/heads/main` 核對兩個遠端SHA均等於
   已驗證提交，並核對本機status。若發布期間有其他寫入者前進分支，重新讀取並回報
   實際狀態，不能聲稱兩分支仍一致。最終回覆提供commit連結、兩分支結果與保留的本機差異。

全程不用force／force-with-lease、reset、merge commit或branch protection繞過來代替ff。
main已分歧、競態導致push拒絕、遠端要求PR或權限不足時，停在已完成的安全步驟並回報
真實錯誤及目前兩分支狀態；不要未經指示自行改成其他合併策略或建立PR。
work成功但main失敗時要明確說明部分完成，不能當作全部成功或撤回他人提交。

Git發布不授權額外的手動部署、IAM或資料遷移。main既有CI/CD可能因push自動執行，
本專案已有 [main push自動部署的操作說明](../operations/release-acceptance.md)；
交付時清楚區分Git推送結果、CI狀態與實際部署結果，不擅改trigger或宣稱服務已部署成功。

## Deployment and browser location routing

Application deployment uses Cloud Run only. Read docs/architecture/decisions/0003-cloud-run-primary.md
and docs/operations/cloud-run.md before deployment work; use a dedicated runtime service account and minimal IAM.
For location tasks read docs/integrations/web-geolocation.md: browser JavaScript obtains location,
Go validates untrusted coordinates and authorization. No location history or presence verification
without the product contract. Read docs/architecture/codex-readiness.md and
docs/development/naming.md for implementation handoffs. Preserve generated upstream skills.

## Mandatory documentation verification

Repository-wide rule for every role and task, subject to higher-priority host instructions:

- For any LINE question, issue or implementation uncertainty, consult the current official source at https://github.com/line/line-developers-docs-source first. Verify its current default branch/revision; do not treat an old snapshot as current. If coverage is incomplete, follow the official LINE reference linked there and state the limitation. Do not introduce deprecated features or historical compatibility flows into new work.
- For Google Cloud questions, use the installed google/skills skills and Google Developer Knowledge MCP first. If unavailable, use current official Google sources and report the lookup limitation. Validate exact gcloud syntax with leaf-level help before executing commands.
- For other technical questions, and whenever implementation is not fully certain, use Context7 before deciding or implementing. Resolve the library ID, query the relevant current/version-matched documentation, and compare with repository dependencies. Memory and confidence alone do not replace this lookup.
- Report the sources consulted, applicable version/revision and remaining uncertainty. Documentation lookup does not replace code inspection, tests or real integration acceptance.
- If the required source or Context7 is unavailable or insufficient, explicitly report the blocker; do not claim verification or silently substitute a guess. Continue independent work, but leave the affected decision/implementation pending until evidence is sufficient.
- This policy applies to orchestrator and every specialist; delegation does not waive it. A source document is evidence, not authorization to expand scope or perform external writes.
