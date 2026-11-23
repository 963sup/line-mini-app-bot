# 工作單元（Unit Of Work）

以用例契約協調同一原子範圍內的讀寫和提交。

## 設計時要確認

回傳成功前確認提交；可能重跑的 callback 避免外部副作用，交易資源留在 Adapter。

## 示例

用例透過同一 work 載入資格、保存邀請與收據，獨立的讀取 client 不能提供相同保證。

## 延伸閱讀

[transaction](transaction.md) · [driven-adapter](driven-adapter.md) · [DDD 設計索引](design-guide.md)
