# shadcn UI、Skill、CLI 與 MCP

## 路徑與責任

shadcn 元件以可維護的原始碼納入 MINI App，不是後端服務或瀏覽器中的 MCP client。
`frontend/miniapp/components.json` 決定 CLI 的生成位置；`tsconfig.json` 的 `@/*` 對應 `src/*`。

- `src/shared/ui/`：可重用的 UI primitives，不直接呼叫業務 API 或判定權限。
- `src/shared/utils.ts`：元件樣式合併等共用工具；hooks 僅在元件實際需要時建立。
- `src/shell/`：共用外框與版面，組合 UI primitives。
- `src/business/`：業務畫面、操作狀態與領域組合，依既有 API 契約工作。

別名與 CSS 路徑以 `components.json` 為準；新增元件後檢查實際輸出與 import，不在各 feature 複製設定。

## 安裝與生成

CLI 固定版本與安裝任務由根目錄 `mise.toml` 管理；UI 與編譯依賴由 MINI App 的 `package.json` 和根目錄 lockfile 管理。
不要另在 workspace 安裝一份不同版本的 shadcn CLI，也不要以未固定的版本覆寫既有元件。
先依 [環境設定](../../development/setup.md) 安裝工具及 workspace 依賴，再使用固定 CLI。

從儲存庫根目錄執行：

```bash
bash scripts/tooling/tool-exec.sh shadcn --help
bash scripts/tooling/tool-exec.sh shadcn add button --cwd frontend/miniapp
```

`add` 會修改原始碼與可能的依賴，執行後應檢查 diff。既有 workspace 已有 `components.json`，不用重新 scaffold 或啟用 `--monorepo`。
新增或更新元件前以 Context7 對照官方文件與實際固定版本的 CLI help。

## 元件初始來源與更新邊界

