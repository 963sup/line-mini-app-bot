# 儲存庫統計

成員可讀取所屬儲存庫的事項狀態、討論、里程碑與專案數。專案數表示在該庫所屬協作空間建立、以該庫為初始來源的專案；後來才連結此庫的跨庫專案不計入。

## 操作契約

| 項目 | 規則 |
| --- | --- |
| Actor／Resource／Action | Session／Bearer 核驗本人，讀取所屬儲存庫的完整統計 |
| Context／Authority | 啟用儲存庫與當下群組成員資格；只可讀取有權範圍 |
| Policy／Decision | 統計完整集合，不以單頁筆數作總數；沿用各功能自己的狀態 |
| Result | 真實數量與觀察時間；讀取失敗顯示錯誤，不偽裝為 0 或沿用過期成功數字 |
| 併發／一致性 | 即時查詢，無快取或背景工作；不提供跨請求統計快照保證 |
| Event／Audit | 唯讀，不建立額外寫入或外部通知 |

## 實作與驗收入口

用例位於 `backend/internal/business/repository/statistics`，不建立通用統計引擎。資料量增長時依實際查詢成本另行評估聚合，不自動增加計數器、排程或趨勢圖。

[OpenAPI 來源](../../../api/contract/repository/statistics.yaml) · [Project 規格](../project/projects.md) · [驗收矩陣](../../agents/acceptance.md)。真實 LINE 身分、群組與正式資料庫須有獨立驗收證據，不沿用舊報告的通過結論。
