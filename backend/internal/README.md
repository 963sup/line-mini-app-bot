# backend/internal 目錄導覽

## 用途與範圍

保存 Go 後端的內部套件。業務功能放在 `business/`，服務組裝放在 `bootstrap/`，其餘套件提供身分驗證、請求傳輸、導覽及資料存取等共用基礎能力。

## 架構與責任邊界

依目錄用途提供後端入口或共用能力；服務接線由 bootstrap 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

先查閱[architecture](architecture)、[browser](transport/http/browser)、[business](business)、[diagnostics](transport/http/geolocation)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：格式與資料相容](../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## Bounded Context 分層範本

以下是單一 Bounded Context 的**標準參考結構，實際按需採用**。此倉庫的業務 Context 位於 `backend/internal/business/<context>/`。簡單功能可維持較少的套件，只有出現明確的分工需求時才新增對應分層或子目錄。無須補齊整棵目錄樹，也無須僅為符合範本而搬移既有程式。

```text
internal/
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

閱讀一項功能時，可從適配器找到入口，再追到應用層的使用案例及領域層的規則。外部請求與資料格式由適配器轉換，跨層接線由 `bootstrap/` 負責。

## 驗證入口

Go 檢查與測試方式見[後端工作指引](../AGENTS.md#驗證)。涉及 Firestore 交易的整合測試，請查閱對應套件的測試與模擬器前置條件。

## 閱讀入口

從 [architecture](architecture)、[browser](transport/http/browser)、[business](business)、[diagnostics](transport/http/geolocation)、[httpapi](httpapi)、[identity](identity) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：格式與資料相容](../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../../docs/diagrams.md#未來設計)

## 共用技術與入口

[platform](platform) 提供文件鍵、分頁游標與 gRPC 錯誤轉換。[bootstrap](bootstrap) 建立 client 並接線。[navigation](navigation) 組裝跨功能選單。[httpapi](httpapi) 維護 OpenAPI 生成契約與路由分派。
