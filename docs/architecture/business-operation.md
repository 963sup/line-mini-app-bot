# 資源導向的通用業務操作模型

狀態：採用於後續新增或修改的業務操作；不代表既有功能已全部符合或完成驗收。
本頁補充 [開發原則](principles.md)，沿用模組化單體與輕量 Hexagonal，不擴大當前產品範圍。

## 模型與語意

系統以資源為管理對象、操作為行為單位、主體為責任來源，以授權與業務規則約束執行。

| 概念 | 必須回答 |
| --- | --- |
| Actor（Subject） | 誰發起？可信身分從何而來？ |
| Resource | 操作什麼？歸屬、租戶／群組及存取範圍是什麼？ |
| Action | 執行哪個明確的業務行為？ |
| Context | 依賴哪些時間、版本、環境及目前狀態？ |
| Authority | 執行、審核與決定權從何取得？適用範圍是什麼？ |
| Policy | 哪些授權條件與業務規則允許或限制操作？ |
| Decision | 系統依上述條件允許或拒絕什麼？ |
| Result | 回傳哪些資料？發生哪些合法狀態轉換與必要後續行為？ |

Authority 表達權限來源與範圍；Policy 包含授權政策及領域規則，實作時仍須區分兩者責任。
系統允許執行 Reject 與業務申請被駁回是不同決定：駁回可以是成功的操作。
Merge 僅在採納提案時作為 Action 或狀態轉換，不是固定 Decision 子項。
讀取結果可只有資料，不必改變狀態；Event／Audit 依風險補充，不強迫每次讀取產生事件。

## 實作邊界

- 後端取得可信 Actor；不採信客戶端自行宣告的操作者、角色或審核權。
- 在使用案例協調授權與業務規則；不同入口的相同業務操作共用必要使用案例。
- 以實際 Resource 歸屬與範圍授權，預設拒絕未授權操作。前端顯示權限不能取代後端檢查。
- 各領域保留自己的規則與狀態模型；重用 identity 與既有規則，有實際跨領域需求才抽取共用機制。
- HTTP、LINE 與資料庫 adapter 隔離外部技術。需與寫入一致的權限、版本及狀態檢查須與交易協調，不先檢查後無條件寫入。
- 重要寫入定義失敗、重送與併發語義；必要追溯記錄操作者、對象、時間與結果，不記錄 token 或不必要個資。
- 外部副作用定義失敗處理；不可在可能重跑的交易 callback 中直接發送通知。
- 不強制每個概念各建模組，不為此建立萬用引擎、固定七層、完整 RBAC 階層或事件平台。

## 功能操作契約範本

在對應 docs/features 文件填寫，每個操作可以使用一張表，不必另建程式檔案：

| 欄位 | 內容 |
| --- | --- |
| Actor / Resource / Action | 可信身分、資源歸屬與行為 |
| Context / Authority | 時間、版本、環境；權限來源與範圍 |
| Policy / Decision | 授權與領域規則分列；允許及拒絕條件 |
| Result | 回傳資料、合法轉換與失敗時不應改變的資料 |
| 重送／併發 | 適用性、重複請求及過期版本語義 |
| Event / Audit | 依風險說明必要紀錄或不適用原因 |
| 實作與證據 | 檔案／符號、實際檢查結果、跳過項目與待確認事項 |

六個必答問題仍是：誰操作、操作什麼、做什麼、有沒有權限、符合哪些業務規則、成功後改變或回傳什麼。
資訊不明確時列為待確認，不自行發明產品規則。

## 持續檢查與修正

依 [代理工作流程](../agents/workflow.md#business-operation-compliance) 執行 Graphify → Repomix／原文核對 → Context7 技術查證 → 範圍內修正 → 驗證。
工具輸出是證據，不是正確性證明或擴大修改範圍的授權。

## 補卡核准對照範例

以下為 2026-11-11 本機原始碼的靜態對照，不是正式環境或整合驗收。產品細節以 [查詢與更正](../features/attendance-queries.md) 及 [工時計算](../features/worktime.md) 為準。

| 欄位 | 現有實作對照 |
| --- | --- |
| Actor / Resource / Action | 後端傳入的 actor；channel 內的補卡申請、群組與申請人出勤；ApproveCorrection |
| Context / Authority | 交易讀取申請與審核群組；審核者須有該群組權限，不能審核自己 |
| Policy / Decision | 檢查既有決定、版本／事件序列、時間與重疊；不符合則拒絕，不更新有效出勤 |
| Result | 套用補卡結果、增加 revision、更新申請狀態；原始事件不改寫 |
| 重送／併發 | 交易內讀取狀態；已存在核准 audit 的再次核准返回成功；不同終態決定被拒絕 |
| Event / Audit | 核准交易建立 proposal、before、after、reviewerKey 與 approvedAt 證據 |
| 實作與證據 | `backend/internal/business/attendance/firestore/attendance_correction.go` 的 `decideCorrection` 與 `backend/internal/business/attendance/corrections/correction.go` 的 `ApplyCorrection` 靜態核對；本次未執行應用測試、正式資料操作或部署 |

此範例不宣稱所有入口已符合模型。後續修改時仍須追查 actor 的可信來源、完整呼叫路徑及適用驗證結果。

Codex 指引機制依據：經 Context7 `/openai/codex` 查得的 [上游 AGENTS.md 規則](https://github.com/openai/codex/blob/main/codex-rs/protocol/src/prompts/base_instructions/default.md)。
2026-11-11 查證，來源為 main，未固定提交；根及巢狀指引按作用範圍適用，使用者與主機指令優先。文件連結不等於自動載入，角色必須明確閱讀。
