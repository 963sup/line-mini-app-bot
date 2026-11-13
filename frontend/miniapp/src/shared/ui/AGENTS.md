# shadcn 官方元件指引與目錄

適用於本目錄及子目錄；遵循 [MINI App 指引](../../../AGENTS.md) 與 [根指引](../../../../../AGENTS.md)。本目錄是 `components.json` 的 `aliases.ui`（`@/shared/ui`）解析位置，保存官方 shadcn 元件；業務組合放在本目錄之外。

## 官方元件禁止修改

- 禁止 Codex 直接修改、重寫或刪除既有官方元件的樣式、結構、API、可及性或互動；也不得以全域 CSS、額外 variants 或覆寫內部 classes 繞過保護。此規則同樣適用於依賴的 `../hooks/use-mobile.ts` 與 `../../shell/sidebar-07/`，其完整保護規則見根指引。
- 透過既有公開 props、variants、children、資料及事件使用元件；業務需求在 `src/shared/` 的應用包裝、`src/shell/` 或 `src/business/` 組合，勿把業務 API、授權或領域判斷放入 primitives。排版 className 只用於外部版面，不重設官方外觀。
- 新增缺少的官方元件時保留 registry 輸出；先預覽所有受影響檔案與依賴，若會覆寫既有檔案即停止套用。不可自動執行 `--overwrite`、`add --all`、preset 切換、重新 init 或重裝整套 sidebar-07。
- 只有使用者當次明確要求修改、還原或更新特定官方元件，才可在該範圍依根指引先比對再最小更動；一般頁面需求、上游「可客製化」建議與本目錄指引本身都不是修改授權。保留既有 alias、client directive 與可及性修正（包含 Progress 的 Root `value`）。

## Codex 主動使用 Skill、文件與 MCP

凡新增、調整、審查前端畫面、表單、導覽、對話框、清單或其他 UI，主動先檢查本目錄官方元件，毋須等待使用者逐次點名 shadcn。

