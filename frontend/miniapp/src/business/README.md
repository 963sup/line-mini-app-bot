# frontend/miniapp/src/business 目錄導覽

## 用途與範圍

遵循 **依功能組織（feature-based architecture）**。將同一功能的畫面、表單、狀態與型別化 API 呼叫就近維護，由 `pages/` 依路由組合。前端呈現輸入與操作結果，業務規則及權限由後端做最終判定。

## 架構與責任邊界

問題、討論、留言、物件、專案及里程碑的內容編輯控制讀取 API 的 `canEdit`。子物件表單使用父物件的編輯能力。問題與討論的分類操作另有授權語意。日誌的關聯選項由群組查詢提供，瀏覽器負責選取、分頁及取消過期請求。

## 閱讀與修改入口

先查閱[attendance](attendance)、[discussion](discussion)、[enterprise](enterprise)、[issue](issue)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：頁面與身分生命週期](../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

企業開通與管理見 [enterprise](enterprise/Enterprise.tsx)，組織團隊列表見 [team](team/OrganizationTeams.tsx)。

從 [attendance](attendance)、[discussion](discussion)、[issue](issue)、[journal](journal)、[object](object) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：頁面與身分生命週期](../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期) · [未來設計圖](../../../../docs/diagrams.md#未來設計)

[上層導覽](../README.md)

## 下層文件入口

- [project](project/README.md)
- [repository](repository/README.md)
- [subobject](subobject/README.md)
