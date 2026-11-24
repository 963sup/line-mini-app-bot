# 工程執行規範

本規範適用於所有貢獻者、Agent、模組與工程工作。執行發布或部署時，另依[發布流程](../workflows/release.md)。

## 任務執行

- MUST identify the task objective, scope, constraints, dependencies, and acceptance criteria before implementation.
- MUST read all applicable AGENTS.md files and authoritative specifications before modifying files.
- MUST inspect the current implementation before making engineering decisions.
- MUST preserve existing user changes and unrelated behavior.
- MUST identify the root cause before implementing corrective changes.
- MUST complete all authorized and feasible requirements before declaring completion.
- MUST explicitly report unresolved blockers and incomplete requirements.

- MUST NOT fabricate implementation results, verification evidence, tool usage, or subagent activity.
- MUST NOT conceal errors, regressions, incomplete work, or verification failures.
- MUST NOT silently omit required tasks or acceptance criteria.
- MUST NOT declare completion while authorized and feasible requirements remain unfinished.
- MUST NOT substitute recommendations for requested and feasible implementation.
- MUST NOT simulate unavailable tools or independent reviewers.
- MUST NOT repeatedly execute failed operations without investigating their cause.

- SHOULD reason from first principles.
- SHOULD identify the underlying problem rather than its symptoms.
- SHOULD prioritize high-impact changes according to evidence, risk, and cost.
- SHOULD choose the simplest solution that satisfies all requirements.
- SHOULD prefer reversible and independently verifiable changes.
- SHOULD validate uncertain assumptions through small experiments.
- SHOULD address related issues sharing the same root cause together.

- SHOULD NOT perform repository-wide exploration for narrowly scoped tasks.
- SHOULD NOT repeatedly generate unchanged repository snapshots.
- SHOULD NOT reread unchanged evidence without a concrete verification need.
- SHOULD NOT delegate simple tasks when direct execution is clearer and faster.
- SHOULD NOT assign overlapping work that creates avoidable conflicts or duplicate effort.
- SHOULD NOT continue exploration after sufficient evidence supports the decision.
- SHOULD NOT redefine task scope or acceptance criteria to justify incomplete work.
- SHOULD NOT present partial completion or a temporary implementation as the final result.
- SHOULD NOT use TODOs, placeholders, or hardcoded responses to claim that functional requirements are satisfied.
- SHOULD NOT treat subagent conclusions as verified facts without cross-validation.
- SHOULD NOT treat generated summaries, code graphs, or cached analysis as substitutes for current source evidence.

- MAY evaluate multiple implementation alternatives.
- MAY create a proof of concept to validate uncertain assumptions.
- MAY use incremental implementation to reduce integration risk.
- MAY defer nonessential improvements that do not affect acceptance criteria.
- MAY propose alternative solutions when existing assumptions are unsupported.

## 實作原則

- MUST follow the established language, naming, directory, and dependency conventions.
- MUST reuse authoritative types, contracts, components, and configurations where applicable.
- MUST keep implementation, contracts, tests, and documentation synchronized.
- MUST preserve supported runtime and deployment compatibility.

- SHOULD apply KISS, YAGNI, DRY, and separation of concerns.
- SHOULD prefer composition over inheritance or unnecessary indirection.
- SHOULD prefer explicit dependencies and behavior over implicit mechanisms.
- SHOULD favor standard solutions over custom implementations.
- SHOULD optimize for correctness, reliability, readability, and maintainability.
- SHOULD preserve reasonable extensibility without speculative implementation.

- SHOULD NOT introduce speculative functionality.
- SHOULD NOT design for hypothetical requirements without supporting evidence.
- SHOULD NOT introduce abstractions before concrete shared requirements exist.
- SHOULD NOT introduce additional layers solely to satisfy a reference directory structure.
- SHOULD NOT replace straightforward implementations with generic frameworks without demonstrated benefit.
- SHOULD NOT introduce configuration options for behavior that does not require variability.
- SHOULD NOT optimize for theoretical scalability at the expense of current correctness and maintainability.

