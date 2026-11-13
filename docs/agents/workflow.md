# Agent delivery workflow

Registered roles are capabilities, not mandatory execution stages.
The orchestrator can complete small work directly. Select the relevant specialist
for substantial changes and use tester/reviewer when their checks add evidence.
Add architect/researcher/explorer for a specific unresolved question and integrator
when separate patches require it.
Delegate only when a bounded independent subproblem or context isolation justifies
the handoff. When parallel work is useful, start with at most two or three workers
within host limits. One writer owns each file, including test fixtures.

For errors, logs and unexplained test failures, use debugger under the
[systematic debugging contract](debugging.md); orchestrator keeps coordination
and acceptance ownership.

## Delegation depth and role selection

The repository's delegation policy is primary agent (depth 0) → orchestrator
(depth 1) → specialist (depth 2). Specialists return results to orchestrator and
MUST NOT spawn agents. Keep `agents.max_depth = 2` for V1 threads; the official
schema says V2 ignores this setting, so routing instructions and actual host
controls must enforce the no-spawn boundary there. More registered roles do not
require deeper nesting. `agents.max_concurrent_threads_per_session = 6` caps spawned
threads, excluding the primary; the orchestrator occupies one of those threads.
The current host's actual limits take precedence, and configuration cannot raise
them. Schedule independent work within both limits instead of launching every role.

