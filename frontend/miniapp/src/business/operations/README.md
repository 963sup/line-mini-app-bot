<a id="文件工具"></a>

# Document tools

`Poppler.tsx` provides PDF extraction through `/miniapp/poppler`. Reuse sessions/generated clients; show loading, error, empty/success states and page markers, preserving page breaks in downloads. No persistence or AI submission.

Verification: run `pnpm --filter @repo/miniapp check`, `test` and `build` from root through `scripts/tooling/tool-exec.sh`.
