# Discussion 契約指引

## 用途與範圍

定義 Repository 討論、分類與留言的 API 契約。

## 架構與責任邊界

維護本能力的操作與 schema；根規格負責組合，公開格式由契約產生。

## 閱讀與修改入口

先查閱[categories.yaml](categories.yaml)、[comments.yaml](comments.yaml)、[discussions.yaml](discussions.yaml)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[討論操作規格](../../../docs/domains/discussion/design.md) · [現行實作圖：最小實作邊界與驗證](../../../docs/domains/discussion/design.md#最小實作邊界與驗證)

適用於本目錄及其子目錄；繼承 [上層指引](../AGENTS.md)，並遵循其引用的根目錄規則。此處只補充本範圍責任；使用者與主機指令優先，不增加外部寫入授權。

## 職責與邊界

- 狀態：基礎主題與單層留言已接入；先讀 [討論操作規格](../../../docs/domains/discussion/design.md)。categories 仍為雛形，其他擴充另依功能規格。
- 實體用 Discussion、集合用 discussions；每個主題歸屬既有 Repository，DiscussionComment 的回覆關係由本能力定義。
- 建立、讀取、修改、分類、回覆、解答與鎖定按需求逐項決定，不把 GitHub 所有操作當成必建清單。
- 明定可見範圍、編輯權、回覆層級、生命週期、重送與併發；留言、Reaction 或標記解答不等同正式核准。
- 與 Issue 連結或轉換需另定授權、識別及追溯語義，不直接改寫或共用兩者狀態。

[現行實作圖：最小實作邊界與驗證](../../../docs/domains/discussion/design.md#最小實作邊界與驗證) · [未來設計圖](../../../docs/diagrams.md#未來設計)
