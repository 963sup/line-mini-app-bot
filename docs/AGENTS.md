# Documentation module rules

This file adds documentation-specific rules to the [repository rules](../AGENTS.md). Follow the [documentation standard](standards/documentation.md).

- **Authority.** Each rule, business fact, architecture statement and operation MUST have one authoritative document. Other pages link to it rather than restating it.
- **Navigation.** README files own directory purpose and reader navigation. AGENTS files contain only executable constraints that a parent cannot cover.
- **Status.** Documents MUST distinguish current implementation, accepted decisions, proposals and unknowns.
- **Compatibility.** Moves and consolidation MUST preserve valid anchors or update every inbound link.
- **Verification.** Documentation changes MUST verify local links, anchors and sources, run `bash scripts/tooling/tool-exec.sh pnpm docs:check`, and run `git diff --check`.
