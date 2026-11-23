import type { useRepositories } from "../business/repository/RepositoryContext";
import type { NavItem } from "./sidebar-07/nav-main";

export function repositoryNavigation(spaces: ReturnType<typeof useRepositories>, workflow: string): NavItem {
  return {
    title: "儲存庫", url: spaces.scope ? spaces.href(spaces.scope) : "/miniapp/repositories", isActive: workflow === "repositories", onExpand: spaces.ensure,
    items: [
      { title: "儲存庫總覽", url: "/miniapp/repositories", isActive: workflow === "repositories" && !spaces.selectedID && !spaces.scope },
      ...spaces.items.map((item) => ({ title: item.name || "群組空間", url: spaces.href(item.id), isActive: item.id === spaces.selectedID })),
      ...(spaces.status ? [{ title: spaces.status, status: true }] : []),
      ...(spaces.cursor ? [{ title: "載入更多空間", onSelect: spaces.loadMore, disabled: spaces.busy, keepOpen: true }] : []),
      ...(spaces.catalogFailed ? [{ title: "重試讀取空間", onSelect: spaces.retryCatalog, keepOpen: true }] : []),
    ],
  };
}
