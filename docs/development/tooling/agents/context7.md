# Context7 文件查證

查套件、語言、框架、SDK 或工具行為時，先對照 `go.mod`、套件 manifest、鎖檔或 `mise.toml` 的版本，再選目前聊天實際曝光的工具與 schema。沿用版本和範圍仍適用的既有證據；查詢時不傳機密或私人原始碼。

先呼叫 `resolve-library-id`，再以精確 library ID 和聚焦的版本問題呼叫 `query-docs`。記錄來源 URL、文件版本與適用限制；找不到已安裝版本時明列落差，不把其他版本當成已驗證。文件結論還須回到目前來源及實際行為核對。

## Anonymous fallback

優先使用主機 plugin。它因配額失敗或未提供時，若主機已曝光專案的匿名 Context7 服務，先在該連線重新解析 library ID，再重試一次；再次限流即停止並記錄阻塞。兩個連線彼此獨立，專案服務不使用 API key 或憑證標頭。工具未曝光時依[MCP 診斷](../../environment/troubleshooting.md#設定與-mcp-診斷)處理，不把配額失敗當成設定缺失。

官方來源路由依[根指引](../../../../AGENTS.md#mandatory-documentation-verification)。Context7 不定義產品需求；查詢成功也不代表主機載入或應用行為已驗證。缺少必要證據時，記錄阻塞並保留受影響決策待確認。
