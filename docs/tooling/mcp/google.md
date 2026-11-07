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

## Verification at installation

- gcloud MCP: initialization and tools/list passed; one tool exposed.
- gcloud: installed; no active authenticated account was found.
- Developer Knowledge: configured; lookup verification requires `DEVELOPERKNOWLEDGE_API_KEY`.
- Codex host loading and cloud resource operations still require separate verification.

References: [official Codex setup](https://github.com/google/skills/blob/2194d3b8b1f31975fef2f1a0ae6b9ab55bd7ceee/plugins/cloud/google-cloud-developer/.codex-plugin/README.md),
[Google Developer Knowledge MCP](https://developers.google.com/knowledge/mcp).
