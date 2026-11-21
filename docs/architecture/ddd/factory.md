# 工廠（Factory）

封裝領域物件的建立過程，讓呼叫端取得符合建立時不變條件的完整物件，無須知道內部組裝細節。當 Aggregate 或複雜 Value Object 的建立需要多個步驟、規則或內部物件時，可採用 Factory；簡單建立可保留具名建構函數，不要求每個 Entity 都有獨立工廠。

Factory 屬於戰術設計概念，不新增一個限界上下文，也不等同 GoF 的 Abstract Factory 或 Factory Method 類別結構。

## 責任與邊界

| 概念 | 負責什麼 |
| --- | --- |
| Factory／具名建構函數 | 建立新的完整領域物件，驗證必要輸入與初始狀態；建立失敗不回傳可用的半成品。 |
| Aggregate Root | 保護 Aggregate 在建立後各次狀態變更的不變條件；Factory 不能取代這項責任。 |
| Application | 取得可信身分、時間、識別碼及必要外部事實，協調授權、Factory、交易與保存。 |
| Repository | 載入與保存 Aggregate。重建既有物件須保留原識別、版本與歷史狀態，不重跑「新建」業務流程。 |
| Bootstrap | 組裝服務與技術依賴；建立 SDK client 或注入 adapter 不等於領域 Factory。 |

本專案的採用原則：純領域 Factory 不直接呼叫 HTTP、LINE 或 Firestore，不在建立過程發通知或提交交易。需要最新資料才能成立的規則，由 Application 在一致性邊界內取得事實並協調；Factory 對傳入值驗證成功，不代表資料庫唯一性、授權或併發檢查已完成。

## 如何選擇實作形式

- 建立規則簡單時，用模型所在套件的具名建構函數即可。
- 建立 Aggregate 內部 Entity 時，優先透過 Aggregate Root 的業務方法，避免呼叫端繞過根直接插入物件。
- 組裝複雜且不適合放在模型本身時，再抽出獨立 Factory；建立語意仍應使用通用語言。
- 不強制建立 `domain/factory/` 目錄、Factory interface 或一對一 Factory 類別；依實際責任與依賴決定位置。

## 成員管理設計例子

以下為未來設計示例，不表示 Membership 已實作或 Aggregate 邊界已定案。

若成員管理規則要求「新範圍建立時必須有首位 Owner」，可由 `OpenScope` 的建立函數一次產生範圍與初始 Owner 關係，拒絕缺少建立者的輸入。Application 再協調資源建立及成員資料原子保存。

如果 Owner 關係與範圍屬於同一 Aggregate，Factory 必須回傳完整 Aggregate；如果設計為不同 Aggregate，跨兩者的一致性由用例與交易策略明確處理，不能只因共用一個 Factory 就宣稱原子性成立。後續移除 Owner，仍由對應模型與交易規則保護最後一位 Owner。

## 設計產出與核對

記錄建立入口、輸入事實的提供者、建立時必須成立的規則、回傳的 Aggregate 邊界與失敗結果。區分新建與持久資料重建，說明哪些規則仍需交易保證；驗收時核對無效輸入不產生可用半成品、重建不觸發新建副作用。

概念依據：Eric Evans，[*Domain-Driven Design Reference*，2015-03，Factories](https://www.domainlanguage.com/wp-content/uploads/2016/05/DDD_Reference_2015-03.pdf)。2026-11-21 核對；上述純領域與套件安排是本專案的採用原則。

[七階段設計流程](design-process.md) · [限界上下文](bounded-context.md) · [Aggregate](aggregate.md) · [Invariant](invariant.md) · [Repository](repository.md) · [DDD 設計索引](README.md)
