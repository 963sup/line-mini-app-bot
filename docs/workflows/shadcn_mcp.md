# 使用 shadcn MCP 實作 Mini App 元件

1. 在 `frontend/miniapp` 使用 `shadcn` MCP，呼叫 `get_project_registries`，核對專案的 registry。確認 [components.json](../../frontend/miniapp/components.json) 的樣式與 `ui` alias。
2. 在 [共用元件目錄](../../frontend/miniapp/src/shared/ui) 找到現有元件並檢查其 props。若缺少元件，使用 `search_items_in_registries` 搜尋，透過 `view_items_in_registries` 檢查來源與相依，再以 `get_add_command_for_items` 取得安裝指令。執行前核對會影響哪些本地檔案。
3. 在功能元件組合業務互動。例如候選匯入由功能元件管理勾選筆數、可匯入項目與重試；共用 Checkbox 呈現已選、未選和部分選取三種狀態。
4. 測試使用者可見的操作、鍵盤可用性、可存取名稱與狀態，再執行 `pnpm --filter @repo/miniapp check` 及受影響的測試。回報採用的元件、通過的檢查及尚未實測的互動。
