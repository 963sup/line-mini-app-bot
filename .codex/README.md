# Codex 專案配置

## Codex 環境初始化

從儲存庫根目錄開啟 Codex。在此 Codex Cloud 環境中，將 Install script 設為下列命令。它會安裝鎖定的 Go／Node／pnpm／API 產生器、依鎖檔安裝依賴，並下載與驗證 Go modules；快取恢復時重用已安裝版本。

```bash
bash scripts/tooling/bootstrap.sh cloud
```

發布環境會保存已準備的檔案與快取，程序不會保留。Start skill 指引從 checkout 根目錄使用 tool-exec.sh；若工具設定或依賴鎖檔變更，明確重跑上方命令。此介面沒有獨立 Maintenance script 欄位。

完整 Go 分析工具使用 `mise run install-core`；API 產生器使用 `mise run install-web`；Mermaid 使用 `mise run install-mermaid`，均透過 `bash scripts/tooling/tool-exec.sh` 執行。Cloud setup 不執行測試或註冊使用者層 MCP。

就地開發可使用 `bash scripts/tooling/bootstrap.sh core`。從全新 Cloud 環境驗收時，確認 setup 成功，再從 checkout 根目錄執行 `bash scripts/tooling/tool-exec.sh mise run check` 與 `bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp check`。本機 MCP 以 checkout 根目錄為工作目錄，依任務安裝所需套件後，在新對話確認 MCP 已載入。詳見[本機與 Cloud 初始化](../docs/guides/local-development/setup.md)。

