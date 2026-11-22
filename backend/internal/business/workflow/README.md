# workflow

[領域文件](../../../../docs/domains/workflow/README.md) · [全域索引](../../../../docs/domains/README.md#domain-matrix)

管理明確的步驟順序、等待條件及執行進度。

[確認與續跑用例](application/confirmation.go)已定義流程儲存與業務命令的介面。先保存操作者、候選與提交識別碼，再允許續跑。業務命令負責重新授權及回傳可重用的收據。持久化 Adapter 與產品入口尚未接線。

## 共用接線

在具名業務用例中以 [Preparation](application/prepare.go) 接既有的唯讀 Operations、可選的 Knowledge 查詢、可選的 Agent 補查及結果組裝。可信呼叫端帶入操作者與目標。用例以程式條件決定是否查知識或呼叫 Agent；任一步失敗即回傳。只有查閱需求時可直接呼叫原業務查詢，或使用不配置可選步驟的 Preparation。

要等待人工確認時，由呼叫端以持久化介面建立流程及不可變候選，再以流程版本呼叫 `Stage`。順序為 `Prepare → Stage → Confirm → Resume`。準備結果本身不會寫入業務。各業務只需接自己的具名讀取、候選組裝和既有命令。通用流程不依賴日誌、出勤或文件資料型別。

[跨能力接線測試](application/prepare_test.go)使用既有 Investigator 和現有 Operations／Knowledge 契約，覆蓋完整輸入略過可選步驟、資料不足時依序補查，以及讀取遭拒後停止。確認與續跑另以測試替身覆蓋。尚未配置任何產品用例或持久化 Adapter。

[確認與恢復流程圖](../../../../docs/proposals/workflow.md#design-diagram) · [維護指引](AGENTS.md)

[上層導覽](../README.md)
