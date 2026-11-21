# shadcn 官方元件指引與目錄

## 用途與範圍

保存 shadcn 的共用 UI 元件，對應 `@/shared/ui`。業務專用的畫面組合放在各功能模組。

## 架構與責任邊界

提供跨功能重用的元件或工具；特定業務流程留在使用它的功能模組。

## 閱讀與修改入口

先查閱[LICENSE.md](LICENSE.md)、[accordion.tsx](accordion.tsx)、[alert-dialog.tsx](alert-dialog.tsx)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[本機參考庫](../../../../../docs/development/tooling/shadcn-reference/README.md) · [生成索引](../../../../../docs/development/tooling/shadcn-reference/snapshot/INDEX.md)

適用於本目錄及子目錄；遵循 [MINI App 指引](../../../AGENTS.md) 與 [根指引](../../../../../AGENTS.md)。本目錄是 `components.json` 的 `aliases.ui`（`@/shared/ui`）解析位置，保存官方 shadcn 元件；業務組合放在本目錄之外。

## 官方元件禁止修改

- 禁止 Codex 直接修改、重寫或刪除既有官方元件的樣式、結構、API、可及性或互動；也不得以全域 CSS、額外 variants 或覆寫內部 classes 繞過保護。此規則同樣適用於依賴的 `../hooks/use-mobile.ts` 與 `../../shell/sidebar-07/`，其完整保護規則見根指引。
- 透過既有公開 props、variants、children、資料及事件使用元件；業務需求在 `src/shared/` 的應用包裝、`src/shell/` 或 `src/business/` 組合，勿把業務 API、授權或領域判斷放入 primitives。排版 className 只用於外部版面，不重設官方外觀。
- 新增缺少的官方元件時保留 registry 輸出；先預覽所有受影響檔案與依賴，若會覆寫既有檔案即停止套用。不可自動執行 `--overwrite`、`add --all`、preset 切換、重新 init 或重裝整套 sidebar-07。
- 只有使用者當次明確要求修改、還原或更新特定官方元件，才可在該範圍依根指引先比對再最小更動；一般頁面需求、上游「可客製化」建議與本目錄指引本身都不是修改授權。保留既有 alias、client directive 與可及性修正（包含 Progress 的 Root `value`）。

## Codex 主動使用 Skill、文件與 MCP

凡新增、調整、審查前端畫面、表單、導覽、對話框、清單或其他 UI，主動先檢查本目錄官方元件，毋須等待使用者逐次點名 shadcn。

