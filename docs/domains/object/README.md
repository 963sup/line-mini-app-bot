# 工項

## 領域設計

**子域定位**　工程數量與進度的核心候選。

**通用語言**　總量、完成量與狀態共同描述工項進度；未提供完成量與明確填零有不同語意。

**上下文與契約**　Repository 提供操作範圍，Object 負責數量與進度規則。匯入與手動建立都應保留相同的數量語意。

**驗收案例**　總量為 5、完成量為 8 的工項應被拒絕；等值十進位表示不應產生不同的完成狀態。

[對照實作或測試](../../../backend/internal/business/object/progress_test.go) · [戰略設計與分類準則](../../architecture/ddd/README.md)

工項規格涵蓋數量、進度與狀態。

[相關業務規格](../repository/objects.md) · [後端object 實作說明](../../../backend/internal/business/object/README.md)
