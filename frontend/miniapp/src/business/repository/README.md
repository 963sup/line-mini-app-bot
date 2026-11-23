<a id="frontendminiappsrcbusinessrepository-目錄導覽"></a>

# frontend/miniapp/src/business/repository directory guide

Manages Repository/space selection, page navigation and feature composition.

Colocate feature UI/interactions; pages compose routes and backend APIs determine permissions/results.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [NavigationGuard.tsx](NavigationGuard.tsx), [RepositoryContext.tsx](RepositoryContext.tsx), [Spaces.tsx](Spaces.tsx), [View.tsx](View.tsx), [labels](labels), [milestones](milestones), [navigation.ts](navigation.ts); before editing, read [local agent guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: space loading sequence](../../../../../docs/domains/repository/README.md#空間載入時序)

[Parent directory guide](../README.md)

<a id="下層文件入口"></a>

## Child documentation entry points

- [insights](insights/README.md)

Tabs use `next/dynamic` to load features on opening, preserving unsaved-form guards/loading feedback.