需要 Google Cloud CLI、Firebase CLI 或本機 gcloud MCP 套件時，再執行 `bash scripts/tooling/tool-exec.sh mise run install-cloud`。初始化後，用 `bash scripts/tooling/tool-exec.sh mise ls --installed` 查看實際安裝版本。Codex 每次開啟專案時可依[根目錄指引](../AGENTS.md#cli-discovery)找到工具入口。


## 用途與範圍

保存 Codex 角色、MCP 與 hooks 設定。

## 架構與責任邊界

此處定義專案的代理配置，工具版本與任務由根目錄 mise 設定管理。

## 閱讀與修改入口

先查閱[config.toml](config.toml)、[agents](agents)、[hooks.json](hooks.json)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[操作模型](../docs/architecture/contracts/business-operations.md) · [符合性流程](../docs/development/workflow.md#business-operation-compliance)


## 目錄導覽與維護

config.toml 定義角色與 MCP 連線，agents/ 定義角色指引，hooks.json 保存 hooks；skills/ 為安裝產物。

連線設定只在 config.toml 維護；工具版本、任務與非機密環境以 ../mise.toml 為準。不要手改上游技能或寫入憑證。

從根目錄執行 bash scripts/tooling/tool-exec.sh python3 scripts/tooling/check-config.py；MCP transport 與聊天主機曝光需分別驗證。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)

## 確認專案設定載入

以儲存庫根目錄開啟 Codex；在可信任的 checkout 確認專案信任狀態。以新工作階段實際列出的專案專屬 MCP 或角色驗證，並分別檢查工具能否呼叫。使用者層級已註冊的同名 MCP、mise trust、角色 TOML 解析成功，都無法單獨證明聊天主機載入了專案設定。[官方設定層級](https://developers.openai.com/codex/config-basic/)

## Business Operation

開始相關任務前，必須閱讀 [操作模型](../docs/architecture/contracts/business-operations.md) 與 [符合性流程](../docs/development/workflow.md#business-operation-compliance)。

現有角色依 Business Operation 契約分工；角色開始相關任務時須讀取模型及符合性流程，文件路徑不是自動 include。

## 角色配置與派工

角色檔位於 `agents/`，由 `config.toml` 註冊；完整職責、派工順序、深度與單一寫入者規則見[代理工作流程](../docs/development/workflow.md#delegation-depth-and-role-selection)。`graphify`、`repomix` 提供探索證據，`context7` 查證適用版本文件，`golang`、`typescript`、`nextjs`、`shadcn` 接受明確範圍的實作指派，`documentation` 維護 Markdown。工具操作依[MCP 指南](../docs/development/tooling/mcp/README.md)，不在角色檔重複連線設定。

支援自訂代理的 Codex 可選用已註冊角色；其他主機依工作流程採用明示的 fallback。配置存在不代表目前聊天主機已載入角色或 MCP。角色沿用父代理權限，不增加推送、發布或部署授權。

## Mermaid

`mermaid` 角色負責本專案的架構、資料關係、操作流程與狀態圖。協調者在圖表有助於釐清設計或說明變更時主動選用，明確指派文件／圖表 hunks；小任務可直接採用該角色指引，不必為每張圖啟動代理。角色註冊與 MCP 連線以 `config.toml` 為準，職責細節以 `agents/mermaid.toml` 為準；不在應用程式加入 Mermaid runtime。

### CLI 安裝與發布信任

Mermaid v13 的 `mermaid` 套件內建 `mermaid render`，與舊版 `@mermaid-js/mermaid-cli` 的 `mmdc` 分開。版本與安裝任務以 [mise.toml](../mise.toml) 為準。`bootstrap.sh core` 安裝 Mermaid；`mcp`、`all` 另外安裝 shadcn 與 Next DevTools；`cloud` 只安裝應用程式工具，且 mise 不自動補裝缺少的 CLI。單項安裝可執行 `mise run install-mermaid`、`install-shadcn`、`install-next-devtools`，均透過 `bash scripts/tooling/tool-exec.sh`。

2026-11-20 核對 npm `mermaid@13.0.0-alpha.3` 的 tarball SHA-512 與 registry integrity 一致，套件版本、CLI 入口及 scripts 符合官方 [release revision](https://github.com/mermaid-js/mermaid-v13/tree/beaa8010fc218d7db123ec4084ca987964acf9c2/packages/mermaid)。該 alpha 發布者為現有 maintainer `ashishjain0512`，有 registry 簽章但未提供 provenance；mise/aube 因較早版本有較強信任證明而拒絕。依 [mise npm 指引](https://mise.jdx.dev/dev-tools/backends/npm.html)，僅對這個確切版本設定 `trust_policy_excludes`，不關閉其他套件的信任檢查。未來換版須重新查證。

已有 `mise.lock` 時，mise 安裝單一工具也可能解析整份工具設定；任何失敗會中止後續批次步驟，不能只憑套件下載完成宣稱整個安裝任務成功。診斷時分別驗證任務 exit code、CLI 執行與 MCP handshake；MCP 通過不代表聊天已載入或 Next.js runtime 已啟動。

### 專案適用範圍

| 圖表目的 | 必讀來源與邊界 |
| --- | --- |
| 儲存庫與專案關係 | [Project](../docs/domains/project/project.md)：個人或群組擁有 Project，引用可跨庫，來源讀權逐庫核驗 |
| 物件與子工序 | [Object](../docs/domains/object/design.md)、[Subobject](../docs/domains/object/design.md)：單層子項，數量／單位／進度獨立 |
| 問題與討論 | [Issue](../docs/domains/issue/design.md)、[Discussion](../docs/domains/discussion/design.md)：各自狀態與資料，不自動推進物件進度 |
| 授權與寫入流程 | [操作模型](../docs/architecture/contracts/business-operations.md)：區分主體、資源、權限、領域規則、交易與結果 |

以 `GroupID` 識別 LINE 群組，以 `SpaceID` 識別 Repository 內的工區。[庫內工區](../docs/domains/repository/spaces.md)

圖表須註明來源版本、相關未提交變更及已實作／提案／待確認狀態，不以舊圖譜或規劃文件代替現行程式證據。

### 操作與驗收

1. 核對來源並依[文件圖表規範](../docs/standards/documentation.md#獨立圖檔)維護內嵌區塊或獨立圖檔。
2. 使用專案鎖定的本機 Mermaid v13 CLI。需要安裝時執行 `bash scripts/tooling/tool-exec.sh mise run install-mermaid`。
3. 執行 `bash scripts/tooling/tool-exec.sh mise run render-mermaid -- render diagram.mmd --format svg > diagram.svg`，核對圖中文字與箭頭符合來源。

[開發與維運指引圖索引](../docs/diagrams.md#開發與維運指引) · [未來設計圖](../docs/diagrams.md#未來設計)。

### Codex Cloud 新環境驗收

從目前 checkout 根目錄執行下列命令。第一個命令會依專案設定啟動 Repomix 並讀回本 checkout 的檔案，可在聊天未提供同名工具時使用。第二個會檢查已啟用的 MCP。

```sh
python3 scripts/tooling/mcp-doctor.py --server repomix --call file_system_read_file --arguments '{"path":"README.md"}'
bash scripts/tooling/tool-exec.sh mise run doctor-mcp
```

使用 Cloud Run 或 Firestore Toolbox 的 Cloud 環境需由主機提供 ADC。若使用憑證檔，讓 GOOGLE_APPLICATION_CREDENTIALS 指向任務執行時可讀的檔案；若主機提供預設身分，移除指向不存在檔案的舊設定。doctor 會確認 ADC 能取得 token，資源 IAM 仍需以目標專案的唯讀操作驗收。新聊天的原生工具曝光須在該聊天內核對。
