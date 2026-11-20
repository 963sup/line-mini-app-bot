# 業務功能

## 用途與範圍

保存各業務領域的行為、權限與驗收規格。

## 架構與責任邊界

領域文件維護產品規則，公開的請求與回應由 OpenAPI 契約定義。

## 閱讀與修改入口

[DDD 戰略設計概念、上下文映射與設計流程](../architecture/ddd/README.md)

各領域 README 說明通用語言、責任邊界與驗收案例。核心／支援分類是產品投資假設，應以客戶採用理由驗證；internal 記錄共用實作的技術邊界。各目錄入口如下。文件維護方式見[工作指引](AGENTS.md)。

- [attendance · 出勤](attendance/README.md)
- [discussion · 討論](discussion/README.md)
- [enterprise · 企業](enterprise/README.md)
- [group · 群組](group/README.md)
- [internal · 共用實作](internal/README.md)
- [issue · 議題](issue/README.md)
- [journal · 日誌](journal/README.md)
- [object · 工項](object/README.md)
- [objectimport · 工項匯入](objectimport/README.md)
- [organization · 組織](organization/README.md)
- [project · 專案](project/README.md)
- [repository · 儲存庫](repository/README.md)
- [subobject · 子工項](subobject/README.md)
- [summary · 群聊摘要](summary/README.md)
- [team · 團隊](team/README.md)
- [weather · 天氣](weather/README.md)

## 相關規格

[Organization 所有權與 Team 歸屬（目標規格）](repository/organization.md) · [功能規格的歸屬與維護方式見功能文件指引。](AGENTS.md)


<a id="業務功能"></a>

## 業務功能

[儲存庫](#儲存庫)
[Organization 所有權與 Team 歸屬（目標規格）](repository/organization.md)

[專案](#專案)

[出勤](#出勤)

[日誌](#日誌)

[群聊摘要](#群聊摘要)

[天氣](#天氣)

[功能規格的歸屬與維護方式見功能文件指引。](AGENTS.md)

現行實作圖索引 · 未來設計圖。

[現行實作圖索引](../diagrams.md#現行實作)

[未來設計圖](../diagrams.md#未來設計)

<a id="出勤"></a>

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

現行實作圖索引 · 未來設計圖。

[現行實作圖索引](../diagrams.md#現行實作)

[未來設計圖](../diagrams.md#未來設計)

<a id="日誌"></a>

## 日誌

[獨立日誌：Feed 與 Dashboard 雛型](journal/journal.md)

現行實作圖索引 · 未來設計圖。

[現行實作圖索引](../diagrams.md#現行實作)

[未來設計圖](../diagrams.md#未來設計)

<a id="專案"></a>

## 專案

[個人或群組擁有的獨立 Project](project/project.md)

現行實作圖索引 · 未來設計圖。

[現行實作圖索引](../diagrams.md#現行實作)

[未來設計圖](../diagrams.md#未來設計)

<a id="儲存庫"></a>

## 儲存庫

空間討論與單層留言 · 留言互鎖時序

[空間討論與單層留言](repository/discussions.md)

[留言互鎖時序](repository/discussions.md#最小實作邊界與驗證)

[空間基礎雛形完成界線](repository/acceptance.md)

[空間問題第一階段](repository/issues.md)

[空間標籤目錄](repository/labels.md)

[里程碑](repository/milestones.md)

[Repository 共用資源編號基礎](repository/numbering.md)

[物件自動匯入](repository/object-imports.md)

儲存庫物件 · 父子歸屬圖

[儲存庫物件](repository/objects.md)

[父子歸屬圖](repository/objects.md#物件與子物件歸屬)

[群組 Repository](repository/README.md)

[儲存庫洞察（Insights）](repository/insights.md)

[物件的單層子物件](repository/objects.md)

[儲存庫業務空間與 ID 責任](repository/spaces.md)

現行實作圖索引 · 未來設計圖。

[現行實作圖索引](../diagrams.md#現行實作)

[未來設計圖](../diagrams.md#未來設計)

<a id="群聊摘要"></a>

## 群聊摘要

[基礎群聊摘要](summary/summary.md)

現行實作圖索引 · 未來設計圖。

[現行實作圖索引](../diagrams.md#現行實作)

[未來設計圖](../diagrams.md#未來設計)

<a id="天氣"></a>

## 天氣

[群組位置天氣](weather/weather.md)

現行實作圖索引 · 未來設計圖。

[現行實作圖索引](../diagrams.md#現行實作)

[未來設計圖](../diagrams.md#未來設計)
