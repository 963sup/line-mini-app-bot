# 組織、企業與團隊的共用 Adapter

Enterprise、Organization 與 Team 共用 HTTP 和 Firestore 轉接工具。授權、所有權與移轉資格由各領域依[歸屬契約](../../contracts/organization-team-ownership.md)決定。共用交易工具不形成獨立業務領域。

| 元件                                                                               | 責任                                                                                            |
| ---------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| [orghttp](../../../backend/internal/business/internal/adapter/http/response.go)    | 驗證呼叫者、解碼請求並映射回應。未通過身分驗證時，不呼叫業務操作；資源歸屬衝突映射為 HTTP 409。 |
| [orgstore](../../../backend/internal/business/internal/adapter/firestore/store.go) | 提供 Channel 範圍的資料讀取與交易工具。操作權限與一致性範圍由呼叫端決定。                       |

[後端功能邊界](backend-boundaries.md) · [依賴規則](dependencies.md) · [執行架構索引](./)
