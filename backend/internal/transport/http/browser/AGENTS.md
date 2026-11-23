<a id="backendinternaltransporthttpbrowser-工作指引"></a>

# backend/internal/transport/http/browser agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [files.go](files.go), [routes.go](routes.go), [directory guide](README.md), [parent guidance](../../../AGENTS.md).

<a id="工作方式"></a>

## Working approach

[Current implementation diagram: formats and data compatibility](../../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

[parent guidance](../AGENTS.md)

## Applicable engineering conventions

- **Routes and compatibility:** Keep static exports, route registration and legacy redirects synchronized; preserve allowed query parameters without permitting external redirects.
- **Security and caching:** Preserve CSP, private-page cache rules and asset integrity while changing page dispatch.
- **Acceptance:** Build static assets and test real registered routes, redirects and missing-page behavior before removing an old page.
