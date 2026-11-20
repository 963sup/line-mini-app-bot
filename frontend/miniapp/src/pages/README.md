# frontend/miniapp/src/pages 目錄導覽

## 用途與範圍

保存 Next.js Pages Router 的頁面入口及全站初始化，從各功能模組組合畫面。

## 架構與責任邊界

解析路由並組合功能畫面；業務狀態與請求邏輯由對應功能模組維護。

## 閱讀與修改入口

先查閱[_app.tsx](_app.tsx)、[_document.tsx](_document.tsx)、[index.tsx](index.tsx)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：頁面與身分生命週期](../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

從 [_app.tsx](_app.tsx)、[_document.tsx](_document.tsx)、[index.tsx](index.tsx)、[location.tsx](location.tsx)、[miniapp](miniapp) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：頁面與身分生命週期](../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
