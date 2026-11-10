# 素材與腳本路徑整理

## 計畫與範圍

將 LINE Rich Menu 素材集中於 `assets/line/richmenu/`，將原本平鋪的腳本依責任
分為 `scripts/tooling/`、`scripts/analysis/`、`scripts/api/`、`scripts/release/`。
`docs/`、`.github/`、`.vscode/`、`api/` 保持根目錄；不新增框架或變更 API 契約。

1. 先用 Graphify CLI 與 Repomix CLI 建立來源覆蓋與引用基線。
2. 搬移檔案，修正 Bash、Python、Node 的儲存庫根目錄定位。
3. 同步 mise、package scripts、CI、VS Code、MCP、Docker 與來源上傳排除設定。
4. 更新維護中的文件引用並執行設定、腳本、API、前端與 Go 相關檢查。
5. 重建圖譜及交叉稽核，確認來源覆蓋與剩餘舊路徑。

## 使用方式

```bash
bash scripts/tooling/bootstrap.sh core
bash scripts/tooling/tool-exec.sh pnpm build
python3 scripts/tooling/check-config.py
bash scripts/tooling/tool-exec.sh mise run analysis-audit
```

完整腳本分類見 [scripts 索引](../../scripts/README.md)。不保留舊路徑轉接腳本；
外部環境中保存的 bootstrap 命令需改為新入口，已執行中的 MCP 程序需重新載入設定。

容器內素材也改為 `/assets/line/richmenu/`。外部 Rich Menu 工作若明確指定舊的
`-images` 路徑，需配合更新；本次不修改外部工作或執行 LINE 發布。

## 驗收與來源

驗收包含遞迴 shell 語法檢查、工具 wrapper 根目錄、API 生成一致性、前端建置與檢查、
Rich Menu 素材預覽，以及 Graphify／Repomix 無漏檔、無懸空關係或重複 ID。
本機驗收不代表已完成 Cloud Build、Cloud Run 或 LINE 實機驗證。

2026-11-10 本機結果：設定與 8 份 shell 語法檢查、API lint／生成一致性／相容性、
前端建置／TypeScript／12 項既有測試、Go format／lint／vet／race tests／build
均通過。Firestore 測試未另外啟動 emulator；套件通過不作為此次資料庫整合證據。
`govulncheck` 未找到實際呼叫到的已知弱點；依賴模組另有 3 項未被目前程式呼叫的弱點。
工具 doctor 與 Graphify／Repomix MCP handshake 通過。

Docker 映像建置、非 root Rich Menu 素材預覽、27 項本機 HTTP smoke checks 通過。
HTTP smoke 使用無真實權限的 MINI App 測試 ID，允許未啟用的出勤儲存層；不使用
token 或資料庫。`gcloud meta list-files-for-upload` 本機清單確認新素材與必要前端腳本
會納入，其他工具腳本與環境檔被排除；未執行上傳。

兩張 JPEG 內容搬移前後完全一致。維護中的來源已無舊腳本或素材路徑引用；
Graphify／Repomix 交叉稽核的來源完整性與圖譜結構檢查通過。

Node 路徑解析已透過 Context7 核對 Node.js 24 官方
[ES modules](https://nodejs.org/docs/latest-v24.x/api/esm.html) 與
[fileURLToPath](https://nodejs.org/docs/latest-v24.x/api/url.html)：
`new URL(..., import.meta.url)` 相對於模組，因此搬深一層後需同步調整相對路徑。
目前工具釘選仍由 `mise.toml` 管理。

Docker 的 `COPY --chmod` 已透過 Context7 核對
[官方 Dockerfile reference](https://docs.docker.com/reference/dockerfile/#copy---chmod)：
採用 `u=rwX,go=rX`，確保目錄可穿越而普通素材可讀，避免 `644` 套到目錄後
造成非 root runtime 無法讀取。符號權限需 Dockerfile frontend 1.14 以上；
本專案使用 `docker/dockerfile:1`。
