# explorer 能力指引

探索時先回答兩件事：證據在哪裡、哪些行為還要實測。由目前代理依[能力入口](./)執行。

## 執行要求

- 已知路徑或符號時直接查原始碼。Go 符號與引用用 gopls；設定用對應解析器核對，執行行為用實測確認。
- 關係不清楚時，用 [Graphify](../../tooling/agents/graphify.md) 定位，再從目前 checkout 核對呼叫端、影響範圍及驗收入口。只有任務指定更新圖譜時才重建對應產物。
- 跨檔查詢可搜尋既有的 [$repomix-reference-line-mini-app-bot](../../tooling/agents/repomix-skills.md)。需要限定來源時，依 [Repomix 指南](../../tooling/agents/repomix.md)與本機 `.agents/skills/repomix-explorer/SKILL.md` 打包，保留排除規則與安全掃描，只寫入指定的未追蹤產物。
- 查外部行為時，依[官方來源規則](../../../../AGENTS.md#mandatory-documentation-verification)與 [Context7 指南](../../tooling/agents/context7.md)比對版本、做法及限制。記錄 library ID、網址、支持結論的原文與待確認處。LINE SDK 與 Firestore 範例須對照 go.mod；LINE ID token 使用 LINE 驗證。已發布的 Git ref 依 [GitHub 指南](../../tooling/agents/github.md)另行核對。

探索期間保持原始碼與設定唯讀，不安裝工具、登入、發布或寫入外部服務。交付調查時分清已確認的原因與推測。若任務已要求修復或更新參考技能，完成定位後再進入對應的實作與驗收步驟。
