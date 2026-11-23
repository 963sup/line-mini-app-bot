<a id="設定檢查工作指引"></a>

# Configuration check guidance

Inherit [parent guidance](../AGENTS.md); failures must exit nonzero for CI detection.

After changes, run checks from the [script overview](../); see the [local README](./) for entry points.

## Applicable engineering conventions

- **Automated enforcement:** Validate the authoritative configuration/schema relationships and fail nonzero for drift, missing consumers or invalid definitions.
- **Test integrity:** Keep checks deterministic and offline where specified; do not relax assertions, exclusions or baselines to hide a regression.
- **Diagnostics:** Report the concrete offending path/rule while redacting secrets and leaving source files unchanged.
