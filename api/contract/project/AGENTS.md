# Project 契約指引

適用於本目錄及其子目錄；繼承 [上層指引](../AGENTS.md)，並遵循其引用的根目錄規則。此處只補充本範圍責任；使用者與主機指令優先，不增加外部寫入授權。

## 職責與邊界

- 狀態：Project／items 已依 [專案與來源引用](../../../docs/features/project/projects.md) 接入；fields／views 仍為雛形。完整管理模型以 [業務模型設計](../../../docs/foundation/architecture/domain-model.md) 作後續參考。
- 本專案名稱使用 Project、ProjectItem、ProjectField、ProjectView；GitHub ProjectV2 只作參考，不據此加入 V2 字尾。
- 個人或工程群組擁有 Project，每個 Project 有預設來源庫，並可連結多庫。references 支援物件、子物件、問題與討論。既有 items 保留問題引用的相容語意。
- 項目 ID 與來源 Issue ID 分開；欄位定義、選項與項目欄位值分開；來源事項狀態不等於專案 Status。
- Table、Board、Roadmap 共用項目資料，View 保存呈現與查詢設定；篩選或隱藏欄位不是授權。
- 不因雛型需求自動加入跨群組、草稿、排程、自動化、分析或範本；來源資源與專案權限分別核驗。
