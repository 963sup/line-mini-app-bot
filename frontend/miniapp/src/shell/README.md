# frontend/miniapp/src/shell 目錄導覽

## 用途與範圍

組合全站版面、側欄與服務導覽。

## 架構與責任邊界

維護共用版面與導覽；業務互動由功能模組提供。

## 閱讀與修改入口

先查閱[AppFrame.tsx](AppFrame.tsx)、[OrganizationSwitcher.tsx](OrganizationSwitcher.tsx)、[ServiceSidebar.tsx](ServiceSidebar.tsx)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：頁面與身分生命週期](../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

從 [AppFrame.tsx](AppFrame.tsx)、[ServiceSidebar.tsx](ServiceSidebar.tsx)、[Shell.tsx](Shell.tsx)、[repository-navigation.ts](repository-navigation.ts)、[sidebar-07](sidebar-07) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：頁面與身分生命週期](../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

[上層導覽](../README.md)
