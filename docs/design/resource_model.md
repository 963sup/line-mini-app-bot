# 資源模型

系統以 bounded context 擁有資源與規則。API resource、Domain model 與 Firestore document 可以有不同形狀，但語意與擁有者必須一致。

## 主要資源範圍

| 範圍                                | 主要責任                            |
| ----------------------------------- | ----------------------------------- |
| Organization／Enterprise／Team      | 組織結構、狀態與管理歸屬            |
| Membership                          | actor 與組織資源之間的資格          |
| Repository                          | 工作空間、分類、里程碑、wiki 與統計 |
| Project／Issue／Discussion／Journal | 專案協作與記錄                      |
| Attendance                          | 打卡、紀錄、請假、更正與審核        |
| Object                              | 文件候選、物件範本與 PDF 擷取       |
| Summary／Weather                    | 衍生摘要與外部天氣資料              |

## 識別與歸屬

資源 ID 必須穩定且不可直接由未驗證輸入拼成資料庫路徑。每個資源明確記錄租戶、organization、team、group、repository 或其他作用域；跨作用域引用使用公開 ID 與契約，不共享可變 entity。

## 操作模型

每個業務操作至少定義：

- Actor、Resource、Action；
- 輸入與可信度；
- 前置狀態與授權來源；
- 成功結果、拒絕與失敗狀態；
- 重送、版本衝突及結果未知；
- 必要 Event、Audit 與外部副作用。

前端只收集輸入與呈現結果；OpenAPI 發布操作契約；後端根據目前資料執行權威判斷。

## 跨資源合作

跨 bounded context 的操作透過使用端 port 取得必要事實。不得直接讀寫另一 context 的 document schema。若多個資源需要同一原子結果，先確認它們是否屬於同一一致性邊界；否則明確設計補償、工作狀態或最終一致性。

## 物件欄位與來源識別

| 欄位                                                       | 規則                                                                                                                                                                                               |
| ---------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| name、description                                          | 名稱 1–120 字元、完整說明最多 8000 字元；首尾空白移除，不靜默截斷                                                                                                                                  |
| type、unit                                                 | 新UI使用 work（施作）／non_work（非施作）；可留空表示未分類，原自由類型保留相容。單位可留空，填寫時最多32字元；追蹤進度時必填，不換算                                                              |
| quantity                                                   | 可留空表示未知；填寫時為精確正十進位字串，最多 18 位整數及 6 位小數；不用浮點數。拒絕零、負數、指數、千分位，整數前導零只允許單一 0，小數尾零正規化後保存                                          |
| completedQuantity                                          | 可選精確非負十進位字串，精度同quantity；施作已知完成量須介於0與總量，非施作不保存完成量。缺值不是0                                                                                                 |
| status                                                     | 施作有完成量時由後端switch衍生：0為not_started、中間為in_progress、等總量為completed；新UI不可另選狀態。未分類物件不顯示進度與狀態；非施作在UI顯示不適用，底層三值僅相容保留；完成不代表驗收／付款 |
| id、repositoryId、creatorId、version、createdAt、updatedAt | 後端可信來源，建立不可自行指定；不提供額外 #number                                                                                                                                                 |

來源欄位可獨立留白，只是追溯資料，不承諾 sourceDocument＋revision＋itemNo 唯一。同來源換建立鍵仍可新增另一筆；不自動認定為同一物件。新版 PDF 不會自動比對、合併或覆寫既有物件與狀態。來源項次不作為資料庫 ID。

手動新增物件僅顯示名稱，不要求或預填類型、數量、單位、空間與完成數；也不讀取空間選項。只含 `name` 的 API 請求可成功建立，未提供的資料以空值保存，不補施作、1 或 SET。既有資料不遷移。

建立者開啟編輯後可逐步補資料；名稱仍必填，數量若填寫仍須為精確正數。要追蹤施作完成量時才核驗總量、單位與完成數，既有有進度物件仍不能藉清空欄位繞過進度保護。來源文件覆核建立也只要求名稱，保留來源與引文，不暗中套入未確認的類型或數量。更新仍使用原有完整內容與 expectedVersion 契約。

