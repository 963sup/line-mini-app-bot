# LINE 工程平台架構

以下圖表記錄工程平台提案。現行服務與已接線的功能見 [執行架構](../../architecture/runtime/overview.md) 與 [LINE 旅程](../../integrations/line/journey.md)。

## 平台總覽

未來設計

```mermaid
flowchart TD
    U["LINE 使用者"]

    subgraph L["LINE 平台"]
        D["LINE Developers · Provider"]
        M["Messaging API Channel"]
        C["LINE MINI App Channel"]
        OA["LINE Official Account"]
        DM["1:1 私訊"]
        GR["LINE 工程群組"]
        LIFF["MINI App · LIFF SDK"]
        D --> M
        D --> C
        M --- OA
        OA --- DM
        OA --- GR
        C --> LIFF
    end

    subgraph APP["Cloud Run · 單一 Go 服務"]
        WH["Webhook Handler"]
        WEB["Next.js / React 靜態頁面"]
        AUTH["Identity / Session / Authorization"]
        API["HTTP API"]
        CORE["Business Use Cases"]
        FILE["File Service"]
        JOB["通知與背景工作入口"]
    end

    subgraph DATA["Google Cloud"]
        DB[("Firestore · 結構化資料")]
        ST[("Firebase Storage / Cloud Storage · 檔案")]
        TASK["Cloud Tasks / Scheduler"]
        SM["Secret Manager"]
    end

    U --> DM
    U --> GR
    U --> LIFF

    DM --> M
    GR --> M
    M -->|"Webhook"| WH
    WH --> CORE
    WH --> JOB
    CORE -->|"Reply / Push"| M

    LIFF -->|"取得網頁"| WEB
    LIFF -->|"ID Token / Session"| AUTH
    WEB -->|"API 請求"| API
    API --> AUTH
    API --> CORE
    API --> FILE

    CORE <--> DB
    AUTH <--> DB
    FILE <--> DB
    FILE <--> ST

    TASK --> JOB
    JOB --> DB
    JOB --> M
    SM -.->|"機密設定"| APP
```

## LINE 渠道與身分

未來設計

```mermaid
flowchart TD
    P["Provider · 同一服務提供者"]

    M["Messaging API Channel"]
    A["LINE MINI App Channel"]

    OA["LINE Official Account"]
    BOT["Webhook / Reply / Push"]
    MINI["LIFF ID / MINI App URL"]
    ID["LINE UserID · 同 Provider 識別"]

    P --> M
    P --> A
    M --> OA
    M --> BOT
    A --> MINI
    M -.-> ID
    A -.-> ID
```

## 資源階層

未來設計

```mermaid
flowchart TD
    U1["User A · UserID"]
    U2["User B · UserID"]
    U3["User C · UserID"]

    G["LINE 工程群組 · GroupID"]
    R["Repository · RepositoryID"]

    A["工區 · A 區"]
    B["工區 · B 區"]
    C["工區 · C 區"]

    O1["Object · 第一套電力系統"]
    O2["Object · 第二套電力系統"]
    O3["Object · 第三套電力系統"]

    S1["SubObject · 配線"]
    S2["SubObject · 測試"]

    P["Project · 電力系統驗收"]
    I["Issue · 測試異常"]
    D["Discussion · 複測討論"]
    F["Attachment · 工程附件"]
    J["Journal · 施工日誌"]

    U1 -->|ownerId| PR["個人 Repository"]
    U1 -->|ownerId| PP["個人 Project"]
    PP -.->|連結| PR
    U1 --> G
    U2 --> G
    U3 --> G

    G -->|"1:1 綁定"| R
    R --> A
    R --> B
    R --> C
    G -->|擁有| P
    P -.->|連結來源庫| R
    P -.->|連結來源庫| R2["另一個 Repository"]
    R --> I
    R --> D
    R --> F
    R --> J

    A --> O1
    B --> O2
    C --> O3
    O1 --> S1
    O1 --> S2

    P --> E["工程引用（每筆指向一個對象）"]
    E -.->|Object| O1
    E -.->|SubObject| S2
    E --> V["InspectionSubmission · 驗收送件"]
    P --> IR["問題引用"]
    IR -.-> I
    P --> DR["討論引用"]
    DR -.-> D
```