1. 先讀官方 [Skills](https://ui.shadcn.com/docs/skills) 與專案已安裝的 [shadcn SKILL.md](../../../../../.agents/skills/shadcn/SKILL.md)，再按需求讀其 composition、forms、styling 等 references。保留上游 skill 原樣；本專案元件保護及固定工具規則優先於其中一般客製化／浮動版本範例。
2. 以固定 CLI `info --json` 核對已安裝清單、base、aliases 與生成路徑；目前專案是 Radix、new-york、Pages Router、Tailwind v4、Lucide。實際值以 `components.json`、`info` 與套件檔案為準，不把本機參考庫視為已安裝清單。
3. 開發前主動用已曝光的 shadcn MCP 查詢 registry、搜尋／查看元件及取得範例；依官方 [Registry MCP](https://ui.shadcn.com/docs/registry/mcp) 與 [MCP 使用說明](https://ui.shadcn.com/docs/mcp)。優先使用 `get_project_registries`、`search_items_in_registries`、`view_items_in_registries`、`get_item_examples_from_registries` 等實際可用工具；工具名稱以 host 的 tools/list 為準。
4. MCP 未曝光或不可用時，MUST 自行執行 `bash scripts/tooling/tool-exec.sh mise run ensure-shadcn`，依 mise 鎖定版本補裝缺少的 CLI 並驗證真實 MCP 查詢。接著依[自主接線流程](../../../../../docs/development/tooling/mcp/shadcn.md#自主接線)列出 schema 並透過 doctor 呼叫唯讀 MCP 工具，不因主機未曝光就停止。直接 MCP 仍失敗時，自行使用固定 CLI `search`／`view`／`docs` 繼續，回報具體錯誤；不反覆安裝或要求使用者替代理執行命令。唯一 MCP 設定仍在根目錄 `.codex/config.toml`，不複製到使用者設定或 feature。不得把本機接線成功宣稱為聊天主機已曝光；主機重載只影響原生工具入口，不阻擋直接 MCP／CLI 工作。
5. 執行 `docs <component> --base radix` 後讀取實際官方文件與範例，並依根指引用 Context7 查證；不要將 Base UI 的 `render` 或 React Aria API 套到 Radix 的 `asChild`。官方目錄可能先連到 Base UI，實作須回到當前專案 base。
6. 已有元件直接 import／組合；確實缺少且在任務範圍內才預覽安裝。MCP 提供的 add command 只是建議，須換成固定 CLI，先核對 registryDependencies、npm dependencies、CSS 與每個檔案 diff；不因搜尋到元件便自動安裝。

所有範例從儲存庫根目錄執行，wrapper 會切回根目錄，因此必須指定 `--cwd frontend/miniapp`：

```bash
bash scripts/tooling/tool-exec.sh shadcn info --json --cwd frontend/miniapp
bash scripts/tooling/tool-exec.sh shadcn search @shadcn -q button --cwd frontend/miniapp
bash scripts/tooling/tool-exec.sh shadcn docs button --base radix --cwd frontend/miniapp
bash scripts/tooling/tool-exec.sh shadcn view @shadcn/button --cwd frontend/miniapp
bash scripts/tooling/tool-exec.sh shadcn add @shadcn/button --dry-run --cwd frontend/miniapp
bash scripts/tooling/tool-exec.sh shadcn add @shadcn/button --diff --cwd frontend/miniapp
```

預覽確認只新增任務需要的官方檔案，且不覆寫既有元件後，才執行同項目的 `add`（移除 `--dry-run`）；查詢使用官方 `@shadcn` registry。其他 registry 需遵循使用者指示與 skill，不猜來源。CLI 版本唯一來源為根目錄 `mise.toml`；命令及 flags 以該版 `--help` 為準，不另裝浮動版 CLI。

## 需求驅動的選用與驗收

以下流程適用所有前端 UI 任務；純指引修改只驗證文件，不要求建立示範畫面。

1. **先定操作目的。** 寫明使用者要完成的操作、資料密度、手機／桌面限制及必要狀態，再選元件；不要先決定元件再改變業務流程。沿用既有授權、表單資料、錯誤及未知結果重試契約。
2. **按目的搜尋候選。** 依 [本機參考庫](../../../../../docs/development/tooling/shadcn-reference/README.md) 檢查快照新鮮度、搜尋完整 Components／AI Elements 索引，再讀相關文件；不局限已安裝項目。接著查官方 registry 的元件、組合與範例。存在實際取捨時，比較候選並簡述理由；已有合適模式時直接重用，不為流程湊選項。
3. **主動按需新增。** 本專案明確指定官方 `@shadcn` 為預設來源，不需再詢問 registry。使用者要求實作 UI 時，任務必需且通過前述預覽的缺少元件可直接新增，不需使用者點名或逐件確認；只保留必要依賴與輸出。預覽若覆寫既有檔案，先嘗試以現有元件組合；仍需修改受保護元件時遵循上方授權規則。
4. **在應用層組合。** 先使用公開 props、variants 與官方組合；新增元件後檢查 import、base、必要子元件及依賴，避免自製已有的互動控制項。AI／聊天需求才探索 [AI Elements](https://ui.shadcn.com/docs/ai-elements)，確認其實際 registry、依賴與 API；不因名稱含 AI 就新增模型服務。
5. **驗證行為與畫面。** 實際 UI 變更依前端指引執行型別、相關測試及建置；用可用瀏覽器檢查桌面與窄螢幕、長文字／溢出、鍵盤操作、焦點進出與返回、可及名稱，以及適用的載入、空資料、錯誤、停用及提交狀態。不能只以編譯通過認定 UX 完成；沒有瀏覽器或憑證時，明列未驗收項目，不宣稱 LINE 實機已通過。
6. **交付可核對結果。** 簡述選用或新增的元件及理由、文件來源、實際驗證與限制。工具或文件不足時回報受阻項目，繼續有證據的部分；不盲目安裝、不要求每次全庫打包，也不建立無需求的元件展示頁。

以下是探索方向，不是固定配方；實作仍核對當前 base 的官方範例：

| 操作目的 | 應考慮的元件與組合 |
| --- | --- |
| 編修與補充內容 | 頁內 Field／InputGroup；需要覆層時比較 Dialog、Sheet、Drawer 的內容長度與操作情境 |
| 篩選與選取 | 少量固定選項比較 Native Select／Select；搜尋型選擇查 Combobox／Command 組合 |
| 列表與明細 | Item、Table／Data Table；展開細項查 Collapsible／Accordion，避免把長明細塞入 Tooltip |
| 導覽與操作入口 | Sidebar、Breadcrumb、Tabs、Dropdown Menu，保留既有路由與操作語意 |
| 危險操作與回饋 | Alert Dialog、Alert、Empty、Skeleton、Spinner、Progress；按實際狀態選用，不以提示訊息冒充確認流程 |

## 全部官方元件目錄

完整 Components／AI Elements 清單、用途、兩種 base 的文件與內嵌範例由 [本機參考庫](../../../../../docs/development/tooling/shadcn-reference/README.md) 的 [生成索引](../../../../../docs/development/tooling/shadcn-reference/snapshot/INDEX.md) 統一維護，不在本檔複製快照或數量。先搜尋索引，再按需讀相關 Radix 文件；未安裝、棄用及 base 不符都不能憑目錄推定可直接使用。

更新與完整性檢查依參考庫說明執行；最新官方目錄仍是發現新增／改名項目的來源。應用檢查、安裝來源與 MCP 診斷見 [shadcn 工具指南](../../../../../docs/development/tooling/mcp/shadcn.md) 與 [前端入口](../../../../README.md)。純指引修改驗證連結及 whitespace；實際安裝／組合再依前端指引檢查，不把文件查證當作 UI、LINE 裝置或整合驗收。

[現行實作圖：頁面與身分生命週期](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期) · [未來設計圖](../../../../../docs/diagrams.md#未來設計)
