# frontend/miniapp/src/business/repository 目錄導覽

## 用途與範圍

管理 Repository 與空間的前端選取、頁內導覽及功能畫面組合。

## 架構與責任邊界

在本功能就近維護畫面與互動，路由由 pages 組合，權限與結果以後端 API 為準。

## 閱讀與修改入口

先查閱[NavigationGuard.tsx](NavigationGuard.tsx)、[RepositoryContext.tsx](RepositoryContext.tsx)、[Spaces.tsx](Spaces.tsx)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：空間載入時序](../../../../../docs/domains/repository/README.md#空間載入時序)

從 [NavigationGuard.tsx](NavigationGuard.tsx)、[RepositoryContext.tsx](RepositoryContext.tsx)、[View.tsx](View.tsx)、[labels](labels)、[milestones](milestones)、[navigation.ts](navigation.ts) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：空間載入時序](../../../../../docs/domains/repository/README.md#空間載入時序) · [未來設計圖](../../../../../docs/diagrams.md#未來設計)
