# Script module rules

This file adds script-specific rules to the [repository rules](../AGENTS.md). Script purposes and entry points belong in the [root README](../README.md#巢狀目錄導覽). Detailed path constraints are indexed in [module constraints](../docs/standards/module-constraints.md).

- **Configuration authority.** `mise.toml` and existing lockfiles own tool versions and non-secret settings. Scripts MUST NOT introduce hidden upgrades or duplicate configuration sources.
- **Reliability.** Failures MUST exit nonzero and preserve the originating diagnostic. External calls and child processes MUST have bounded lifetimes, cancellation and cleanup.
- **Reproducibility.** Generation and build scripts MUST use declared inputs and pinned tools. Generated outputs MUST be repaired through their source or generator.
- **Security.** Scripts MUST redact credentials, verify targets before side effects and keep temporary artifacts out of published outputs.
- **Verification.** Script changes MUST run the relevant configuration, unit or smoke checks. A successful probe MUST NOT be reported as deployment or integration acceptance.
