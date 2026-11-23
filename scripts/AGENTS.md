<a id="scripts-工作指引"></a>

# scripts agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Select subdirectories through the [script overview](./), read their README/AGENTS, then verify under [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Module entries](../docs/) · [Documentation maintenance](../docs/AGENTS.md#主題子目錄)

Applies here and below; follow [root guidance](../AGENTS.md). User/host instructions take precedence.

<a id="職責與邊界"></a>

## Responsibilities and boundaries

File scripts by purpose through [categories](./); each shell script owns one operation.
mise.toml owns versions, non-secret settings/tasks. Scripts implement tasks without duplicating secrets or granting deployment authority.

<a id="文件與驗證"></a>

## Documentation and verification

Read [related documents](./)/[module entries](../docs/); classify cross-module specifications under docs. Follow [maintenance guidance](../docs/AGENTS.md#主題子目錄) for indexes.
Run from root through scripts/tooling/tool-exec.sh. Script changes run configuration/relevant smoke checks; prose-only changes check links/git diff --check. Never claim unexecuted deployments/acceptance.

## Applicable engineering conventions

- **Structure and configuration:** Keep scripts grouped by operation, with versions/settings owned by the existing configuration sources.
- **Reliability and errors:** Propagate failures with nonzero status, bound external calls and clean up temporary resources on cancellation or partial failure.
- **Security and delivery:** Redact credentials and verify the intended environment before side effects; script execution does not expand publication authority.
