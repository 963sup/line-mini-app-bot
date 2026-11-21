# 空間標籤目錄

提供 Repository 標籤的列表與建立畫面。

在本功能就近維護畫面與互動，路由由 pages 組合，權限與結果以後端 API 為準。

## 閱讀與修改入口

先查閱[Picker.tsx](Picker.tsx)、[View.tsx](View.tsx)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[標籤規格](../../../../../../docs/domains/repository/labels.md) · [現行實作圖：空間載入時序](../../../../../../docs/domains/repository/README.md#空間載入時序)

僅提供列出及建立，操作、權限、資料與驗收依 [標籤規格](../../../../../../docs/domains/repository/labels.md)；不提供改名或刪除。分類引用由問題／討論自己的作者更新流程保存。

`Picker` 共用目錄查詢及選取 UI，只修改父層草稿；透過問題／討論既有更新 API 保存。

[現行實作圖：空間載入時序](../../../../../../docs/domains/repository/README.md#空間載入時序) · [未來設計圖](../../../../../../docs/diagrams.md#未來設計)

[上層導覽](../README.md)
