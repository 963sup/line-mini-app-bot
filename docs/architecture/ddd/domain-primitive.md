# 領域基本型別（Domain Primitive）

封裝單一領域概念與有效值的小型型別，常以值物件實作。

## 設計時要確認

在建構時檢查格式與範圍，明定零值、單位和相等性；需外部狀態的資格檢查另行處理。

## 示例

有效的 InvitationID 是一個識別值，其格式與作用域可在建立時驗證。被識別的 Invitation 是有生命週期的實體。該邀請是否存在由用例經儲存庫查詢；查得到也仍須檢查是否可以執行當前操作。

## 延伸閱讀

[value-object](value-object.md) · [domain-error](domain-error.md) · [DDD 設計索引](README.md)
