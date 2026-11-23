# documentation 能力指引

由目前代理按需採用；這是工作階段，不是可啟動的子代理。共同規則見[能力入口](./)。審查結果為自我審查，不宣稱獨立驗收。

## 執行要求

Apply these documentation capabilities in the current chat. Read AGENTS.md, the applicable
nested instructions, docs/standards/documentation.md and docs/development/tooling/testing/commands.md
before editing. The documentation policy is authoritative; reference it rather
than duplicating its principles in role instructions or other documents.

Own only the documentation files or diagram hunks within the authorized task and the necessary
direct references included in that ownership. Apply the policy's workflow and
autonomous-repair boundaries. You may directly write and repair those documents;
preserve existing user changes and coordinate ownership before touching any
additional file. Record required code, tooling or configuration changes and check their task scope before implementation.

For codebase questions follow the root Graphify and current-source corroboration
rules. For technical claims follow the root official-source/Context7 routing;
consult docs/development/tooling/agents/openai-docs.md for Codex questions. Reuse still-applicable
evidence, distinguish product decisions from technical facts, and leave unsupported
or conflicting claims unresolved with a concrete explanation. For business
documentation also read docs/contracts/business-operations.md and the compliance
workflow in docs/development/workflow/implementation.md#business-operation-compliance. Do not invent or change product rules.

Validate the assigned diff, local links and heading anchors, source references,
and relevant commands using checks actually available in the repository. Run
git diff --check; run the existing configuration check when configuration changes
are part of the assignment. Do not claim a Markdown linter, command, runtime or
integration passed unless it was actually exercised. Check that consolidation
preserved valid knowledge, constraints and navigation.

Report changed paths, repaired violations, sources/revisions, actual checks and
remaining uncertainties in the final delivery. Do not spawn subagents. Publishing and deployment require task authorization.
External publication is limited to explicitly authorized Wiki edits. Use the current session model, permissions and MCP servers;
these instructions do not create a separate sandbox or additional authorization.
Use scripts/tooling/tool-exec.sh for repository tooling.

For assigned Mermaid diagrams follow docs/development/tooling/shared/mermaid.md: retain editable source and distinguish current, proposed and unknown relationships. Match diagram type to the question, check authority, references and transaction boundaries against the domain contract, render locally and review semantic accuracy independently. For authorized Wiki maintenance follow docs/development/tooling/shared/github.md#wiki-publication and report the published URL and verified revision. Wiki authorization does not extend to repository pushes or deployment.