## 加入工程群組

未來設計

```mermaid
sequenceDiagram
    actor User as LINE 使用者
    participant LINE as LINE Platform
    participant Go as Cloud Run Go
    participant DB as Firestore

    User->>LINE: 邀請官方帳號加入工程群組
    LINE->>Go: Join Webhook (GroupID)
    Go->>Go: 驗證 Webhook 簽章
    Go->>DB: 查詢 GroupID 綁定
    alt 首次加入
        Go->>DB: 交易建立 Repository 與綁定
    else 重新加入
        Go->>DB: 重新啟用既有 Repository
    end
    Go-->>LINE: Messaging API 回覆
    LINE-->>User: 顯示工程儲存庫入口
```

## 私訊查詢工程

未來設計

```mermaid
flowchart TD
    U["使用者傳送：我的工程"]
    W["Messaging API Webhook"]
    I["取得 UserID"]
    DB["查詢使用者曾關聯的 Repository 候選"]
    V["逐一確認目前群組成員與讀取權限"]
    L["顯示可存取 Repository 清單"]
    S["使用者選擇工程"]
    M["開啟 MINI App"]
    R["讀取工程資訊"]

    U --> W --> I --> DB --> V --> L --> S --> M --> R
```

## MINI App 入口

未來設計

```mermaid
sequenceDiagram
    actor User as 工程人員
    participant LINE as LINE 群組
    participant Go as Cloud Run Go
    participant App as MINI App
    participant DB as Firestore

    User->>LINE: 點選工程管理卡片
    LINE->>App: 開啟 MINI App 連結
    App->>Go: 提交啟動參照與 LINE ID Token
    Go->>Go: 驗證 Token / Channel
    Go->>DB: 查詢啟動參照對應的 GroupID
    Go->>DB: 查詢 GroupID 對應 RepositoryID
    Go->>LINE: 查核 UserID 的群組成員資格
    LINE-->>Go: 成員核驗結果
    Go->>Go: 檢查 Resource / Action / Policy
    Go-->>App: 建立 Session / 返回授權範圍
    App->>Go: 載入 Repository 資料
    Go->>DB: 讀取庫內工區與物件，以及群組有權查看的 Project
    DB-->>Go: 業務資料
    Go-->>App: 顯示工程管理畫面
```

## 工程通知

未來設計

```mermaid
sequenceDiagram
    actor User as 工程人員
    participant App as MINI App
    participant Go as Go API
    participant DB as Firestore
    participant ST as Firebase Storage
    participant LINE as LINE Messaging API

    User->>App: 選擇 Project 與測試工作
    App->>Go: 讀取工作及工程物件
    Go->>DB: 查詢授權與資料
    DB-->>Go: 回傳內容
    Go-->>App: 顯示工作表單

    User->>App: 填寫結果並選擇照片
    App->>Go: 建立附件上傳請求
    Go->>Go: 驗證使用者、範圍與檔案限制
    Go-->>App: 核准受限上傳
    App->>ST: 上傳照片
    ST-->>App: 上傳完成

    App->>Go: 提交附件與測試結果
    Go->>ST: 核對附件實際存在
    Go->>DB: 交易保存結果、附件參照、通知事件
    DB-->>Go: 提交完成
    Go-->>App: 返回成功結果

    Go->>LINE: 非同步送出群組通知
    LINE-->>User: 群組顯示驗收待審卡片
```

## 背景派送

未來設計

```mermaid
flowchart TD
    A["業務操作成功"]
    B["Firestore Transaction"]
    C["保存業務資料及 Outbox"]
    D["可靠派送工作"]
    E["Cloud Tasks / Worker"]
    F["Messaging API"]
    G{"發送結果"}
    H["標記完成"]
    I["依規則重試／人工處理"]

    A --> B --> C --> D --> E --> F --> G
    G -->|"成功"| H
    G -->|"失敗"| I
    I -.->|"可重試"| E
```
