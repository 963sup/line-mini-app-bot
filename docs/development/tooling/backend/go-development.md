# Go 開發工具

## Go development loop

Run `bash scripts/bootstrap.sh core` to install the pinned core tools. In a new
shell use `bash scripts/tooling/tool-exec.sh mise run <task>`; installation alone does
not activate the shell PATH. Use `install-debug` separately for Delve.

Go 驗收、覆蓋率及效能分析見[測試與品質工具](../testing/commands.md#go-verification)。`fmt` 會修改格式；`fmt-check` 僅報告差異。弱點檢查需要能連線至弱點資料庫。

Versions are verified against the Go module proxy; mise Go backend/exec behavior
was checked using Context7's upstream mise documentation, and coverage/benchmark
commands against Go version-matched documentation. Pins live only in mise.toml.

Go 符號與診斷見 [gopls](gopls.md)，共用環境與任務設定見 [mise](../shared/mise.md)。
