# Repository scripts

保存建置、檢查、分析與發布的工具腳本。

腳本負責開發與維護流程，應用程式規則由前後端套件維護。

## 閱讀與修改入口

從儲存庫根目錄執行。每支 Shell 腳本負責一項操作，多步流程由 bootstrap 或 mise 任務組合。工具版本與任務定義見 [mise.toml](../mise.toml)。

| 位置 | 職責 | mise 任務或執行入口 |
| --- | --- | --- |
| [bootstrap.sh](bootstrap.sh) | 初始化指定環境 | core 使用 mise run bootstrap；直接執行預設 cloud，也可傳入 core、cloud、mcp、all |
| [setup/](setup/README.md) | 初始化工具環境、準備 Go 依賴 | deps |
| [dev/](dev/README.md) | 載入本機環境、啟動服務 | dev |
| [generate/](generate/README.md) | 生成 API 程式碼、前端資源與文件圖 | pnpm api:generate、graphify-docs |
| [check/](check/README.md) | 設定檢查 | check-config |
| [contract/](contract/README.md) | API 能力與路由契約 | pnpm api:check |
| [test/](test/README.md) | 路由、MCP 與容器驗收 | pnpm api:check、unittest、container-build |
| [perf/](perf/README.md) | Go 效能剖析 | profile |
| [build/](build/README.md) | 建置容器映像 | container-build |
| [release/](release/README.md) | 發布驗收 | release-smoke.py |
| [ops/](ops/README.md) | 查閱 Cloud Run 狀態 | cloud-run-snapshot.py |
| [db/](db/README.md) | 維護 Firestore 索引 | ensure-attendance-indexes.py |
| [tooling/](tooling/README.md) | 開發工具與環境診斷 | doctor、tool-exec.sh |
| [mcp/](mcp/README.md) | 專案維護的 MCP 設定核對、連線與診斷腳本 | check-mcp-config、doctor-mcp |

```bash
bash scripts/bootstrap.sh core
bash scripts/tooling/tool-exec.sh mise run doctor
python3 scripts/check/check-config.py
bash scripts/tooling/tool-exec.sh python3 -m unittest discover -s scripts/test -p 'test_*.py'
bash scripts/tooling/tool-exec.sh pnpm api:check
```

Tool versions and task definitions remain in `mise.toml`. Frontend-only script
paths are wired through `frontend/miniapp/package.json`; repository API commands
through root `package.json`. Release scripts do not grant deployment permission.

`tooling/shadcn-reference.py` downloads and verifies the official Components／AI
Elements documentation snapshot; it never installs application components.
See [local reference usage and refresh](../docs/development/tooling/frontend/shadcn-reference/README.md).

Rich Menu images are stored under `assets/line/richmenu/`.

Go 弱點檢查請執行 `mise run vuln`；工具版本由 [mise.toml](../mise.toml) 固定。

## 相關文件

- [MCP 診斷期限與清理](../docs/development/tooling/agents/README.md)
- [文件索引：Scripts](../docs/README.md#按模組閱讀)
- [mise 任務與工具](../docs/development/tooling/shared/mise.md)

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md)
