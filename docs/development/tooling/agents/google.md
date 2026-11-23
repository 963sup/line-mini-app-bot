# Google Cloud skills and MCP

Use the three project-installed skills from [google/skills](https://github.com/google/skills):
`gcloud`, `cloud-run-basics`, and `retrieving-developer-knowledge`.
The exact source revision is recorded in `skills-lock.json`.
Other development skills remain installed. Do not install the entire Google catalog by default.

## Google Developer Knowledge

The remote endpoint is `https://developerknowledge.googleapis.com/mcp`.
Codex reads `DEVELOPERKNOWLEDGE_API_KEY` from its launch environment and sends it
as `X-Goog-Api-Key` via `env_http_headers`. No key belongs in project configuration.

Enable the Developer Knowledge API and create an API-restricted key in the selected
Google Cloud project using the [official quickstart](https://developers.google.com/knowledge/quickstart).
Supply the key through the host's secret/environment settings before launching Codex.
Restart or reload Codex after configuration or environment changes.

Open `/mcp` in Codex and check that the Developer Knowledge tools appear.
Missing credentials and HTTP authentication errors are failures, not successful verification.
Once loaded, use the server's actual exposed search/retrieval tools; names may vary by release.

## Google Cloud CLI MCP

Use the [Google-maintained local server](https://github.com/googleapis/gcloud-mcp/tree/main/packages/gcloud-mcp)
with the existing Google Cloud CLI. Both versions are pinned in `mise.toml`.
Install with `bash scripts/tooling/tool-exec.sh mise run install-cloud`, then verify discovery
with `/mcp` in Codex.

The local server uses the authenticated gcloud account and the `CLOUDSDK_CONFIG`
directory supplied by the runtime. It exposes `run_gcloud_command`; approval mode
is `prompt`. This is the local stdio server, not the Cloud CLI Execution API remote server.
The upstream server is currently a preview solution; fixed package versions avoid
unreviewed changes on every launch.

CLI 登入、ADC 與執行環境設定見 [Google Cloud CLI](../shared/gcloud.md)。代理操作仍受當前聊天的權限與任務授權限制。

## gcloud sign-in

依[共用登入流程](../shared/gcloud.md#gcloud-sign-in)執行；由目前代理持有同一程序並呈現驗證網址，不啟動子代理處理 OAuth。

## Verification at installation

- gcloud MCP: initialization and tools/list passed; one tool exposed.
- gcloud: installed; no active authenticated account was found.
- Developer Knowledge: configured; lookup verification requires `DEVELOPERKNOWLEDGE_API_KEY`.
- Codex host loading and cloud resource operations still require separate verification.

References: [official Codex setup](https://github.com/google/skills/blob/2194d3b8b1f31975fef2f1a0ae6b9ab55bd7ceee/plugins/cloud/google-cloud-developer/.codex-plugin/README.md),
[Google Developer Knowledge MCP](https://developers.google.com/knowledge/mcp).
