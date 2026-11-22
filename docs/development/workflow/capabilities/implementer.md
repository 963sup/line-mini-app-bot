# implementer 能力指引

由目前代理按需採用；這是工作階段，不是可啟動的子代理。共同規則見[能力入口](README.md)。審查結果為自我審查，不宣稱獨立驗收。

## 執行要求

依使用者要求的行為與架構限制作業。修復時重現原始症狀、追查原因、做最小修正，再重跑原案例與相關檢查；無法重現時說明證據與限制。見[除錯流程](../debugging.md)。

保留範圍外的公開行為與其他聊天的變更，遇到同檔衝突先釐清所有權。業務操作依契約維持後端授權及領域邊界，驗證失敗與重試語意。Go 程式明確處理錯誤、取消及資源所有權。修復開發環境時保留 TLS、校驗碼與來源驗證，可寫工具狀態放在工作區。

For HTTP handlers bound request bodies, propagate request contexts and return safe errors. Verify LINE webhook signatures before processing; design duplicate event handling. Do not add an application module or change dependencies unless the task requires it.

For Go use docs/standards/go.md, docs/development/tooling/backend/gopls.md and .agents/skills/golang-gopls/SKILL.md; inspect all references before a rename and change only owned files. Coordinate race checks once and use docs/development/tooling/testing/commands.md for delivery.
For TypeScript verify tsconfig, package versions and aliases with docs/development/tooling/agents/context7.md; preserve narrowing and cancellation and regenerate API types from the contract. For Next.js read frontend/miniapp/AGENTS.md and installed next/dist/docs. Preserve Pages Router static export, session lifecycle and Go server boundaries. Use docs/development/tooling/frontend/nextjs-devtools.md for scoped read-only runtime inspection.
Before UI work read .agents/skills/shadcn/SKILL.md, frontend/miniapp/src/shared/ui/AGENTS.md and docs/development/tooling/frontend/shadcn.md. Compose through public props and variants, protect upstream primitives and use dry-run/diff before authorized updates. One writer owns overlapping files. Report type, build and behavior checks separately from real browser/device results.
