# 使用 Next.js MCP 診斷 Mini App

1. 從儲存庫根目錄啟動 `pnpm --filter @repo/miniapp dev`，或沿用已啟動的專案開發伺服器。從終端輸出確認實際連接埠，專案預設為 3000。
2. 用 `next-devtools` MCP 的 `nextjs_index` 找出伺服器與它提供的工具，核對專案目錄及連接埠。若未發現服務，傳入實際連接埠再查，並查看開發伺服器的啟動錯誤。
3. 開啟受影響頁面並重現操作，再以 `nextjs_call` 指定該連接埠與上一步列出的工具。例如伺服器提供 `get_errors` 時，用它取得編譯或執行期錯誤。需要瀏覽器互動證據時可使用 `browser_eval`。需要確認框架行為時可透過 `nextjs_docs` 查證，並核對專案使用的版本與 Pages Router。
4. 修正後在同一頁面重做相同操作，再讀取診斷。執行受影響的測試；涉及路由或正式輸出時，另執行 `pnpm --filter @repo/miniapp build`。回報測試路徑、實際結果及需要登入或實機驗證的部分。

[Mini App 指令與框架版本](../../frontend/miniapp/package.json) · [驗證流程](verification.md)
