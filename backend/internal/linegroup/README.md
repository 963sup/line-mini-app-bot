# LINE 群組相容整合

[領域文件](../../../docs/domains/group/README.md) · [全域索引](../../../docs/domains/README.md#domain-matrix)

保存從已驗簽 LINE 事件取得的群組資料與成員候選資訊，提供群組查詢。

本目錄維護 LINE 群組觀測、舊群組查詢與相容 API。Team 管理團隊狀態與歸屬，Membership 管理正式資格。群組候選索引只用於發現資料，讀取仍須驗證即時 LINE 成員資格。

## 閱讀與修改入口

先查閱[group.go](application/query/group.go)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[專案設計](../../../docs/domains/project/project.md) · [現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 閱讀入口

Group 直接以 LINE GroupID 為文件 ID，保存已驗簽的群組事件與成員候選清單。Project 歸屬與授權見[專案設計](../../../docs/domains/project/project.md)。

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 領域追蹤入口

[領域規格](../../../docs/domains/group/README.md) · [判定流程](../../../docs/domains/group/README.md#decision-flow) · [代表測試](application/query/group_test.go)

[上層導覽](../README.md)
