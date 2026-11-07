# 專案 Skills

> 安裝範圍：本儲存庫 `.agents/skills/`。新安裝來源與固定 commit 見 `.agents/skills-sources.json`。
> 安裝檔案不代表當前聊天已重新載入，亦不代表 MCP 連線與實際操作已驗收。

| Skill | 來源 | 用途 |
| --- | --- | --- |
| fly | superfly/skills（官方） | Fly CLI／MCP、Machines、部署與維運 |
| brand | superfly/skills（官方） | 僅需 Fly 品牌素材／文案時使用 |
| golang-gopls | samber/cc-skills-golang（使用者指定社群） | Go 語意探索、診斷、重構 |
| context7-mcp | upstash/context7（官方） | 版本相符文件查詢 |
| repomix-explorer | yamadashy/repomix（官方） | CLI 打包、搜尋、分析 |
| openai-docs | openai/skills（官方） | OpenAI／Codex 官方文件 |
| graphify | 既有 Graphify CLI 安裝 | 程式關係圖查詢 |

## MCP 對應與搜尋範圍

依本地 `.codex/config.toml` 對應；安裝時未變更 MCP 設定。
Graphify 已由現有 install-graphify 任務管理，不覆蓋現有兩處技能。
Repomix 與 repomix_skills 共用 repomix-explorer；generate_skill 是工具能力，不等於另一個 skill。
截至 2026-11-08，在 github/github-mcp-server 與 golang/tools 的預設分支遞迴檔案樹未找到 SKILL.md；不聲稱整個 GitHub 沒有官方技能。
golang-gopls 是 Samber skill，不能標為 Go 官方發布。

## 專案規則優先

技能上游檔案保留原樣；使用者指示與系統／宿主限制優先。尤其：
- LINE 問題先查 LINE 官方來源；其他技術問題與實作不確定先查 Context7。
- 工具由 mise 管理，不照上游範例擅自 go install @latest 或修改宿主全域設定。
- 使用實際暴露的 MCP 工具名稱與能力；上游工具矩陣不是本宿主一定支援的證據。
- 不因安裝 skill 就登入、部署、建立雲資源或輸出 token；外部操作仍依當次授權。
- `.agents/skills` 是新技能的單一來源，不再複製到 `.codex/skills` 產生雙份漂移。
- 上游相對 references/scripts 會一併保存；跨其他 skill 的引用不代表那些 skill 已安裝。

更新時依 manifest 的 repo/path 先查核新 commit 與差異，檢查完整子目錄、授權與腳本，再更新來源鎖定；目前安裝不會在 bootstrap 自動追最新版本。
