# MINI App 開發指引

<!-- BEGIN:nextjs-agent-rules -->

## This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->

## 架構與責任邊界

依層級維護本目錄的共用能力；功能畫面由 business 組織，頁面入口由 pages 接入。

## 閱讀與修改入口

先查閱[components.json](components.json)、[next-env.d.ts](next-env.d.ts)、[next.config.ts](next.config.ts)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## shadcn 與 sidebar-07 保護

任何 UI 任務先讀 [shared/ui/AGENTS.md](src/shared/ui/AGENTS.md) 的官方元件目錄、保護規則及 Skill／MCP 主動查詢流程；先探索、核對已安裝項目，再於應用層組合。

遵循[前端元件保護規則](../AGENTS.md#shadcn-元件保護)。Codex 預設不得改寫 shared/ui 與 shell/sidebar-07 的樣式、結構或互動；新增需求在應用層組合，只有使用者當次明確授權才能修改官方元件。

[現行實作圖：頁面與身分生命週期](../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)
