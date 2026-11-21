# 業務功能

<a id="domain-matrix"></a>

## 18 個業務領域的追蹤矩陣

沿用六邊形架構：從入口適配器追到使用案例或既有實作，再核對輸出介面、儲存適配器及業務判定。各列連到代表路徑；實作連結與提案驗收均不代表已執行測試。

| 領域入口 | 狀態 | 核心圖與契約 | 實作、測試或驗收 |
|---|---|---|---|
| [agent](agent/README.md) | 提案 | [核心流程](../proposals/agent.md#design-diagram) | [驗收條件](../proposals/domain-collaboration.md) |
| [attendance](attendance/README.md) | 已有實作 | [核心流程](attendance/README.md#decision-flow) | [測試](../../backend/internal/business/attendance/adapter/firestore/attendance_integration_test.go) |
| [discussion](discussion/README.md) | 已有實作 | [核心流程](discussion/README.md#decision-flow) | [測試](../../backend/internal/business/discussion/domain/entity/classification_test.go) |
| [enterprise](enterprise/README.md) | 已有實作 | [核心流程](enterprise/README.md#decision-flow) | [實作](../../backend/internal/business/enterprise/adapter/firestore/store.go) |
| [group](group/README.md) | 已有實作 | [核心流程](group/README.md#decision-flow) | [測試](../../backend/internal/business/group/application/query/group_test.go) |
| [issue](issue/README.md) | 已有實作 | [核心流程](issue/README.md#decision-flow) | [測試](../../backend/internal/business/issue/domain/entity/issue_test.go) |
| [journal](journal/README.md) | 已有實作 | [核心流程](journal/README.md#decision-flow) | [測試](../../backend/internal/business/journal/domain/entity/entry_test.go) |
| [knowledge](knowledge/README.md) | 提案 | [核心流程](../proposals/knowledge.md#design-diagram) | [驗收條件](../proposals/domain-collaboration.md) |
| [object](object/README.md) | 已有實作 | [核心流程](object/README.md#decision-flow) | [測試](../../backend/internal/business/object/application/progress_test.go) |
| [operations](operations/README.md) | 提案 | [核心流程](../proposals/operations.md#design-diagram) | [驗收條件](../proposals/domain-collaboration.md) |
| [organization](organization/README.md) | 已有實作 | [核心流程](organization/README.md#decision-flow) | [實作](../../backend/internal/business/organization/adapter/firestore/store.go) |
| [project](project/README.md) | 已有實作 | [核心流程](project/README.md#decision-flow) | [測試](../../backend/internal/business/project/application/reference_source_test.go) |
| [repository](repository/README.md) | 已有實作 | [核心流程](repository/README.md#decision-flow) | [實作](../../backend/internal/business/repository/domain/entity/repository.go) |
| [subobject](subobject/README.md) | 歸入 Object | [Object 核心流程](object/README.md#decision-flow) | [Object 樹儲存](../../backend/internal/business/object/adapter/firestore/store.go) |
| [summary](summary/README.md) | 已有實作 | [核心流程](summary/README.md#decision-flow) | [實作](../../backend/internal/business/summary/application/summarization/validation.go) |
| [team](team/README.md) | 已有實作 | [核心流程](team/README.md#decision-flow) | [實作](../../backend/internal/business/team/adapter/firestore/store.go) |
| [weather](weather/README.md) | 已有實作 | [核心流程](weather/README.md#decision-flow) | [測試](../../backend/internal/business/weather/adapter/line/forecast/forecast_test.go) |
| [workflow](workflow/README.md) | 提案 | [核心流程](../proposals/workflow.md#design-diagram) | [驗收條件](../proposals/domain-collaboration.md) |

## 業務功能

依上表的「狀態」區分現行實作與提案。各領域擁有通用語言、規則及驗收案例；公開格式以 [API 契約](../../api/README.md) 為準。核心／支援子域的分類需依客戶採用理由驗證。跨模組問題查[共用實作邊界](internal/README.md)、[DDD 設計方法](../architecture/ddd/README.md)或[Organization 與 Team 目標契約](../architecture/contracts/organization-team-ownership.md)。

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

空間討論與單層留言 · 留言互鎖時序

[空間討論與單層留言](discussion/design.md)

[留言互鎖時序](discussion/design.md#最小實作邊界與驗證)

[空間基礎雛形完成界線](repository/acceptance.md)

[空間問題第一階段](issue/design.md)

[空間標籤目錄](repository/labels.md)

[里程碑](repository/milestones.md)

[Repository 共用資源編號基礎](repository/numbering.md)


儲存庫物件 · 父子歸屬圖

[儲存庫物件](object/design.md)

[父子歸屬圖](../architecture/contracts/object-subobject-ownership.md#物件與子物件歸屬)

[群組 Repository](repository/README.md)

[儲存庫洞察（Insights）](repository/insights.md)

[物件的單層子物件](object/design.md)

[儲存庫業務空間與 ID 責任](repository/spaces.md)

<a id="群聊摘要"></a>

## 群聊摘要

[基礎群聊摘要](summary/summary.md)

<a id="天氣"></a>

## 天氣

[群組位置天氣](weather/weather.md)

[領域維護規則](AGENTS.md) · [圖表索引](../diagrams.md) · [上層導覽](../README.md)
