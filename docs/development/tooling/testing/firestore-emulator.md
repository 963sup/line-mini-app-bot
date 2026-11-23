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

若連接埠已被其他服務占用，先確認服務身分，保留既有程序。測試驗收門檻見[整合驗收流程](commands.md#firestore-integration)。

[Firestore 整合索引](../../../integrations/firestore/README.md)

## CI 全套驗收

CI 在 emulator 就緒及靜態前端建置後，以已鎖定的 Playwright 環境執行 `go test -race -count=1 -json ./...`。由 `go list ./...` 取得預期套件，交給[結果檢查器](../../../../scripts/check/check-go-test-results.py)核對完成事件；任一測試／子案例失敗、跳過或套件缺席均失敗。沒有測試檔的套件層級 skip 可以保留。這使 Enterprise、Team 與其他 Firestore 案例不再只靠 Project 的兩個案例代表整合通過。

本機重用此門檻時須準備 `MINIAPP_BROWSER_EXPORT` 與固定版本 Playwright。局部檢查仍可依上方篩選，但不能把它回報成全套驗收。
