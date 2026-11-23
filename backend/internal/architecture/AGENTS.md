<a id="backendinternalarchitecture-工作指引"></a>

# backend/internal/architecture agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with the [directory guide](./) and [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: formats and data compatibility](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## Applicable engineering conventions

- **Dependency enforcement:** Express architectural boundaries through actual package imports and focused tests, without exemptions that hide new cycles.
- **Structure:** Verify callers and allowed dependency direction after package moves; a matching directory name alone is not evidence of a valid boundary.
- **Documentation:** Keep boundary tests, architecture diagrams and accepted decisions consistent with the same implementation revision.
