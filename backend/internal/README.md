<a id="backendinternal-目錄導覽"></a>

# backend/internal directory guide

Stores internal Go backend packages. Features live in `business/`, service composition in `bootstrap/`; other packages provide shared identity, transport, navigation and data-access foundations.

Directories expose backend entry points or shared capabilities by purpose. Bootstrap wires services; domains own resource authorization and use cases.

[linegroup](linegroup/README.md) owns LINE group observation and legacy query compatibility.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [architecture](architecture), [static pages](transport/http/browser), [business](business), [location diagnostics](transport/http/geolocation), [httpapi](transport/http/api) and [identity](identity). Read the [local guidance](AGENTS.md) before editing.

<a id="bounded-context-分層範本"></a>

## Bounded context layering reference

This is a **standard reference structure, adopted as needed** for one bounded context. Business contexts live in `backend/internal/business/<context>/`. Simple features may use fewer packages; add layers/subdirectories only for clear responsibilities. Do not fill the entire tree or move code merely to match the template.

```text
internal/business/
└── <Bounded Context>/
    ├── domain/                  # Domain layer
    │   ├── entity/              # Entity / Aggregate Root
    │   ├── valueobject/         # Value Object
    │   ├── service/             # Domain Service
    │   ├── event/               # Domain Event
    │   └── policy/              # Domain Policy
    ├── application/             # Application layer
    │   ├── command/             # Write operations / Use cases
    │   ├── query/               # Queries / Use cases
    │   ├── port/                # Output Port / Interface
    │   └── dto/                 # Application DTO
    └── adapter/                 # Adapter layer
        ├── http/                # HTTP Handler
        ├── line/                # LINE Webhook / API
        ├── firestore/           # Database Adapter
        ├── messaging/           # Message Queue Adapter
        └── scheduler/           # Scheduled Trigger
```

Trace features from adapter entry points into application use cases and domain rules. Adapters convert external requests/data formats; `bootstrap/` wires layers together.

<a id="驗證入口"></a>

## Verification entry points

See [backend guidance](../AGENTS.md#驗證) for Go checks/tests. For Firestore transaction integration tests, check package tests and emulator prerequisites.

<a id="閱讀入口"></a>

## Reading entry points

[Current implementation diagram: formats and data compatibility](../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

<a id="共用技術與入口"></a>

## Shared technology and entry points

[platform](platform) provides document keys, pagination cursors and gRPC error mapping. [bootstrap](bootstrap) creates clients/wiring. [navigation](transport/line/navigation) builds cross-feature menus. [httpapi](transport/http/api) maintains generated OpenAPI contracts and dispatch.

[Parent directory guide](../README.md)
