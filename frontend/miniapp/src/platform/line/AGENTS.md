<a id="frontendminiappsrcplatformline-工作指引"></a>

# frontend/miniapp/src/platform/line agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [sdk.ts](sdk.ts), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: MINI App trust flow](../../../../../docs/contracts/http.md#mini-app-信任流程)

## Applicable engineering conventions

- **Identity and integration:** Use the existing LIFF initialization/login boundary; provider profile data does not establish backend authorization.
- **Lifecycle and reliability:** Keep initialization and callbacks safe across remounts, and preserve intended navigation through supported login flows.
- **Verification:** Distinguish SDK mocks from actual LINE WebView/device acceptance and test unsupported or rejected SDK states.
