# frontend/miniapp/src 工作指引

## 閱讀與修改入口

先查閱[business](business)、[diagnostics](diagnostics)、[identity](identity)、[navigation](navigation)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：頁面與身分生命週期](../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期) · [未來設計圖](../../../docs/diagrams.md#未來設計)

## 驗證入口

從儲存庫根目錄執行 `bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp check` 做型別檢查。修改流程或互動時，執行相同命令並以 `test` 取代 `check`，核對受影響的登入狀態、請求失敗或重試行為。需要 LINE WebView 的流程另在對應環境驗證。
