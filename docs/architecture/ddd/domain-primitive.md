# 領域基本型別（Domain Primitive）

封裝單一領域概念與有效值的小型型別，常以值物件實作。

## 設計時要確認

在建構時檢查格式與範圍，明定零值、單位和相等性；需外部狀態的資格檢查另行處理。

## 示例

有效的 InvitationID 只代表格式有效，是否存在由儲存庫查詢。

## 延伸閱讀

[value-object](value-object.md) · [domain-error](domain-error.md) · [DDD 設計索引](README.md)
