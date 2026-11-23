# 領域狀態機

狀態機描述資源在具名操作後可達的狀態，以及轉換成立的條件。適用於有生命週期的模型；單純查詢或由其他欄位唯一推導的狀態，不必另建可任意修改的狀態欄位。本頁為設計方法，具體轉換以所屬領域規格為準。

## 轉換契約

| 欄位 | 必須寫清楚的內容 |
|---|---|
| 狀態 | 每個值代表的業務事實、初始值與舊資料缺值語意 |
| 操作 | 觸發意圖、可信 Actor、輸入與適用來源狀態 |
| Guard | 轉換前須成立的資格、版本、時間與領域條件 |
| 結果 | 目標狀態、同步更新的資料及必要事件／稽核 |
| 拒絕 | 哪個條件失敗、呼叫端如何區分、原狀態如何保留 |
| 終止狀態 | 哪些操作永遠不再允許；是否允許重送取得原結果 |

授權與 Guard 分開描述：具有操作權的人仍可能因目前狀態而被拒絕。Guard 在權威狀態上求值，持久化時需防止過期讀取覆寫。

## 示例：可撤銷的邀請

以下僅為方法示例，未宣稱專案已有 Invitation 功能。

| 來源 | 操作 | Guard | 目標／拒絕 |
|---|---|---|---|
| pending | 接受 | 接受者相符、未到期、版本相符 | accepted |
| pending | 撤銷 | 有撤銷權、版本相符 | revoked |
| pending | 接受 | 已到期 | 拒絕並保留 pending；不假造已持久化的 expired 狀態 |
| revoked | 接受 | 無可成立條件 | 拒絕並保留 revoked |
| accepted | 再次接受 | 原操作身分與重送條件相符 | 依契約回傳原結果，不再建立成員 |

accepted 與 revoked 在此示例為終態；若需求允許恢復，應另定具名轉換。時間到期可只是 Guard，不必假定有背景工作改寫狀態。

## 現行程式對照

[Issue 規格](../../domains/issue/design.md#issue-lifecycle)允許關閉後重新開啟，因此 closed 不能直接當作終態。[ApplyUpdate](../../../backend/internal/business/issue/domain/entity/issue.go)核對轉換與版本。[Object.WorkStatus](../../../backend/internal/business/object/domain/entity/object.go)從總量和完成量推導狀態，不能額外保存互相矛盾的進度判定。

## 驗證與設計產出

列出狀態集合及轉換表，為未列出的操作定義拒絕。驗證每條合法轉換、非法來源、Guard 的邊界時間、終態、同狀態重送、拒絕後資料保留，以及「先接受後撤銷」和相反順序的結果。併發測試須證明兩個競爭請求不能提交互斥結果；單元測試只證明模型判定。

產出須包含規則擁有者、轉換表、Guard 所需事實與交易邊界、可辨識錯誤及驗收案例。狀態值新增時，依[契約演進](model-evolution.md)核對舊讀端行為。

[不變條件](invariant.md) · [領域授權](domain-authorization.md) · [冪等](idempotency.md) · [領域測試](domain-testing.md) · [DDD 索引](design-guide.md)
