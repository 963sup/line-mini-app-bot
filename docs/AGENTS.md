# docs 工作指引

適用於本目錄及其子目錄；先遵循 [根目錄 AGENTS.md](../AGENTS.md)，再套用較近的子目錄補充。使用者及主機指令優先；本檔不增加外部寫入授權。

## 職責與邊界

README.md 是文件索引；architecture/、development/、features/、integrations/、operations/、tooling/、agents/ 各自保存決策、開發、功能、整合、營運、工具與代理流程。

以已接受 ADR 與實際來源為準，區分已實作、待驗收與未來規劃。新增文件需連回索引；不在多處複製版本、連線設定、機密或完整政策。

文件維持主題分類；docs/README.md 提供 Backend、API、Frontend、Scripts、Deploy 閱讀入口。新增文件同步更新索引與相關模組 README，以連結共用跨模組規格。

## 驗證

驗證本機連結與來源路徑；工具圖譜不證明語意正確。純文件變更執行 git diff --check，相關設定變更再執行設定檢查。

命令除另有標示外均在儲存庫根目錄執行。閱讀 [本目錄 README](README.md) 與 [文件索引](../docs/README.md)；回報實際執行結果與限制。

## 主題子目錄

各主題目錄 README 列出完整文件入口，AGENTS.md 補充歸屬與維護規則；入口見 [文件歸屬](README.md#文件歸屬與維護)。新增文件需同步該層索引與總索引，模組相關內容另更新模組 README。
