# 規劃與提案

這裡保存待評估的業務能力、平台目標與歷史交付設計。確認「今天做得到什麼」時，從[領域狀態](../domains/README.md#domain-matrix)、[實作位置](../architecture/runtime/code-layout.md#從需求定位修改範圍)及[已採用的 ADR](../decisions/README.md)查證。

## 按問題查閱

| 設計問題 | 提案或對照 |
| --- | --- |
| 持久 Inbox、Outbox、Worker 與管理後台如何分工？ | [平台提案](platform/README.md) · [套件配置](platform/package-layout.md) · [Webhook 可靠交付](webhook-delivery/README.md) |
| LINE 與跨業務能力的目標架構是什麼？ | [工程平台設計](line-engineering/README.md) |
| Repository、Issue 與 Discussion 的目標模型有何差異？ | [平台資源模型](platform/domain-model.md) |
| 如何看待早期平台切片的交付順序？ | [歷史交付階段](platform/delivery.md)；目前變更以[驗收矩陣](../development/acceptance.md)選擇檢查。 |

## 相鄰的設計參考

[Firestore 資料契約](firestore-data-contracts.md)描述規劃中的技術基線。[可靠性設計](../architecture/README.md#可靠性設計參考)連到通用 Inbox／Outbox 模型與現行接收契約。

## 維護邊界

每項提案標示狀態與現行契約來源。能力獲採用後，把生效規則寫入領域規格或 ADR，保留必要的提案背景與引用。變更公開契約、正式資料或部署仍依各自的流程。

圖表分為[未來設計](../diagrams.md#未來設計)與[現行實作](../diagrams.md#現行實作)。

## 規劃中的業務領域

[Membership 統一成員關係設計](membership.md)：使用者與資源的成員、邀請、角色及撤權生命週期。

[Operations 設計](operations.md) · [Workflow 設計](workflow.md) · [Agent 設計](agent.md) · [Knowledge 設計](knowledge.md)

[四領域架構、共用詞條與落地驗收](domain-collaboration.md)

[適用工作指引](../AGENTS.md) · [上層導覽](../README.md)
