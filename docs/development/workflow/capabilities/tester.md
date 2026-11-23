# tester 能力指引

由目前代理按需採用；這是工作階段，不是可啟動的子代理。共同規則見[能力入口](./)。審查結果為自我審查，不宣稱獨立驗收。

## 執行要求

Verify contract outcomes using applicable checks, including denial, illegal transitions and write retry/concurrency when relevant. Current acceptance phase restrictions override general test-authoring duties; do not invent passing evidence.

Design and execute meaningful tests for the assigned behavior and failure
modes. May edit only assigned test/fixture files; record application fixes before switching to the authorized implementation stage. Validate bugs with
reproduction evidence and verify fixes against it. Choose focused Go tests,
race checks and integration checks according to risk, avoiding redundant suites.
For environment/MCP work distinguish configured, installed, started and
functionally verified states. Record commands, actual outcomes, skips and
external prerequisites. Never weaken assertions or fabricate passing results.

Use the acceptance matrix in docs/development/tooling/testing/commands.md. Exercise invalid webhook signatures, duplicate events, cancelled contexts and authorization failures when relevant. Use emulator-backed Firestore tests without production credentials; report missing services as skipped.
