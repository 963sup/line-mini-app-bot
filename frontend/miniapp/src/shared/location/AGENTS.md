<a id="frontendminiappsrcsharedlocation-工作指引"></a>

# frontend/miniapp/src/shared/location agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [position.ts](position.ts), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: attendance integration responsibilities](../../../../../docs/integrations/geolocation/browser-api.md#現有出勤整合檔案與責任) · [Geolocation specification](../../../../../docs/integrations/geolocation/browser-api.md)

[Current implementation diagram: attendance integration files and responsibilities](../../../../../docs/integrations/geolocation/browser-api.md#現有出勤整合檔案與責任)

Browser JavaScript obtains location; Go validates untrusted coordinates/permissions. Location history/presence validation requires product contracts. Read the [geolocation specification](../../../../../docs/integrations/geolocation/browser-api.md) first.

## Applicable engineering conventions

- **Types and time:** Preserve contract coordinate units, precision and sample timestamps; client checks provide feedback rather than trusted location authorization.
- **Lifecycle and reliability:** Stop watchers/timers on completion, cancellation or unmount and ignore obsolete callbacks.
- **Privacy:** Keep samples purpose-limited and do not add retention, background tracking or sensitive diagnostic logging implicitly.
