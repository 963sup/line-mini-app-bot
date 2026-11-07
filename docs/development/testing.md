# Testing

Use `bash scripts/tool-exec.sh <command>` for all Go tooling.
Run gofmt checks on modified files and focused package tests during development.
Before delivery of Go changes run `mise run check` and `mise run build` through
that wrapper. These tasks verify modules, vet, race tests and compilation.

Use emulator-backed Firestore tests with explicit local configuration and test
project IDs. Never fall back to production when an emulator is missing.
Test invalid LINE signatures/tokens, duplicate events, authorization failures and
cancelled contexts when affected. Report unrun integration checks explicitly.
Go doctor uses a temporary fixture; it does not test the application.
See the [acceptance matrix](../agents/acceptance.md).
