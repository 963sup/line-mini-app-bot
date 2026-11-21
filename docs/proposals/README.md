# 規劃與提案

[平台能力提案](platform/README.md) · [Webhook 可靠交付提案](webhook-delivery/README.md)

## 用途與範圍

保存待評估或演進中的設計提案。

## 架構與責任邊界

提案描述目標與取捨，現行行為以程式、契約與領域規格為準。

## 閱讀與修改入口

先查閱[line-engineering](line-engineering/README.md)。

## 相關規格

[本層規格與指引](../architecture/README.md)


<a id="規劃與提案"></a>

```mermaid
flowchart LR
  h["規劃與提案"]
  h --- n0["本目錄保存未來能力、目標模型與歷史平台交付設<br/>計。這些文件不表示功能已實作，也不構成目前任<br/>務的必要前置條件。"]
  h --- n1["現行功能先讀功能索引，實作位置見需求導航，已<br/>接受決策見 ADR。提案中已落地的部分以所連結的<br/>現行契約為準。"]
  n1 --- n1l0["功能索引"]
  click n1l0 "../domains/README.md"
  n1 --- n1l1["需求導航"]
  click n1l1 "../architecture/runtime/code-layout.md#從需求定位修改範圍"
  n1 --- n1l2["ADR"]
  click n1l2 "../architecture/README.md#architecture-decision-records"
```

<a id="按問題查閱"></a>

```mermaid
flowchart LR
  h["按問題查閱"]
  h --- n0["文件：未來平台目錄<br/>用途：評估持久 Inbox、Outbox、Worker 與管理<br/>後台的責任配置"]
  click n0 "platform/package-layout.md"
  h --- n1["文件：工程平台目標架構<br/>用途：查閱 LINE 平台與跨業務能力的提案模型"]
  click n1 "line-engineering/README.md"
  h --- n2["文件：Repository、Issue 與 Discussion 模型<br/>用途：比較後續資源模型與產品設計參考"]
  click n2 "platform/domain-model.md"
  h --- n3["文件：平台交付階段<br/>用途：查閱舊平台切片與未來驗收方向；目前檢查<br/>依驗收矩陣"]
  n3 --- n3l0["平台交付階段"]
  click n3l0 "platform/delivery.md"
  n3 --- n3l1["驗收矩陣"]
  click n3l1 "../development/acceptance.md"
```

<a id="相鄰的設計參考"></a>

```mermaid
flowchart LR
  h["相鄰的設計參考"]
  h --- n0["資料契約提案：尚未實作的技術基線。"]
  click n0 "firestore-data-contracts.md"
  h --- n1["可靠性設計參考：通用 Inbox／Outbox 與狀態機<br/>，與現行功能區分。"]
  click n1 "../architecture/README.md#可靠性設計參考"
```

<a id="維護邊界"></a>

```mermaid
flowchart LR
  h["維護邊界"]
  h --- n0["提案須標示狀態與現行契約來源。能力被接受後，<br/>將有效規則落在功能規格或 ADR，保留必要的提案<br/>背景與引用；不以提案直接修改公開契約、資料或<br/>部署。"]
  h --- n1["文件總入口 · 文件維護指引"]
  n1 --- n1l0["文件總入口"]
  click n1l0 "../README.md"
  n1 --- n1l1["文件維護指引"]
  click n1l1 "../AGENTS.md"
  h --- n2["未來設計圖索引 · 對照現行實作。"]
  n2 --- n2l0["未來設計圖索引"]
  click n2l0 "../diagrams.md#未來設計"
  n2 --- n2l1["對照現行實作"]
  click n2l1 "../diagrams.md#現行實作"
```

## 規劃中的業務領域

[Membership 統一成員關係設計](membership.md)：使用者與資源的成員、邀請、角色及撤權生命週期。

[Operations 設計](operations.md) · [Workflow 設計](workflow.md) · [Agent 設計](agent.md) · [Knowledge 設計](knowledge.md)

[四領域架構、共用詞條與落地驗收](domain-collaboration.md)

[適用工作指引](../AGENTS.md) · [上層導覽](../README.md)
