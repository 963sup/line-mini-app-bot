# frontend/miniapp/src 目錄導覽

MINI App 採用 **依功能組織（feature-based architecture）**。`pages/` 負責路由與頁面組合，`business/` 負責各業務功能的畫面與互動，`shell/` 負責共用版面與導覽，`shared/` 提供跨功能元件與工具。 **功能內部按需拆分。** 簡單畫面可集中在一個元件，表單、狀態或請求邏輯有獨立責任時再拆檔。

依層級維護本目錄的共用能力；功能畫面由 business 組織，頁面入口由 pages 接入。

## 閱讀與修改入口

從 [business](business)、[diagnostics](diagnostics)、[identity](identity)、[navigation](navigation)、[pages](pages)、[platform](platform) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：頁面與身分生命週期](../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期) · [未來設計圖](../../../docs/diagrams.md#未來設計)

## 驗證入口

從儲存庫根目錄執行 `bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp check` 做型別檢查。修改流程或互動時，執行相同命令並以 `test` 取代 `check`，核對受影響的登入狀態、請求失敗或重試行為。需要 LINE WebView 的流程另在對應環境驗證。

[上層導覽](../README.md)

## 下層文件入口

- [styles](styles/README.md)
