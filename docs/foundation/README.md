# 共用系統基礎

只讀任務涉及的共用能力。業務規則在 [features](../features/README.md)，未來平台設計在 [planning](../planning/README.md)。

| 分類 | 內容 | 何時閱讀 |
| --- | --- | --- |
| [architecture](architecture/README.md) | 現行架構、模組責任、依賴與已接受決策 | 定位程式、調整邊界時 |
| [identity](identity/README.md) | Session 與可信身分 | 修改登入、續期與認證時 |
| [api](api/README.md) | 共用 HTTP 與身分契約 | 修改入口、安全與錯誤語意時 |
| [frontend](frontend/README.md) | React 共用生命週期 | 修改前端非同步請求與頁面狀態時 |
| [persistence](persistence/README.md) | Firestore 整合與資料設計參考 | 修改儲存與交易時；先確認各頁狀態 |
| [integrations](integrations/README.md) | LINE 與瀏覽器能力整合 | 修改平台接入或定位時 |
| [reliability](reliability/README.md) | 事件、狀態機與可靠發送設計 | 涉及重試或持久工作時；不預設所有設計已實作 |
| [operations](operations/README.md) | Cloud Run、發布驗收與排錯 | 涉及部署與真實環境問題時 |

[文件總入口](../README.md)
