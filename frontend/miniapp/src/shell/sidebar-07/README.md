<a id="sidebar-07-專案整合"></a>

# sidebar-07 project integration

Stores the shadcn sidebar-07 block for application-shell composition.

Owns shared layout/navigation; features own business interactions.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [app-sidebar.tsx](app-sidebar.tsx), [nav-main.tsx](nav-main.tsx), [nav-projects.tsx](nav-projects.tsx), [local agent guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

Installed from official `@shadcn/sidebar-07` through pinned CLI `add`, then moved here.
Official entry: https://ui.shadcn.com/blocks/sidebar . mise and project shadcn guidance own versions/installation.

Preserve AppSidebar, NavMain, NavProjects, NavUser, SidebarRail/official header structure/interactions. AppFrame is the sole application shell using it.

- NavMain: existing service entries and collapsible attendance/space submenus.
- NavProjects: real tools/new-tab actions, not fake project delete/share features.
- [OrganizationSwitcher](../OrganizationSwitcher.tsx): authentication status/organization menu; switching persists preferences then navigates client-side. NavProjects provides location diagnostics.
- NavUser: actual Shell identity/login/verify/logout actions; location diagnostics creates no LINE sessions.

Remove upstream demo users, companies, subscriptions, billing/empty dashboards. Business content retains existing routes.
Do not overwrite shared/ui primitives; add only Breadcrumb/Collapsible required by this block.

<a id="元件保護與官方比對"></a>

## Protection and official comparison

Follow [frontend protection](../../../../AGENTS.md#shadcn-元件保護); page requests do not authorize appearance/interaction changes. Real data/routes/identity are required integration, never restore demo accounts, billing/fake links.

Historical comparison used CLI 4.23.0 new-york for 24 primitives. Apart from aliases/client directives, Button returned to radix-ui Slot.Root; Progress retains accessible Root value. sidebar.tsx matches upstream. NavUser restores official avatar/two-line header while retaining real login. Registries are mutable; CLI versions do not pin registry content. No pixel/device acceptance claimed.

After server identity verification, obtain LINE avatars through LIFF `getProfile()` into frontend memory only. Clear on logout, reverification/pagehide; ignore stale responses. Never send profiles backend or use them to authorize. Allow only official `profile.line-scdn.net` HTTPS sources; Console/`profile` scope authorization required. Missing/failed images use AvatarFallback. Sidebar/account menus use AvatarImage without modifying Avatar.

Sources: official LINE main `dea09dc6e63fc5c6a3754abfbb304fc6fce4cc54` getProfile and Context7 `/shadcn-ui/ui` AvatarImage/AvatarFallback guidance; CLI stays 4.23.0.

[Current implementation diagram: page and identity lifecycle](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

[Parent directory guide](../README.md)
