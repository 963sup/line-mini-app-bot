# Firestore emulator

<a id="firestore-emulator"></a>

以下命令從儲存庫根目錄執行。先檢查 `curl -fsS http://127.0.0.1:8080/`，確認端點是已啟動的 emulator 後可直接沿用。若尚未啟動，在獨立終端機或背景工作階段按需準備並啟動：

```bash
bash scripts/tooling/tool-exec.sh mise run firestore-emulator
```

等本機端點可達後，在另一個終端機指定 `FIRESTORE_EMULATOR_HOST`，用受影響的案例驗證。例如：

```bash
FIRESTORE_EMULATOR_HOST=127.0.0.1:8080 \
  bash scripts/tooling/tool-exec.sh go -C backend test \
  ./internal/business/repository/adapter/http/wiki \
  -run '^TestWikiMembershipAndRevisions$' -count=1 -v
```

若連接埠已被其他服務占用，先確認服務身分，保留既有程序。測試驗收門檻見[整合驗收流程](../../development/verification/commands.md#firestore-integration)。

[Firestore 整合索引](README.md)
