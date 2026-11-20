# Issue 契約指引

## 用途與範圍

定義 Repository 事項、事項類型與留言的 API 契約。

## 架構與責任邊界

維護本能力的操作與 schema；根規格負責組合，公開格式由契約產生。

## 閱讀與修改入口

先查閱[comments.yaml](comments.yaml)、[issues.yaml](issues.yaml)、[types.yaml](types.yaml)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[事項操作規格](../../../docs/domains/repository/issues.md) · [業務模型設計](../../../docs/proposals/platform/domain-model.md)

適用於本目錄及其子目錄；繼承 [上層指引](../AGENTS.md)，並遵循其引用的根目錄規則。此處只補充本範圍責任；使用者與主機指令優先，不增加外部寫入授權。

## 職責與邊界

- 狀態：基礎事項 API 已接入；先讀 [事項操作規格](../../../docs/domains/repository/issues.md)。types／comments 仍為雛形，[業務模型設計](../../../docs/proposals/platform/domain-model.md) 不取代已確認規格。
- 實體用 Issue、集合用 issues；每個 Issue 歸屬既有 Repository。正式路由、欄位、識別碼與保存位置須在功能規格確認。
- 建立契約前先建立 Issue 功能規格，定義建立、查詢、修改與生命週期的操作者、權限、狀態轉換、版本、重送及必要稽核。
- Issue 的關閉不自動代表施工驗收或業務核准；不沿用出勤補卡的狀態或審核權。
- IssueComment 與事項關聯歸本能力；Label、Milestone 等引用歸其擁有者，不在多個能力複製定義。未確認的操作不預建空端點。

[現行實作圖：基礎入口](../../../docs/architecture/contracts/http.md#基礎入口) · [未來設計圖](../../../docs/diagrams.md#未來設計)
