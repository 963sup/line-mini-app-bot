# 樂觀並行控制（Optimistic Concurrency）

更新時檢查讀取版本以偵測並行修改。

## 設計時要確認

比對預期版本並原子更新；重試前重讀並重新授權及判斷。

## 示例

兩個請求讀到版本 5，其中一個提交版本 6，另一個應回衝突或重讀。

## 圖解

**設計參考。**

```mermaid
sequenceDiagram
  participant A as 用例 A
  participant B as 用例 B
  participant DB as 儲存庫
  A->>DB: 讀取版本 5
  B->>DB: 讀取版本 5
  A->>DB: 預期版本 5，保存版本 6
  DB-->>A: 提交成功
  B->>DB: 預期版本 5，嘗試保存
  DB-->>B: 版本衝突，無寫入
```

## 延伸閱讀

[transaction](transaction.md) · [aggregate-root](aggregate_root.md) · [DDD 設計索引](design_guide.md)