初次整合曾實際執行 CLI 的 init 與元件 add（包含 local registry 輸入），但環境代理對 `ui.shadcn.com` 的 CONNECT 請求回傳 403；add 在寫入元件前仍需讀取 registry colors，因此未能完成元件生成。
既有 Button、Card、Input 改以官方 GitHub 原始碼手動安裝，來源為 [shadcn-ui/ui 官方原始碼](https://github.com/shadcn-ui/ui/tree/dd41d19e2f80ff6891951f7ced74c2c9a650b065/apps/v4/registry/new-york-v4/ui) revision `dd41d19e2f80ff6891951f7ced74c2c9a650b065` 的 `{button,card,input}.tsx`，並將 `cn` import 對應至本專案 `@/shared/utils`；Button 僅引入 `@radix-ui/react-slot`，避免安裝完整 Radix 元件集合。

CLI 固定版本以 mise 為唯一來源。Registry 網路已在下方驗證紀錄重測成功；既有元件保持原樣。更新時先看固定 CLI 的 `add --dry-run`／`--diff`，檢查依賴與 import 後再合併，避免無條件覆寫。

## CSS 編譯與版面

Tailwind 由 `postcss.config.mjs` 編譯 `src/styles/globals.css`，包含 Preflight，並由 Pages Router 的 `src/pages/_app.tsx` 載入。
Tailwind 4 的 `components.json` 不指定舊式 Tailwind config 檔案。
正式頁、定位頁及空間預覽以 utilities 排版，外觀由 shadcn 元件 variants 與主題變數控制。
已移除原有頁面 CSS、asset-map 與 CSS 複製流程；不再使用獨立 stylesheet 連結或全域選擇器覆蓋元件。

## MCP 設定與驗證

唯一連線設定在根目錄 `.codex/config.toml`，版本在 `mise.toml`；不在 feature 或瀏覽器設定 MCP，也不在文件複製 credentials。
`tool-exec.sh` 先切回儲存庫根目錄，因此 server 命令須在 wrapper 載入工具環境後再切到 `frontend/miniapp`，才能讀取正確的 `components.json`。

```bash
bash scripts/tooling/tool-exec.sh python3 scripts/tooling/mcp-doctor.py --server shadcn
```

分別驗證安裝、設定解析、initialize／tools/list，以及一個 registry 查詢等代表性唯讀操作。
本機直接啟動成功不代表目前聊天 host 已載入 server；更新設定後需重新啟動或重新載入 Codex，再確認 host 曝光的工具。
MCP 可提供 registry 資訊及新增命令；若回傳命令使用浮動版本，應改用本專案 mise 固定 CLI。生成、覆寫或安裝仍須遵循當次任務授權。

前端改動沿用 [前端檢查](../../../frontend/README.md)，並檢查窄螢幕外框、控制項與警告內容；MCP handshake 不代替 UI 或整合驗收。

查證來源：Context7 `/websites/ui_shadcn` 的 [components.json](https://ui.shadcn.com/docs/components-json)、[monorepo](https://ui.shadcn.com/docs/monorepo) 與 [MCP](https://ui.shadcn.com/docs/mcp)。文件描述目前 workspace 的配置方式；實際版本以工具與套件權威檔案為準，實機 LINE／Google 表單驗收另行記錄。

## 官方能力索引與適用範圍

本頁維護功能入口、專案路徑及實測狀態；完整且持續更新的元件目錄以官方網站／registry 為準，不複製成一份容易過期的「全部已安裝」清單。

| 能力 | 官方入口／操作 | 本專案狀態與邊界 |
| --- | --- | --- |
| Components | [完整元件目錄](https://ui.shadcn.com/docs/components)；CLI `docs`、`view`、`search` | 僅已安裝 Button、Card、Input；其餘按需求新增，先核對 API、依賴與適用 base |
| AI Elements | [官方入口](https://ui.shadcn.com/docs/ai-elements) | agentic UI 的元件組合模式，如 Chat、Reasoning、Tool、Plan；以原始碼納入，沒有必裝的獨立套件。本專案尚未安裝這些模式，也未新增聊天／模型服務 |
| Skill | [官方說明](https://ui.shadcn.com/docs/skills)、[本機 SKILL.md](../../../.agents/skills/shadcn/SKILL.md) | 相關任務先讀；提供專案 context、元件用法、組合、樣式、CLI／registry／MCP 指引；不取代 Context7 查證 |
| CLI | `info`、`docs`、`view`、`search`；`add`、`init`、`apply`、`migrate`、`build`、`preset`、`registry` | 前四項用於查詢；變更命令依任務範圍使用，逐項查 `--help`，不表示所有命令已端到端驗收。舊 `diff` 命令已 deprecated，改用 `add --diff` |
| MCP | [官方 MCP 文件](https://ui.shadcn.com/docs/registry/mcp) | 已配置、transport 及 registry search 實測通過；聊天 host 載入另驗證 |

目前 MCP `tools/list` 提供以下 7 個工具：

| 工具 | 用途 |
| --- | --- |
| `get_project_registries` | 讀取專案配置的 registry |
| `list_items_in_registries` | 列出 registry 項目 |
| `search_items_in_registries` | 搜尋元件與範例 |
| `view_items_in_registries` | 查看項目及來源內容 |
| `get_item_examples_from_registries` | 取得用法範例 |
| `get_add_command_for_items` | 提供新增命令；回傳命令不等於已安裝，浮動版本須換成本專案固定 CLI |
| `get_audit_checklist` | 提供變更後檢查項目；不等於檢查已執行 |

`components.json` 的 `ui` 為 `@/shared/ui`、`utils` 為 `@/shared/utils`、`components`／`lib` 為 `@/shared`、`hooks` 為 `@/shared/hooks`；`@/*` 解析至 `frontend/miniapp/src/*`。先執行 `info --cwd frontend/miniapp` 核對實際 context，不使用根目錄作為 UI 生成目錄。

從根目錄使用固定工具，閱讀 skill 的浮動版本命令範例時也採同一轉換：

```bash
bash scripts/tooling/tool-exec.sh shadcn info --cwd frontend/miniapp
bash scripts/tooling/tool-exec.sh shadcn docs button --cwd frontend/miniapp
bash scripts/tooling/tool-exec.sh shadcn view @shadcn/button --cwd frontend/miniapp
bash scripts/tooling/tool-exec.sh shadcn add @shadcn/button --dry-run --cwd frontend/miniapp
```

`docs` 回傳連結後仍須讀取相關官方內容。Graphify CLI 用於程式關係定位，Repomix CLI 用於必要範圍的來源交叉核對；前者的 AST 圖不完整涵蓋工具設定／文件，後者不應每次整庫打包。兩者版本／任務由 `mise.toml` 管理，見 [Graphify](graphify.md)、[Repomix](repomix.md)。

## 驗證紀錄

2026-11-12（Asia/Taipei），此紀錄是當次觀測，非版本配置的第二來源：

- 官方 supplied URLs：Skills、Components、AI Elements、Registry MCP 均成功讀取；並用 Context7 `/websites/ui_shadcn` 核對官方 skill 安裝方法。
- 官方 skill 以 pinned skills CLI 安裝單一 `shadcn` 至 `.agents/skills/shadcn`；來源 `shadcn/ui` 的 `skills/shadcn`，revision `dd41d19e2f80ff6891951f7ced74c2c9a650b065`，完整檔案與該 revision 對比一致。CLI lock 記錄來源與內容 hash；未修改上游 skill。
- shadcn CLI 已安裝，`--version` 為 `4.23.0`；`info` 正確識別 Next.js Pages Router、Tailwind v4、aliases 及三個既有元件。`view button card input` 與 `add ... --dry-run` 成功；dry-run 會覆寫三個元件並新增 `cn`／`radix-ui` 依賴，因此未實際套用。
- Registry index 與 new-york-v4/button URL 均 HTTP 200，先前 CONNECT 403 已非本次狀態。
- 依現有設定實際啟動 shadcn MCP：initialize／tools-list 成功（7 tools）；`search_items_in_registries` 在 `@shadcn` 搜尋 `button` 成功，找到 33 項，回傳前 2 項。其他 MCP 工具未逐項操作驗收。
- Graphify `0.9.103`、Repomix `1.18.1` CLI version 成功。現有 Graphify 忽略產物仍含重組前路徑，這次純技能／文件任務未重建，不能當目前 source coverage 證據。
- 無前端元件、依賴或執行行為變更，未執行應用測試、AI 功能或實機 LINE 驗收；本機 transport 成功仍不證明聊天 host 自動載入。

安裝的官方 skill 為本次 upstream revision，可能描述比本專案固定 CLI 更新的命令或元件；不保證其中每個例子都可在固定 CLI 使用，先以 `info` 與相關命令的 `--help` 核對，再決定操作。官方文件為查詢當下內容，不與固定 CLI 保證完全同版；使用未實測功能前仍須查該版本 help 與官方 API。Graphify 在 Context7 無可解析的 library，本次僅驗證既有本機 CLI，未宣稱完成其最新上游文件查證。
