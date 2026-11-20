# 四領域架構與 PDF 匯入驗收

狀態：設計提案。現行入口見 [HTTP handlers](../../backend/internal/business/objectimport/adapter/http/routes.go)。現有 Parse 接受使用者同意與選定頁面，Get 讀取草稿，Review 與 Confirm 接受 ExpectedVersion。下列跨領域契約描述目標行為；上線前應從實際入口逐項驗證。

## 從選檔到工項

使用者的目標是建立可追溯且數量正確的工項。前端選擇檔案並呈現資料傳送方式；HTTP 驗證身分及解碼請求。自動匯入入口交由 Workflow 選取 Operations，objectimport 負責核對資源存取、凍結本次候選及處理重送。Excel 使用已授權的工作表讀取器；文字型 PDF 沿用頁面辨識及模型分類。Object 擁有最終工項寫入規則。

目前 Excel 的讀取與判定在後端。PDF 的本機選頁資料及原始活頁簿各自有明確輸入格式；Workflow 的分派不代表已辨識所有檔案。需要人工確認的列會保留候選，其他依現行自動匯入契約處理。

目標流程中，只有「缺哪個已發布的版型對應」才查 Knowledge；只有「還能去哪個允許的來源找到證據」才委派 Agent。補查結果仍回 objectimport 驗證。證據不足時交回人工，不讓模型直接寫入。這兩個選擇性分支尚未接到自動匯入入口。

## 責任與依賴

```mermaid
flowchart TD
    ui["使用者或 API"] --> wf["Workflow：流程進度、等待與恢復"]
    ui --> op["Operations：具名操作入口"]
    wf --> op
    wf -->|委派補查| ag["Agent：選擇下一步、回傳證據"]
    ag -->|限權操作| op
    op --> oi["objectimport：授權、候選版本、驗證與寫入"]
    op --> kn["Knowledge：授權、審核與發布版本"]
    oi --> model["模型轉接器：產生候選"]
```

箭頭表示執行時呼叫。Operations 委派資源領域執行檢查；引用知識的任務記下版本。固定流程可直接執行擷取；表頭位置確定時可直接讀取，只有補查方式不明確時才委派 Agent。函數契約分見 [Operations](operations.md)、[Workflow](workflow.md)、[Agent](agent.md)、[Knowledge](knowledge.md)。

<a id="document-ingestion"></a>

## 多格式文件的入口與證據

現況與提案對照。現行文字型 PDF 在瀏覽器讀成選定頁面，後端 Parse 接收頁碼與文字。以下虛線是提案接點；來源讀取器放在實際讀檔的一端，bootstrap 只組裝後端依賴。

```mermaid
flowchart TD
    pdf["文字型 PDF"] --> pages["既有前端 PDF adapter：頁碼與文字"]
    pages --> api["既有 Parse API：選定頁面"]
    api --> wf["Workflow：準備待核對候選"]
    excel["XLSX：使用者選取活頁簿"] --> run["自動匯入 HTTP：核驗身分與資源"]
    run --> xlsx["Go workbook adapter：核對表頭、資料列與儲存格"]
    xlsx --> review["保存待核對類型，確認後以原識別碼寫入"]
    csv["CSV、TSV、TXT、Markdown 表格"] --> run
    run --> cells["Go 文字表格讀取器：欄位與來源行號"]
    cells --> review
    scan["掃描 PDF／圖片"] -.-> ocr["OCR adapter：頁面或影像區域"]
    source["多格式輸入契約：結構化內容與來源定位"]
    ocr -.-> source
    pages -.-> source
    source -.-> wf
    wf --> op["Operations：具型別的擷取操作"]
    op --> owner["objectimport：核驗身分、來源、候選與證據"]
```

目標契約以文件識別與修訂識別綁定內容。每個區塊保存段落或含表頭的表格；證據引用區塊及原始位置。PDF 用頁碼，Excel 用工作表與儲存格範圍，CSV 用記錄列與欄位，影像用影像識別與區域。讀取器無法解讀格式、頁面沒有文字，或工作表缺少必要欄位時，回傳可辨識的原因。授權與同意應先於後端付費辨識及模型呼叫。

### 手動解析呼叫時序

現行實作（本地骨架）。確認寫入有獨立入口。Knowledge 目前只有查詢介面。Agent 已提供可選的模型證據補查器，這條手動解析路徑尚未呼叫它。

```mermaid
sequenceDiagram
    actor U as 使用者
    participant F as 前端 PDF adapter
    participant H as HTTP adapter
    participant W as Workflow
    participant O as Operations 擷取操作
    participant I as objectimport
    participant M as 模型轉接器
    U->>F: 選擇文件與頁面
    F->>H: 送出頁碼、文字與同意狀態
    H->>W: 以已驗證身分呼叫 Parse
    W->>O: Extract
    O->>I: 委派現有解析用例
    I->>I: 檢查同意、資源權限與來源
    alt 通過且已設定模型
        I->>M: 擷取候選
        M-->>I: 欄位與引文
        I->>I: 驗證並整理候選
        I-->>H: 經原呼叫鏈回傳候選或驗證錯誤
    else 前置條件失敗
        I-->>H: 經原呼叫鏈回傳錯誤
    end
    H-->>U: 顯示候選供核對或提示失敗原因
```

