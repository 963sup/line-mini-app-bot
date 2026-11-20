# 組織

## 領域設計

**子域定位**　組織治理的支援候選。

**通用語言**　組織擁有者可發起移轉；成員資格決定誰能成為新擁有者。企業連結與團隊歸屬另有規則。

**上下文與契約**　Organization 決定組織移轉，Team 決定團隊管理範圍。企業與團隊可出現在同一導覽結果，各自的管理權仍需核對。

**驗收案例**　將組織移轉給他人時，在同一交易以目前擁有者與成員資料判斷；非擁有者或非成員受讓者不能完成移轉。

[對照實作或測試](../../../backend/internal/business/organization/adapter/firestore/store.go) · [戰略設計與分類準則](../../architecture/ddd/README.md)

組織所有權、移轉及團隊歸屬的目標規格集中於此入口。

[相關業務規格](../repository/organization.md) · [後端organization 實作說明](../../../backend/internal/business/organization/domain/entity/model.go)
