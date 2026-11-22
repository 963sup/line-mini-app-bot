# reviewer 能力指引

由目前代理按需採用；這是工作階段，不是可啟動的子代理。共同規則見[能力入口](README.md)。審查結果為自我審查，不宣稱獨立驗收。

## 執行要求



Compare the operation contract with actual entry points, authorization, domain rules and results. Remain read-only; report bypasses, inconsistencies and unnecessary abstraction with evidence.

Review in a separate read-only pass the authorized diff for correctness, maintainability,
security and architectural compliance. Examine relevant tests and contracts;
check Go errors, contexts, concurrency and external boundaries when applicable.
Remain read-only, running only safe scoped checks. Report actionable findings
with file/line, severity, rationale, reproduction evidence and a minimal remedy.
Distinguish confirmed defects from uncertainties and avoid speculative warnings.
Record findings and residual risks in the final delivery as self-review, without claiming independent approval.

Inspect LINE identity verification, webhook signature handling, duplicate delivery, request size limits, Firestore user ownership checks and logging of sensitive data when applicable. Review relevant contracts rather than requiring every check for every diff.
