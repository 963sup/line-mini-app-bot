# 業務功能

<a id="domain-matrix"></a>

## 領域與容器的追蹤矩陣

頂層依獨立領域與容器組織；Repository 內的 Issue、Discussion、Object（含子物件）由[儲存庫能力索引](repository/README.md#repository-capabilities)集中導覽。文件歸組不改變各能力的授權、狀態或程式模組責任。

沿用六邊形架構：從入口適配器追到使用案例或既有實作，再核對輸出介面、儲存適配器及業務判定。各列連到代表路徑；實作連結與提案驗收均不代表已執行測試。

| 領域入口 | 狀態 | 核心圖與契約 | 實作、測試或驗收 |
|---|---|---|---|
| [agent](agent/README.md) | 提案 | [核心流程](../proposals/agent.md#design-diagram) | [驗收條件](../proposals/domain-collaboration.md) |
| [attendance](attendance/README.md) | 已有實作 | [核心流程](attendance/README.md#decision-flow) | [測試](../../backend/internal/business/attendance/adapter/firestore/attendance_integration_test.go) |
| [enterprise](enterprise/README.md) | 已有實作 | [核心流程](enterprise/README.md#decision-flow) | [實作](../../backend/internal/business/enterprise/adapter/firestore/store.go) |
| [group](group/README.md) | 已有實作 | [核心流程](group/README.md#decision-flow) | [測試](../../backend/internal/business/group/application/query/group_test.go) |
| [journal](journal/README.md) | 已有實作 | [核心流程](journal/README.md#decision-flow) | [測試](../../backend/internal/business/journal/domain/entity/entry_test.go) |
| [knowledge](knowledge/README.md) | 提案 | [核心流程](../proposals/knowledge.md#design-diagram) | [驗收條件](../proposals/domain-collaboration.md) |
| [業務操作能力](operations/README.md) | 通用操作模型為提案，PDF 文字擷取已實作 | [核心流程](../proposals/operations.md#design-diagram) | [驗收條件](../proposals/domain-collaboration.md) |
| [organization](organization/README.md) | 已有實作 | [核心流程](organization/README.md#decision-flow) | [實作](../../backend/internal/business/organization/adapter/firestore/store.go) |
| [project](project/README.md) | 已有實作 | [核心流程](project/README.md#decision-flow) | [測試](../../backend/internal/business/project/application/reference_source_test.go) |
| [repository](repository/README.md) | 已有實作 | [核心流程](repository/README.md#decision-flow) | [實作](../../backend/internal/business/repository/domain/entity/repository.go) |
| [summary](summary/README.md) | 已有實作 | [核心流程](summary/README.md#decision-flow) | [實作](../../backend/internal/business/summary/application/summarization/validation.go) |
| [team](team/README.md) | 已有實作 | [核心流程](team/README.md#decision-flow) | [實作](../../backend/internal/business/team/adapter/firestore/store.go) |
| [weather](weather/README.md) | 已有實作 | [核心流程](weather/README.md#decision-flow) | [測試](../../backend/internal/business/weather/adapter/line/forecast/forecast_test.go) |
| [workflow](workflow/README.md) | 提案 | [核心流程](../proposals/workflow.md#design-diagram) | [驗收條件](../proposals/domain-collaboration.md) |

## 業務功能

依上表的「狀態」區分現行實作與提案。各領域擁有通用語言、規則及驗收案例；公開格式以 [API 契約](../../api/README.md) 為準。核心／支援子域的分類需依客戶採用理由驗證。跨模組問題查[共用實作邊界](internal/README.md)、[DDD 設計方法](../references/ddd/README.md)或[Organization 與 Team 目標契約](../contracts/organization-team-ownership.md)。

## 出勤

[基礎出勤](attendance/README.md)

[群組出勤看板與本人漏下班檢查](attendance/board.md)

[群組請假登記](attendance/leave.md)

出勤查詢與更正 · 審核畫面流程

[出勤查詢與更正](attendance/queries.md)

[審核畫面流程](attendance/queries.md#審核畫面狀態同步)

工時計算 · 補卡狀態圖

[工時計算](attendance/worktime.md)

[補卡狀態圖](attendance/worktime.md#補卡與人工審核)

<a id="日誌"></a>

## 日誌

[獨立日誌：Feed 與 Dashboard 雛型](journal/journal.md)

<a id="專案"></a>

## 專案

[個人或群組擁有的獨立 Project](project/project.md)

<a id="儲存庫"></a>

## 儲存庫

[Repository 容器與能力索引](repository/README.md#repository-capabilities)

庫內能力包括工區、議題、討論與留言、物件與子物件、物件範本、標籤、里程碑、共用編號、洞察及 Wiki；詳細規格與驗收由 Repository 入口導向各自的單一來源。

跨庫的 [Project](project/README.md) 與獨立的 [Journal](journal/README.md) 保留各自入口。

<a id="群聊摘要"></a>

## 群聊摘要

[基礎群聊摘要](summary/summary.md)

<a id="天氣"></a>

## 天氣

[群組位置天氣](weather/weather.md)

[領域維護規則](AGENTS.md) · [上層導覽](../README.md)
