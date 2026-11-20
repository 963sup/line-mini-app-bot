# frontend/miniapp/src/diagnostics/geolocation 工作指引

## 用途與範圍

提供測試瀏覽器定位並呈現結果的診斷畫面。

## 架構與責任邊界

依層級維護本目錄的共用能力；功能畫面由 business 組織，頁面入口由 pages 接入。

## 閱讀與修改入口

先查閱[View.tsx](View.tsx)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：現有出勤整合檔案與責任](../../../../../docs/architecture/integrations/web-geolocation.md#現有出勤整合檔案與責任)

遵循[上層指引](../AGENTS.md)。修改互動時，核對使用此模組的頁面與 miniapp 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：現有出勤整合檔案與責任](../../../../../docs/architecture/integrations/web-geolocation.md#現有出勤整合檔案與責任) · [未來設計圖](../../../../../docs/diagrams.md#未來設計)
