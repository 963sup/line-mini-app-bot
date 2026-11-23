<a id="設定檢查"></a>

# Configuration checks

Entry: check-config.

- [check-config.py](check-config.py)

Run `bash scripts/tooling/tool-exec.sh python3 scripts/check/check-config.py` from the repository root. It validates configuration and inline local Markdown links across Git-tracked and untracked, non-ignored project documents, including nested README/AGENTS files. Link checks accept directory targets and validate Markdown ATX heading fragments and explicit HTML anchors, including same-file fragments.

Fenced examples and inline code are skipped. Upstream skills under `.agents/` and `.claude/`, and shadcn snapshot pages, are excluded; the snapshot index remains checked. Reference-style links, HTML links, Setext headings and remote URLs are outside this check's coverage.

Regression cases: `bash scripts/tooling/tool-exec.sh python3 -m unittest discover -s scripts/test -p 'test_documentation_links.py'`.

[Script overview](../README.md) · [Local guidance](AGENTS.md)