建立 Object 時可指定同一 Repository 中已存在的 `parentId`。省略時建立根節點。父、子和葉是依階層推導的角色。父連結建立後固定，因此新節點可以接到現有節點，更新內容不會改寫階層或產生循環。每個節點各自保存數量、單位與進度。一般 Object 清單仍使用既有分頁，回傳包含各層的 Object。

[物件契約](../../api/contract/object/objects.yaml) · [數量格式](../../backend/internal/business/object/domain/valueobject/quantity.go)

## 儲存庫空間與物件分配

| 操作                                                                                                                        | Result／重送                                                                                                  |
| --------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| 成員讀取列表／單筆Context／Authority／Policy：可信 actor、目前 repository 成員；容器 ACTIVE；不另設 Space 成員              | GET /api/v1/repositories/{repositoryId}/spaces[/id]，列表每頁 20 筆，游標綁 actor/channel/repository。        |
| 成員建立空間Context／Authority／Policy：同上；名稱 trim 後 1–120 字；寫入交易核 repository ID/GroupID/ACTIVE                | POST collection；Idempotency-Key 綁 actor/repository 與名稱。同鍵同內容回原空間，異內容衝突；不使用共用編號。 |
| 物件建立／修改分配Context／Authority／Policy：沿物件既有權限與 expectedVersion；同交易核業務空間屬於目前 channel/repository | 可選 spaceId：建立省略為未分配；PATCH 省略保留、空字串解除。失敗不寫物件或收據。                              |

[空間契約](../../api/contract/repository/spaces.yaml)

## 物件進度與舊資料相容

- quantity是唯一總數，completedQuantity是完成數；畫面進度由兩者組成，不另外保存百分比或重複總量。後端以math/big精確比較18位整數／6位小數，前端以固定小數位的BigInt預驗，不使用float或int64保存數量。更新是設定絕對完成數，不是每次加一；同一值不製造版本變更。

- 施作已有完成量時，省略完成量保留現值；減少總量到完成數以下拒絕，除非同次明確提交合法新完成數。目前完成數大於0時拒絕轉為非施作或舊類型；須先明確將施作完成數調為0保存，再改類型。從非施作或舊自由類型轉施作必須明確提供完成數，不靜默歸零；新施作已提供總量與單位時，建立省略值才預設0。

- 歷史自由type（包含原字串恰為work／non_work）及缺完成量的紀錄仍可讀，沒有批次遷移或猜測Engineering就是施作。畫面顯示類型待確認／進度待補；轉施作時可依舊not_started／completed提供0／總量預填，使用者可改；舊in_progress不捏造數字。舊PATCHstatus相容路徑保留；已追蹤施作若另送與完成數矛盾的status則拒絕。非施作底層原status不因改名被重設，UI一律不適用，不能把technical status解釋成施作完成。

- 既有expectedVersion、建立者與當下成員資格不變；交易讀取目前總量／完成量／類型／版本後同時驗證與更新。舊建立鍵的v1內容指紋保持相容，缺完成量／新施作預設0可取回舊資源，不重設原紀錄。

[進度轉換測試](../../backend/internal/business/object/domain/entity/conversion_test.go)

## 儲存庫、群組與工區識別

Space 是 Repository 內的業務分類空間，具有獨立的 32 位小寫十六進位 ID 與名稱。它不是 LINE 群組，也不產生新的成員或授權範圍；既有物件可保持未分配。

| 識別碼       | 權威與責任                                                                                                                          |
| ------------ | ----------------------------------------------------------------------------------------------------------------------------------- |
| GroupID      | 簽章驗證的 LINE 群組互動綁定；C 加 32 位十六進位。它不是 Owner，也不直接授予 Repository 權限。                                      |
| UserID       | 後端驗證 LINE 身分取得的 actor，不接受 body 宣告。LINE User ID 以 provider 為範圍；同 provider 的相關 channels 才能對應同一使用者。 |
| RepositoryID | 伺服器管理的資料來源容器，與 LINE GroupID 分開。                                                                                    |
| Owner        | Repository 的業務擁有者，只能是 User 或 Organization；授權由 Owner 所屬領域判定。                                                   |
| SpaceID      | 本功能的 repository-owned 業務空間 ID；同一空間不能掛到其他 repository 或 channel。                                                 |

[工區契約](../../api/contract/repository/spaces.yaml)
