# Troubleshooting

1. Run `python3 scripts/tooling/check-config.py` for syntax and document references.
2. Run Go doctor for compiler, race toolchain and writable paths.
3. Diagnose one MCP with `python3 scripts/tooling/mcp-doctor.py --server <name>` through
   tool-exec.sh. Missing optional credentials need not block another service.
4. For Codex configuration problems use [OpenAI Docs MCP](../tooling/mcp/openai-docs.md)
   search/fetch. Use [Context7](../tooling/mcp/context7.md) for version-specific APIs.
5. Repair repository-owned settings within task scope, then rerun the affected check.

Separate configured, installed, transport-ready, host-loaded and functional states.
Inspect host config/trust diagnostics when direct transport works but chat tools
are missing. Do not rewrite HOME/CODEX_HOME or platform credentials to force it.
Never print authentication headers or secret values while debugging.


## Cloud Build、Cloud Run 與部署後同步

先記錄 project、region、build ID、來源 commit 與查詢時間，再依下表定位。
部署前後 metadata 的取得方式見[發布證據](release-acceptance.md#發布證據與前置條件)。
只保存必要 metadata 與匿名化錯誤，不輸出 Secret payload 或登入 token。

| 階段／症狀 | 核對證據與處理 |
| --- | --- |
| Build 失敗，尚未產生新版映像 | 查看第一個失敗步驟及來源 SHA；若出現 `--mount option requires BuildKit`，核對實際 inline Build env 與 [BuildKit 契約](cloud-run.md#docker-buildkit-建置契約)。 |
| Deploy 失敗或 revision 未 Ready | 核對該 revision 的啟動日誌、映像 digest、PORT、runtime service account 與 Secret references；與前一版 metadata 比較，勿用本機 YAML 推定實際設定。 |
| Deploy 成功、SyncRichMenus 失敗 | 分別確認 Deploy 結果與目前流量；整體 Build 失敗不代表前面已回滾。核對實際 Trigger 的 `-images /assets/line/richmenu`，再核對同一映像中的目錄可穿越及 JPEG 可由 nonroot 讀取。詳見[圖片路徑與權限紀錄](cloud-run.md#managed-trigger-的-rich-menu-圖片路徑)。 |
| Build 成功，但登入／出勤／摘要失敗 | 依請求對應 revision 查執行期錯誤，再核對 Session、Firestore、Cloud Tasks、模型與 LINE 訊息流程；Ready 不代表端到端驗收完成。 |

managed trigger 使用 inline build；修改 repository YAML 不會修正該 Trigger。
修正後以新 Build 的來源 SHA、每個步驟結果、Ready revision、traffic 及真實功能結果驗收。
Rich Menu 同步會修改 LINE 外部狀態，重跑前先確認既有選單與目標帳號；
Cloud Run 回滾不會撤銷已同步的 LINE 選單。

## 歷史快照與查證界線

文件中的 revision、Secret 綁定與 IAM 紀錄只代表其標示日期；不能當作即時現況。
本次文件整理未查詢雲端服務，未執行部署、同步或清理。
Google skills 與 Developer Knowledge MCP 未在本側聊提供可用入口，改查
Google 官方 [Cloud Run 回滾](https://docs.cloud.google.com/run/docs/rollouts-rollbacks-traffic-migration)
及 [Artifact Registry 映像管理](https://docs.cloud.google.com/artifact-registry/docs/docker/manage-images)。
本頁沒有新增 gcloud 指令；執行命令前仍須依安裝版本的 leaf help 核對。
