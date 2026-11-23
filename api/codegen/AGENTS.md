<a id="共用契約生成設定"></a>

# Shared contract generation settings

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [Redocly lint/bundle settings](redocly.yaml), [go.yaml](go.yaml), [internal-go.yaml](internal-go.yaml), the [directory guide](./) and [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

Applies to this directory and descendants. Inherit [parent guidance](../AGENTS.md) and its referenced root rules. This file adds only local responsibilities. User and host instructions take precedence; it grants no additional authorization for external writes.

<a id="職責與邊界"></a>

## Responsibilities and boundaries

- Maintain public and internal contract generation settings. Root scripts, package.json and mise.toml define workflows and versions; do not create per-capability generators.
- Keep public and internal entry points separate. Generate HTTP adapter DTOs without making the business core depend on generated packages.
- Change source contracts and configuration first, then generate from the root and inspect the diff. Do not edit artifacts or refresh compatibility baselines to conceal problems.
- Run `bash scripts/tooling/tool-exec.sh pnpm api:check` after changing generation settings or workflows to verify synchronization and compatibility.

[Current implementation diagram: basic entry points](../../docs/contracts/http.md#基礎入口)

## Applicable engineering conventions

- **Reproducible generation:** Use pinned generators and the existing entry points to rebuild artifacts from schemas; review output drift before committing.
- **Structure and dependencies:** Keep public/internal generation separate and generated transport types outside domain packages.
- **Compatibility:** Preserve operation names, field casing and identifier mappings unless the source contract explicitly changes them.