Configuration semantics were checked against the [official subagent guide](https://learn.chatgpt.com/docs/agent-configuration/subagents)
and [Codex schema revision](https://github.com/openai/codex/blob/4413deac572eb5d4cdb49295794cbe30f8622935/codex-rs/core/config.schema.json)
on 2026-11-13. The existing depth and thread values support this topology;
adding specialist roles is not a reason to increase them.

| Task | Specialist and handoff |
| --- | --- |
| Locate symbols and relationships | `graphify` returns a scoped graph and current-source evidence; it remains read-only unless graph output updates are explicitly assigned |
| Cross-file source corroboration | `repomix` returns a scoped bundle, included/excluded paths and source checks; no source edits or skill generation |
| Version-matched library documentation | `context7` returns library IDs, versions, sources and applicability limits; root official-source routes take priority |
| Go implementation | `golang` owns assigned Go files and uses gopls evidence plus applicable Go checks |
| TypeScript implementation | `typescript` owns assigned language-level changes and verifies version-matched APIs |
| Next.js application changes | `nextjs` owns assigned framework changes and checks the installed version and router conventions |
| UI composition and registry work | `shadcn` owns assigned UI integration, preserving official components under the existing protection policy |
| Markdown maintenance | `documentation` owns assigned Markdown and necessary direct-reference updates under the documentation policy |

Use evidence specialists first when needed, then assign implementation to an
explicit owner. For a mixed Next.js/TypeScript/shadcn task, choose one primary
writer; the other specialists provide read-only advice. Parallel writers require
disjoint file ownership. Return cross-scope changes to orchestrator for reassignment.
Tool operations and fallbacks remain in the [MCP guides](../tooling/mcp/README.md).

The orchestrator or current owner MAY call the relevant tools directly; a tool
does not require its namesake specialist. Do not build a mandatory
graphify → repomix → context7 → implementer chain. Select only the steps required
by the question and root policies. The root Graphify-first and official-source
requirements, and the Business Operation compliance steps, still apply.

Custom roles inherit the parent's model, permissions and MCP servers unless a
supported setting explicitly changes behavior. A role description does not grant
tool access or create a sandbox. If the host cannot select a registered role,
pass its role instructions to an available worker or have the current owner adopt
them, and report the fallback rather than claiming native role execution.

## Go verification

For Go changes use the pinned environment through `bash scripts/tooling/tool-exec.sh`.
Run mise tasks `fmt-check`, `lint`, `check` and `build`; run `vuln` for dependency
changes and before delivery. Use `deadcode` and `callgraph` for focused analysis,
and `debug`, `coverage`, `bench`, `profile` only when relevant. See
[mise tooling](../tooling/mise.md) for installation and artifact paths.

## Assignment contract

For Markdown work, apply the [documentation policy](../development/documentation.md).
Assign documentation explicit files/hunks and necessary direct-reference updates;
it may edit that scope directly. Keep one writer per file and return code/tooling
changes to their assigned owner. Small edits may stay with the current owner.

Provide the goal, owned files, dependencies, behavior to preserve and acceptance
commands. Return changed paths, actual commands/results, skips and open issues.
Do not repeat an investigation already answered unless new evidence conflicts.

### Minimal handoff and evidence reuse

Use six fields rather than forwarding the full conversation or raw logs:

| Field | Required content |
| --- | --- |
| Goal | One bounded question or change and its completion condition |
| Ownership | Repository, revision, relevant dirty paths, owned files/hunks and read-only boundaries |
| Inputs | Located paths/symbols, applicable instructions and constraints; read required instructions if not already supplied and current |
| Evidence | Only relevant source/query results with version, scope and artifact locations; include host/server/tool identity for MCP evidence |
| Checks | Exact commands already run, actual results and required checks still pending |
| Open issues | Missing evidence, conflicts, failures and cross-scope work needing reassignment |

Reuse evidence only while its repository/revision, relevant dirty paths, query or
bundle coverage, dependency versions and tool configuration remain applicable.
An unchanged HEAD alone is insufficient when relevant working files changed.
Refresh only invalidated evidence after source/version/scope/configuration changes
or conflicting observations; do not repeat unchanged document reads, library
lookups, bundles or whole check suites just because the owner changed.

Every agent that actually explores the codebase MUST still perform its initial
Graphify query under the root rule. Supply the previous focused query and located
paths to avoid repeating broad exploration; reuse that agent's still-valid query
for the same question. A worker consuming supplied evidence without new codebase
exploration need not invent an additional exploration task. New questions or stale
graphs require focused queries and current-source corroboration.

For Context7 reuse a prior resolved library ID and versioned answer only for the
same covered question; new or unsupported questions still require resolve/query
and the root official-source route. Reuse a scoped Repomix bundle only while its
source revision, dirty paths and inclusion/exclusion coverage match. Reviewers
retain independent judgment and perform checks needed to challenge a conclusion.

The [MCP operating guide](../tooling/mcp/README.md#task-scoped-tool-use) owns tool
inventory and connection reuse. A service, role or source mismatch invalidates
the relevant evidence; successful registration is never runtime acceptance.

Cache a diagnosed tool failure with its host/environment, server configuration
and failure stage. Use the documented fallback for the rest of that unchanged
task; another role is not a reason to repeat the same timeout. Retry only after
a relevant condition changes or a bounded diagnostic is explicitly assigned.


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

All roles MUST follow the [root source-verification policy](../../AGENTS.md#mandatory-documentation-verification)
and the [Markdown policy](../development/documentation.md). Reuse applicable
evidence; report unavailable sources and leave affected decisions pending.
Documentation evidence does not replace implementation checks or authorize
additional writes.

## Business Operation compliance

業務開發、審查及修正必須先讀 [操作模型](../architecture/business-operation.md)。任務分派附上操作契約、規格來源、既有實作、檔案所有權與驗收條件。

1. **定位**：先執行 Graphify query，必要時 path／explain；依符號定位目前程式。圖譜可能過期或遺漏動態關係，不以無節點判定無實作。
2. **核對**：執行 Repomix CLI，核對打包範圍與對應原文，確認根／巢狀指引、契約、規則和入口。Markdown／TOML 不保證被 AST 收錄，遺漏者直接讀取並註明。產物使用任務專屬且不提交的目錄，不覆蓋他人分析。
3. **查證**：涉及技術判斷或修改時，Context7 先 resolve library ID 再查版本相符文件；LINE、Google Cloud、Codex 問題仍遵循根指引的官方來源路由。重用本任務已查證且仍適用的證據，不重查未變內容。產品決策取自使用者及已接受規格，不從套件文件推導。
4. **判定**：記錄違反哪條規格、檔案／符號、可觀察影響、最小修正與驗證方式。區分確認缺陷、工具盲點與待確認需求。
5. **修正**：explorer／reviewer 保持唯讀，將發現交 orchestrator；範圍內由 implementer 修正，必要時 debugger 定位。保留既有變更；範圍外問題列為後續工作，不自動全庫改寫。主機不支援角色時由主代理承擔責任並說明，不聲稱有委派。
6. **驗收**：依 [acceptance](acceptance.md) 檢查契約、實作與證據。程式變更後依根指引更新 Graphify，受影響內容再打包；已驗證且未變內容沿用證據。必需來源不可用時保留受影響決策待確認，繼續可獨立完成部分。

交付列出：發現 → 規格依據 → 修改位置 → 檢查結果 → 剩餘限制。不把工具成功等同業務正確，不無限重跑全庫檢查。此流程不增加 push、部署或資料遷移授權。

沿用已註冊角色即可，不新增 compliance agent。協調者可直接完成小任務；角色是責任分工，不是必須逐一啟動的階段。
