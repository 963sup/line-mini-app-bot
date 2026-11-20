# backend/internal/business/objectimport 工作指引

## 目錄用途

處理工程清單匯入，從來源解析、候選驗證和人工審核，到確認後透過 Object 使用案例建立物件。

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：自動匯入操作契約](../../../../docs/domains/repository/object-imports.md#自動匯入操作契約) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
