# 四領域架構與 PDF 匯入驗收

狀態：設計提案。現行入口見 [HTTP handlers](../../backend/internal/business/objectimport/adapter/http/routes.go)。現有 Parse 接受使用者同意與選定頁面，Get 讀取草稿，Review 與 Confirm 接受 ExpectedVersion。下列跨領域契約描述目標行為；上線前應從實際入口逐項驗證。

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

1. 先對照[前端擷取測試](../../frontend/miniapp/tests/pdf-extraction.test.tsx)、[表格測試](../../frontend/miniapp/tests/pdf-table.test.tsx)與[匯入測試](../../frontend/miniapp/tests/object-import.test.tsx)，從現有入口跑通擷取、確認、重送。以實際儲存的工項數與來源頁驗收。
2. 接入知識查詢後，以同一批樣本比較每百筆候選的人工修正數與錯誤自動採納數；另測撤回、跨組織隔離及新版取代舊版。
3. 只對仍需動態補查的案例接入 Agent，記錄找到證據、證據不足、越權拒絕及步數耗盡的結果。以相同樣本對照固定流程的正確率、耗時與模型呼叫數。
4. 用匯入 ID 關聯候選版本、執行識別碼、來源頁、知識版本、模型與提示版本及結果。日誌記錄識別資訊與計數，避免記錄整頁原文。上線時保留停用補查、撤回知識版本及回到人工確認的路徑。
