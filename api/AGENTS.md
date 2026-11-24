# API module rules

This file adds API-specific rules to the [repository rules](../AGENTS.md). Directory purposes and entry points belong in the [root README](../README.md#巢狀目錄導覽). Detailed path constraints are indexed in [module constraints](../docs/standards/module-constraints.md).

- **MCP tools.** Contributors SHOULD use `context7` for version-specific OpenAPI tooling and generator behavior, and `gopls` to trace affected generated Go types into their adapter callers.
- **Contract authority.** OpenAPI MUST remain the sole authority for HTTP requests, responses, status codes, authentication and stable errors. Business rules and authorization remain in the backend domain.
- **Ownership.** Each operation and DTO MUST have one owning capability. Shared transport definitions MAY be referenced, but business models MUST NOT be moved into common schemas merely because several consumers use them.
- **Wire conventions.** New or changed operations MUST follow the [resource and operation naming rules](../docs/standards/naming.md#業務實體集合與-api-命名) and the [HTTP conventions](../docs/contracts/http.md#api-一致性). Existing published identifiers retain the compatibility rule below.
- **DTO boundary.** OpenAPI schemas MUST describe the public request and response shapes. Generated DTOs MUST be mapped at the HTTP adapter boundary before entering the application or domain, and MUST NOT serve as domain models or Firestore records.
- **Compatibility.** Public paths, fields, types, error meanings and existing consumers MUST remain compatible unless an explicit migration defines the break.
- **Generation.** Contract changes MUST flow through Contract → Generate → Backend Adapter → Frontend Consumer. Generated Go and TypeScript files MUST NOT be edited manually.
- **Validation.** Contracts MUST distinguish missing, null and zero values and define pagination scope, filters, stable ordering and projections where applicable.
- **Operation semantics.** Business operations MUST define required authentication, scoped resource identifiers, request validation, the observable success result, and applicable denial or conflict responses. Use the published [session and CSRF contract](../docs/contracts/identity.md) and [shared error envelope](contract/common/transport.yaml) for applicable operations. Safe HTTP methods MUST NOT request a business state change. When a write can be retried or can race with another write, its contract MUST state the supported retry and concurrency behavior.
- **Verification.** API changes MUST run `bash scripts/tooling/tool-exec.sh pnpm api:check`. Changes to published HTTP contracts MUST also run `bash scripts/tooling/tool-exec.sh pnpm api:compat --base <revision>` against a named baseline. Contract changes MUST verify affected adapter and consumer behavior and update any capability entries or documentation whose references or behavior changed.

Business operations follow [the operation model](../docs/contracts/business-operations.md).
