# frontend/miniapp/tests 工作指引

## 用途與範圍

保存 MINI App 的互動、請求生命週期與資料處理測試，以及測試用 DOM 設定。

## 架構與責任邊界

維護測試環境與可觀察的行為驗證，應用程式實作位於 src。

## 閱讀與修改入口

先查閱[dom-setup.ts](dom-setup.ts)、[session-controls.tsx](session-controls.tsx)、[sidebar-dom.ts](sidebar-dom.ts)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：頁面與身分生命週期](../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

遵循[上層指引](../AGENTS.md)。修改互動時，核對使用此模組的頁面與 miniapp 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：頁面與身分生命週期](../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期) · [未來設計圖](../../../docs/diagrams.md#未來設計)
