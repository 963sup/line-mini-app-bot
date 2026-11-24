# explorer 能力指引

探索時先回答兩件事：證據在哪裡、哪些行為還要實測。由目前代理依[能力入口](./)執行。

## 執行要求

- 已知路徑或符號時直接查原始碼。Go 符號與引用用 gopls；設定用對應解析器核對，執行行為用實測確認。
- 關係不清楚時，用 [Graphify](../../tooling/shared/graphify.md) 定位，再從目前 checkout 核對呼叫端、影響範圍及驗收入口。圖譜更新時機依[根圖譜規則](../../../../AGENTS.md#graphify)，不因探索而一律重建產物。
- 跨檔查詢可搜尋既有的 [$repomix-reference-line-mini-app-bot](../../tooling/shared/repomix.md#查閱現有技能)。需要限定來源時，依 [Repomix 指南](../../tooling/shared/repomix.md)與本機 `.agents/skills/repomix-explorer/SKILL.md` 打包，保留排除規則與安全掃描，只寫入指定的未追蹤產物。
- 查外部行為時，依[官方來源規則](../../../../AGENTS.md#mandatory-documentation-verification)與 [Context7 指南](../../tooling/agents/context7.md)比對版本、做法及限制。記錄 library ID、網址、支持結論的原文與待確認處。LINE SDK 與 Firestore 範例須對照 go.mod；LINE ID token 使用 LINE 驗證。已發布的 Git ref 依 [GitHub 指南](../../tooling/shared/github.md)另行核對。

唯讀調查保持原始碼與設定不變；工具準備、登入及外部寫入依任務授權與根指引處理。分清已確認的原因與推測，使用者要求修復時持續處理至驗收。
