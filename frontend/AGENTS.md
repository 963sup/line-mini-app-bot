<a id="frontend-工作指引"></a>

# frontend agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [miniapp](miniapp), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Documentation index](../docs/) · [Module entry points](../docs/)

Applies to this directory and its descendants. Follow the [root AGENTS.md](../AGENTS.md), then nearer directory guidance. User and host instructions take precedence; this file grants no additional authorization for external writes.

<a id="職責與邊界"></a>

## Responsibilities and boundaries

miniapp/ is the Next.js/React workspace. src/pages/ owns routes; business/attendance/ owns attendance UI; identity/, platform/line/, shared/ and shell/ own identity, platform, shared capabilities and framing.

Read miniapp/AGENTS.md and installed Next.js guidance first. Use @repo/api-client for HTTP; browsers obtain location and backends validate it. Do not edit build outputs.

## Thin Client & UI Gate

Applies to Next.js Pages Router, React/LINE MINI App components, pages, interactions and data flows. Follow [Mini App guidance](miniapp/AGENTS.md) and shadcn protection below.

<a id="職責與資料"></a>

### Responsibilities and data

- MUST use `@repo/api-client`/generated types for API reads/writes. Frontends own presentation, interaction, screen data/UI state. MUST NOT duplicate business DTOs, clients, authorization, authoritative calculations, consistency or transitions. Field checks provide immediate feedback only.
- MUST reuse sessions, LINE identity, location and retries. Distinguish loading, success, empty, denied, failed and unknown results; contracts determine retries. Buttons/sent requests do not prove business success.
- MUST preserve routing, Shell, Identity, Platform, Business/Shared boundaries. Investigate sources/callers/impact before consolidation. Move business authority backend; change API contracts before transport structures.

<a id="元件渲染與狀態"></a>

### Components, rendering and state

- MUST follow installed Pages Router guidance and [production static export](miniapp/next.config.ts). Build precomputable content; fetch sessions, LIFF, location and authenticated data through browser flows. Do not introduce per-request SSR or router migrations solely for new patterns.
- MUST compose existing shadcn/design-system/shared components, preserving keyboard access, accessible names, responsiveness/loading feedback. Respect primitive, Sidebar/generated-file restrictions. SHOULD use semantic HTML, existing Tailwind utilities/tokens/variants and avoid redundant DOM/CSS/global overrides.
- MUST separate server data, local UI state and derived values. SHOULD colocate state, derive during render, handle user actions in events and reserve Effects for external synchronization. Avoid synchronized state copies, duplicate requests, unnecessary waterfalls/polling.
- MUST NOT duplicate mappings, sessions, errors or interactions across pages, or add hooks/wrappers/providers/stores/adapters/factories/generic components solely for one case, hypotheticals or uniformity. Do not mix responsibilities/add branches just to reduce lines.
- SHOULD remove dead code/reuse components to reduce hydration, client JS/rerenders. Use actual needs/measurements for memoization, dynamic imports, dependencies or special performance mechanisms.

<a id="變更與驗收"></a>

### Changes and acceptance

