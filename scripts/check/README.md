<a id="設定檢查"></a>

# Configuration checks

Entry: check-config.

- [check-config.py](check-config.py)
- [check-go-test-results.py](check-go-test-results.py) · Validate a `go test -json` report against `go list ./...` package output; failed, missing or skipped test cases fail acceptance.

Run `bash scripts/tooling/tool-exec.sh python3 scripts/check/check-config.py` from the repository root. It validates configuration, Cloud Build script upload coverage and supported local documentation links across Git-tracked and untracked, non-ignored project documents, including nested README/AGENTS files. Link checks accept directory targets and validate Markdown ATX heading fragments and explicit HTML anchors, including same-file fragments.

Reference definitions, HTML hrefs, Mermaid click links and standalone `.mermaid` sources are also checked. Other fenced examples and inline code are skipped. Upstream skills under `.agents/` and `.claude/`, and shadcn snapshot pages, are excluded; the snapshot index remains checked. Setext headings and remote URL reachability remain outside coverage. The document graph reuses the same link extraction; fragment validation belongs to this checker.

Cloud Build coverage checks literal `scripts/...` entries and statically discoverable sibling Python helpers against explicit `.gcloudignore` rules using Git’s ignore engine. SDK include directives fail explicitly until supported; computed runtime paths require separate release verification. Verify SDK behavior with `gcloud meta list-files-for-upload` before publishing a new build path.

Regression cases: `bash scripts/tooling/tool-exec.sh python3 -m unittest discover -s scripts/test -p 'test_documentation_links.py'`.

[Script overview](../README.md) · [Local guidance](AGENTS.md)
