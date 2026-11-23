<a id="查閱-cloud-run-狀態工作指引"></a>

# Cloud Run inspection guidance

Inherit [parent guidance](../AGENTS.md); Cloud Run status queries stay read-only.

After changes, run checks from the [script overview](../README.md); see the [local README](README.md) for entry points.

## Applicable engineering conventions

- **Target configuration:** Identify project, region, service and revision explicitly before inspecting operational state.
- **Observability:** Correlate structured logs, traces, metrics and health evidence for the same revision rather than mixing historical results.
- **Security:** Keep inspection read-only and redact secrets; diagnosis does not authorize changing IAM, deployment or resources.
