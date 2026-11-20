# frontend/miniapp/src/shared/ui 目錄導覽

## 目錄用途

保存 shadcn 的共用 UI 元件，對應 `@/shared/ui`。業務專用的畫面組合放在各功能模組。

從 [LICENSE.md](LICENSE.md)、[alert.tsx](alert.tsx)、[avatar.tsx](avatar.tsx)、[badge.tsx](badge.tsx)、[breadcrumb.tsx](breadcrumb.tsx)、[button.tsx](button.tsx) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

## 已安裝元件

2026-11-19：已補齊固定版 shadcn CLI 的 `@shadcn` registry 全部 61 個 `registry:ui` 項目，另安裝官方 Radix 的 Code Block、Questionnaire、Trace，共 64 個 TSX 元件模組。後三項及 Code Block 的 hook／highlighter 由官方 `radix-rhea` registry 產生；專案仍維持 `new-york` 設定，既有元件與全域樣式未覆寫。Combobox 的官方實作使用 Base UI，其他元件不因此切換 base。

查閱實際已安裝清單：`bash scripts/tooling/tool-exec.sh shadcn info --json --cwd frontend/miniapp`。完整用途與 Radix 範例見[官方文件參考庫](../../../../../docs/development/tooling/shadcn-reference/README.md)，不要再次安裝整套或覆寫既有檔案。

Data Table、Date Picker、Typography 是組合／樣式配方，沿用已安裝的 Table、Calendar、Popover 等元件與官方範例；不額外建立同名業務框架。已棄用的 Toast 由 Sonner 替代。安裝不會自動更換業務畫面，也不代表所有互動已完成整合驗收。

[現行實作圖：頁面與身分生命週期](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期) · [未來設計圖](../../../../../docs/diagrams.md#未來設計)

## 前端使用位置

以下清單依 `src` 的靜態 import 分類。直接引用是指元件由 `shared/ui` 外的模組匯入；間接引用則沿這些模組使用的 UI 元件繼續追蹤。是否在畫面呈現取決於路由、資料與互動狀態。

| 主要用途 | 直接引用的 30 個元件模組 |
|---|---|
| 導覽與外框 | accordion、avatar、breadcrumb、collapsible、dropdown-menu、separator、sidebar、tabs |
| 表單與操作 | button、button-group、checkbox、combobox、field、input、input-group、native-select、questionnaire、radio-group、textarea、toggle-group |
| 資料與回饋 | alert、alert-dialog、badge、card、chart、empty、item、progress、spinner、table |

| 間接引用 | 入口與實際用途 |
|---|---|
| label | field 的 FieldLabel 使用 Label，業務表單經 FieldLabel 或 FormField 取得標籤語意。 |
| sheet | sidebar 在手機寬度以 Sheet 顯示側欄，含標題、說明與焦點管理。 |
| tooltip | sidebar 提供 TooltipProvider；有 tooltip 設定的選單按鈕可在圖示收折模式提示名稱。 |
| skeleton | sidebar 的 SidebarMenuSkeleton 使用 Skeleton。此路徑代表原始碼依賴，只有渲染 SidebarMenuSkeleton 才會出現佔位內容。 |
| toggle | toggle-group 共用 toggle 的 variants；這是樣式依賴，不代表另外渲染 Toggle 控制項。 |

上述 30 個直接引用與 5 個間接依賴，共 35 個入口可達模組；不是畫面上的元件實例數。其餘 29 個模組目前未從上述入口引用，仍屬已安裝的官方目錄：

- 表單、日期、分頁與浮層：calendar、dialog、drawer、form、pagination、popover、select、slider、sonner、switch。
- 導覽及進階互動：command、context-menu、hover-card、input-otp、kbd、menubar、navigation-menu、resizable、scroll-area。
- 媒體、訊息與追蹤：aspect-ratio、attachment、bubble、carousel、code-block、direction、marker、message、message-scroller、trace。

異動引用後可用 `rg -n 'from .*ui/' frontend/miniapp/src` 查找畫面與共用入口，再查元件本身的 import（含相對路徑）。更換控制項時核對操作語意：工作關聯在完整清單內搜尋，來源儲存庫仍採游標載入；Issue 狀態在套用篩選後查詢；儲存庫頁籤在使用者確認切換後載入。

表單群組與標籤使用 FieldSet／FieldLegend／FieldLabel（含 FormField 包裝），載入與訊息沿用 ReadFeedback／Status，適用的可展開明細使用 Collapsible。日期、月份、數字、檔案輸入仍保留原生型別與驗證；表單、文字、連結、資料列表、圖表圖例及診斷結果的語意 HTML 不為增加元件數而替換。未登入頁面的說明可保留文字，避免與 Shell 的身分提示重複。
