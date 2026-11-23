<a id="契約來源"></a>

# Contract sources

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [internal-openapi.yaml](internal-openapi.yaml), [openapi.yaml](openapi.yaml), the [directory guide](README.md) and [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Projects and source references](../../docs/domains/project/project.md) · [Milestones](../../docs/domains/repository/milestones.md)

Applies to this directory and descendants. Inherit [parent guidance](../AGENTS.md) and its referenced root rules. This file adds only local responsibilities. User and host instructions take precedence; it grants no additional authorization for external writes.

<a id="職責與邊界"></a>

## Responsibilities and boundaries

- `openapi.yaml` composes public HTTP; `internal-openapi.yaml` composes internal worker HTTP. Entry points maintain only metadata and `$ref`; operations and DTOs belong to their capabilities.
- Maintain unique ownership through the [capability index](README.md). Capability files must not reference composition entry points. Cross-capability references follow ownership direction without cycles.
- Put shared transport definitions in common, but do not move business models there merely because they have several users. Name folders by capability; follow parent entity and REST resource naming rules.
- Keep scaffolds for Issue types/comments, Discussion categories and Project fields/views. Do not compose empty paths/schemas. Project/items and Repository milestones/statistics are wired according to [projects and source references](../../docs/domains/project/project.md), [milestones](../../docs/domains/repository/milestones.md) and [repository insights](../../docs/domains/repository/insights.md). Specify other operations before adding contracts or implementations.
- For new operations, update the [capability inventory](capabilities.json), adapters, consumers and feature specifications together. Do not add HTTP APIs for event entry points merely for directory consistency.
- Run `bash scripts/tooling/tool-exec.sh pnpm api:check` from the root after contract changes. For guidance-only changes, check links and whitespace. Specify the base revision separately for historical compatibility comparisons.

[Current implementation diagram: basic entry points](../../docs/contracts/http.md#基礎入口)

## Applicable engineering conventions

- **Schema and ownership:** Keep one owning capability for each operation and DTO, with shared transport definitions referenced rather than copied.
- **HTTP and querying:** Specify request validation, stable errors, pagination scope and ordering before wiring an operation or registering consumers.
- **Evolution:** Synchronize affected consumers and generated artifacts after schema or call-site changes, then run the existing API contract checks; retain compatible readers during expand-contract migrations.
- **Types and compatibility:** Preserve field casing, optionality and existing operation/error semantics; make incompatible schema evolution an explicit migration.