- MUST run `pnpm --filter @repo/miniapp check`, relevant tests and necessary static builds per the [matrix](../docs/development/tooling/testing/commands.md#變更驗收矩陣) and verification below.
- MUST check affected pages' normal/loading/empty/error/denied states and key interactions. UI changes also verify accessibility, responsiveness and affected performance metrics.

**Completion:** business authority remains backend; UI retains necessary state, logic/dependencies only; existing interactions, API compatibility and applicable acceptance pass. Continue fixing failures; report blockers for unverified behavior.

<a id="驗證"></a>

## Verification

Check consuming pages and miniapp tests when changing interactions.

From root, use tool-exec.sh for pnpm --filter @repo/miniapp check and relevant existing tests; build when static output is needed. API changes also follow ../api/README.md.

Run commands from the repository root unless stated otherwise. Read the [local README](./) and [documentation index](../docs/); report actual results and limitations.

<a id="文件閱讀路由"></a>

## Documentation routing

Choose sources through [related module documentation](./) and [module entry points](../docs/). Maintain local purposes and commands in this module's README; organize cross-module architecture, feature and operations documents by topic under docs. Follow the [documentation maintenance guidance](../docs/AGENTS.md#主題子目錄) when updating indexes.

## Business Operation

Before related work, read the [operation model](../docs/contracts/business-operations.md) and [compliance workflow](../docs/development/workflow/implementation.md#business-operation-compliance).

UI presents availability, confirmation/results; hiding buttons is not backend authorization. Handle rejections, stale versions/failures by API contract. Unknown results follow retry contracts without self-declared permissions/success.

[Current implementation diagram: page and identity lifecycle](../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

<a id="前端操作與業務契約"></a>

## Frontend operations and business contracts

Review operation results using [frontend, API and backend responsibilities and retry examples](../docs/contracts/business-operations.md#前端操作與業務契約).

<a id="shadcn-任務入口"></a>

## shadcn task entry point

For every frontend UI task, proactively read [component protection/Skill/MCP workflow](miniapp/src/shared/ui/AGENTS.md), then check [reference freshness](../docs/development/tooling/frontend/shadcn-reference/) and search complete Components/AI Elements indexes/docs. No repeated user invocation is needed. The reference library alone generates indexes; follow UI host-exposure diagnostics for absent MCP tools.
Follow its requirements-driven selection/acceptance section for component choices, additions and verification. Complete the current task without limiting candidates to installed components or expanding into unsolicited redesigns.

Before shadcn/ui, components.json, registry, preset or MCP work, all roles read [.agents/skills/shadcn/SKILL.md](../.agents/skills/shadcn/SKILL.md) and relevant references. See the [tool guide](../docs/development/tooling/frontend/shadcn.md) for usage/verification.

For statistics, reports/dashboards, consult [official chart examples](../docs/development/tooling/frontend/shadcn.md#官方圖表範例), select for data needs and reuse shared Chart components.
Preserve upstream skills. Replace examples such as `npx shadcn@latest` with `bash scripts/tooling/tool-exec.sh shadcn`, targeting `frontend/miniapp`. Read actual docs/installed components; preview updates with `--dry-run`/`--diff`. Examples do not authorize project recreation or wholesale installation.

<a id="shadcn-元件保護"></a>

## shadcn component protection

By default, Codex must not directly modify official styles, structure/interactions in `frontend/miniapp/src/shared/ui/*.tsx`, `shared/hooks/use-mobile.ts` or `src/shell/sidebar-07/*.tsx`, change primitives for individual pages or override appearance globally.
Compose business needs through public props, variants, data/events in `shell/AppFrame.tsx`, `shell/Shell.tsx` or business layers. Keep real sidebar-07 navigation/identity/login; do not restore demo accounts, fake links or subscription/billing.
Only explicit current-task requests to modify/restore/update these components authorize changes. Compare with pinned CLI `add --dry-run`/`--diff`, make minimal edits and run checks. Do not independently run `add sidebar-07 --overwrite`; it creates separate shared/dashboard files rather than updating integrated shell code.
Preserve aliases, client directives and recorded accessibility fixes. Do not remove `value` passed to Root by `Progress` when applying upstream code. Preserve skills; local protection overrides general customization advice.

Before adding, editing or reviewing business models, API contracts, resource paths, collections or query fields, all roles must read the [naming standards](../docs/standards/naming.md#業務實體集合與-api-命名). Use singular names for individual business entities and plural names for collections and REST resource paths, respecting existing compatibility exceptions. Maintain the full rules only in that document; do not directly rename existing public contracts or persisted data.

## Applicable engineering conventions

- **Types and forms:** Use generated API types and existing field components; client validation supplies feedback while the backend decides business validity and permissions.
- **UI and state:** Compose the design system and distinguish loading, empty, denied, failed and unknown outcomes with accessible controls.
- **Recovery and privacy:** Keep draft, request and idempotency identity intact after an unknown write; discard scoped data when identity or authorization changes.
- **Acceptance:** Validate types, affected interactions and required static routes, including keyboard/focus behavior and narrow layouts.
