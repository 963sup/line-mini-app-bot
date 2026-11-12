# Discussion 契約預備指引

適用於本目錄及其子目錄；繼承 [上層指引](../AGENTS.md)，並遵循其引用的根目錄規則。此處只補充本範圍責任；使用者與主機指令優先，不增加外部寫入授權。

## 職責與邊界

- 狀態：尚未建立正式 Discussion API；先讀 [業務模型設計](../../../docs/architecture/domain-model.md)，先完成對應功能規格。
- 實體用 Discussion、集合用 discussions；每個主題歸屬既有 Repository，DiscussionComment 的回覆關係由本能力定義。
- 建立、讀取、修改、分類、回覆、解答與鎖定按需求逐項決定，不把 GitHub 所有操作當成必建清單。
- 明定可見範圍、編輯權、回覆層級、生命週期、重送與併發；留言、Reaction 或標記解答不等同正式核准。
- 與 Issue 連結或轉換需另定授權、識別及追溯語義，不直接改寫或共用兩者狀態。
