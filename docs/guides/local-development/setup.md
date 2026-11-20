# Linux x64／arm64：Bash、curl、tar、sha256sum、Python 3.11+

## Codex Cloud 環境設定

在 Codex Cloud 環境編輯頁，將 **Setup script** 與選用的 **Maintenance script** 設為以下命令。首次建立容器時安裝固定版本的 Go、Node、pnpm 與 API 產生器，依鎖檔安裝 workspace 套件並下載、驗證後端 Go modules。恢復快取時，Maintenance script 以同一命令重新核對目前 checkout，重用工具與套件快取；pnpm 使用 frozen lockfile 與 prefer-offline。Cloud 模式不執行測試，也不註冊使用者層 MCP。

Go 分析工具、API 產生器、Mermaid 分別使用 mise 的 install-core、install-web、install-mermaid 任務，透過 scripts/tooling/tool-exec.sh 執行。MCP 與部署工具依各自任務安裝。一般 core 模式仍保留完整工具與 doctor 驗證。執行 pnpm check 前需先 install-core，因後端檢查包含格式與 lint；pnpm build／pnpm test 使用 Cloud 模式的工具即可。

Cloud 必需工具由 mise.lock 鎖定 Linux x64／arm64 下載資訊，npm 工具的相依鎖保存於 .mise/locks；兩者依 [mise 官方鎖檔指引](https://mise.jdx.dev/dev-tools/mise-lock.html)一併提交。應用依賴由 pnpm-lock.yaml 與 backend/go.sum 驗證。Go 原始碼工具使用模組版本與 Go checksum database；選用工具須依各自安裝任務與鎖檔內容核對。

```bash
bash scripts/tooling/bootstrap.sh cloud
```

若未設定 Maintenance script，工具設定或鎖檔變更時明確重跑初始化。Agent 階段從 checkout 根目錄使用 scripts/tooling/tool-exec.sh。驗收需分別記錄空工具／依賴快取的冷啟動與同一快取的再次執行；現有機器成功不代表已驗證全新環境。[官方環境設定](https://developers.openai.com/codex/cloud/environments)

## 一般開發與 MCP 安裝

```mermaid
flowchart TD
  A["Linux x64／arm64：Bash、curl、tar、sha256sum、Python 3.11+"] --> B["核對固定版 mise，信任專案設定，安裝 core 工具"]
  B --> W{"core（預設）或 all？"}
  W -->|是| D["安裝 Web 工具與鎖定依賴，執行 doctor"]
  W -->|否| M["安裝 Mermaid"]
  D --> M
  M --> C{"mcp 或 all？"}
  C -->|否| E["core 完成"]
  C -->|是| G["安裝 MCP 與 Cloud 工具，更新 Graphify，執行 doctor-mcp"]
  G -->|探測失敗| F["依各 server 結果檢查連線、設定與執行期憑證"]
  G -->|探測通過| H["在載入專案設定的新對話確認工具，實際呼叫任務所需工具"]
```

```mermaid
flowchart TB
  root["開發環境與驗收"]
  root --- b0["從儲存庫根目錄執行"]
  b0 --- n0_0["選擇一個模式：<br/>bash scripts/tooling/bootstrap.sh core<br/>bash scripts/tooling/bootstrap.sh mcp<br/>bash scripts/tooling/bootstrap.sh all"]
  b0 --- n0_1["bash scripts/tooling/tool-exec.sh mise run check-config"]
  b0 --- n0_2["bash scripts/tooling/tool-exec.sh mise run doctor-mcp"]
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
