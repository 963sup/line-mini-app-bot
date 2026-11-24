# 領域錯誤（Domain Error）

表示模型拒絕某個業務操作的穩定原因。

## 設計時要確認

為呼叫端需要區分的拒絕定義型別或代碼，由 Adapter 映射傳輸格式，基礎設施不可用單獨分類。

## 示例

邀請已撤銷可由 UI 顯示替代操作，儲存逾時則交由用例處理結果未知。

## 延伸閱讀

[invariant](invariant.md) · [adapter](adapter.md) · [DDD 設計索引](design_guide.md)
