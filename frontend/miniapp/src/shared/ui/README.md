# frontend/miniapp/src/shared/ui 目錄導覽

從 [LICENSE.md](LICENSE.md)、[alert.tsx](alert.tsx)、[avatar.tsx](avatar.tsx)、[badge.tsx](badge.tsx)、[breadcrumb.tsx](breadcrumb.tsx)、[button.tsx](button.tsx) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

## 已安裝元件

2026-11-19：已補齊固定版 shadcn CLI 的 `@shadcn` registry 全部 61 個 `registry:ui` 項目，另安裝官方 Radix 的 Code Block、Questionnaire、Trace，共 64 個 TSX 元件模組。後三項及 Code Block 的 hook／highlighter 由官方 `radix-rhea` registry 產生；專案仍維持 `new-york` 設定，既有元件與全域樣式未覆寫。Combobox 的官方實作使用 Base UI，其他元件不因此切換 base。

查閱實際已安裝清單：`bash scripts/tooling/tool-exec.sh shadcn info --json --cwd frontend/miniapp`。完整用途與 Radix 範例見[官方文件參考庫](../../../../../docs/development/tooling/shadcn-reference/README.md)，不要再次安裝整套或覆寫既有檔案。

Data Table、Date Picker、Typography 是組合／樣式配方，沿用已安裝的 Table、Calendar、Popover 等元件與官方範例；不額外建立同名業務框架。已棄用的 Toast 由 Sonner 替代。安裝不會自動更換業務畫面，也不代表所有互動已完成整合驗收。

[現行實作圖：頁面與身分生命週期](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期) · [未來設計圖](../../../../../docs/diagrams.md#未來設計)
