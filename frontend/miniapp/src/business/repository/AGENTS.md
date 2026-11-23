<a id="frontendminiappsrcbusinessrepository-工作指引"></a>

# frontend/miniapp/src/business/repository agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [NavigationGuard.tsx](NavigationGuard.tsx), [RepositoryContext.tsx](RepositoryContext.tsx), [Spaces.tsx](Spaces.tsx), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: space loading sequence](../../../../../docs/domains/repository/design.md)

## Applicable engineering conventions

- **Data, scope and state:** Keep repository selection and navigation guards synchronized; invalidate stale scopes on identity changes and retain unresolved write protection.
