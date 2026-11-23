# api/contract/project 目錄導覽

定義工程專案、欄位、項目、來源引用與檢查送件的 API 契約。

維護本能力的操作與 schema；根規格負責組合，公開格式由契約產生。

## 閱讀與修改入口

從 [fields.yaml](fields.yaml)、[inspection-submissions.yaml](inspection-submissions.yaml)、[items.yaml](items.yaml)、[owner-projects.yaml](owner-projects.yaml)、[projects.yaml](projects.yaml)、[references.yaml](references.yaml) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## Owner、空白專案與複製 API

專案列表依 Owner 查詢。Owner 僅接受使用者或組織，後端依對應實體核驗權限。複製操作須識別來源專案、目的地 Owner 與新專案名稱，核驗來源讀取權與目的地建立權，成功時回傳新專案 ID 與目的地 Owner。相同請求以同一防重鍵重試應取得同一副本；同鍵改用其他參數應回報衝突。複製內容依[領域規格](../../../docs/domains/project/project.md#複製到另一個-owner)，副本可在沒有初始來源的情況下開啟。

建立請求可省略 defaultRepositoryId；省略時建立零個來源的專案，詳情回應省略該欄位，已連結來源清單為空。提供初始來源時仍核驗資格，既有專案讀回原有來源。後續連結其他儲存庫時沿用逐庫授權。

[建立請求](owner-projects.yaml) 的 copyFrom 指定來源 Owner 與專案。驗收涵蓋[專案入口與跨庫案例](../../../docs/domains/project/journey.md)。

## 相關規格

[現行實作圖：最小來源引用骨架](../../../docs/domains/project/project.md#最小來源引用骨架)

[上層導覽](../README.md)
