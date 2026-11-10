# Google Cloud skills and MCP

Use the three project-installed skills from [google/skills](https://github.com/google/skills):
`gcloud`, `cloud-run-basics`, and `retrieving-developer-knowledge`.
The exact source revision is recorded in `skills-lock.json` and `.agents/skills-sources.json`.
Other development skills remain installed. Do not install the entire Google catalog by default.

## Google Developer Knowledge

The remote endpoint is `https://developerknowledge.googleapis.com/mcp`.
Codex reads `DEVELOPERKNOWLEDGE_API_KEY` from its launch environment and sends it
as `X-Goog-Api-Key` via `env_http_headers`. No key belongs in project configuration.

Enable the Developer Knowledge API and create an API-restricted key in the selected
Google Cloud project using the [official quickstart](https://developers.google.com/knowledge/quickstart).
Supply the key through the host's secret/environment settings before launching Codex.
Restart or reload Codex after configuration or environment changes.

Verify with `bash scripts/tool-exec.sh mise run doctor-google-knowledge`.
Missing credentials and HTTP authentication errors are failures, not successful verification.
Once loaded, use the server's actual exposed search/retrieval tools; names may vary by release.

## Google Cloud CLI MCP

Use the [Google-maintained local server](https://github.com/googleapis/gcloud-mcp/tree/main/packages/gcloud-mcp)
with the existing Google Cloud CLI. Both versions are pinned in `mise.toml`.
Install with `bash scripts/tool-exec.sh mise run install-cloud`, then verify discovery
with `bash scripts/tool-exec.sh mise run doctor-gcloud`.

The local server uses the authenticated gcloud account and the `CLOUDSDK_CONFIG`
directory supplied by the runtime. It exposes `run_gcloud_command`; approval mode
is `prompt`. This is the local stdio server, not the Cloud CLI Execution API remote server.
The upstream server is currently a preview solution; fixed package versions avoid
unreviewed changes on every launch.

Authenticate through the official gcloud login flow on the host before cloud resource
operations. CLI login and application ADC are separate: running the Go application
locally may additionally require ADC, while Cloud Run uses its attached service account.
Never copy credential files into this repository or print tokens in doctor output.
Check leaf-level `gcloud help` before commands, specify project and region explicitly,
and constrain list output. Configuration does not authorize deploying resources or changing IAM.

## gcloud sign-in

登入任務採「先重用、再網址、最後驗證」；不執行部署、IAM 或其他資源變更。

1. **先排除專案環境誤值。** `CLOUDSDK_CORE_PROJECT` 若是字面值
   `$GOOGLE_CLOUD_PROJECT`，不會自動展開，可能讓成功登入或憑證檢查仍回傳專案錯誤。
   在執行以下 auth 命令前，先確認 `GOOGLE_CLOUD_PROJECT` 是本環境原先選定的具體
   project ID；確認後可用
   `bash scripts/tool-exec.sh env CLOUDSDK_CORE_PROJECT="$GOOGLE_CLOUD_PROJECT" gcloud ...`
   作命令層覆寫（放在 wrapper 後，避免啟動環境再覆寫）。永久修正應在 Cloud
   environment settings 填具體 ID，不在登入任務擅自改專案或 repository 設定。

2. **先檢查，不直接重登。** 先讀 `gcloud auth list --help` 與
   `gcloud auth print-access-token --help`，再透過專案 wrapper 查帳號 metadata：

   ```bash
   bash scripts/tool-exec.sh gcloud auth list --format='json(account,status)'
   ```

   有第 1 步誤值時，以下命令均套用該覆寫。設定了 `core/account` 不代表已有憑證；
   預期帳號已列出時，無論是否 ACTIVE，都先私密確認
   `gcloud auth print-access-token --account="<預期帳號>"` 能成功（stdout 導向
   `/dev/null`，不可顯示或保存 token）。成功但帳號未啟用時，先讀
   `gcloud config set --help`，再以 `gcloud config set account "<預期帳號>"`
   選取該帳號，並檢查環境 `CLOUDSDK_CORE_ACCOUNT` 沒有覆寫成其他帳號。
   憑證可用就直接回報已登入，不因不是 ACTIVE 而重啟 OAuth。

3. **需要登入時立即給網址。** 先讀 `gcloud auth login --help`，使用預期帳號：

   ```bash
   bash scripts/tool-exec.sh gcloud auth login "<預期帳號>" --no-launch-browser
   ```

   若有第 1 步的誤值，套用該命令層覆寫。立刻把產生的 Google OAuth 網址以可點擊連結
   給使用者，請其登入、同意授權，再提供本次驗證碼。預設不探測或啟動共享桌面，
   不先試 `BROWSER`、DISPLAY、loopback 或 Playwright；使用者要求代開瀏覽器時才使用。
   不改用 `--no-browser`，該流程需要另一台已安裝 gcloud 的可信機器。

4. **同一執行者保留同一程序。** 由 orchestrator 的單一 executor 持有登入 session，
   top-level 只負責傳達網址及將使用者驗證碼交回原 executor，原 executor 直接寫入
   該程序 stdin。不可假設其他 agent 能以相同 session ID 呼叫 `write_stdin`。
   等待使用者期間保留程序；不另開第二次登入、不重複讀瀏覽器狀態。
   若程序已結束或碼過期，只重建一次並提供新網址，舊碼不能送進新流程。

5. **以憑證結果判斷成功。** 完成後再查 ACTIVE 帳號並私密確認 token 可取得。
   若登入退出碼非零，先區分 OAuth 交換失敗和登入後專案設定錯誤；已有 ACTIVE
   帳號且 token 驗證成功時，不要求使用者重做 OAuth。回報帳號與任何仍存在的專案誤值。

驗證碼只送原程序，不在回覆中重述；token、完整 OAuth URL、驗證碼、憑證檔不得進入
tracked 文件或一般 log。需要暫存時只用 repository 外的私密檔，完成後清除。
CLI 登入不等於 ADC：只有使用者要求或已授權工作明確需要 client-library 憑證時，
才另外執行 `gcloud auth application-default login`；不要預設加 `--update-adc`。

依據：[官方 headless 登入流程](https://cloud.google.com/sdk/docs/authorizing#authorizing_with_a_user_account)
及本機 pinned CLI 的 leaf-level help；本流程已在此環境驗證。安裝時的歷史結果如下，
不是目前帳號狀態，登入前仍需即時檢查。

## Verification at installation

- gcloud MCP: initialization and tools/list passed; one tool exposed.
- gcloud: installed; no active authenticated account was found.
- Developer Knowledge: configured; lookup verification requires `DEVELOPERKNOWLEDGE_API_KEY`.
- Codex host loading and cloud resource operations still require separate verification.

References: [official Codex setup](https://github.com/google/skills/blob/2194d3b8b1f31975fef2f1a0ae6b9ab55bd7ceee/plugins/cloud/google-cloud-developer/.codex-plugin/README.md),
[Google Developer Knowledge MCP](https://developers.google.com/knowledge/mcp).
