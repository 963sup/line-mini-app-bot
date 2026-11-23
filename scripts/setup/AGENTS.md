<a id="初始化工具環境準備-go-依賴工作指引"></a>

# Tool setup and Go dependency guidance

Inherit [parent guidance](../AGENTS.md); keep setup rerunnable and verify mise download checksums.

After changes, run checks from the [script overview](../README.md); see the [local README](README.md) for entry points.

## Applicable engineering conventions

- **Pinned configuration:** Install declared tool/dependency versions, validate platform/checksums and preserve manifest/lockfile consistency.
- **Idempotency:** Make repeated setup safe and failures explicit; do not silently upgrade tools or leave partially configured resources.
- **Security and lifecycle:** Keep writable state in the declared workspace paths and credentials out of installation logs or tracked configuration.
