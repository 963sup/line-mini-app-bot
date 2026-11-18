# 可靠性設計參考

本目錄的設計文件屬於通用 Inbox／Outbox 與後續平台提案，不表示目前已有持久接收、租約 Worker 或完整可靠投遞能力。現行行為以[功能規格](../../features/README.md)及 [HTTP 契約](../api/http-contracts.md)為準。

- [事件與發送狀態機提案](state-machines.md)

只在任務涉及持久工作、重試或平台設計時載入；目前必要檢查見[驗收矩陣](../../agents/acceptance.md)。[規劃入口](../../planning/README.md)
