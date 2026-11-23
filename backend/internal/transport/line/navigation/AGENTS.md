<a id="backendinternaltransportlinenavigation-工作指引"></a>

# backend/internal/transport/line/navigation agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [navigation events](events.go), [menu presentation](menu.go), the [directory guide](./) and [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: formats and data compatibility](../../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## Applicable engineering conventions

- **UI and identifiers:** Keep menu actions, aliases and deep-link targets consistent with shared menu definitions and registered page routes.
- **Events and safety:** Route verified postbacks to existing operations; navigation hints never replace session or resource authorization.
- **Compatibility:** Test old entry links and menu-action mappings when changing navigation, without publishing remote menus as a side effect.
