# 防腐層（Anti Corruption Layer）

在模型邊界轉譯語意以保護下游模型。

## 設計時要確認

為枚舉、識別、單位和錯誤建立明確映射；未知值依契約拒絕或隔離。

## 示例

上游的 active 只表示帳戶已啟用，轉譯後仍須依本地規則判定能否接受邀請。

## 圖解

**設計參考。**

```mermaid
flowchart LR
  U["上游公開契約"] --> A["ACL：轉譯識別、狀態與錯誤"]
  A -->|可映射| F["下游模型所需事實"]
  A -->|未知值| E["依契約拒絕或隔離"]
  F --> P["下游政策判斷"]
```

## 延伸閱讀

[bounded-context-integration](bounded-context-integration.md) · [driven-adapter](driven-adapter.md) · [DDD 設計索引](README.md)
