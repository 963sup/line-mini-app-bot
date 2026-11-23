<a id="mini-app-開發指引"></a>

# MINI App development guidance

<!-- BEGIN:nextjs-agent-rules -->

## This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->

<a id="架構與責任邊界"></a>

## Architecture and responsibility boundaries

Maintain shared capabilities by layer; business owns feature UI and pages owns entry points.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [components.json](components.json), [next-env.d.ts](next-env.d.ts), [next.config.ts](next.config.ts), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="shadcn-與-sidebar-07-保護"></a>

## shadcn and sidebar-07 protection

Before UI work, read the catalog, protection and proactive Skill/MCP workflow in [shared/ui/AGENTS.md](src/shared/ui/AGENTS.md). Explore, check installed items, then compose application UI.

Follow [frontend protection](../AGENTS.md#shadcn-元件保護). Codex must not rewrite shared/ui or shell/sidebar-07 styles, structure/interactions by default. Compose new needs in applications; official changes require explicit current-task authorization.

[Current implementation diagram: page and identity lifecycle](../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

## Applicable engineering conventions

- **Structure and builds:** Keep route composition in Pages Router and preserve static export; never make a page depend on unavailable per-request server execution.
- **Components and forms:** Reuse protected components through public props and application composition, with accessible field labels and error feedback.
- **Configuration and types:** Keep public runtime configuration separate from secrets and use the pinned workspace dependency/type-checking workflow.
