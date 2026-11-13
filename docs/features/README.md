# 功能規格

出勤、工時計算、天氣與摘要等跨模組功能流程及業務規則。

## 文件索引

- [Space 擁有的獨立 Project](projects.md)：獨立業務、側欄與頁面，來源儲存庫唯讀，包含legacy相容與資料遷移界線。
- [出勤查詢與更正](attendance-queries.md)
- [基礎出勤](attendance.md)
- [群組出勤看板與本人漏下班檢查](attendance-board.md)
- [群組請假登記](leave.md)
- [日誌表單](journal.md)
- [基礎群聊摘要](summary.md)
- [群組位置天氣](weather.md)
- [工時計算第一階段](worktime.md)

## 歸屬與維護

同一功能的前端、API、後端規則集中描述；實作位置以連結指向模組，不複製多份規格。區分已實作、待驗收與未來規劃。
新增或搬移文件時更新本索引及 [總索引](../README.md)，並同步相關模組 README。

[本目錄指引](AGENTS.md) · [父層指引](../AGENTS.md) · [文件總索引](../README.md)

新增或修改業務操作時，依 [操作契約範本](../architecture/business-operation.md#功能操作契約範本) 在對應功能文件補齊六個問題與證據；未知規則明列待確認。

- [群組 Repository](repository.md)
- [空間基礎雛形完成界線](repository-prototype.md)
- [空間標籤目錄](labels.md)：最基本的列出、建立及事項／討論作者分類。
- [Repository 共用資源編號基礎](repository-numbering.md)
- [空間事項第一階段](issues.md)：成員可讀取及建立，僅建立者可修改及變更狀態。
- [空間討論與單層留言](discussions.md)：成員可發起及留言，各自作者可編輯。

後續業務規劃：[Repository、Issue 與 Discussion 業務模型](../architecture/domain-model.md)；Issue 與 Discussion 已依各自功能規格接入；標籤分類與里程碑／專案基礎已接入，完整專案管理仍屬後續規劃。

- [里程碑、專案與總覽統計](planning.md)：同空間最小追蹤流程與完成邊界。
