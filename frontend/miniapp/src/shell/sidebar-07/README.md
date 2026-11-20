# sidebar-07 專案整合

## 用途與範圍

保存 shadcn sidebar-07 版型，供應用程式外框組合側欄。

## 架構與責任邊界

維護共用版面與導覽；業務互動由功能模組提供。

## 閱讀與修改入口

先查閱[app-sidebar.tsx](app-sidebar.tsx)、[nav-main.tsx](nav-main.tsx)、[nav-projects.tsx](nav-projects.tsx)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：頁面與身分生命週期](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

來源為官方 `@shadcn/sidebar-07`，透過專案固定版本 CLI 的 `add` 安裝，再搬入本目錄。
官方入口：https://ui.shadcn.com/blocks/sidebar 。版本與安裝方式由 mise 及專案 shadcn 指引管理。

保留 AppSidebar、NavMain、NavProjects、NavUser、TeamSwitcher、SidebarRail 與官方頁首的結構及互動；AppFrame 是唯一使用此結構的應用外框。

- NavMain：既有服務入口與可收合的出勤／正式空間子選單。
- NavProjects：改用實際工具入口與新分頁動作，不提供虛構的專案刪除或分享功能。
- TeamSwitcher：呈現正式服務入口，不宣稱切換群組成員資格或授權；定位診斷入口由 NavProjects 提供。
- NavUser：接收 Shell 傳入的真實身分狀態、登入／驗證／登出動作；定位診斷不建立 LINE Session。

移除上游示例帳號、公司、訂閱、帳單及空白 dashboard；業務內容沿用既有路由。
既有 shared/ui primitives 不覆寫；只新增本 block 需要的 Breadcrumb、Collapsible。

## 元件保護與官方比對

Codex 遵循[前端元件保護規則](../../../../AGENTS.md#shadcn-元件保護)，不得為頁面需求擅改此 block 的外觀或互動。資料、真實路由與身分動作屬必要整合，不能還原成官方示例的帳號、帳單及假連結。

本次以專案 CLI 4.23.0 的 new-york registry 比對全部 24 個已安裝 primitives；除 alias、client directive 外，Button 改回官方 radix-ui Slot.Root；Progress 保留 Root value 的可及性修正。sidebar.tsx 本體與上游一致。NavUser 選單標頭恢復官方的頭像及雙行排版，保留真實登入操作。Registry 是可更新來源，CLI 版本不等於 registry 內容版本；未宣稱像素或手機實機驗收。

LINE 頭像經伺服器身分驗證成功後，以 LIFF `getProfile()` 取得，僅存在前端記憶體；登出、重新驗證及 pagehide 時清除，過期回應不更新畫面。不把 profile 回傳後端或當作授權依據。頭像限官方 `profile.line-scdn.net` HTTPS 來源；需 LINE Console 啟用並取得 `profile` scope 授權，失敗或未設定圖片時使用 AvatarFallback。側欄及帳號選單均使用 AvatarImage，未修改共用 Avatar 元件。

來源：LINE 官方 docs source main `dea09dc6e63fc5c6a3754abfbb304fc6fce4cc54` 的 LIFF getProfile 參考，以及 Context7 `/shadcn-ui/ui` AvatarImage／AvatarFallback 文件；shadcn CLI 維持 4.23.0。

[現行實作圖：頁面與身分生命週期](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期) · [未來設計圖](../../../../../docs/diagrams.md#未來設計)

[上層導覽](../README.md)
