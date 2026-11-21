# 開始任務與分工

開始修改、審查或接手前，先確認目標、工作樹與本次允許的操作。

1. 從儲存庫根目錄確認目前分支、HEAD、相關未暫存與未追蹤檔案。接手時比對交接版本。
2. 用一句話寫出要改的行為與完成條件。確認相關目錄、已有變更、局部指引與操作授權。
3. 選擇[操作配方](../recipes/README.md)與[驗收命令](../verification/commands.md)。跨模組作業再定義檔案所有權與交付證據。

Registered roles are capabilities, not mandatory execution stages.
The primary agent completes single-owner work directly. Select the relevant specialist
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

When coordination is needed, the delegation topology is primary agent (depth 0) → orchestrator
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
| Architecture, relationship, operation or state diagrams | `mermaid` owns assigned diagram hunks and checks source semantics and rendering separately; follow the [Mermaid workflow](../../../.codex/README.md#mermaid) |

Use evidence specialists first when needed, then assign implementation to an
explicit owner. For a mixed Next.js/TypeScript/shadcn task, choose one primary
writer; the other specialists provide read-only advice. Parallel writers require
disjoint file ownership. A single writer owns the requested behavior across all necessary layers. Coordinate reassignment when another writer owns a required file or the requested behavior changes.
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

## Assignment contract

For Markdown work, apply the [documentation policy](../../standards/documentation.md).
Assign documentation explicit files/hunks and necessary direct-reference updates;
it may edit that scope directly. Keep one writer per file and return code/tooling
changes to their assigned owner. Small edits may stay with the current owner.

Provide the goal, owned files, dependencies, behavior to preserve and acceptance
commands. Return changed paths, actual commands/results, skips and open issues.
Do not repeat an investigation already answered unless new evidence conflicts.

[六欄交接與證據重用](handoff.md) · [開發流程索引](README.md)