跨格式驗收使用內容相同的 PDF 與 Excel 驗收：比較項次、名稱、數量、單位，再分別核對來源頁與儲存格。保留既有頁面 API 的相容映射；跨格式契約的驗證、API 與前端證據顯示須一併完成，才開放新入口。

## 一次確認和兩種競爭結果

```mermaid
sequenceDiagram
    actor U as 使用者
    participant W as Workflow
    participant O as Operations
    participant I as objectimport
    U->>W: 確認草稿 V1
    W->>W: 保存請求 K 與預期版本 V1
    W->>O: 寫入 K、V1
    O->>I: 帶可信身分執行
    alt 補查修訂 V2 先保存
        I-->>W: 版本衝突，無寫入
        W-->>U: 顯示 V2 並重新確認
    else V1 先提交成功
        I->>I: 原子保存工項與執行結果
        Note over W,I: 回應遺失，流程恢復
        W->>O: 重送同一 K、V1
        O->>I: 驗證授權並查已提交結果
        I-->>W: 原執行結果
        W-->>U: 匯入完成
    end
```

重新授權後才可回傳既有結果。執行識別碼綁定資源、操作與請求內容；同碼不同內容回報衝突。首次寫入在同一一致性邊界檢查版本與可寫狀態。已完成的匯入拒絕遲到的修訂。

## 共用詞條

| 詞條 | 在本設計中的含義 |
|---|---|
| 操作／操作執行 | 前者定義能力與契約；後者是一次具身分與資源範圍的呼叫。重試沿用同一執行識別碼。 |
| 候選版本／知識版本 | 前者識別本次待匯入內容；後者識別跨任務引用的已發布規則。分別由 objectimport 與 Knowledge 管理。 |
| 使用者確認／發布核准 | 確認允許寫入指定候選版本；發布核准允許特定範圍的後續任務引用規則。 |
| Knowledge／Persistent Memory | Knowledge 具備來源、版本與適用條件。持久化記憶是跨任務保留經驗；經審核的版型規則可成為知識。Workflow 的恢復點只屬本次執行。 |
| Retrieval／RAG | Retrieval 按權限與條件查回資料；將查回的證據交給模型生成候選，即為 RAG。 |
| Context Engineering | 使用案例選擇傳給模型的指令、證據、知識版本與上下文大小。文件內容當作資料，引用規則與原文衝突時交回確認。 |

## 可重現的驗收樣本

以下是合成頁面文字與預期結果，可作為 PDF 測試素材。進行端到端驗收時，將樣本排成指定頁數的 PDF，從實際上傳入口執行。

| 樣本 | 頁面輸入或事件 | 預期結果 |
|---|---|---|
| 一般工項 | 第 1 頁：項次 01，模板，12.5 m² | 一列，保留數量、單位及第 1 頁來源。 |
| 跨頁 | 第 1 頁末：02 不鏽鋼；第 2 頁首：欄杆 8 m | 同一項次一列，名稱「不鏽鋼欄杆」，來源涵蓋兩頁。 |
| 相似欄名 | 同列契約數量 20、累計完成數量 30；任務要求契約數量 | 取 20。其他組織的同名版型規則不可見。 |
| 缺證據 | 去掉表頭，留下兩個數字且沒有可信對應 | 標示待確認，不自動挑選數字或發布規則。 |
| 重送與競爭 | 確認後斷線重送；另測確認與修訂同時發生 | 重送只產生一次業務寫入；競爭結果符合上圖。 |

## 落地與觀測

1. 先對照[後端 PDF 擷取測試](../../backend/internal/business/objectimport/adapter/workbook/pdf_test.go)與[匯入測試](../../frontend/miniapp/tests/object-import.test.tsx)，從現有入口跑通擷取、確認、重送。以實際儲存的工項數與來源頁驗收。
2. 接入知識查詢後，以同一批樣本比較每百筆候選的人工修正數與錯誤自動採納數；另測撤回、跨組織隔離及新版取代舊版。
3. 只對仍需動態補查的案例接入 Agent，記錄找到證據、證據不足、越權拒絕及步數耗盡的結果。以相同樣本對照固定流程的正確率、耗時與模型呼叫數。
4. 用匯入 ID 關聯候選版本、執行識別碼、來源頁、知識版本、模型與提示版本及結果。日誌記錄識別資訊與計數，避免記錄整頁原文。上線時保留停用補查、撤回知識版本及回到人工確認的路徑。

<a id="storage-boundaries"></a>

## 儲存接線與權限邊界

現況與提案對照。實線是後端接線，虛線是後續的匯入決策。

```mermaid
flowchart TD
    W["Workflow：來源歧義"] -.-> K["Knowledge：依組織與版型查已發布規則"]
    K --> F["Firestore：檢查組織成員與發布投影"]
    I["objectimport：已授權 XLSX 匯入"] --> J["Firestore：匯入紀錄與來源雜湊"]
    I -->|設定來源 bucket| B["Cloud Storage：依內容雜湊保存原檔"]
    A["Agent：需要額外證據"] -.-> R["objectimport：重讀指定儲存格"]
    R -->|核對目前權限、建立者與位置| J
    R -->|核對下載內容雜湊| B
```

[規則投影與發布邊界](knowledge.md) · [附件留存與重讀契約](../domains/objectimport/design.md)
