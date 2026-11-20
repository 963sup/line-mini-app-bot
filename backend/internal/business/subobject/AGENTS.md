# backend/internal/business/subobject 工作指引

## 目錄用途

管理工程物件下的子物件，維護它與上層 Object 的歸屬關係。

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：物件與子物件歸屬](../../../../docs/domains/repository/objects.md#物件與子物件歸屬) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
