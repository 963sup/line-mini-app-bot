# Repository agent guidance

<a id="&#24478;&#36889;&#35041;&#38283;&#22987;"></a>

## Start here

Follow rule 18 to start a task. See the [nested directory guide](README.md#%E5%B0%88%E6%A1%88%E7%B5%90%E6%A7%8B) for directory guidance.

- Codex MUST run with the repository root as its working directory. In this workspace, launch `codex -C /workspace/line-mini-app-bot` to use the project's `.codex/config.toml`. Verify the checkout with `pwd` and `git rev-parse --show-toplevel`. If the background daemon cannot start, retry with `codex --no-daemon -C /workspace/line-mini-app-bot`. Both modes require a writable Codex state directory.
- On the first user message in every new conversation, use the `repomix-explorer` skill to generate a fresh Repomix snapshot of the current local working tree. Consult that snapshot when handling the request. On later turns, refresh it as needed.
- This project does not use Codex hooks. Use Graphify or source documents as needed for the task.

## Normative engineering requirements

All contributors and agents MUST read and comply with the repository-wide
[MUST](docs/standards/must.md), [MUST NOT](docs/standards/must_not.md),
[SHOULD](docs/standards/should.md), [SHOULD NOT](docs/standards/should_not.md), and
[MAY](docs/standards/may.md) standards before performing engineering work. LINE
interaction changes MUST also follow the [LINE interaction standard](docs/standards/line_interactions.md).
These keywords use the RFC 2119 and RFC 8174 meanings. A SHOULD or SHOULD NOT
may be overridden only when the exception, rationale, impact, and selected
alternative are recorded. MAY does not grant authority or override a MUST or
MUST NOT requirement. Nearer module instructions may add constraints but MUST
NOT weaken these repository-wide standards.

<a id="&#20849;&#36890;&#24037;&#31243;&#35215;&#21063;"></a>

## Common engineering rules

1. **Single Source of Truth**: Assign one authoritative source to each business rule, data model, API contract, configuration and document. Keep other locations synchronized through references or generation.
2. **Backend → API → Frontend**: The backend executes business rules, data access, authorization and state transitions, and exposes results through APIs. The frontend handles presentation, interaction, immediate input feedback and necessary client state. Centralize business calculations, validation and data consistency checks in the backend; APIs provide the results each use case needs. Compute display values derivable from existing data during rendering rather than adding state or Effects.
3. **Contract First**: Define schemas, serialization, requests, responses and error formats in OpenAPI before implementation. Prefer generated types and clients over a separately maintained contract. Validate external data at API boundaries and use contract types internally; distinguish missing values, null and zero values. Follow the contract's HTTP statuses and stable error codes; clients handle errors by code. Specify pagination, filtering, stable ordering and returned fields for list endpoints as needed.
4. **Standard First**: Prefer HTTP semantics, REST conventions, idiomatic Go and the existing architecture. Reuse established standards and tools; avoid custom protocols, formats, frameworks and unnecessary dependencies.
5. **Explicit Boundaries**: Use DDD bounded contexts to define domain models, responsibilities, data ownership, module interfaces and dependency direction. Cross-domain operations must use explicit interfaces rather than accessing internals or creating dependency cycles. Use the domain's ubiquitous language in names. Supply external dependencies at composition entry points so domain rules can be verified independently.
6. **Minimal Abstraction**: Follow KISS, YAGNI and DRY. Prefer direct, explicit implementations; extract abstractions only after confirming a real shared need. Before extraction, assess correctness, reliability, readability, cognitive load and maintenance cost.
7. **Consistent Conventions**: Use the same naming, directories, formats, types, identifiers, time representations, serialization, validation and error conventions for the same semantics. Follow existing conventions for queries, pagination, sorting, filtering, state and resource lifecycles. Follow the [architecture standard](docs/standards/architecture.md) for module boundaries and the language conventions of the affected module. Keep public resource IDs stable. Specify timestamp formats and time zones in contracts, such as RFC 3339 UTC; determine business dates in the domain's time zone.
8. **Explicit Business Behavior**: The responsible domain centrally defines and executes business operations, permission decisions, state machines, valid transitions and data validation. Specify invariants, preconditions, side effects and error semantics. For supported operations such as creation, updates, archival and deletion, define allowed initial and final states and responsibility for cleaning up related data.
9. **Reliable Operations**: Define transaction and consistency requirements for each operation. Handle concurrency, idempotency, timeouts, cancellation, deduplication and recovery. Define backoff and stopping conditions for retryable operations to avoid duplicate side effects. Protect invariants with transactions or version conditions when writes compete. Reuse the same operation identifier when retrying operations with side effects. When external dependencies fail persistently, assess whether their impact warrants circuit breaking.
10. **Security by Default**: Apply secure defaults and least privilege at trust boundaries. Handle authentication, authorization, sessions, tokens, input validation, sensitive data and secrets consistently; the backend makes security decisions. Validate sessions or tokens according to the [security standard](docs/standards/security.md), then authorize operations in the backend based on the operation, resource and membership.
11. **Event & Integration Consistency**: Define synchronous and asynchronous integration boundaries. Standardize event names, schemas, versions, publication, consumption and processing semantics. Handle ordering, duplicate delivery, retries and compatibility according to delivery guarantees. Before adopting CloudEvents for cross-system events, verify the receiver's supported formats and versions.
12. **Observable Operations**: Standardize structured logs, traces, metrics, correlation IDs and error classification. Prefer OpenTelemetry so key operations, performance and failures can be measured and cross-service requests traced and diagnosed. Correlate logs and traces for the same operation through identifiers. Use bounded categories for metric labels rather than individual resource IDs.
13. **Compatibility & Lifecycle**: Standardize versioning, API compatibility, data migration, configuration management and resource creation and deletion. Prefer backward-compatible changes and gradual migrations to protect existing users, data and integrations. For data changes spanning deployed versions, first expand compatible structures, migrate data and confirm old readers and writers have been retired before removing old structures. Describe changes in external releases according to their compatibility policy.
14. **Automated Enforcement**: Automate mechanical engineering checks through formatting, linting, type checks, schema validation, code generation, dependency checks, tests and CI. Regenerate after contract changes and inspect the diff so committed artifacts can be rebuilt from their sources.
15. **Minimal Change & Continuous Validation**: Analyze dependencies, root causes and impact before applying the smallest necessary fix. Validate types, contracts, test results, security, compatibility and core behavior according to each change's impact.
16. **Continuous Simplification**: Continuously remove or consolidate duplicate, conflicting, outdated, invalid and unnecessary code, abstractions, dependencies, configuration, documentation and processes. Measure before and after improvements and verify behavior with regression checks.
17. **Task Scope & Acceptance**: Understand the user's intended outcome, context and constraints before choosing an approach. Use available evidence to resolve routine ambiguity and continue independent work; ask only when a missing decision blocks the outcome. Verify results against the request and report material limitations.
18. **Applicable Instructions**: Select relevant guidance from the [engineering guidance](README.md#%E5%B7%A5%E7%A8%8B%E8%A6%8F%E7%AF%84) as needed. Before editing, read applicable AGENTS.md files along the target path. Reuse current guidance and evidence while applicable. Within host permissions and the user's task scope, nearer AGENTS.md files take precedence for local matters.
19. **Agent Execution**: Follow the [multi-agent collaboration guidelines](#multi-agent-collaboration-guidelines). Preserve existing changes and upstream skills; identify self-review accurately. Tools and capabilities do not grant additional authorization for external writes.
20. **Source Verification**: Verify local behavior against the current checkout's source and actual execution; verify external behavior against official documentation for the applicable version. Prioritize assumptions and failure risks that affect decisions. Reuse verified evidence while its version, source and applicability remain valid. Documentation research does not replace behavioral acceptance.
21. **Targeted Exploration**: Start exploration with a concrete question. Prefer available Graphify tools to locate unfamiliar dependencies; search source directly for known identifiers. Check graph and Repomix findings against the current working tree. Stop expanding searches once the evidence answers the question.
22. **Documentation Ownership**: README files own directory purposes and navigation. Authoritative topic documents own detailed specifications. Nested AGENTS.md files exist only for mandatory constraints that their parent cannot cover; they MUST NOT repeat, paraphrase or copy parent rules. Merge or remove instructions without independent value. Preserve safety, compatibility, business and architectural constraints in the nearest module rule or authoritative topic document. When behavior changes, check entry points, links and commands. API documentation references OpenAPI. Record reasons and tradeoffs for architectural decisions affecting multiple modules or difficult to reverse in existing decision documents or ADRs.
23. **Regression Verification**: Rerun the original case after a fix and follow the [verification workflow](docs/workflows/verification.md). For cross-directory moves, check the source, destination and callers. For Firestore transactions, concurrency, retries or persisted-data compatibility, proactively prepare and rerun applicable emulator tests. Choose verification by risk: unit tests for domain rules, integration or contract tests for storage and service boundaries, and end-to-end tests for critical user flows.
24. **Failure Isolation**: If tools are unavailable, graphs are stale or packages omit content, verify against available sources and record limitations. Defer only decisions dependent on missing evidence; continue independent work.
25. **Verified Delivery**: At delivery, compare results with specifications and acceptance criteria. Report implementation locations, capabilities used, actual checks, results and applicable versions. Distinguish completed changes, verified behavior and remaining limitations.
26. **Verified Publication**: Use `work` for changes submitted to `main` through the repository publication workflow. Each commit must be a complete, verified and independently reviewable milestone. Push only within the authorized scope and follow publication workflows and branch protection. Review staged changes before committing, confirm acceptance criteria and describe concrete changes in the commit message.
27. **Explicit Handoff**: When the user requests a handoff, takeover or transfer to another sidebar chat, provide a concise summary of the goal, current state, decisions, changed files, verification evidence and remaining work, plus the sources needed to take over.
28. **Configuration Boundaries**: Define configuration sources, precedence, required values and safe defaults. Parse and validate types at the loading boundary. Clearly identify missing required configuration by name; never include secret values in diagnostics.
29. **UI & Form Consistency**: Reuse components and interaction conventions from the [frontend standard](docs/standards/frontend.md). Map form fields to API contracts, distinguishing unfilled, cleared and unchanged values. Present immediate validation and backend field errors consistently; preserve correctable input after submission failures.
30. **Cache Consistency**: Before introducing caching, define acceptable staleness, resource and authorization scope of keys, expiration and invalidation conditions. Verify reads after writes or permission changes, and behavior on cache misses.
31. **Deployment & Recovery**: Before deployment, verify authorization, target environment and validated artifacts. After deployment, check health and core flows using the [release workflow](docs/workflows/release.md). Confirm the rollback version and data compatibility in advance; use acceptance results to decide whether to continue or recover.

## Backend → API → Frontend

- Prefer Next.js, React and existing components. Choose components and data-fetching approaches for the router and deployment model. The current Mini App uses the Pages Router and static export; verify router support before considering Server Components or Client Components.
- When responsibilities are misplaced, first check callers, dependencies and impact, then move, merge or remove duplicate implementations in Backend → API → Frontend order. Synchronize affected contracts and tests. Preserve functionality, API compatibility, security and performance. Report concrete risks first if the change's safety cannot be established.

<a id="&#24034;&#29376;&#25991;&#20214;&#32173;&#35703;&#27161;&#28310;"></a>

## Nested instruction standard

- Root AGENTS.md owns global execution rules. Top-level module AGENTS.md files contain only mandatory module constraints.
- Nested AGENTS.md files are exceptional. Keep one only when a narrower path has a safety, compatibility, business or generated-source restriction that cannot be expressed by its module parent or authoritative specification.
- Every rule uses `MUST`, `MUST NOT`, `SHOULD` or `MAY`, names its scope and states a verifiable condition. Do not include directory tours, file lists, start-here links or copied parent rules.
- Module-specific constraints live in the nearest module rules. Hexagonal backend boundaries remain in [backend module rules](backend/AGENTS.md#hexagonal-architecture) and the [architecture boundary specification](docs/architecture/boundaries.md).

<a id="cli-discovery"></a>
<a id="&#25353;&#38656;&#23433;&#35037;&#24037;&#20855;"></a>

## Tool installation and selection

### MCP selection and registration

- Contributors SHOULD use `context7` for dependency documentation, `openaiDeveloperDocs` for Codex or OpenAI API documentation, `graphify` to trace relationships and `repomix` to inspect a scoped source bundle. Select additional servers by the target module's AGENTS.md.
- A server is required only when the user explicitly requests it or the task’s acceptance criteria depend on a capability that available tools and source inspection cannot provide. Otherwise, contributors SHOULD use the available evidence.
- When a required MCP tool is unavailable, contributors MUST record the limitation, use current source or another authoritative source where possible, and follow the [troubleshooting guide](docs/operations/troubleshooting.md).

### Mandatory documentation verification

- **External behavior**: Choose version-matched official sources according to the [external systems reference](docs/references/external_systems.md). For Context7, resolve the library ID first, then query for the project's dependency version.

<a id="&#21934;&#19968;&#20195;&#29702;&#33287;&#33021;&#21147;&#36984;&#29992;"></a>

## **Multi-Agent Collaboration Guidelines**

- The primary agent is responsible for task decomposition, delegation, decision-making, coordination, integration, and final acceptance.
- Proactively use multi-agent collaboration for complex tasks, with a maximum of four concurrently running subagents.
- Dynamically create subagents based on task requirements, defining their responsibilities and scope without relying on predefined or custom roles.
- Dynamically determine the number of subagents based on task complexity and dependencies.
- Execute independent tasks in parallel and dependent tasks sequentially.
- Assign each subagent explicit objectives, scope, deliverables, constraints, and acceptance criteria.
- Prioritize existing Skills, MCP servers, tools, and project standards.
- Avoid redundant exploration, duplicate implementation, overlapping responsibilities, and file modification conflicts.
- Assign implementation and review to separate agents to ensure independent verification.
- The primary agent must wait for all relevant subagents to complete their tasks, collect results, cross-validate findings, resolve conflicts, and complete integration.
- If acceptance criteria are not met, delegate corrective actions and repeat verification until all criteria are satisfied. Explicitly report any blockers.
- Do not terminate prematurely or merely provide recommendations while tasks remain incomplete and execution is still feasible.
- Execute simple tasks directly to avoid unnecessary agent orchestration.
- Use actual subagent tools. Do not simulate multi-agent execution through a single agent.
- If the current environment does not support subagents, explicitly report the limitation and do not claim that subagents were used.
- The final report must include the actual number of subagents used, task assignments, execution results, and acceptance outcomes.

## Capability selection

Choose guidance from the [engineering guidance](README.md#%E5%B7%A5%E7%A8%8B%E8%A6%8F%E7%AF%84) when relevant to the task.

When diagrams clarify architecture, data relationships, flows or states, follow the [documentation standard](docs/standards/documentation.md) and validate Mermaid syntax with an available compatible renderer.

## graphify

The code graph lives under `graphify-out/`. Treat it as a derived navigation aid and corroborate findings with current source.

After completing the task-start reading in rule 18, if the user specifies `/graphify`, read the installed Graphify skill or operating guide before performing graph operations.

Rules:

- When exploring dependencies under rule 21, supplement graph results with rg, language references and source inspection for callers, routes, configuration, code-generation inputs and fixtures to cover paths omitted by the graph.
- Local graph changes after incremental updates are expected and do not prevent queries.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- Contributors MUST run `graphify extract . --force --code-only --allow-dedup-shrink --out . && graphify cluster-only .` when an edit changes indexed structure (for example, a package move or a changed call relationship), or when acceptance relies on a current code graph and the local graph is missing or stale. A copy-only edit that leaves indexed relationships intact MAY skip the rebuild.
- Code graphs are generated locally and ignored by Git. During read-only queries, update the current checkout's graph only if it is stale and current dependency information is needed.

For cross-file queries, use scoped search, graph or Repomix packaging as appropriate. Report stale or incomplete tool output and corroborate findings with current sources.

## Documentation routing

Before creating, editing or reviewing any owned Markdown, MUST read and follow the [documentation standards](docs/standards/documentation.md). The current agent uses documentation capability guidance as needed, confirms file ownership and independently fixes evidence-backed issues within scope without expanding into a repository-wide rewrite.

Use the [workflow index](docs/README.md#%E5%B7%A5%E4%BD%9C%E6%B5%81%E7%A8%8B) to select development, refactoring, verification and release guidance.

For implementation handoffs, preserve the goal, decisions, changed files, verification evidence and remaining limitations.

## Repository publication default

When the user asks to follow the GitHub MCP push workflow (「走 GitHub MCP 推送流程」), follow the [release workflow](docs/workflows/release.md) through its branch synchronization check. Honor any narrower endpoint the user specifies.

<a id="&#24037;&#31243;&#27770;&#31574;&#33287;&#20132;&#20184;&#35201;&#27714;"></a>

## Engineering decisions and delivery requirements

- Before engineering work, read the applicable [architecture](docs/standards/architecture.md), [backend](docs/standards/backend.md), [frontend](docs/standards/frontend.md), [API](docs/standards/api.md), [security](docs/standards/security.md) and [testing](docs/standards/testing.md) standards.
- For interfaces, interactions or user flows, follow the [frontend standard](docs/standards/frontend.md).
- First check domain models and key dependencies. Prioritize by impact and risk, and address issues sharing a root cause together.

<a id="&#20849;&#36890;&#39511;&#35657;&#33287;&#20132;&#20184;"></a>

## Common verification and delivery

For Mini App component work, follow the [shadcn MCP workflow](docs/workflows/shadcn_mcp.md). For Next.js runtime and routing diagnostics, follow the [Next.js MCP workflow](docs/workflows/nextjs_mcp.md).

For failing pull request checks, use GitHub MCP and follow the [PR verification workflow](docs/workflows/verification.md#用-github-mcp-處理-pr-紅燈) through the final check result.

Documentation-only and configuration-only changes do not require unrelated application tests. Behavior changes require the corresponding checks.
