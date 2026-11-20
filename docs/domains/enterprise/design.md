# 企業帳戶設計

## Enterprise（企業帳戶）

| 資料或操作 | 規則與驗收條件 |
|---|---|
| `enterprises/{enterpriseId}` | 企業集中管理旗下組織與 Enterprise Team。Organization 可獨立存在，加入企業後以 enterpriseId 指向一個 Enterprise。 |
| 指派企業 Team | 具企業 Team 指派權的管理者指定旗下 Organization。後端核對 Team 的管理企業與目標組織的企業相同，再依組織與 Team 的組合鍵寫入；跨企業一律拒絕，重送不產生第二份名單。 |
| 移除指派 | 刪除指定組織的指派。依賴這筆指派的授權即失效，其他組織的指派與企業成員名單維持有效。 |

### 企業自助開通與共用頁面

| 操作 | 後端規則與前端呈現 |
|---|---|
| 顯示入口 | 組織 context 一併回傳 enterprises 與 canCreateEnterprise；空企業也對其 Owner 顯示。側欄直接使用回應。 |
| POST /api/v1/enterprises | 名稱為 1–120 字。後端於交易內核對登入者資格、建立企業、加入第一位 Owner 並記錄 enterpriseId。重送同一開通請求回傳既有企業。 |
| GET /api/v1/enterprises/{enterpriseId} | 企業 Owner 與管理者可看旗下組織及企業團隊清單；一般組織成員只看本人可見組織及已指派且通過 LINE 成員核驗的 Team。只有 Owner 收到 canManage=true。 |
| POST /api/v1/enterprises/{enterpriseId}/owners | 現任企業 Owner 可新增另一位 Owner；交易保留原有 Owner，支援既有 ownerUserId。 |
| POST /api/v1/enterprises/{enterpriseId}/organizations | 操作者須同時擁有企業與目標組織；已屬其他企業回 409。獨立組織可繼續單獨運作。 |

開通資格由服務端配置於同一渠道下的 `enterpriseEntitlements/{SHA-256(userId)}`（文件鍵使用專案 firestorekey.Digest）。設定 `enabled: true` 提供一次企業開通資格。建立成功後由後端保存 enterpriseId；缺少資格時直接呼叫建立 API 也回 403。頁面使用 OpenAPI 生成的型別，顯示後端回傳的清單與管理操作。


[跨領域歸屬契約](../../architecture/contracts/organization-team-ownership.md) · [領域入口](README.md)
