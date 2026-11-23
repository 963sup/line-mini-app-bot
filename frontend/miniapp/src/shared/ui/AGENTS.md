<a id="shadcn-官方元件指引與目錄"></a>

# Official shadcn component guidance and catalog

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [LICENSE.md](LICENSE.md), [accordion.tsx](accordion.tsx), [alert-dialog.tsx](alert-dialog.tsx), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Local references](../../../../../docs/development/tooling/frontend/shadcn-reference/README.md) · [Generated index](../../../../../docs/development/tooling/frontend/shadcn-reference/snapshot/INDEX.md)

Applies here and below; follow [MINI App guidance](../../../AGENTS.md)/[root guidance](../../../../../AGENTS.md). This is the `aliases.ui` (`@/shared/ui`) resolution directory for `components.json`, storing official shadcn components. Keep business composition outside it.

<a id="官方元件禁止修改"></a>

## Official component modification prohibited

- Codex must not directly modify, rewrite/delete existing official styles, structure, APIs, accessibility/interactions or bypass protection with global CSS, extra variants/internal class overrides. Also protects `../hooks/use-mobile.ts`/`../../shell/sidebar-07/`; see root guidance for full rules.
- Use existing public props, variants, children, data/events. Compose business needs in `src/shared/` application wrappers, `src/shell/` or `src/business/`. No business APIs, authorization/domain decisions in primitives. Layout className controls outer layout only, never official appearance.
- Preserve registry output for missing official components. Preview every affected file/dependency; stop if existing files would be overwritten. Never automatically run `--overwrite`, `add --all`, preset switches, reinitialization or full sidebar-07 reinstallation.
- Modify specific official components only under explicit current-task update/restore/change requests, after comparison and minimal edits per root rules. General page needs, upstream customization advice and this guide grant no such authority. Preserve aliases, client directives/accessibility fixes, including Progress Root `value`.

<a id="codex-主動使用-skill文件與-mcp"></a>

## Proactive Skill, documentation and MCP usage

Before adding, changing or reviewing any UI, proactively inspect official components here without waiting for repeated shadcn requests.

