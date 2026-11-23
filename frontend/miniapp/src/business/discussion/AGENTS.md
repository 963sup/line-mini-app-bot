<a id="frontendminiappsrcbusinessdiscussion-工作指引"></a>

# frontend/miniapp/src/business/discussion agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [Comments.tsx](Comments.tsx), [View.tsx](View.tsx), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: minimal boundaries and verification](../../../../../docs/domains/repository/discussion/design.md#最小實作邊界與驗證)

## Applicable engineering conventions

- **Data, scope and state:** Keep reply/edit drafts and pagination independent, and use separate server capabilities for editing and classification.
