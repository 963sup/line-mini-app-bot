# shadcn UI 與 MCP

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

## 本次安裝來源與限制

已實際執行 CLI 的 init 與元件 add（包含 local registry 輸入），但環境代理對 `ui.shadcn.com` 的 CONNECT 請求回傳 403；add 在寫入元件前仍需讀取 registry colors，因此未能完成元件生成。
本次 Button、Card、Input 改以官方 GitHub 原始碼手動安裝，來源為 [shadcn-ui/ui 官方原始碼](https://github.com/shadcn-ui/ui/tree/dd41d19e2f80ff6891951f7ced74c2c9a650b065/apps/v4/registry/new-york-v4/ui) revision `dd41d19e2f80ff6891951f7ced74c2c9a650b065` 的 `{button,card,input}.tsx`，並將 `cn` import 對應至本專案 `@/shared/utils`；Button 僅引入 `@radix-ui/react-slot`，避免安裝完整 Radix 元件集合。

CLI 固定版本仍以 mise 為唯一來源。這次替代來源不代表 registry 網路已恢復，也不代表 CLI add 的端到端流程已通過；網路可用後再以固定 CLI 核對更新 diff，避免無條件覆寫本地元件。
MCP 的 transport、registry 操作與聊天 host 載入仍須各自記錄驗證結果。

## CSS 編譯與既有版面

Tailwind CSS 由 `postcss.config.mjs` 編譯 `src/styles/globals.css`，並由 Pages Router 的 `src/pages/_app.tsx` 載入。
Tailwind 4 的 `components.json` 不指定舊式 Tailwind config 檔案。
本次整合不啟用 preflight，避免重設既有表單、按鈕與六頁 compact 版面的瀏覽器樣式。

既有 `src/shell/app.css` 仍由 asset-map 複製到 `/assets/app.css`；該流程不執行 PostCSS。
不可把未編譯的 Tailwind import／指令放進這份直接複製的 CSS。後續遷移應逐個元件進行，保留既有頁面行為並檢查 CSS 優先序。

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