1. Read official [Skills](https://ui.shadcn.com/docs/skills) and installed [shadcn SKILL.md](../../../../../.agents/skills/shadcn/SKILL.md), then relevant composition/forms/styling references. Preserve upstream skills; local protection/pinned-tool rules override general customization/floating-version examples.
2. Use pinned CLI `info --json` for installed items, base, aliases/paths. Current setup is Radix, new-york, Pages Router, Tailwind v4/Lucide. `components.json`, `info` and package files determine actual values; references are not installation inventories.
3. Before development, proactively query available shadcn MCP registries/components/examples per official [Registry MCP](https://ui.shadcn.com/docs/registry/mcp)/[MCP guide](https://ui.shadcn.com/docs/mcp). Prefer available `get_project_registries`, `search_items_in_registries`, `view_items_in_registries`, `get_item_examples_from_registries`; host tools/list defines actual names.
4. For absent MCP tools, repair/verify through [host exposure diagnostics](../../../../../docs/development/tooling/frontend/shadcn.md#自主接線).
5. Run `docs <component> --base radix`, read actual official docs/examples and verify via Context7 per root guidance. Do not apply Base UI `render` or React Aria APIs to Radix `asChild`. Official catalogs may default to Base UI; return to project base.
6. Import/compose installed components. Preview only genuinely missing, in-scope additions. MCP add commands are suggestions: replace with pinned CLI and check registryDependencies, npm dependencies, CSS/every diff. Search results do not automatically justify installation.

Examples run from root; the wrapper returns there, so specify `--cwd frontend/miniapp`:

```bash
bash scripts/tooling/tool-exec.sh shadcn info --json --cwd frontend/miniapp
bash scripts/tooling/tool-exec.sh shadcn search @shadcn -q button --cwd frontend/miniapp
bash scripts/tooling/tool-exec.sh shadcn docs button --base radix --cwd frontend/miniapp
bash scripts/tooling/tool-exec.sh shadcn view @shadcn/button --cwd frontend/miniapp
bash scripts/tooling/tool-exec.sh shadcn add @shadcn/button --dry-run --cwd frontend/miniapp
bash scripts/tooling/tool-exec.sh shadcn add @shadcn/button --diff --cwd frontend/miniapp
```

Only after previews confirm necessary new official files without overwrites, run the same `add` without `--dry-run`. Use official `@shadcn`; other registries follow user/skill guidance, never guesses. Root `mise.toml` alone defines CLI versions; use that version's `--help`, not floating installations.

<a id="需求驅動的選用與驗收"></a>

## Requirements-driven selection and acceptance

Applies to all frontend UI work. Guidance-only changes verify documentation without demo screens.

1. **Define purpose first.** Specify operations, density, mobile/desktop constraints/states before component choices. Do not reshape business flows around components. Preserve authorization, forms, errors/unknown-result retry contracts.
2. **Search by purpose.** Check [snapshot freshness](../../../../../docs/development/tooling/frontend/shadcn-reference/README.md), full Components/AI Elements indexes and relevant docs beyond installed items; then official registry compositions/examples. Compare candidates when tradeoffs exist; directly reuse suitable patterns without artificial alternatives.
3. **Add as needed.** Official `@shadcn` is the specified default; no registry clarification needed. UI requests authorize required missing components after safe preview, without per-item confirmation. Keep necessary dependencies/output only. For overwrites, first compose existing components; otherwise apply protection authorization above.
4. **Compose applications.** Prefer public props, variants/official patterns. Check imports, base, child components/dependencies after adding; do not rebuild existing controls. Explore [AI Elements](https://ui.shadcn.com/docs/ai-elements) only for AI/chat needs, verifying registries/dependencies/APIs; names do not justify model services.
5. **Verify behavior/screens.** Run frontend types, tests/builds for UI changes. Use browsers for desktop/narrow layouts, long text/overflow, keyboard, focus entry/return, accessible names and applicable loading/empty/error/disabled/submitting states. Compilation alone proves no UX completion. Report browser/credential gaps; do not claim LINE-device acceptance.
6. **Deliver verifiable results.** State choices/additions, reasons, sources, actual checks/limitations. Report tool/source blockers and continue supported work. No blind installation, mandatory whole-repo bundles or unsolicited component galleries.

These are exploration directions, not fixed recipes; verify official examples for the current base:

| Purpose                    | Candidate components/compositions                                                                     |
| -------------------------- | ----------------------------------------------------------------------------------------------------- |
| Edit/add content           | Inline Field/InputGroup; compare Dialog/Sheet/Drawer by content length/context                        |
| Filter/select              | Native Select/Select for small fixed sets; Combobox/Command for search                                |
| Lists/details              | Item, Table/Data Table; Collapsible/Accordion for expansion, not long Tooltip details                 |
| Navigation/actions         | Sidebar, Breadcrumb, Tabs, Dropdown Menu; preserve routes/semantics                                   |
| Dangerous actions/feedback | Alert Dialog, Alert, Empty, Skeleton, Spinner, Progress by state; notices do not replace confirmation |

<a id="全部官方元件目錄"></a>

## Complete official catalog

The [local references](../../../../../docs/development/tooling/frontend/shadcn-reference/README.md)/[generated index](../../../../../docs/development/tooling/frontend/shadcn-reference/snapshot/INDEX.md) exclusively own complete Components/AI Elements lists, purposes, both bases/examples. Do not duplicate snapshots/counts. Search first, then read Radix pages; listings do not establish installed, supported or matching-base usability.

Follow reference update/integrity instructions; latest official catalogs reveal additions/renames. See [shadcn tooling](../../../../../docs/development/tooling/frontend/shadcn.md)/[frontend entry](../../../../README.md) for checks, sources/MCP diagnostics. Guidance-only work checks links/whitespace; installation/composition follows frontend checks. Source research does not establish UI, device or integration acceptance.

[Current implementation diagram: page and identity lifecycle](../../../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

## Applicable engineering conventions

- **Design-system compatibility:** Preserve protected component props, variants, semantics and upstream accessibility behavior; this guidance does not authorize primitive changes.
- **Forms and interaction:** Compose labels, descriptions, errors, focus and disabled states through public APIs rather than embedding domain rules in primitives.
- **Types and verification:** Keep generated component dependencies and types consistent and verify keyboard/responsive behavior for explicitly authorized updates.
