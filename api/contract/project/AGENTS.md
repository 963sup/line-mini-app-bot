# Project 契約預備指引

適用於本目錄及其子目錄；繼承 [上層指引](../AGENTS.md)，並遵循其引用的根目錄規則。此處只補充本範圍責任；使用者與主機指令優先，不增加外部寫入授權。

## 職責與邊界

- 狀態：尚未建立正式專案管理 API；先讀 [業務模型設計](../../../docs/architecture/domain-model.md)，現有 YAML 僅為雛形，建立功能規格後才新增操作與 schema。
- 本專案名稱使用 Project、ProjectItem、ProjectField、ProjectView；GitHub ProjectV2 只作參考，不據此加入 V2 字尾。
- 先決定 Project 擁有者、Repository 歸屬與跨容器引用規則；頂層 projects 路徑不表示專案沒有歸屬。items、fields、views 是 Project 下的資源集合，正式路由待契約確認。
- 項目 ID 與來源 Issue ID 分開；欄位定義、選項與項目欄位值分開；來源事項狀態不等於專案 Status。
- Table、Board、Roadmap 共用項目資料，View 保存呈現與查詢設定；篩選或隱藏欄位不是授權。
- 不因雛型需求自動加入跨群組、草稿、排程、自動化、分析或範本；來源資源與專案權限分別核驗。
