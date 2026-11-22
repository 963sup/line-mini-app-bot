# DDD 與六邊形架構圖解

開發與維運指引。依問題選圖，主題入口見[DDD 設計索引](README.md#mermaid-圖表)。

下圖為設計參考，從業務問題推進至模型、用例與驗證。模型和上下文可以隨反例反覆調整。

```mermaid
flowchart LR
  Problem["領域問題與通用語言"] --> Boundary["限界上下文與整合契約"]
  Boundary --> Model["領域模型與不變條件"]
  Model --> UseCase["用例與 Port"]
  UseCase --> Adapter["Adapter 與組裝"]
  Adapter --> Evidence["領域、契約與架構測試"]
  Evidence -->|反例| Model
```

### 戰略設計：從問題到模型邊界

開發與維運指引。子域劃分業務問題，限界上下文劃分模型適用範圍。先用語言、規則與責任決定哪些模型應分開，再畫出具體契約。子域分類不會自動決定上下文的數量或邊界。

```mermaid
flowchart TB
    DOMAIN["業務領域"] --> SUB["子域：核心、支援、通用"]
    SUB --> DESIGN["依通用語言、業務規則及責任劃分模型"]
    DESIGN --> A["限界上下文 A"]
    DESIGN --> B["限界上下文 B"]
    A -->|"示例：A 提供已確認的資格事實，B 使用"| B
```

最後一條箭頭表示契約的影響方向，由提供者指向使用者。若 B 另有契約提供給 A，應以另一條具名關係表達。詳見 [限界上下文](bounded-context.md)與[上下文映射圖](context-map.md)。

### 戰術設計：從不變條件到聚合

開發與維運指引。圖中連線表示設計推導，無程式呼叫含義。實體和值物件各按業務語意建模，只有共同保護不變條件所需的狀態才納入同一聚合。

```mermaid
flowchart TB
    BC["限界上下文"] --> MODEL["領域模型"]
    MODEL --> ENTITY["實體：依身分追蹤生命週期"]
    MODEL --> VALUE["值物件：依內容判等"]
    MODEL --> RULE["找出操作成功後必須成立的不變條件"]
    RULE --> BOUND["推導保護規則所需的最小一致性邊界"]
    BOUND --> ROOT["聚合根：對外異動入口"]
```

聚合可以只有根實體。工廠、領域服務、政策與事件按需要採用，詳見[聚合](aggregate.md)與[不變條件](invariant.md)。

### 六邊形架構：一次用例如何執行

開發與維運指引。實線表示執行時呼叫。HTTP、LINE 和排程是可能的入站方式；下圖以讀寫資料為例，外部服務與訊息系統也可由各自的出站 Adapter 接入。

```mermaid
flowchart LR
    IN["入站 Adapter<br/>HTTP、LINE 或排程"]
    subgraph CORE["應用核心"]
        API["入站 Port：用例方法"]
        APP["Application：協調用例"]
        DOMAIN["Domain：模型與規則"]
        PORT["出站 Port：用例所需能力"]
        API --> APP
        APP --> DOMAIN
        APP --> PORT
    end
    OUT["出站 Adapter<br/>例如 Firestore Adapter"]
    DB["Firestore"]
    IN --> API
    PORT --> OUT
    OUT --> DB
```

程式碼依賴方向需另外核對：出站 Adapter 依賴並實作核心需要的契約，由 Bootstrap 注入。參見[程式碼依賴圖](hexagonal-architecture.md#程式碼依賴)。

[六邊形架構](hexagonal-architecture.md#執行與程式碼依賴)分開呈現執行與程式碼依賴，[交易](transaction.md#提交與發布)呈現提交和發布的時間差。可將 Markdown 圍欄內的圖表原始碼存為 `diagram.mmd`，從儲存庫根目錄執行。

```sh
bash scripts/tooling/tool-exec.sh mermaid render diagram.mmd --format svg > diagram.svg
```

使用專案鎖定的 `mermaid@13.0.0-alpha.3`，版本及操作方式見 [Mermaid 專案指引](../../development/tooling/shared/mermaid.md)。概念資料以匿名 Context7 查閱 [Domain-Driven Hexagon](https://github.com/sairyss/domain-driven-hexagon)，文件索引未固定 revision。Context7 另查得 [Mermaid CLI](https://github.com/mermaid-js/mermaid-cli) 的 `mmdc` 用法，渲染命令以專案鎖定版本為準。
