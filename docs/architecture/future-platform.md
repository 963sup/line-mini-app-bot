# 未來平台目錄規劃

> 狀態：規劃。以下是完整責任配置，非現有檔案清單；按交付階段建立，避免空目錄與空介面。
> 路徑從 repository root 起算；每個 Go 子目錄是獨立 package，不假設父目錄自動包含子目錄。

本頁是後續平台參考配置，不是首版必建目錄。依 [ADR 0002](decisions/0002-baseline.md)
及 [分階段落地](../development/foundation.md) 先穩定基礎，再新增出勤需要的套件。
Worker、Garr、管理後台與通用模板系統目前不實作，也不新增測試工程。
已提供的接入能力位於 cmd/server、cmd/richmenu、cmd/firestorecheck 與 internal/lineui；
Rich Menu 僅提供明確執行的 CLI，不預建管理後台或完整四層模組。

## 第二層與第三層目錄

```text
cmd/
  server/
    main.go                 # HTTP 組裝與啟停
  worker/
    main.go                 # 常駐工作處理入口
internal/
  direct/
    domain/                 # User、DirectConversation、個人值型別與規則
    application/            # 個人 Use Cases；ports.go 定義所需介面
  group/
    domain/                 # Group、Role、群組狀態與規則
    application/            # 群組 Use Cases；ports.go 定義所需介面
  identity/
    domain/                 # Actor、Session、channel 身分
    application/            # 登入、session、驗證需求介面
  templates/
    domain/                 # 模板、版本、發布狀態
    application/            # 驗證、發布、回復、讀取已發布內容
  richmenu/
    domain/                 # Menu、Alias、指派目標
    application/            # 建立、上傳、發布、指派、回復
  jobs/
    domain/                 # Inbox、Outbox、UserTarget、GroupTarget、Lease
    application/            # 領取、續租、重試、完成；ports.go
  adapters/
    http/                   # 第三層：入站 HTTP adapter 群組
    line/                   # 第三層：LINE SDK、token 驗證與訊息呈現
    firestore/              # 第三層：repository、交易、租約與 DTO 轉換
    garr/                   # 第三層：pool、backoff、breaker 包裝
  web/
    templates/              # Go html/template：MINI App 與 admin 頁面
    static/                 # CSS、LIFF JS glue；使用 embed 管理
  bootstrap/
    server/                 # HTTP clients、routes、依賴與 shutdown 組裝
    worker/                 # clients、dispatch、pool 與 shutdown 組裝
  config/
    config.go               # 非機密設定解析與必要值驗證
```

## Adapter 的下一層拆分

adapters 本身已是第二層，以下明列第四層，避免把所有 adapter 塞進巨大 package。

```text
internal/adapters/http/
  webhook/                  # 原始 body 驗簽、事件轉換、Inbox 接收
  miniapi/                  # MINI App 個人／群組 API
  admin/                    # 管理頁面、發布操作
  middleware/               # session、CSRF、request ID、body limit
  health/                   # liveness/readiness
internal/adapters/line/
  messaging/                # Reply、個人 Push、群組 Push
  rendering/                # 業務結果 → Flex/Text/Quick Reply；不放授權
  identity/                 # LINE token 驗證 HTTP client
  richmenu/                 # Rich Menu API client
internal/adapters/firestore/
  direct/                   # 個人資料 repository
  group/                    # 群組資料 repository
  identity/                 # session 與身分關聯
  templates/                # 模板與發布版本
  richmenu/                 # 選單狀態與發布紀錄
  jobs/                     # Inbox、Outbox、租約與原子提交
internal/adapters/garr/
  execution/                # 程序內併發、退避、斷路器
```

MINI App Service Message adapter 在資格與需求確立後新增，不預建空實作。
每個 package 內以 `model.go`、`service.go`、`ports.go`、`*_test.go` 等責任命名；不要求每包都有同名檔案。

## 擴充規則

- direct 與 group 不互相 import；兩者相同的業務能力只有在實際成立後才抽成獨立模組。
- 不建立 shared/utils/common 雜物包；共用型別須有清楚責任，例如 jobs/domain 的發送工作目標。
- application 不能 import adapters；SDK 與 Firestore 型別留在 adapter。
- 機密由部署環境注入，config 不保存 token；工具設定仍由 mise.toml 管理。

依賴允許表見 [dependencies.md](dependencies.md)，實作順序見 [delivery.md](delivery.md)。
