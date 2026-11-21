# backend/internal 工作指引

## 閱讀與修改入口

先查閱[architecture](architecture)、[靜態頁面](transport/http/browser)、[business](business)、[定位診斷](transport/http/geolocation)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## Bounded Context 分層範本

以下是單一 Bounded Context 的**標準參考結構，實際按需採用**。此倉庫的業務 Context 位於 `backend/internal/business/<context>/`。簡單功能可維持較少的套件，只有出現明確的分工需求時才新增對應分層或子目錄。無須補齊整棵目錄樹，也無須僅為符合範本而搬移既有程式。

```text
internal/business/
└── <Bounded Context>/
    ├── domain/                  # 領域層
    │   ├── entity/              # Entity / Aggregate Root
    │   ├── valueobject/         # Value Object
    │   ├── service/             # Domain Service
    │   ├── event/               # Domain Event
    │   └── policy/              # Domain Policy
    ├── application/             # 應用層
    │   ├── command/             # 寫入操作 / Use Case
    │   ├── query/               # 查詢操作 / Use Case
    │   ├── port/                # Output Port / Interface
    │   └── dto/                 # Application DTO
    └── adapter/                 # 適配器層
        ├── http/                # HTTP Handler
        ├── line/                # LINE Webhook / API
        ├── firestore/           # Database Adapter
        ├── messaging/           # Message Queue Adapter
        └── scheduler/           # Scheduled Trigger
```

依賴由外向內。應用層協調領域規則並定義所需介面，適配器呼叫使用案例或實作輸出介面，領域層保持獨立於 HTTP、LINE 與 Firestore。由 `bootstrap/` 組裝依賴。

## 修改與驗證

業務授權由資源所屬領域判定。涉及一致性的讀寫，應指出授權與狀態檢查所在的交易邊界。修改公開 HTTP 格式時先更新 OpenAPI，修改保存格式時則核對既有資料的相容性。Go 變更依[後端指引](../AGENTS.md#驗證)執行檢查，並跑受影響的測試。依賴 Firestore 模擬器的測試需先設定 `FIRESTORE_EMULATOR_HOST`。

## 相關規格

[現行實作圖：格式與資料相容](../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../docs/diagrams.md#未來設計)