- SHOULD NOT duplicate equivalent functions, components, or utilities without justified differences.
- SHOULD NOT introduce interfaces with only speculative consumers.
- SHOULD NOT introduce wrapper functions that add no semantic or technical value.
- SHOULD NOT create unnecessary mapping layers between equivalent representations.
- SHOULD NOT use reflection, dynamic dispatch, or metaprogramming where explicit code is sufficient.
- SHOULD NOT retain obsolete compatibility code after its supported consumers have been retired.
- SHOULD NOT introduce dependencies for capabilities adequately provided by existing dependencies or standard libraries.

- MAY introduce helper functions for meaningful reuse or readability.
- MAY introduce interfaces for established capability boundaries.
- MAY introduce adapters for additional external providers.
- MAY introduce configuration options for legitimate environment or operational differences.
- MAY introduce dependencies when existing capabilities cannot adequately satisfy requirements.

- MAY refactor implementations when measurable maintainability or reliability benefits exist.
- MAY remove obsolete abstractions, dependencies, and compatibility mechanisms within authorized scope.
- MAY improve nonfunctional behavior when existing acceptance criteria remain satisfied.
- MAY propose future improvements without implementing them outside the authorized task scope.

## 交付與發布

- MUST distinguish completed, verified, unverified, and blocked work.
- MUST provide evidence for material implementation and verification claims.
- MUST preserve a recoverable state before authorized high-risk operations.
- MUST follow applicable publication, deployment, and recovery workflows.

- MUST NOT bypass branch protection, required approvals, or publication workflows.
- MUST NOT publish unverified artifacts as verified releases.
- MUST NOT deploy to an unauthorized environment.
- MUST NOT perform irreversible migrations or publication without required recovery safeguards.
- MUST NOT conceal deployment failures or unresolved production risks.

## 授權與範圍

- MUST obtain explicit authorization for external or destructive operations.
- MUST NOT exceed the authorized task scope.
- MUST NOT overwrite, revert, or discard unrelated user changes.
- MUST NOT perform unauthorized external writes, deletions, commits, pushes, merges, or deployments.
- MUST NOT modify protected specifications or acceptance criteria without authorization.
- MUST NOT treat tool availability as authorization.

## 工具選擇

- SHOULD use Graphify for architectural relationships and dependency exploration.
- SHOULD use Repomix for scoped repository and documentation analysis.
- SHOULD use gopls for Go symbol-aware navigation and diagnostics.
- SHOULD use Context7 for version-specific dependency documentation.
- SHOULD use official OpenAI developer documentation for OpenAI and Codex behavior.
- SHOULD use GitHub MCP for repository and pull request operations.
- SHOULD use existing Skills and repository scripts before introducing new workflows.
- SHOULD corroborate tool-derived findings against current authoritative sources.

- SHOULD NOT treat AI-generated output or model confidence as verified evidence.
- SHOULD NOT publish generated content as authoritative knowledge without traceable evidence and applicable review.
- SHOULD NOT overwrite original evidence during knowledge extraction, consolidation, or distillation.
- SHOULD NOT resolve conflicting evidence by silently discarding contradictory sources.
- SHOULD NOT reuse knowledge outside its authorized resource or tenant scope.
- SHOULD NOT use revoked, expired, or superseded knowledge as active guidance.

- MAY use Graphify, Repomix, direct source search, or language-aware tools according to the exploration question.
- MAY use additional MCP servers when they provide distinct relevant capabilities.
- MAY use local scripts for repetitive engineering operations.
- MAY use alternative verification tools when required capabilities are unavailable and acceptance requirements remain satisfied.
- MAY reuse current and applicable evidence instead of repeating equivalent investigations.

## 協作與審查

- SHOULD delegate independent complex tasks when supported.
- SHOULD execute independent tasks in parallel.
- SHOULD execute dependent tasks in dependency order.
- SHOULD assign explicit scope, deliverables, and acceptance criteria to subagents.
- SHOULD use independent review for high-risk or cross-module changes.
- SHOULD consolidate evidence before making final engineering decisions.

- MAY execute simple tasks directly.
- MAY delegate independent research, implementation, or verification tasks when supported.
- MAY use multiple agents for complex cross-module investigations.
- MAY request independent review for material architectural or security decisions.
- MAY adjust task decomposition according to discovered dependencies and risks.
