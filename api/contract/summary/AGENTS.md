# 摘要事件與內部工作契約

## 閱讀與修改入口

先查閱[worker.yaml](worker.yaml)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[摘要功能](../../../docs/domains/summary/summary.md) · [現行實作圖：資料與可靠性](../../../docs/domains/summary/summary.md#資料與可靠性)

適用於本目錄及其子目錄；繼承 [上層指引](../AGENTS.md)，並遵循其引用的根目錄規則。此處只補充本範圍責任；使用者與主機指令優先，不增加外部寫入授權。

## 職責與邊界

- 先讀 [既有入口契約](README.md) 與 [摘要功能](../../../docs/domains/summary/summary.md)。訊息收集、摘要指令與內部 worker 分開描述。
- worker 契約由 `../internal-openapi.yaml` 組合，生成內部 DTO；不放入公開瀏覽器 client，不把工作身分替換成瀏覽器 Session。
- 保留既有輸入、請求上限、大小寫相容、回應、重送、租約及失敗語義；不複製 LINE 官方事件 schema。
- 只有新功能規格明確需要時才新增公開 Summary API。

[現行實作圖：資料與可靠性](../../../docs/domains/summary/summary.md#資料與可靠性)
