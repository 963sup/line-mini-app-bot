# frontend/miniapp/src/business/attendance 工作指引

## 用途與範圍

組合打卡、補卡、請假、審核與紀錄等出勤畫面。

## 架構與責任邊界

在本功能就近維護畫面與互動，路由由 pages 組合，權限與結果以後端 API 為準。

## 閱讀與修改入口

先查閱[launchToken.ts](launchToken.ts)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：定位前預檢、狀態與時序](../../../../../docs/domains/attendance/README.md#定位前預檢狀態與時序)

遵循[上層指引](../AGENTS.md)。修改互動時，核對使用此模組的頁面與 miniapp 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：定位前預檢、狀態與時序](../../../../../docs/domains/attendance/README.md#定位前預檢狀態與時序) · [未來設計圖](../../../../../docs/diagrams.md#未來設計)
