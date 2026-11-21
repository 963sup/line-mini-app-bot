# repomix_skills

Owner: the primary agent preparing the handoff, or the assigned skill-generation owner.

## When to use

每次送出使用者訊息時，專案 Hook 刷新開發入口技能。需要交接或反覆查閱一組程式時，再按任務選擇來源。

## Prerequisites and configuration

Shares Repomix installation but has a separate tool allowlist and prompting policy. Handshake checks do not invoke generate_skill or write files.
Connection, timeouts and enabled tool names are owned by `.codex/config.toml`.
Tool versions are owned by `mise.toml`; never copy token values into documentation.

<a id="operation"></a>

## 每次交談前重新產生

[專案 Hooks](../../../../.codex/hooks.json) 在 `UserPromptSubmit` 呼叫[刷新腳本](../../../../scripts/tooling/repomix-prompt-hook.py)。腳本使用固定版本 CLI 和 `--force`，將開發手冊入口打包到 `.agents/skills/repomix-reference-development-entry/`，並把 `SKILL.md` 路徑交給當輪 Codex。來源由腳本限定。使用者訊息不傳給 Repomix。

在支援專案 Hooks 的宿主開啟並信任此 checkout。Codex 首次載入這項 Hook 時要審核其定義，CLI 可用 `/hooks`。驗收時在一個新聊天先後送出兩則訊息。兩次都應看到刷新，且 `SKILL.md` 的修改時間都應更新。生成失敗時由 Hook 顯示原因，聊天可繼續查原始文件。[OpenAI Hooks](https://developers.openai.com/codex/hooks/#userpromptsubmit)

## 按需打包任務來源

開發入口技能提供導覽。需要更深入或要[交接原始碼](../../workflow/handoff.md#sidebar-handoff)時，可另外產生限定於該任務的參考包。

以下為 Codex Cloud 設定參考的範例；其他任務 MUST 同時調整技能名稱與 `--include`，只涵蓋本次問題需要的來源：

```bash
bash scripts/tooling/tool-exec.sh repomix . \
  --skill-generate repomix-reference-codex-cloud \
  --skill-output .agents/skills/repomix-reference-codex-cloud \
  --include '.codex/config.toml,.codex/README.md,mise.toml,scripts/bootstrap.sh,scripts/setup/**,scripts/dev/**,scripts/mcp/**,scripts/tooling/tool-exec.sh' \
  --ignore '**/node_modules/**,**/__pycache__/**,**/.env*,**/*credentials*.json,.agents/**,.claude/**,graphify-out/**' \
  --force
```

- `--skill-output` 是技能目錄本身，MUST 包含個別技能子目錄；不要只傳 `.agents/skills/`，以免把 `SKILL.md` 與 `references/` 寫到共用根目錄。
- `--skill-output` 跳過位置選擇；`--force` 跳過覆寫確認，不會縮小打包範圍。範圍由 `--include`／`--ignore` 控制，需核對實作時保留完整內容，不加壓縮。
- `--force` 只可更新本任務擁有的生成目錄。MUST 先核對既有來源、版本與未提交內容。需要保留的舊包則改用新技能名稱，不覆寫上游或其他工作的技能。
- 成功後核對 `SKILL.md`、`references/project-structure.md` 與 `references/files.md`；Evidence 記錄來源版本、未提交狀態、範圍與可讀路徑。使用 `bash scripts/tooling/tool-exec.sh skills list --agent codex` 確認本機探索；當前聊天載入及自動觸發仍須另行驗證。
- `.agents/skills/` 受 Git 忽略。另一個 checkout 或新環境須重新生成或明確保存產物；不因本機生成成功就宣稱已隨 Git 或發布環境提供。

指令已對照專案固定 Repomix 1.18.1 的 `--help` 與輸出路徑實作；參見 [官方非互動產生說明](https://repomix.com/guide/agent-skills-generation#non-interactive-usage)。

### 明確指定 MCP 時

- 主責代理在目前聊天工具目錄找專案 `repomix_skills` 的 `generate_skill`，依主機曝光的實際名稱呼叫。使用者明確指定 MCP 時選用，遵循主機的工具核准設定。專案的唯讀 `repomix` 角色只負責一般打包。
- 依原生工具的 input schema 傳入正規化的 checkout 絕對路徑 `directory`、任務專用的 `skillName`，以及逗號分隔的 `includePatterns`，只涵蓋本次問題需要的來源、契約或測試。需核對實作細節時保留 `compress=false`。
- 此版工具在來源目錄下建立 `.claude/skills/<skillName>/`。同名目錄已存在時，先檢查版本、相關未提交內容與涵蓋範圍；可重用就引用，來源有變才用新名稱產生。既有包保留。
- 成功後讀回工具回傳的 `skillPath` 下的 `SKILL.md` 與 `references/files.md`，核對來源清單和所需實作，將可讀位置、版本與範圍放入 Evidence。此目錄受 Git 忽略；另一個 checkout 需要另行取得該產物或在相同來源重新生成。
- 工具未曝光時，依[主機載入診斷](README.md#project-loading-check)檢查專案選擇、信任與工具清單，使用主機支援的重新探索。尚未恢復時，在 Open issues 記錄失敗階段，附上可追的來源路徑。握手成功只證明伺服器能啟動。

## Verification and failures

```bash
bash scripts/tooling/tool-exec.sh python3 scripts/mcp/mcp-doctor.py --server repomix_skills
```

This checks initialize, tools/list and configured allowlists. Verify a representative
operation separately. For failures distinguish missing binary, network, credential,
workspace prerequisites and host loading; see [troubleshooting](../../../guides/release/troubleshooting.md).
