<a id="shadcn-本機參考庫"></a>

# Local shadcn references

Local entry points for official shadcn documentation.

Existing workflows generate snapshots/indexes; frontend UI guidance owns project component rules.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [snapshot/INDEX.md](snapshot/INDEX.md).

<a id="相關規格"></a>

## Related specifications

[Generated index](snapshot/INDEX.md) · [MIT license](LICENSE.md)

Contains every item linked by official [Components](https://ui.shadcn.com/docs/components)/[AI Elements](https://ui.shadcn.com/docs/ai-elements) entry pages for Codex discovery. The [generated index](snapshot/INDEX.md) alone maintains the full list, purposes, Base UI/Radix links and deprecations; do not duplicate catalogs manually.

<a id="搜尋與選用"></a>

## Search and selection

For UI tasks, search the complete index, read relevant Radix docs/examples, then verify installed content with pinned CLI/registry. Follow [UI guidance](../../../../../frontend/miniapp/src/shared/ui/AGENTS.md#需求驅動的選用與驗收) for protection, proactive selection and acceptance.

Run from the repository root:

```bash
rg -n -i 'select|search|dialog|chat|reasoning' docs/development/tooling/frontend/shadcn-reference/snapshot/INDEX.md
rg -n '## |```tsx|Installation' docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/chat.md
bash scripts/tooling/tool-exec.sh shadcn info --json --cwd frontend/miniapp
```

Do not inject the whole reference library into every task. It is not an installed-component list. Upstream aliases, routers, `@latest`, model services and install commands are examples, not project policy/authorization. Adapt to this project's base, aliases, Pages Router and pinned CLI.

<a id="更新與完整性"></a>

## Updates and integrity

The [download script](../../../../../scripts/tooling/shadcn-reference.py) updates references without extra Python packages:

```bash
# Download the current official catalog and Base UI/Radix Markdown; replace snapshots only after validation.
bash scripts/tooling/tool-exec.sh python3 scripts/tooling/shadcn-reference.py

# Offline verification: coverage, hashes, sources, base markers, index and seven-day freshness.
bash scripts/tooling/tool-exec.sh python3 scripts/tooling/shadcn-reference.py --check
```

Run offline checks before UI work; proactively update missing, damaged or over-seven-day snapshots before searching. Reuse results if unchanged within a task. Failed updates preserve old snapshots; report dates/blockers. Old data supports exploration only; affected decisions require current official docs/registry evidence.

[manifest.json](snapshot/manifest.json) records requested/final URLs, fetch time, ETag/Last-Modified when present, SHA-256, descriptions, base and deprecation. Indexes come from current official entries without fixed counts; abort if catalogs change mid-update. HTTP snapshots are not immutable revisions or guaranteed to share a deployment version; do not claim matching pinned CLI versions.

<a id="內容邊界與授權"></a>

## Content boundaries and licensing

`snapshot/pages/` preserves upstream Markdown/examples without manual edits. It does not recursively download external APIs, images, videos, demo projects, all dependencies or unexpanded sources such as `<ComponentSource>`. Use pinned CLI previews for component source. Coverage is the complete documentation listed by two official entry pages, not a mirror of all registries or shadcn's entire website.

Both bases are retained to distinguish differences; this project prefers Radix. Index base mismatches or deprecations rather than presenting them as installable Radix implementations; Radix Toast, for example, recommends Sonner.

The upstream is shadcn/ui. Preserve its [MIT license](LICENSE.md), byte-compared against the fetched [official license](https://github.com/shadcn-ui/ui/blob/main/LICENSE.md); source URLs live in the manifest. Generated snapshots retain upstream headings/links, resolving relative URLs against the official site. Graphify/Repomix normally exclude snapshots; use `rg` or read relevant pages directly. Reference guidance, scripts and project rules remain analysis inputs.

<a id="本次查證"></a>

## Verification for this update

2026-11-18 (Asia/Taipei): read official Markdown entries/pages directly; scripts checked coverage/hashes. Graphify located tool relationships but its AST omits documentation rules; Repomix cross-checked guidance. Context7 resolve succeeded but queries hit monthly limits; no new results claimed. CLI previews reused earlier chat evidence; current official responses verified download formats. No components installed or UI/LINE acceptance performed.

Passed: download/offline integrity, `check-config.py`, maintained-document local links/anchors and `git diff --check`; temporary copies verified rejection of damaged/stale documents and preservation on download failure. Repomix cross-checked eight files. Graphify AST updated without LLM extraction.

[Tool index](../../README.md) · [shadcn guide](../shadcn.md) · [Documentation index](../../../../README.md)

[applicable guidance](../../AGENTS.md)
