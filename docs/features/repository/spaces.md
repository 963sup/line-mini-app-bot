# 儲存庫業務空間

Space 是 Repository 內的業務分類空間，具有獨立的 32 位小寫十六進位 ID 與名稱。它不是 LINE 群組，也不產生新的成員或授權範圍；既有物件可保持未分配。

## 身分與相容邊界

| 識別碼 | 權威與責任 |
| --- | --- |
| GroupID | 簽章驗證的 LINE 群組上下文與目前成員資格；`C` 加 32 位十六進位。 |
| UserID | 後端驗證 LINE 身分取得的 actor，不接受 body 宣告。LINE User ID 以 provider 為範圍；同 provider 的相關 channels 才能對應同一使用者。 |
| RepositoryID | 伺服器管理的資料來源容器，與 LINE GroupID 分開。 |
| SpaceID | 本功能的 repository-owned 業務空間 ID；同一空間不能掛到其他 repository 或 channel。 |

工程群組透過 `/api/v1/groups` 存取，Project 的 `ownerId` 指向該群組。庫內工區只使用 Repository 的 `spaces` 路徑及 `spaceId`，移動工區不會改變 Project 的歸屬。

LINE 規則核對官方 [User IDs](https://developers.line.biz/en/docs/messaging-api/getting-user-ids/)，官方 source main revision `259d1fa7698d597f95a7f044f2592379a63defbe`，2026-11-18。不同 provider 的 User ID 不得僅憑同一真人自行合併。

## 操作契約

| 操作 | Context／Authority／Policy | Result／重送 |
| --- | --- | --- |
| 成員讀取列表／單筆 | 可信 actor、目前 repository 成員；容器 ACTIVE；不另設 Space 成員 | GET `/api/v1/repositories/{repositoryId}/spaces[/id]`，列表每頁 20 筆，游標綁 actor/channel/repository。 |
| 成員建立空間 | 同上；名稱 trim 後 1–120 字；寫入交易核 repository ID/GroupID/ACTIVE | POST collection；`Idempotency-Key` 綁 actor/repository 與名稱。同鍵同內容回原空間，異內容衝突；不使用共用編號。 |
| 物件建立／修改分配 | 沿物件既有權限與 expectedVersion；同交易核業務空間屬於目前 channel/repository | 可選 `spaceId`：建立省略為未分配；PATCH 省略保留、空字串解除。失敗不寫物件或收據。 |

資料位於 `channels/{channelId}/repositories/{repositoryId}/spaces/{spaceId}`；建立收據位於同庫 `spaceCreationReceipts`。首版只提供建立／讀取，不提供刪除、獨立成員管理或正式資料遷移。Subobject 沿父物件的空間，不另存 SpaceID 或建立權限。

## 實作與驗收範圍

後端由 `backend/internal/business/repository/space` 擁有，OpenAPI 權威為 `api/contract/repository/spaces.yaml`。物件編輯器提供未分配／空間選取及建立空間。建立結果不明時沿同一 body/key 重試，不因切換選項建立另一意圖。

必要驗收包含 ID 隔離、可信 actor、目前成員／停用容器、同庫同 channel 分配、PATCH 省略與清除、legacy 收據指紋、並行重送與分頁。local emulator 不等於正式 Firestore／LINE 驗收；本輪不操作正式資料。
