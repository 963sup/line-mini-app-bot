# api 工作指引

適用於本目錄及其子目錄；先遵循 [根目錄 AGENTS.md](../AGENTS.md)，再套用較近的子目錄補充。使用者及主機指令優先；本檔不增加外部寫入授權。

## 職責與邊界

contract/openapi.yaml 為入口，contract/attendance/ 下依業務拆分的契約 定義出勤；client/ 保存 TypeScript client，compatibility/baseline/ 保存相容性基準。

先修改契約，再由根目錄 pnpm api:generate 更新 Go/TS 產生碼；不手改 generated 檔或任意刷新 baseline。業務核心不依賴 HTTP DTO。

## 驗證

從根目錄執行 bash scripts/tooling/tool-exec.sh pnpm api:check。歷史比較依 README.md 指定 base revision；缺少 base 不宣稱通過。

命令除另有標示外均在儲存庫根目錄執行。閱讀 [本目錄 README](README.md) 與 [文件索引](../docs/README.md)；回報實際執行結果與限制。

## 文件閱讀路由

先由 [本模組相關文件](README.md#相關文件) 與 [模組閱讀入口](../docs/README.md#按模組閱讀) 選擇文件。局部用途與命令維護在本模組 README；跨模組架構、功能與營運文件依 docs 主題分類保存。新增或搬移相關文件時，同步更新兩層索引與連結。
