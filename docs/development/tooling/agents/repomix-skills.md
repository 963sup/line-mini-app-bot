# repomix_skills

Owner: the primary agent preparing the handoff, or the assigned skill-generation owner.

## When to use

需要交接或反覆查閱一組程式時，按任務選擇來源並產生參考技能。專案不使用 Codex hooks。

## Prerequisites and configuration

[repomix_skills 設定](../../../../.codex/config.toml)以固定版本 Repomix 的 `--mcp` 模式啟動，只曝光 `generate_skill`，依主機核准設定執行。啟動與握手本身不產生或覆寫技能。使用現有參考包請從 [Codex 技能操作](../shared/repomix.md#在-codex-使用本專案的參考技能)進入。

<a id="operation"></a>

## 按需打包任務來源

CLI 的命名、來源選擇、覆寫與驗證規則統一見 [產生參考技能](../shared/repomix.md#產生參考技能)。

### 明確指定 MCP 時

以下流程適用於另行以 `repomix --mcp` 啟動且已曝光 `generate_skill` 的工作階段。

- 主責代理在目前聊天工具目錄找專案 `repomix_skills` 的 `generate_skill`，依主機曝光的實際名稱呼叫。使用者明確指定 MCP 時選用，遵循主機的工具核准設定。專案的唯讀 `repomix` 角色只負責一般打包。
- 依原生工具的 input schema 傳入正規化的 checkout 絕對路徑 `directory`、依[共用命名規則](../shared/repomix.md#產生參考技能)指定的 `skillName`，以及逗號分隔的 `includePatterns`，只涵蓋本次問題需要的來源、契約或測試。需核對實作細節時保留 `compress=false`。
- 此版工具在來源目錄下建立 `.claude/skills/<skillName>/`。同名目錄已存在時，先檢查版本、相關未提交內容與涵蓋範圍；可重用就引用。來源有變時沿用固定名稱；若 MCP 不支援安全更新，回報限制；使用者同意改用 CLI 時，依[共用 CLI 流程](../shared/repomix.md#產生參考技能)更新。
- 成功後讀回工具回傳的 `skillPath` 下的 `SKILL.md` 與 `references/files.md`，核對來源清單和所需實作，將可讀位置、版本與範圍放入 Evidence。此目錄受 Git 忽略；另一個 checkout 需要另行取得該產物或在相同來源重新生成。
- 工具未曝光時，依[主機載入診斷](README.md#project-loading-check)檢查專案選擇、信任與工具清單，使用主機支援的重新探索。尚未恢復時，在 Open issues 記錄失敗階段，附上可追的來源路徑。握手成功只證明伺服器能啟動。

## Verification and failures

下列 doctor 驗證已設定的 MCP 服務。CLI 產物另依共用操作文件的查詢範例核對。

```bash
bash scripts/tooling/tool-exec.sh python3 scripts/mcp/mcp-doctor.py --server repomix_skills
```

This checks initialize, tools/list and configured allowlists. Verify a representative
operation separately. For failures distinguish missing binary, network, credential,
workspace prerequisites and host loading; see [troubleshooting](../../environment/troubleshooting.md).
