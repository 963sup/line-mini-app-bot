# implementer 能力指引

由目前代理按需採用；這是工作階段，不是可啟動的子代理。共同規則見[能力入口](README.md)。審查結果為自我審查，不宣稱獨立驗收。

## 執行要求



Implement the assigned operation contract and confirmed in-scope fixes. Enforce authorization in the backend, preserve domain boundaries and verify failure/retry semantics.

Implement the assigned feature, refactor or bug fix within the requested behavior and
architecture constraints. Preserve user changes and public behavior outside the
task. Handle Go errors, context cancellation and resource ownership explicitly.
May repair development tooling/environment within an assigned scope; preserve
TLS, checksums and provenance checks, and keep writable tool state in workspace.
Run focused checks and report the changes, evidence and unresolved risks.
Preserve changes from other chats; resolve ownership before editing overlapping files.

For HTTP handlers bound request bodies, propagate request contexts and return safe errors. Verify LINE webhook signatures before processing; design duplicate event handling. Do not add an application module or change dependencies unless the task requires it.

For failures follow docs/development/workflow/debugging.md: reproduce the original symptom, trace its cause and verify the minimal fix with regression evidence. Report provisional diagnoses and skips.
For Go use docs/standards/go.md, docs/development/tooling/backend/gopls.md and .agents/skills/golang-gopls/SKILL.md; inspect all references before a rename and change only owned files. Coordinate race checks once and use docs/development/tooling/testing/commands.md for delivery.
For TypeScript verify tsconfig, package versions and aliases with docs/development/tooling/agents/context7.md; preserve narrowing and cancellation and regenerate API types from the contract. For Next.js read frontend/miniapp/AGENTS.md and installed next/dist/docs. Preserve Pages Router static export, session lifecycle and Go server boundaries. Use docs/development/tooling/frontend/nextjs-devtools.md for scoped read-only runtime inspection.
Before UI work read .agents/skills/shadcn/SKILL.md, frontend/miniapp/src/shared/ui/AGENTS.md and docs/development/tooling/frontend/shadcn.md. Compose through public props and variants, protect upstream primitives and use dry-run/diff before authorized updates. One writer owns overlapping files. Report type, build and behavior checks separately from real browser/device results.
