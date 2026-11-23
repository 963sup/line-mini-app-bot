<a id="backendinternal-工作指引"></a>

# backend/internal agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [architecture](architecture), [static pages](transport/http/browser), [business](business), [location diagnostics](transport/http/geolocation), the [directory guide](README.md) and [parent guidance](../AGENTS.md).

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

Dependencies point inward. Applications coordinate domain rules and define required interfaces. Adapters call use cases or implement output ports; domains remain independent of HTTP, LINE and Firestore. `bootstrap/` composes dependencies.

<a id="修改與驗證"></a>

## Changes and verification

The resource-owning domain authorizes business operations. Identify transaction boundaries for consistency-sensitive authorization/state checks. Update OpenAPI before public HTTP formats and check existing data compatibility before persistence formats. Follow [backend guidance](../AGENTS.md#驗證) and affected tests for Go changes; set `FIRESTORE_EMULATOR_HOST` before Firestore emulator tests.

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: formats and data compatibility](../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## Applicable engineering conventions

- **Structure and naming:** Place each change in its existing bounded context and responsibility package; do not create empty layers to match a template.
- **Types and dependencies:** Keep domain code independent of transport/storage SDKs and inject ports at composition entry points.
- **Data evolution:** Preserve resource IDs, stored encodings and old-reader compatibility when moving models or changing serialization.
