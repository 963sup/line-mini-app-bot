# docs 文件維護指引

本指引適用於本目錄及其子目錄。先讀[根層代理指引](../AGENTS.md)與[文件規範](standards/documentation.md)。更深層的 AGENTS.md 僅在其子樹內優先適用。共通規則回到權威文件更新，本頁只補充 docs 的執行要求。

## 撰寫與歸檔

- 依文件規範的權威來源表選定目錄。規格描述必須滿足的條件，設計文件解釋設計與取捨，教學在 workflows 以可重現步驟引導讀者，查詢型資料放在 references 或其專屬契約位置。
- 命名、目錄、標題層級、帶語言標記的圍欄程式碼區塊，以及 OpenAPI 與 schema 的引用，依文件規範辦理。領域術語與所屬 bounded context 的模型及契約一致。
- 採 CommonMark 與 GitHub Flavored Markdown（GFM），站內引用優先使用相對連結。可用文字描述的流程、時序與關係圖採用帶有 mermaid 語言標記的圍欄區塊。

## 版本與審查

- 文件隨實作以 Git 追蹤，重大架構取捨依[決策紀錄索引](decisions/README.md)新增 ADR。
- 需要獨立追蹤的規格或設計文件使用 YAML front matter，以 status 記錄 draft、active 或 deprecated，以 owner 記錄負責的團隊或角色。既有 ADR 依其決策狀態管理。
- 所屬模組維護者負責審查文件與來源是否一致。草案經審查後轉為 active，停用時標記 deprecated 並連到替代文件。契約、行為或責任變更時，在同一個 PR 更新文件並請相關維護者審查。
- 提交前以專案的 Prettier 與 markdownlint 設定檢查 Markdown，並驗證內部連結的檔案與標題錨點。PR 的 CI 應自動執行這三項檢查；檢查缺失或無法執行時，在驗證結果中明列。
