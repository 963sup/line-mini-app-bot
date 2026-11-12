# sidebar-07 專案整合

來源為官方 `@shadcn/sidebar-07`，透過專案固定版本 CLI 的 `add` 安裝，再搬入本目錄。
官方入口：https://ui.shadcn.com/blocks/sidebar 。版本與安裝方式由 mise 及專案 shadcn 指引管理。

保留 AppSidebar、NavMain、NavProjects、NavUser、TeamSwitcher、SidebarRail 與官方頁首的結構及互動；AppFrame 是唯一使用此結構的應用外框。

- NavMain：既有服務入口與可收合的出勤／示範空間子選單。
- NavProjects：改用實際工具入口與新分頁動作，不提供虛構的專案刪除或分享功能。
- TeamSwitcher：切換正式服務與示範預覽，不宣稱切換群組成員資格或授權。
- NavUser：接收 Shell 傳入的真實身分狀態、登入／驗證／登出動作；預覽與定位診斷不建立 LINE Session。

移除上游示例帳號、公司、訂閱、帳單及空白 dashboard；業務內容沿用既有路由。
既有 shared/ui primitives 不覆寫；只新增本 block 需要的 Breadcrumb、Collapsible。
