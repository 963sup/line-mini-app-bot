# frontend/miniapp/src/business/repository 目錄導覽

管理 Repository 與空間的前端選取、頁內導覽及功能畫面組合。

在本功能就近維護畫面與互動，路由由 pages 組合，權限與結果以後端 API 為準。

## 閱讀與修改入口

從 [NavigationGuard.tsx](NavigationGuard.tsx)、[RepositoryContext.tsx](RepositoryContext.tsx)、[Spaces.tsx](Spaces.tsx)、[View.tsx](View.tsx)、[labels](labels)、[milestones](milestones)、[navigation.ts](navigation.ts) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：空間載入時序](../../../../../docs/domains/repository/README.md#空間載入時序) · [未來設計圖](../../../../../docs/diagrams.md#未來設計)

[上層導覽](../README.md)

## 下層文件入口

- [insights](insights/README.md)

各功能分頁使用 `next/dynamic`，開啟時才載入對應功能程式，沿用原本的未儲存表單保護與載入回饋。
