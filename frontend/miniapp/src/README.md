# frontend/miniapp/src 目錄導覽

## 目錄用途

MINI App 採用 **依功能組織（feature-based architecture）**。`pages/` 負責路由與頁面組合，`business/` 負責各業務功能的畫面與互動，`shell/` 負責共用版面與導覽，`shared/` 提供跨功能元件與工具。

從 [business](business)、[diagnostics](diagnostics)、[identity](identity)、[navigation](navigation)、[pages](pages)、[platform](platform) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：頁面與身分生命週期](../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期) · [未來設計圖](../../../docs/diagrams.md#未來設計)