1. 先讀官方 [Skills](https://ui.shadcn.com/docs/skills) 與專案已安裝的 [shadcn SKILL.md](../../../../../.agents/skills/shadcn/SKILL.md)，再按需求讀其 composition、forms、styling 等 references。保留上游 skill 原樣；本專案元件保護及固定工具規則優先於其中一般客製化／浮動版本範例。
2. 以固定 CLI `info --json` 核對已安裝清單、base、aliases 與生成路徑；目前專案是 Radix、new-york、Pages Router、Tailwind v4、Lucide。實際值以 `components.json`、`info` 與套件檔案為準，不把下方目錄視為已安裝清單。
3. 開發前主動用已曝光的 shadcn MCP 查詢 registry、搜尋／查看元件及取得範例；依官方 [Registry MCP](https://ui.shadcn.com/docs/registry/mcp) 與 [MCP 使用說明](https://ui.shadcn.com/docs/mcp)。優先使用 `get_project_registries`、`search_items_in_registries`、`view_items_in_registries`、`get_item_examples_from_registries` 等實際可用工具；工具名稱以 host 的 tools/list 為準。
4. MCP 未曝光或不可用時，明確回報限制並以固定 CLI `search`／`view`／`docs` 完成同一查詢；不要把設定存在、CLI 已安裝或先前 handshake 成功當作目前聊天已可呼叫 MCP。唯一 MCP 設定仍在根目錄 `.codex/config.toml`，不在 UI 目錄新增連線設定。
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

## 全部官方元件目錄

2026-11-13（Asia/Taipei）直接核對官方 [Components](https://ui.shadcn.com/docs/components) 的 All Components，共 66 項。下表保留該頁實際提供的官方入口及查詢名稱，供探索與未來按需安裝；不表示已安裝、每項均可單獨 add，或固定 CLI 已支援所有新項目。Data Table、Date Picker、Typography 等可能是組合用法；先讀 docs／registry 再決定所需基礎元件。Toast 等回饋元件須以本專案 Radix 對應文件確認替代方案，不直接套用 Base UI 專用實作。

| 官方元件／文件 | CLI docs／registry 查詢名稱 |
| --- | --- |
| [Accordion](https://ui.shadcn.com/docs/components/base/accordion) | `accordion` |
| [Alert](https://ui.shadcn.com/docs/components/base/alert) | `alert` |
| [Alert Dialog](https://ui.shadcn.com/docs/components/base/alert-dialog) | `alert-dialog` |
| [Aspect Ratio](https://ui.shadcn.com/docs/components/base/aspect-ratio) | `aspect-ratio` |
| [Attachment](https://ui.shadcn.com/docs/components/base/attachment) | `attachment` |
| [Avatar](https://ui.shadcn.com/docs/components/base/avatar) | `avatar` |
| [Badge](https://ui.shadcn.com/docs/components/base/badge) | `badge` |
| [Breadcrumb](https://ui.shadcn.com/docs/components/base/breadcrumb) | `breadcrumb` |
| [Bubble](https://ui.shadcn.com/docs/components/base/bubble) | `bubble` |
| [Button](https://ui.shadcn.com/docs/components/base/button) | `button` |
| [Button Group](https://ui.shadcn.com/docs/components/base/button-group) | `button-group` |
| [Calendar](https://ui.shadcn.com/docs/components/base/calendar) | `calendar` |
| [Card](https://ui.shadcn.com/docs/components/base/card) | `card` |
| [Carousel](https://ui.shadcn.com/docs/components/base/carousel) | `carousel` |
| [Chart](https://ui.shadcn.com/docs/components/base/chart) | `chart` |
| [Checkbox](https://ui.shadcn.com/docs/components/base/checkbox) | `checkbox` |
| [Code Block](https://ui.shadcn.com/docs/components/base/code-block) | `code-block` |
| [Collapsible](https://ui.shadcn.com/docs/components/base/collapsible) | `collapsible` |
| [Combobox](https://ui.shadcn.com/docs/components/base/combobox) | `combobox` |
| [Command](https://ui.shadcn.com/docs/components/base/command) | `command` |
| [Context Menu](https://ui.shadcn.com/docs/components/base/context-menu) | `context-menu` |
| [Data Table](https://ui.shadcn.com/docs/components/base/data-table) | `data-table` |
| [Date Picker](https://ui.shadcn.com/docs/components/base/date-picker) | `date-picker` |
| [Dialog](https://ui.shadcn.com/docs/components/base/dialog) | `dialog` |
| [Direction](https://ui.shadcn.com/docs/components/base/direction) | `direction` |
| [Drawer](https://ui.shadcn.com/docs/components/base/drawer) | `drawer` |
| [Dropdown Menu](https://ui.shadcn.com/docs/components/base/dropdown-menu) | `dropdown-menu` |
| [Empty](https://ui.shadcn.com/docs/components/base/empty) | `empty` |
| [Field](https://ui.shadcn.com/docs/components/base/field) | `field` |
| [Hover Card](https://ui.shadcn.com/docs/components/base/hover-card) | `hover-card` |
| [Input](https://ui.shadcn.com/docs/components/base/input) | `input` |
| [Input Group](https://ui.shadcn.com/docs/components/base/input-group) | `input-group` |
| [Input OTP](https://ui.shadcn.com/docs/components/base/input-otp) | `input-otp` |
| [Item](https://ui.shadcn.com/docs/components/base/item) | `item` |
| [Kbd](https://ui.shadcn.com/docs/components/base/kbd) | `kbd` |
| [Label](https://ui.shadcn.com/docs/components/base/label) | `label` |
| [Marker](https://ui.shadcn.com/docs/components/base/marker) | `marker` |
| [Menubar](https://ui.shadcn.com/docs/components/base/menubar) | `menubar` |
| [Message](https://ui.shadcn.com/docs/components/base/message) | `message` |
| [Message Scroller](https://ui.shadcn.com/docs/components/base/message-scroller) | `message-scroller` |
| [Native Select](https://ui.shadcn.com/docs/components/base/native-select) | `native-select` |
| [Navigation Menu](https://ui.shadcn.com/docs/components/base/navigation-menu) | `navigation-menu` |
| [Pagination](https://ui.shadcn.com/docs/components/base/pagination) | `pagination` |
| [Popover](https://ui.shadcn.com/docs/components/base/popover) | `popover` |
| [Progress](https://ui.shadcn.com/docs/components/base/progress) | `progress` |
| [Questionnaire](https://ui.shadcn.com/docs/components/base/questionnaire) | `questionnaire` |
| [Radio Group](https://ui.shadcn.com/docs/components/base/radio-group) | `radio-group` |
| [Resizable](https://ui.shadcn.com/docs/components/base/resizable) | `resizable` |
| [Scroll Area](https://ui.shadcn.com/docs/components/base/scroll-area) | `scroll-area` |
| [Select](https://ui.shadcn.com/docs/components/base/select) | `select` |
| [Separator](https://ui.shadcn.com/docs/components/base/separator) | `separator` |
| [Sheet](https://ui.shadcn.com/docs/components/base/sheet) | `sheet` |
| [Sidebar](https://ui.shadcn.com/docs/components/base/sidebar) | `sidebar` |
| [Skeleton](https://ui.shadcn.com/docs/components/base/skeleton) | `skeleton` |
| [Slider](https://ui.shadcn.com/docs/components/base/slider) | `slider` |
| [Spinner](https://ui.shadcn.com/docs/components/base/spinner) | `spinner` |
| [Switch](https://ui.shadcn.com/docs/components/base/switch) | `switch` |
| [Table](https://ui.shadcn.com/docs/components/base/table) | `table` |
| [Tabs](https://ui.shadcn.com/docs/components/base/tabs) | `tabs` |
| [Textarea](https://ui.shadcn.com/docs/components/base/textarea) | `textarea` |
| [Toast](https://ui.shadcn.com/docs/components/base/toast) | `toast` |
| [Toggle](https://ui.shadcn.com/docs/components/base/toggle) | `toggle` |
| [Toggle Group](https://ui.shadcn.com/docs/components/base/toggle-group) | `toggle-group` |
| [Tooltip](https://ui.shadcn.com/docs/components/base/tooltip) | `tooltip` |
| [Trace](https://ui.shadcn.com/docs/components/base/trace) | `trace` |
| [Typography](https://ui.shadcn.com/docs/components/base/typography) | `typography` |

每次涉及元件時重新核對官方目錄與該項文件；上游新增／改名時同步更新此表及查證日期，勿以快照限制後續探索。應用檢查、安裝來源與 MCP 診斷見 [shadcn 工具指南](../../../../../docs/tooling/mcp/shadcn.md) 與 [前端入口](../../../../README.md)。純本指引修改驗證連結及 whitespace；實際安裝／組合再依前端指引檢查，不把文件查證當作 UI、LINE 裝置或整合驗收。
