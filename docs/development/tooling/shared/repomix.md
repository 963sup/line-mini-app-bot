# Repomix CLI

通用打包與參考技能產生方式由本頁維護；代理與 MCP 接線見 [Repomix MCP](../agents/repomix.md) 與 [Repomix Skills MCP](../agents/repomix-skills.md)。

## 安裝與設定

版本由 `mise.toml` 固定，使用 `scripts/tooling/tool-exec.sh` 載入專案工具環境；初始化見 [工具鏈設定](toolchain-setup.md)。來源排除由 `.repomixignore` 維護。保留 Git ignore、內建排除與安全掃描；安全掃描通過不等於不存在機密。

## 打包

從儲存庫根目錄執行：

```bash
bash scripts/tooling/tool-exec.sh mise run repomix
```

任務輸出 `repomix-output.xml`。限定任務來源時使用 `--include` 與 `--output`，核對輸出清單與來源內容，避免把打包成功當作功能驗收。

## 產生參考技能

### 更新本專案參考技能

從專案根目錄執行：

```bash
bash scripts/tooling/tool-exec.sh repomix \
  --skill-generate repomix-reference-line-mini-app-bot \
  --skill-output .agents/skills/repomix-reference-line-mini-app-bot \
  --ignore '.mise/locks/**' --force
```

### 產生任務專用參考包

需要[交接原始碼](../../workflow/handoff.md#sidebar-handoff)時，可產生限定於該任務的參考包。

用簡短、好識別的用途名稱，例如 `repomix-reference-codex-cloud`。同一用途沿用名稱，來源版本記在 Evidence。

以下為 Codex Cloud 設定參考的範例；其他用途 MUST 同時調整主題名稱與 `--include`，只涵蓋本次問題需要的來源：

```bash
bash scripts/tooling/tool-exec.sh repomix . \
  --skill-generate repomix-reference-codex-cloud \
  --skill-output .agents/skills/repomix-reference-codex-cloud \
  --include '.codex/config.toml,docs/development/tooling/agents/codex.md,mise.toml,scripts/bootstrap.sh,scripts/setup/**,scripts/dev/**,scripts/mcp/**,scripts/tooling/tool-exec.sh' \
  --ignore '**/node_modules/**,**/__pycache__/**,**/.env*,**/*credentials*.json,.agents/**,.claude/**,graphify-out/**' \
  --force
```

- `--skill-output` 是技能目錄本身，MUST 包含個別技能子目錄；不要只傳 `.agents/skills/`，以免把 `SKILL.md` 與 `references/` 寫到共用根目錄。
- `--skill-output` 跳過位置選擇；`--force` 跳過覆寫確認，不會縮小打包範圍。範圍由 `--include`／`--ignore` 控制，需核對實作時保留完整內容，不加壓縮。
- `--force` 只可更新本任務擁有的生成目錄。MUST 先核對既有來源、版本與未提交內容。同一用途的自有生成包可在核對後更新；需要保留的舊快照移至技能目錄外的報告或封存位置，不以版本後綴建立另一個技能。不覆寫上游或其他工作的技能。
- 成功後核對 `SKILL.md`、`references/project-structure.md` 與 `references/files.md`；Evidence 記錄來源版本、未提交狀態、範圍與可讀路徑。使用 `bash scripts/tooling/tool-exec.sh skills list --agent codex` 確認本機探索；當前聊天載入及自動觸發仍須另行驗證。
- `.agents/skills/` 下的本機 Repomix 生成包受 Git 忽略。另一個 checkout 或新環境須重新生成或明確保存產物；不因本機生成成功就宣稱已隨 Git 或發布環境提供。

指令已對照專案固定 Repomix 1.18.1 的 `--help` 與輸出路徑實作；參見 [官方非互動產生說明](https://repomix.com/guide/agent-skills-generation#non-interactive-usage)。

## 更新與清理

來源搬移或刪除後，以原本用途名稱及更新後的來源清單重新產生自有參考技能，連同一般打包輸出核對舊路徑是否仍存在。不要手改生成的 `references/files.md`，以免內容、統計與索引不同步。記錄實際來源與未提交狀態；本機生成物不隨 Git 提交自動同步。
