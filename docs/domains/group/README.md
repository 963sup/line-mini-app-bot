# 群組

## 領域設計

**子域定位**　群組身分與探索的支援能力。

**通用語言**　群組觀測是已驗簽事件帶來的群組狀態；候選成員是可供探索的人員線索。

**上下文與契約**　LINE 提供事件與成員查詢；Group 整理狀態並核驗成員資格。Team 另外決定管理歸屬。

**驗收案例**　曾在事件中出現的使用者即使被列為候選，仍須通過群組成員核驗才能取得受限資料。

[對照實作或測試](../../../backend/internal/business/group/group.go) · [戰略設計與分類準則](../../architecture/ddd/README.md)

管理群組狀態與 LINE 群組事件。

[後端group 實作說明](../../../backend/internal/business/group/README.md)
