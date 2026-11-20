# frontend/miniapp/src/shared/http 工作指引

## 用途與範圍

提供共用的 HTTP 請求生命週期處理。

## 架構與責任邊界

提供跨功能重用的元件或工具；特定業務流程留在使用它的功能模組。

## 閱讀與修改入口

先查閱[lifecycle.ts](lifecycle.ts)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：頁面與身分生命週期](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

遵循[上層指引](../AGENTS.md)。修改互動時，核對使用此模組的頁面與 miniapp 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：頁面與身分生命週期](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期) · [未來設計圖](../../../../../docs/diagrams.md#未來設計)
