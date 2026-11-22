# 修改 MINI App 前端

修改頁面、路由、互動、表單或元件時使用。先找出使用者進入的路由、需要的身分和前置資料。

1. 從 [MINI App 目錄導覽](../../../frontend/miniapp/README.md)找頁面、業務畫面及巢狀指引。元件另核對 [UI 指引](../../../frontend/miniapp/src/shared/ui/AGENTS.md)。
2. 從畫面讀寫或變更狀態的動作追到 API 及身分來源。涉及取消、過期回應或登入切換時，對照 [React 生命週期](../../architecture/runtime/react-lifecycle.md)。
3. 用[環境啟動表](../environment/README.md#啟動應用)開啟目標頁面，驗收主要操作與本次變更涉及的載入、錯誤、空資料或拒絕情境。
4. 按[前端驗收矩陣](../tooling/testing/commands.md#變更驗收矩陣)執行既有類型檢查與測試。路由、資源或靜態輸出有改時再建置。

完成時記錄路由、重現步驟、瀏覽器結果及既有檢查。需驗證 LINE 或定位時，分開記錄實際裝置與瀏覽器的結果。[裝置與定位證據](../verification/evidence.md#cloud-run-and-browser-location)

[操作配方](README.md)
