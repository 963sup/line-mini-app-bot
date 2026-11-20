# backend/internal/server 工作指引

## 目錄用途

組裝後端服務、業務依賴與各功能路由，將 HTTP、LINE 與資料存取轉接器接到業務層。

## 工作方式

遵循[上層指引](../AGENTS.md)。修改套件行為時，檢查呼叫端與受影響的 Go 測試。 實作入口見[本目錄導覽](README.md)。

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../../docs/diagrams.md#未來設計)
