# 開發工具初始化

## Codex Cloud 環境設定

Install script、Start skill、更新與新聊天驗收由 [Codex Cloud 環境手冊](../../environment/codex-cloud.md)維護。

## 一般開發與 MCP 安裝

省略模式時使用 `cloud`，本機開發使用 `core`。初始化只安裝該模式的開發工具。需要 MCP 時，依 [mise 工具任務](mise.md)另行準備，再執行 [MCP 診斷](../../environment/troubleshooting.md#設定與-mcp-診斷)。

**開發與維運指引**

```mermaid
flowchart TB
  root["開發環境與驗收"]
  root --- b0["從儲存庫根目錄執行"]
  b0 --- n0_0["選擇一個模式：<br/>bash scripts/bootstrap.sh cloud<br/>bash scripts/bootstrap.sh core"]
  b0 --- n0_1["bash scripts/tooling/tool-exec.sh mise run check-config"]
  root --- b1["工作目錄"]
  b1 --- n1_0["Go module：backend；直接使用 go 前先 cd backend"]
  b1 --- n1_1["前端與契約：從根目錄執行；pnpm 使用 workspace 鎖檔"]
  b1 --- n1_2["Go race 需 C 編譯器；container-build 需 Docker daemon"]
  root --- b2["Codex Cloud"]
  b2 --- n2_0["從 checkout 根目錄開啟 Codex 並載入 .codex/config.toml；依受支援的信任設定啟用"]
  b2 --- n2_1["套件下載與遠端 MCP 均需網路。GITHUB_MCP_TOKEN 由主機於執行期安全注入"]
  b2 --- n2_2["本機 codex mcp list 只驗本機設定；重載 host 後在新對話實際呼叫"]
  root --- b3["編輯器與工具"]
  b3 --- n3_0["VS Code 使用專案 Go／gopls wrappers；開啟 Go 檔前執行 Go: setup"]
  b3 --- n3_1["shadcn 版本與任務依 mise.toml；CLI 透過 tool-exec.sh"]
  b3 --- n3_2["mise run repomix 封裝原文；graphify-update 更新程式圖；graphify-docs 更新文件圖"]
```
