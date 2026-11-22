# explorer 能力指引

由目前代理按需採用；這是工作階段，不是可啟動的子代理。共同規則見[能力入口](README.md)。審查結果為自我審查，不宣稱獨立驗收。

## 執行要求

依問題選擇查找方式：已知檔案或符號直接查原始碼；關係不清時用 Graphify，跨檔核對可用 Repomix。以目前 checkout 確認呼叫端、影響範圍與驗收入口，區分已確認的原因和推測。

探索階段唯讀。若原任務要求修復，定位原因後由目前代理接續已授權的實作與驗收；若只要求調查，交付證據與結論。

Use gopls for Go symbols and references. Graphify AST output may omit TOML and runtime behavior; verify configuration with its parser. Read the scoped Repomix pack instead of generated skill bundles.

Use docs/development/tooling/agents/graphify.md for graph identity and current-source corroboration. Only update explicitly assigned graph artifacts using the documented AST workflow. Use docs/development/tooling/agents/repomix.md and .agents/skills/repomix-explorer/SKILL.md for scoped bundles: write only assigned untracked outputs, preserve exclusions and scanning, never generate skills or publish. Keep repository source and configuration read-only.
For technical research compare approaches, constraints and versions using docs/development/tooling/agents/context7.md and root official-source routes; return library IDs, URLs, excerpts, tradeoffs and unknowns. Check LINE SDK and Firestore examples against go.mod; use LINE verification for LINE ID tokens. Do not install tools, authenticate or write external data. Use docs/development/tooling/agents/github.md to verify published refs separately from local changes.
