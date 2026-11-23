<a id="frontendminiappsrcstyles-工作指引"></a>

# frontend/miniapp/src/styles agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [globals.css](globals.css), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: page and identity lifecycle](../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

## Applicable engineering conventions

- **UI and formatting:** Use existing design tokens, theme variables and scoped utilities; do not override protected component internals globally.
- **Compatibility:** Check narrow layouts, long content, contrast and focus visibility when changing shared styles.
- **Builds:** Preserve the existing stylesheet/toolchain entry points and verify affected static output without adding a parallel styling system.
