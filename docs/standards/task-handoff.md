# Codex 任務範本

在側欄聊天說「交接目前工作」即可取得以下格式。欄位定義與證據重用條件由[交接規格](../development/workflow/handoff.md#minimal-handoff-and-evidence-reuse)維護；跨聊天操作依[側欄交接流程](../development/workflow/handoff.md#sidebar-handoff)。

```text
Goal｜目標與完成條件
Ownership｜儲存庫、分支、提交、相關未提交檔案與負責範圍
Inputs｜必讀檔案、符號、設計或操作限制
Evidence｜已確認的事實、來源、可讀的產物位置與適用範圍
Checks｜已執行的命令、實際結果、未驗證項目
Open issues｜阻礙、待決事項、下一個動作
```

## 範例：接手補卡重送問題

以下示範唯讀調查剛開始時如何交接。實際交接時填入當下的版本及結果。

```text
Goal｜判斷同一補卡核准請求再次送達時，會不會重複套用。
Ownership｜line-mini-app-bot／<分支>@<提交>；相關未提交檔案 <實際清單或無>；唯讀調查。
Inputs｜從出勤 application/command/review.go 的 ReviewCorrection 追到 domain/policy/review.go。
Evidence｜出勤目錄為 backend/internal/business/attendance/；交易實作在 adapter/firestore/attendance_correction.go。
Checks｜尚未執行測試；交易與重送結果待核對。
Open issues｜先讀 NeedsDecision 對既有核准紀錄的判斷，再對照交易測試。
```

涉及實作時，在 Inputs 補充用例、介面、函數簽名、允許的依賴與交易邊界；Checks 補充適用的成功、拒絕、重送、失敗與取消案例，以及外部驗證前置條件。

## 完整實作就緒條件

核對本次操作的 API、schema、狀態機與驗收案例，逐一確認角色及各分支的結果。涉及[平台交付提案的 P01–P07](../proposals/platform/delivery.md#產品決策閘門)時，只追蹤本功能依賴的決策，現行行為以[領域規格](../domains/README.md)為準。

若要新增到場判定或定位歷史，先依[定位契約](../integrations/geolocation/browser-api.md)確認 P08 的目的、精度與新鮮度、替代流程、保存期及同意。未決前維持既有功能範圍。
