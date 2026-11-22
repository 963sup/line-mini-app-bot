# frontend/miniapp/src/business 目錄導覽

遵循 **依功能組織（feature-based architecture）**。將同一功能的畫面、表單、狀態與型別化 API 呼叫就近維護，由 `pages/` 依路由組合。前端呈現輸入與操作結果，業務規則及權限由後端做最終判定。

問題、討論、留言、物件、專案及里程碑的內容編輯控制讀取 API 的 `canEdit`。子物件表單使用父物件的編輯能力。問題與討論的分類操作另有授權語意。日誌的關聯選項由群組查詢提供，瀏覽器負責選取、分頁及取消過期請求。

## 閱讀與修改入口

從 [attendance](attendance)、[discussion](discussion)、[enterprise](enterprise)、[issue](issue)、[journal](journal)、[object](object) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 相關規格

[MINI App 首頁路由](../pages/miniapp/index.tsx)組合 [Organization 操作](organization/OrganizationPage.tsx)、[團隊清單](team/OrganizationTeams.tsx)及[企業開通與管理](enterprise/Enterprise.tsx)。Organization 和 Team 由路由傳入同一個組織 ID，各自讀取自己的 API。Enterprise 開通提示由路由傳入側欄取得的資格結果。這三個功能模組不互相匯入；各自負責請求與畫面狀態，後端判斷實際權限。

[現行實作圖：頁面與身分生命週期](../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

[上層導覽](../README.md)

## 下層文件入口

- [project](project/README.md)
- [repository](repository/README.md)
- [subobject](subobject/README.md)
