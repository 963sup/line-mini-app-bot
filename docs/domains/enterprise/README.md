# 企業

## 領域設計

**子域定位**　治理支援候選。

**通用語言**　開通資格決定能否建立企業；企業擁有者、管理者與可見的組織連結是不同概念。

**上下文與契約**　Enterprise 決定企業開通與擁有者管理。Organization 的導覽可顯示企業連結；導覽結果不能取代寫入時的授權檢查。

**驗收案例**　使用者能登入但缺少開通資格時，建立企業應失敗；查閱企業詳情也不等於取得變更擁有者的資格。

[對照實作或測試](../../../backend/internal/business/enterprise/adapter/firestore/store.go) · [戰略設計與分類準則](../../architecture/ddd/README.md)

企業與組織連結的目標設計收錄在組織所有權規格。

[相關業務規格](../repository/organization.md) · [後端enterprise 實作說明](../../../backend/internal/business/enterprise/application/dto/model.go)
