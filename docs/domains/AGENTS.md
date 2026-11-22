# 領域規則及驗收工作指引

修改本目錄時沿用[上層指引](../AGENTS.md)，按任務查[本層索引](README.md)。

## 職責與邊界

出勤、工時計算、天氣與摘要等跨模組功能流程及業務規則。
同一功能的前端、API、後端規則集中描述；實作位置以連結指向模組，不複製多份規格。區分已實作、待驗收與未來規劃。

## 維護與驗證

更新索引與驗證連結依[文件維護指引](../AGENTS.md#驗證)。

新增或修改功能文件時，使用 [操作契約範本](../contracts/business-operations.md#功能操作契約範本)，區分既有證據、目標行為與待確認規則，不擅自改變產品範圍。

在對應功能文件回答六個問題：誰操作、操作什麼、做什麼、有沒有權限、符合哪些業務規則、成功後改變或回傳什麼；補充 Context、Authority、Decision 與依風險需要的 Event／Audit。
區分系統允許／拒絕與業務核准／駁回；Merge 僅於採納提案時使用。

業務模型與契約名稱依[命名規範](../standards/naming.md#業務實體集合與-api-命名)。

## 文件歸屬

每個領域以 README.md 導覽責任、核心流程、契約及驗收。單一主規格使用 design.md；已有多個明確主題時保留具體名稱。跨領域保證在 docs/architecture/contracts 維護一份，相關領域引用同一來源。待評估設計放在 docs/proposals。搬移時一併校正章節錨點、圖檔、Mermaid click 與 API 能力清單的路徑。

獨立領域或容器入口固定為本目錄下的「領域名/README.md」。Repository 容器內的 Issue、Discussion、Object（含子物件與範本）文件放在 `repository/<能力>/`，由 Repository README 統一導覽；Project 與 Journal 保留獨立入口。能力子目錄沿用本指引，文件歸組不改變程式模組或業務權責。首頁先列狀態、設計、核心流程、驗收依據、程式目錄與維護指引；提案沿用同一入口，正文連回 proposals 的單一來源。新增或調整狀態時同步更新[全域矩陣](README.md#domain-matrix)。
