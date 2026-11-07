# Go development

Prefer net/http, html/template, encoding/json, context and log/slog where adequate.
Keep interfaces consumer-owned and packages focused on current requirements.
Propagate request contexts, bound external calls and release resources explicitly.
Handle errors with context; do not log secrets, access tokens or raw identity tokens.

Use the pinned mise toolchain and GOTOOLCHAIN=local. Normal verification uses
readonly modules. Change dependencies deliberately and review go.mod/go.sum.
Format modified Go files. Test development is deferred during the current foundation
phase; keep existing CI. Follow the [accepted baseline](../architecture/decisions/0002-baseline.md)
for current scope and [testing guidance](testing.md) when that phase resumes.
Do not globally disable CGO: the race detector needs a working C toolchain.
See [tests](testing.md), [LINE](../integrations/line.md) and
[Firestore](../integrations/firestore.md) for boundary-specific requirements.

架構與抽象選擇依 [開發原則與決策](../architecture/principles.md)；層級責任見 [應用架構](../architecture/overview.md)。

檔案、函數、介面與 DTO 命名依 [命名規則](naming.md)。
