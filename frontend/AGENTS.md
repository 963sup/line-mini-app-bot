# Frontend module rules

This file adds frontend-specific rules to the [repository rules](../AGENTS.md). Directory purposes belong in the [root README](../README.md#巢狀目錄導覽). Detailed path constraints are indexed in [module constraints](../docs/standards/module-constraints.md).

- **Runtime.** The MINI App MUST remain compatible with its installed Next.js Pages Router and static export. Before changing framework behavior, read the matching installed documentation under `frontend/miniapp/node_modules/next/dist/docs/`. Pages MUST NOT depend on per-request server execution.
- **Thin client.** API reads and writes MUST use `@repo/api-client` and generated types. The browser owns presentation, interaction and local UI state. It MUST NOT duplicate authorization, authoritative calculations, transitions or transport DTOs.
- **Boundaries.** Pages compose Shell and feature modules. Identity, Platform, Business and Shared dependencies MUST retain their existing direction. Provider profiles and hidden controls MUST NOT be treated as business authorization.
- **State.** Server data, local UI state and derived values MUST remain distinct. Obsolete requests and watchers MUST be cancelled or ignored. Actor or resource changes MUST clear scoped private data while preserving unresolved-write protection.
- **Recovery.** Unknown writes MUST retain their original request and idempotency identity. A timeout or cancellation MUST NOT be interpreted as confirmed failure.
- **UI compatibility.** Existing shadcn primitives, `shared/ui`, `shared/hooks/use-mobile.ts` and `shell/sidebar-07` MUST be composed through public APIs. Direct upstream-style changes require explicit task scope and MUST preserve accessibility fixes, aliases and client directives.
- **Accessibility.** Changed flows MUST preserve accessible names, keyboard and focus behavior, responsive layouts, and distinct loading, empty, denied, failed and unknown states.
- **Verification.** Frontend changes MUST run `bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp check`, relevant tests and any static build required by the [acceptance matrix](../docs/development/tooling/testing/commands.md#變更驗收矩陣).
