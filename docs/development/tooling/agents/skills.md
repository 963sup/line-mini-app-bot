# 專案 Skills

> 安裝範圍：本儲存庫 `.agents/skills/`。新安裝來源與固定 commit 見 `skills-lock.json`。
> 安裝檔案不代表當前聊天已重新載入，亦不代表 MCP 連線與實際操作已驗收。

| Skill                          | 來源                                            | 用途                                                                             |
| ------------------------------ | ----------------------------------------------- | -------------------------------------------------------------------------------- |
| gcloud                         | google/skills（官方）                           | CLI 指令核對與操作規則                                                           |
| cloud-run-basics               | google/skills（官方）                           | Cloud Run 部署與服務身分                                                         |
| retrieving-developer-knowledge | google/skills（官方）                           | Google 官方文件查詢                                                              |
| golang-gopls                   | samber/cc-skills-golang（使用者指定社群）       | Go 語意探索、診斷、重構                                                          |
| context7-mcp                   | upstash/context7（官方）                        | 版本相符文件查詢                                                                 |
| repomix-explorer               | yamadashy/repomix（官方）                       | CLI 打包、搜尋、分析                                                             |
| openai-docs                    | openai/skills（官方）                           | OpenAI／Codex 官方文件                                                           |
| shadcn                         | shadcn/ui（官方；CLI lock 見 skills-lock.json） | 元件、registry、preset、MCP 工作流程；[專案規則](../../../../frontend/AGENTS.md) |
| graphify                       | 既有 Graphify CLI 安裝                          | 程式關係圖查詢                                                                   |

## MCP 對應與搜尋範圍

依本地 `.codex/config.toml` 對應；Google 設定與驗證見 [Google MCP](google.md)。
`mise run install-graphify` 安裝上游 agents 版本到 [Graphify 技能目錄](../../../../.agents/skills/graphify/SKILL.md)。本專案不執行技能中的子代理派工；採用共用 Graphify CLI 的 AST 與文件圖流程。若任務需要語意分析，由目前代理在授權範圍內逐段核對來源，明列涵蓋限制，不宣稱已執行 worker。
Repomix 與 repomix_skills 共用 repomix-explorer；generate_skill 是工具能力，不等於另一個 skill。
截至 2026-11-08，在 github/github-mcp-server 與 golang/tools 的預設分支遞迴檔案樹未找到 SKILL.md；不聲稱整個 GitHub 沒有官方技能。
Samber 的 golang-* skills 為使用者指定社群技能，不能標為 Go 官方發布。

## 專案規則優先

技能上游檔案保留原樣；使用者指示與系統／宿主限制優先。尤其：

- LINE 問題先查 LINE 官方來源；Google Cloud 先用 Google skills／知識庫，其他技術問題先查 Context7。
- 工具由 mise 管理，不照上游範例擅自 go install @latest 或修改宿主全域設定。
- 使用實際暴露的 MCP 工具名稱與能力；上游工具矩陣不是本宿主一定支援的證據。
- 不因安裝 skill 就登入、部署、建立雲資源或輸出 token；外部操作仍依當次授權。
- `.agents/skills` 是新技能的單一來源，不再複製到 `.codex/skills` 產生雙份漂移。
- 上游相對 references/scripts 會一併保存；跨其他 skill 的引用不代表那些 skill 已安裝。

更新時依 manifest 的 repo/path 先查核新 commit 與差異，檢查完整子目錄、授權與腳本，再更新來源鎖定；目前安裝不會在 bootstrap 自動追最新版本。

## skills.sh CLI

CLI 來源為 [vercel-labs/skills](https://github.com/vercel-labs/skills)，工具版本由 mise.toml 固定。
安裝：`mise run install-skills-cli`；列出 Codex 專案技能：`mise run skills-list`。
也可 `mise exec -- skills --help` 或 `mise exec -- skills add owner/repo --list` 查詢候選。
CLI 安裝不等於批次安裝網站技能；add/update/remove 會修改技能檔案，僅在任務授權範圍使用。
專案技能來源與內容雜湊由根目錄技能鎖檔維護。

## Samber 技能與 Codex plugins

Samber 技能依目前專案需求保留通用 Go 開發指引。已移除目前程式與依賴未使用的 GraphQL、Wire、Dig、Fx、Cobra、Viper、Testify 與 Samber 套件專用技能。資料層使用 Firestore，OpenAPI 由契約產生，因此移除 SQL 資料庫與 swaggo 專用技能。亦移除需要未配置的 godig 工具的查詢技能，以及 Go 社群資訊追蹤技能。Go 技能由任務按需選用；套件配置與六邊形邊界依[後端指引](../../../../backend/AGENTS.md)維護。後端僅引用 gRPC 錯誤狀態與代碼，因此省略 gRPC 服務開發技能。安裝清單以 `skills-lock.json` 及 `.agents/skills/` 為準。
CLI 來源／內容 hash 記錄在根目錄 skills-lock.json。golang-gopls 早期手動安裝的 revision 為 bb1d43d98a3f8ab639a0042f90d0ec524e39a477，現行安裝以鎖檔內容雜湊為準。
`--all` 同時選擇其他 agent 目標；本 repo 提交 canonical .agents/skills，Claude/Eve 額外副本留在本機。

本機使用者層已安裝（不隨 Git 複製；當前聊天載入與實際功能尚未驗收）：

```bash
codex plugin marketplace add wshobson/agents
codex plugin add python-development@claude-code-workflows
```

本次觀測 python-development 為 1.2.4；這是安裝紀錄，不是重建版本鎖定。
安裝不授權 plugin 自動進行對外操作，目前代理依[能力指引](../../workflow/capabilities/)選用所需能力，文件查證依根目錄指引。

## Next.js 官方技能

`next-dev-loop` 用於開發伺服器上的行為驗證，`next-bundle-optimizer` 用於前端 bundle 分析。兩者取自 [Next.js 官方儲存庫](https://github.com/vercel/next.js/tree/v16.4.5/skills) 對應專案版本的標籤，來源與內容雜湊由 `skills-lock.json` 保存。Mini App 使用 Pages Router 靜態匯出。框架一般指引見專案所安裝版本的 `next/dist/docs/`。Go 技能用於後端 Go 模組。

## shadcn 官方 skill

依 [官方安裝文件](https://ui.shadcn.com/docs/skills) 使用本專案固定 skills CLI：

```bash
bash scripts/tooling/tool-exec.sh skills add shadcn/ui --skill shadcn --agent codex --yes
```

單一安裝位置為 `.agents/skills/shadcn/`，保留完整 rules、references、assets 與 scripts，不另複製到 `.codex/skills`。來源與內容 hash 由 `skills-lock.json` 保存；來源 revision／檔案核對見 [shadcn 驗證紀錄](../frontend/shadcn.md#驗證紀錄)。安裝不代表當前聊天已自動載入；根目錄 AGENTS.md 明確要求相關任務先讀 SKILL.md。上游浮動版本範例須依專案規則換成 mise 固定 CLI，完整功能與狀態見同一工具指南。
