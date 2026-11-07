# Go development

Prefer net/http, html/template, encoding/json, context and log/slog where adequate.
Keep interfaces consumer-owned and packages focused on current requirements.
Propagate request contexts, bound external calls and release resources explicitly.
Handle errors with context; do not log secrets, access tokens or raw identity tokens.

Use the pinned mise toolchain and GOTOOLCHAIN=local. Normal verification uses
readonly modules. Change dependencies deliberately and review go.mod/go.sum.
Format modified Go files; run focused tests before shared acceptance checks.
Do not globally disable CGO: the race detector needs a working C toolchain.
See [tests](testing.md), [LINE](../integrations/line.md) and
[Firestore](../integrations/firestore.md) for boundary-specific requirements.
