# 儲存庫洞察（Insights）

## 用途與範圍

呈現所選 Repository 的統計資料。

## 架構與責任邊界

在本功能就近維護畫面與互動，路由由 pages 組合，權限與結果以後端 API 為準。

## 閱讀與修改入口

先查閱[View.tsx](View.tsx)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[基礎規格](../../../../../../docs/domains/repository/insights.md) · [現行實作圖：空間載入時序](../../../../../../docs/domains/repository/README.md#空間載入時序)

操作、權限、資料與驗收依 [基礎規格](../../../../../../docs/domains/repository/insights.md)。不擴充未授權功能；核心、HTTP 與 Firestore 各自保留責任，前端組合既有元件。

[現行實作圖：空間載入時序](../../../../../../docs/domains/repository/README.md#空間載入時序) · [未來設計圖](../../../../../../docs/diagrams.md#未來設計)
