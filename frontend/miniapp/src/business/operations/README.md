# 文件工具

`Poppler.tsx` 提供 PDF 文字擷取測試，由 `/miniapp/poppler` 接入。沿用登入 Session 與產生的 API client；顯示載入、失敗、空文字及成功結果，分頁以標記顯示，下載保留原分頁符號。文件不保存，也不送至 AI。

驗證：從儲存庫根目錄執行 `pnpm --filter @repo/miniapp check`、`test`、`build`，均透過 `scripts/tooling/tool-exec.sh`。
