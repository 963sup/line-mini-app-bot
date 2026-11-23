# 儲存庫模式（Repository）

DDD Repository 提供以 Aggregate 為單位的載入與保存能力，讓業務操作不用解釋資料庫文件或查詢語法。它與產品的 Repository 業務領域是不同概念。

## 邊界與契約

以業務識別及需要的操作定義介面，避免將 collection 名稱、SDK transaction 或通用 CRUD 查詢器傳入核心。用例需要的 Port 可由 Application 定義，儲存 Adapter 提供實作。

新增與重建必須分開：載入既有 Aggregate 不得重新分配 ID、重設 joinedAt 或發出「已建立」通知。重建也不能直接信任損壞資料；須有相容性檢查與失敗策略，避免把舊資料當成新建輸入重新套用所有現行規則。

保存契約明列 expectedVersion、原子範圍、not_found、conflict、unavailable 與取消語意。若多個 Repository 呼叫必須一起提交，另由 Unit of Work 協調；介面名稱相同或共用 client 不代表交易相同。

## 讀取投影與例子

People 清單可由 Reader Port 直接回傳投影，不必為顯示姓名載入完整 Aggregate；讀取仍須授權、分頁與租戶隔離。寫入側可依確認後的 Aggregate 邊界定義 Repository，與 People Reader 共用儲存來源。兩者契約用途不同；People 的畫面範圍也不能用來推定 Aggregate 的邊界。

## 設計產出與核對

列出載入／保存單位、重建責任、版本衝突、交易生命週期與格式轉換。不要為每張資料表建立一個 Repository，也不因採用此概念引入通用 Repository 框架。

[Aggregate](aggregate.md) · [Factory](factory.md) · [六邊形架構](hexagonal-architecture.md) · [DDD 設計索引](design-guide.md)
