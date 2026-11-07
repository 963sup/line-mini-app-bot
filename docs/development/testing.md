# Testing

階段狀態：依 [已接受的基線](../architecture/decisions/0001-baseline.md)，目前先穩定基礎，
不新增測試方案、案例或工具。以下內容保留為後續階段指引；本輪文件同步不執行測試。
既有測試與 CI 保留。

Use `bash scripts/tool-exec.sh <command>` for all Go tooling.
Run gofmt checks on modified files and focused package tests during development.
Before delivery of Go changes run `mise run check` and `mise run build` through
that wrapper. These tasks verify modules, vet, race tests and compilation.

Test invalid LINE signatures/tokens, duplicate events, authorization failures and
cancelled contexts when affected. Report unrun integration checks explicitly.
Go doctor uses a temporary fixture; it does not test the application.
See the [acceptance matrix](../agents/acceptance.md).
