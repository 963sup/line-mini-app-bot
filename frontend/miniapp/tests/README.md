# frontend/miniapp/tests 目錄導覽

## 用途與範圍

保存 MINI App 的互動、請求生命週期與資料處理測試，以及測試用 DOM 設定。

## 架構與責任邊界

維護測試環境與可觀察的行為驗證，應用程式實作位於 src。

## 閱讀與修改入口

先查閱[dom-setup.ts](dom-setup.ts)、[session-controls.tsx](session-controls.tsx)、[sidebar-dom.ts](sidebar-dom.ts)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：頁面與身分生命週期](../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

從 [dom-setup.ts](dom-setup.ts)、[lifecycle.test.tsx](lifecycle.test.tsx)、[object-import.test.tsx](object-import.test.tsx)、[object-spaces.test.tsx](object-spaces.test.tsx)、[objects.test.tsx](objects.test.tsx) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：頁面與身分生命週期](../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期) · [未來設計圖](../../../docs/diagrams.md#未來設計)

[上層導覽](../README.md)
