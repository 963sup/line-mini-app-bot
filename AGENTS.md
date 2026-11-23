# Repository agent guidance

<a id="從這裡開始"></a>

## Start here

Follow rule 18 to start a task. See the [nested directory guide](README.md#巢狀目錄導覽) for directory guidance.

- Codex MUST run with the repository root as its working directory. In this workspace, launch `codex -C /workspace/line-mini-app-bot` to use the project's `.codex/config.toml`. Verify the checkout with `pwd` and `git rev-parse --show-toplevel`. If the background daemon cannot start, retry with `codex --no-daemon -C /workspace/line-mini-app-bot`. Both modes require a writable Codex state directory.
- This project does not use Codex hooks. Use Repomix, Graphify or source documents as needed for the task.

<a id="共通工程規則"></a>

## Common engineering rules

1. **Single Source of Truth**: Assign one authoritative source to each business rule, data model, API contract, configuration and document. Keep other locations synchronized through references or generation.
2. **Backend → API → Frontend**: The backend executes business rules, data access, authorization and state transitions, and exposes results through APIs. The frontend handles presentation, interaction, immediate input feedback and necessary client state. Centralize business calculations, validation and data consistency checks in the backend; APIs provide the results each use case needs. Compute display values derivable from existing data during rendering rather than adding state or Effects.
3. **Contract First**: Define schemas, serialization, requests, responses and error formats in OpenAPI before implementation. Prefer generated types and clients over a separately maintained contract. Validate external data at API boundaries and use contract types internally; distinguish missing values, null and zero values. Follow the contract's HTTP statuses and stable error codes; clients handle errors by code. Specify pagination, filtering, stable ordering and returned fields for list endpoints as needed.
4. **Standard First**: Prefer HTTP semantics, REST conventions, idiomatic Go and the existing architecture. Reuse established standards and tools; avoid custom protocols, formats, frameworks and unnecessary dependencies.
5. **Explicit Boundaries**: Use DDD bounded contexts to define domain models, responsibilities, data ownership, module interfaces and dependency direction. Cross-domain operations must use explicit interfaces rather than accessing internals or creating dependency cycles. Use the domain's ubiquitous language in names. Supply external dependencies at composition entry points so domain rules can be verified independently.
6. **Minimal Abstraction**: Follow KISS, YAGNI and DRY. Prefer direct, explicit implementations; extract abstractions only after confirming a real shared need. Before extraction, assess correctness, reliability, readability, cognitive load and maintenance cost.
7. **Consistent Conventions**: Use the same naming, directories, formats, types, identifiers, time representations, serialization, validation and error conventions for the same semantics. Follow existing conventions for queries, pagination, sorting, filtering, state and resource lifecycles. Follow the [naming standards](docs/standards/naming.md) for names and package layout. Keep public resource IDs stable. Specify timestamp formats and time zones in contracts, such as RFC 3339 UTC; determine business dates in the domain's time zone.
8. **Explicit Business Behavior**: The responsible domain centrally defines and executes business operations, permission decisions, state machines, valid transitions and data validation. Specify invariants, preconditions, side effects and error semantics. For supported operations such as creation, updates, archival and deletion, define allowed initial and final states and responsibility for cleaning up related data.
9. **Reliable Operations**: Define transaction and consistency requirements for each operation. Handle concurrency, idempotency, timeouts, cancellation, deduplication and recovery. Define backoff and stopping conditions for retryable operations to avoid duplicate side effects. Protect invariants with transactions or version conditions when writes compete. Reuse the same operation identifier when retrying operations with side effects. When external dependencies fail persistently, assess whether their impact warrants circuit breaking.
10. **Security by Default**: Apply secure defaults and least privilege at trust boundaries. Handle authentication, authorization, sessions, tokens, input validation, sensitive data and secrets consistently; the backend makes security decisions. Validate sessions or tokens against the [identity contract](docs/contracts/identity.md), then authorize operations in the backend based on the operation, resource and membership.
11. **Event & Integration Consistency**: Define synchronous and asynchronous integration boundaries. Standardize event names, schemas, versions, publication, consumption and processing semantics. Handle ordering, duplicate delivery, retries and compatibility according to delivery guarantees. Before adopting CloudEvents for cross-system events, verify the receiver's supported formats and versions.
12. **Observable Operations**: Standardize structured logs, traces, metrics, correlation IDs and error classification. Prefer OpenTelemetry so key operations, performance and failures can be measured and cross-service requests traced and diagnosed. Correlate logs and traces for the same operation through identifiers. Use bounded categories for metric labels rather than individual resource IDs.
13. **Compatibility & Lifecycle**: Standardize versioning, API compatibility, data migration, configuration management and resource creation and deletion. Prefer backward-compatible changes and gradual migrations to protect existing users, data and integrations. For data changes spanning deployed versions, first expand compatible structures, migrate data and confirm old readers and writers have been retired before removing old structures. Describe changes in external releases according to their compatibility policy.
14. **Automated Enforcement**: Automate mechanical engineering checks through formatting, linting, type checks, schema validation, code generation, dependency checks, tests and CI. Regenerate after contract changes and inspect the diff so committed artifacts can be rebuilt from their sources.
15. **Minimal Change & Continuous Validation**: Analyze dependencies, root causes and impact before applying the smallest necessary fix. Validate types, contracts, test results, security, compatibility and core behavior according to each change's impact.
16. **Continuous Simplification**: Continuously remove or consolidate duplicate, conflicting, outdated, invalid and unnecessary code, abstractions, dependencies, configuration, documentation and processes. Measure before and after improvements and verify behavior with regression checks.
17. **Task Scope & Acceptance**: Clarify the actual goal, constraints and acceptance criteria from the user's description. Use first principles to choose the smallest reliable solution and suitable tools. Prioritize by impact and risk; investigate root causes when acceptance fails. Finish when criteria are met or a blocker cannot be resolved independently. Report evidence and follow-up work for out-of-scope findings separately.
18. **Applicable Instructions**: Before starting each task, MUST read the [environment and capability entry point](README.md#依任務選擇能力), select applicable capabilities and read their full guidance before substantive work. Do not read only capability names or skip guidance from memory. Within the same task, reuse guidance already read if it is unchanged; reselect when scope changes. Before editing, read every applicable AGENTS.md along the target file's directory path. Within host permissions and the user's task scope, the root AGENTS.md defines common rules; nearer AGENTS.md files take precedence for local matters.
19. **Single-Agent Execution**: The current agent handles planning, exploration, design, implementation, self-review, testing and documentation without starting subagents or delegating to an orchestrator. Preserve existing changes and upstream skills; identify self-review accurately. Tools and capabilities do not grant additional authorization for external writes.
20. **Source Verification**: Verify local behavior against the current checkout's source and actual execution; verify external behavior against official documentation for the applicable version. Prioritize assumptions and failure risks that affect decisions. Reuse verified evidence while its version, source and applicability remain valid. Documentation research does not replace behavioral acceptance.
21. **Targeted Exploration**: Start exploration with a concrete question. Prefer available Graphify tools to locate unfamiliar dependencies; search source directly for known identifiers. Check graph and Repomix findings against the current working tree. Stop expanding searches once the evidence answers the question.
22. **Pinned Toolchain**: Prefer existing tools, use [mise.toml](mise.toml) as the version source and run tools through project wrappers or existing installation tasks. If a tool is missing, check its declared version and installation state; do not independently upgrade tools, change lockfiles or install undeclared packages. Build with repository tooling and dependency locks so a clean checkout can be rebuilt from the same inputs.
23. **Documentation Ownership**: README files own directory purposes and navigation. Authoritative topic documents own detailed specifications. Nested AGENTS.md files exist only for mandatory constraints that their parent cannot cover; they MUST NOT repeat, paraphrase or copy parent rules. Merge or remove instructions without independent value. Preserve safety, compatibility, business and architectural constraints in the nearest module rule or authoritative topic document. When behavior changes, check entry points, links and commands. API documentation references OpenAPI. Record reasons and tradeoffs for architectural decisions affecting multiple modules or difficult to reverse in existing decision documents or ADRs.
24. **Regression Verification**: Rerun the original case after a fix. For new or moved files, verify behavior with the [acceptance matrix](docs/development/tooling/testing/commands.md#變更驗收矩陣). For cross-directory moves, check the source, destination and callers. For Firestore transactions, concurrency, retries or persisted-data compatibility, proactively check, prepare and rerun emulator tests using the [integration acceptance workflow](docs/development/tooling/testing/commands.md#firestore-integration). Choose verification by risk: unit tests for domain rules, integration or contract tests for storage and service boundaries, and end-to-end tests for critical user flows.
25. **Failure Isolation**: If tools are unavailable, graphs are stale or packages omit content, verify against available sources and record limitations. Defer only decisions dependent on missing evidence; continue independent work.
26. **Verified Delivery**: At delivery, compare results with specifications and acceptance criteria. Report implementation locations, capabilities used, actual checks, results and applicable versions. Distinguish completed changes, verified behavior and remaining limitations.
27. **Verified Publication**: Develop on `main` by default. Each commit must be a complete, verified and independently reviewable milestone. Push only within the authorized scope and follow publication workflows and branch protection. Review staged changes before committing, confirm acceptance criteria and describe concrete changes in the commit message.
28. **Explicit Handoff**: When the user requests a handoff, takeover or transfer to another sidebar chat, directly use the [sidebar handoff workflow](docs/development/workflow/handoff.md#sidebar-handoff) and provide its six-field summary and the sources needed to take over.
29. **Configuration Boundaries**: Define configuration sources, precedence, required values and safe defaults. Parse and validate types at the loading boundary. Clearly identify missing required configuration by name; never include secret values in diagnostics.
30. **UI & Form Consistency**: Reuse components and interaction conventions from the [Mini App UI standards](docs/standards/miniapp-ui.md). Map form fields to API contracts, distinguishing unfilled, cleared and unchanged values. Present immediate validation and backend field errors consistently; preserve correctable input after submission failures.
31. **Cache Consistency**: Before introducing caching, define acceptable staleness, resource and authorization scope of keys, expiration and invalidation conditions. Verify reads after writes or permission changes, and behavior on cache misses.
32. **Deployment & Recovery**: Before deployment, verify authorization, target environment and validated artifacts. After deployment, check health and core flows using [release acceptance](docs/operations/release-acceptance.md). Confirm the rollback version and data compatibility in advance; use acceptance results to decide whether to continue or recover.

## Backend → API → Frontend

- Prefer Next.js, React and existing components. Choose components and data-fetching approaches for the router and deployment model. The current Mini App uses the Pages Router and static export; verify router support before considering Server Components or Client Components.
- When responsibilities are misplaced, first check callers, dependencies and impact, then move, merge or remove duplicate implementations in Backend → API → Frontend order. Synchronize affected contracts and tests. Preserve functionality, API compatibility, security and performance. Report concrete risks first if the change's safety cannot be established.

<a id="巢狀文件維護標準"></a>

## Nested instruction standard

- Root AGENTS.md owns global execution rules. Top-level module AGENTS.md files contain only mandatory module constraints.
- Nested AGENTS.md files are exceptional. Keep one only when a narrower path has a safety, compatibility, business or generated-source restriction that cannot be expressed by its module parent or authoritative specification.
- Every rule uses `MUST`, `MUST NOT`, `SHOULD` or `MAY`, names its scope and states a verifiable condition. Do not include directory tours, file lists, start-here links or copied parent rules.
- Consolidated path-specific constraints live in [module constraints](docs/standards/module-constraints.md) and are read only for the affected path. Hexagonal backend boundaries remain in [backend module rules](backend/AGENTS.md#hexagonal-architecture) and the [runtime boundary specification](docs/architecture/runtime/backend-boundaries.md).

<a id="cli-discovery"></a>
<a id="按需安裝工具"></a>

## Install tools as needed

- In a new environment, follow the [Cloud environment setup](docs/development/environment/codex-cloud.md) and run `scripts/bootstrap.sh` to prepare mise, system Python and networking.
- Before locating or running a CLI, check installed versions from the repository root with `bash scripts/tooling/tool-exec.sh mise ls --installed`. If the wrapper cannot find mise, check `/workspace/tooling/bin/mise` and `/workspace/tooling/mise-data/installs`.
- Run an individual tool with `bash scripts/tooling/tool-exec.sh <command> [arguments]`; the wrapper prepares tools at declared versions. `mise run` may prepare the entire task environment. For compound commands, list declared tools explicitly, for example `bash scripts/tooling/tool-exec.sh --tool node --tool npm:repomix bash -c 'repomix --version'`.
- If Graphify is missing, run `mise run install-graphify`.
- Confirm separately which MCP tools the chat can call. If a service is missing, a call fails or comprehensive acceptance is needed, follow the [shared troubleshooting workflow](docs/development/tooling/agents/) and continue independent work.

Sources: mise 2026.11.3 `exec --help` and [official settings](https://mise.jdx.dev/configuration/settings.html#exec_auto_install).

## Mandatory documentation verification

- **External behavior**: Choose official sources using the [tooling and sign-in guidance](docs/AGENTS.md). For Context7, resolve the library ID first, then query for the project's dependency version. On quota limits, use the [fallback workflow](docs/development/tooling/agents/context7.md#anonymous-fallback).

<a id="單一代理與能力選用"></a>

## Single-agent execution and capability selection

Read professional guidance by task through the [capability entry point](docs/development/workflow/capabilities/); see [starting a task](docs/development/workflow/task-start.md) for the workflow. Control tool calls, context and intermediate artifacts according to acceptance needs; avoid repeated analysis and unnecessary work splitting.

When diagrams clarify architecture, data relationships, flows or states, proactively validate and render them with the project's pinned local Mermaid v13 CLI. See the [Mermaid tooling guide](docs/development/tooling/shared/mermaid.md).

## graphify

The code graph lives under `graphify-out/`. For Markdown headings and links, use the separate [docs-only graph](docs/development/tooling/shared/graphify.md#docs-only-graph).

After completing the task-start reading in rule 18, if the user specifies `/graphify`, read the installed Graphify skill or operating guide before performing graph operations.

Rules:

- When exploring dependencies under rule 21, supplement graph results with rg, language references and source inspection for callers, routes, configuration, code-generation inputs and fixtures to cover paths omitted by the graph.
- Local graph changes after incremental updates are expected and do not prevent queries.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `bash scripts/tooling/tool-exec.sh mise run graphify-update` to refresh the code graph without API calls.
- Code graphs are generated locally and ignored by Git. During read-only queries, update the current checkout's graph only if it is stale and current dependency information is needed.

For cross-file queries, use the [exploration guide](docs/development/workflow/capabilities/explorer.md) to select existing reference skills or scoped packaging. Report stale-graph limitations and corroborate findings with current sources.

## Documentation routing

Before creating, editing or reviewing any owned Markdown, MUST read and follow the [documentation standards](docs/standards/documentation.md). The current agent uses documentation capability guidance as needed, confirms file ownership and independently fixes evidence-backed issues within scope without expanding into a repository-wide rewrite.

Use the [development handbook](docs/development/) to select environment guidance, workflows, recipes and acceptance checks. See the [MCP guide](docs/development/tooling/agents/) for tool selection and diagnostics.

For implementation handoffs, follow [handoff and evidence reuse](docs/development/workflow/handoff.md#minimal-handoff-and-evidence-reuse) and the [naming standards](docs/standards/naming.md).

## Repository publication default

See the [publication workflow](docs/development/workflow/publication.md#repository-publication) for the operating steps of rule 27.

<a id="工程決策與交付要求"></a>

## Engineering decisions and delivery requirements

- Before engineering work, read and follow the [engineering principles and decisions](docs/standards/engineering.md).
- For interfaces, interactions or user flows, also read and follow the [product and experience design principles](docs/standards/design.md).
- First check domain models and key dependencies. Prioritize by impact and risk, and address issues sharing a root cause together.

<a id="共通驗證與交付"></a>

## Common verification and delivery

Documentation-only and configuration-only changes do not require unrelated application tests. Behavior changes require the corresponding checks.
