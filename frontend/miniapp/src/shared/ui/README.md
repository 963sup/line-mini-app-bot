<a id="frontendminiappsrcsharedui-目錄導覽"></a>

# frontend/miniapp/src/shared/ui directory guide

Shared shadcn components resolve `@/shared/ui`; feature modules own business screen composition.

Cross-feature reusable utilities/components; specific workflows remain in consuming features.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [LICENSE.md](LICENSE.md), [accordion.tsx](accordion.tsx), [alert-dialog.tsx](alert-dialog.tsx), [alert.tsx](alert.tsx), [avatar.tsx](avatar.tsx), [badge.tsx](badge.tsx), [breadcrumb.tsx](breadcrumb.tsx), [button.tsx](button.tsx); before editing, read [local agent guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Official references](../../../../../docs/development/tooling/frontend/shadcn-reference/README.md) · [Current implementation diagram: page/identity lifecycle](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

<a id="已安裝元件"></a>

## Installed components

2026-11-19: all 61 `registry:ui` items from pinned shadcn CLI `@shadcn` registry installed, plus official Radix Code Block, Questionnaire/Trace: 64 TSX modules. The latter and Code Block hooks/highlighter use official `radix-rhea`; project retains `new-york`, without overwriting existing components/global styles. Official Combobox uses Base UI without switching other components.

Inspect installed items with `bash scripts/tooling/tool-exec.sh shadcn info --json --cwd frontend/miniapp`. See [official references](../../../../../docs/development/tooling/frontend/shadcn-reference/README.md) for purposes/Radix examples. Do not reinstall everything or overwrite files.

Data Table, Date Picker/Typography are composition/style recipes using Table, Calendar, Popover/examples, not new business frameworks. Sonner replaces deprecated Toast. Installation neither replaces screens nor proves integration acceptance.

[Current implementation diagram: page and identity lifecycle](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

<a id="前端使用位置"></a>

## Frontend usage

Classified by static imports in `src`. Direct means imported outside `shared/ui`; indirect traces those UI dependencies. Actual rendering depends on routes, data/state.

| Purpose | 30 directly imported modules |
|---|---|
| Navigation/framing | accordion, avatar, breadcrumb, collapsible, dropdown-menu, separator, sidebar, tabs |
| Forms/actions | button, button-group, checkbox, combobox, field, input, input-group, native-select, questionnaire, radio-group, textarea, toggle-group |
| Data/feedback | alert, alert-dialog, badge, card, chart, empty, item, progress, spinner, table |

| Indirect | Entry/purpose |
|---|---|
| label | FieldLabel uses Label; forms obtain labels through FieldLabel/FormField. |
| sheet | Mobile sidebar uses Sheet with title, description/focus management. |
| tooltip | Sidebar TooltipProvider enables configured icon-collapse menu labels. |
| skeleton | SidebarMenuSkeleton uses Skeleton; dependencies render placeholders only when that component renders. |
| toggle | toggle-group shares variants; styling dependency does not mean separate Toggle rendering. |

30 direct plus 5 indirect dependencies yield 35 reachable modules, not rendered instance counts. Remaining 29 installed official modules are not referenced from those entries:

- Forms, dates, pagination/overlays: calendar, dialog, drawer, form, pagination, popover, select, slider, sonner, switch.
- Navigation/advanced interaction: command, context-menu, hover-card, input-otp, kbd, menubar, navigation-menu, resizable, scroll-area.
- Media, messages/tracing: aspect-ratio, attachment, bubble, carousel, code-block, direction, marker, message, message-scroller, trace.

After import changes, use `rg -n 'from .*ui/' frontend/miniapp/src` for application/shared entries, then inspect component imports including relative paths. Preserve semantics: work links search complete lists; repositories load cursors; Issue status queries apply filters; tabs load after confirmed switches.

Forms use FieldSet/FieldLegend/FieldLabel including FormField; loading/messages use ReadFeedback/Status; expandable details use Collapsible. Date/month/number/file inputs retain native types/validation. Do not replace semantic forms, text, links, lists, legends or diagnostics to increase component counts. Unauthenticated explanations may remain text without duplicating Shell identity notices.

[Parent directory guide](../README.md)
