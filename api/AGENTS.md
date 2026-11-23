# API module rules

This file adds API-specific rules to the [repository rules](../AGENTS.md). Directory purposes and entry points belong in the [root README](../README.md#巢狀目錄導覽). Detailed path constraints are indexed in [module constraints](../docs/standards/module-constraints.md).

- **Contract authority.** OpenAPI MUST remain the sole authority for HTTP requests, responses, status codes, authentication and stable errors. Business rules and authorization remain in the backend domain.
- **Ownership.** Each operation and DTO MUST have one owning capability. Shared transport definitions MAY be referenced, but business models MUST NOT be moved into common schemas merely because several consumers use them.
- **Compatibility.** Public paths, fields, types, error meanings and existing consumers MUST remain compatible unless an explicit migration defines the break.
- **Generation.** Contract changes MUST flow through Contract → Generate → Backend Adapter → Frontend Consumer. Generated Go and TypeScript files MUST NOT be edited manually.
- **Validation.** Contracts MUST distinguish missing, null and zero values and define pagination scope, filters, stable ordering and projections where applicable.
- **Verification.** API changes MUST run `bash scripts/tooling/tool-exec.sh pnpm api:check`; compatibility checks MUST name their base revision.

Business operations follow [the operation model](../docs/contracts/business-operations.md). Naming follows [API naming standards](../docs/standards/naming.md#業務實體集合與-api-命名).
