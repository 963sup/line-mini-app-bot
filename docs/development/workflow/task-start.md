# 開始任務與能力選用

目前代理直接完成任務，不啟動子代理。

1. 確認交付結果、操作範圍與驗收方式。「繼續」時接續未完成項目。
2. 核對分支與既有變更，依[任務選擇表](../../../README.md#依任務選擇能力)選讀所需指引。
3. 修復問題時，用原始案例定位偏差、修改來源並重驗；詳見[除錯流程](debugging.md#diagnostic-loop)。

<a id="delegation-depth-and-role-selection"></a>

## 單一代理執行

`.codex/config.toml` 不覆寫 `multi_agent` 功能開關，也不註冊子代理；不啟動子代理的工作規則由根 AGENTS.md 維護。專業能力保存為普通 Markdown，按需讀取；不依賴角色選擇或額外模型。MCP 與 Skills 保持各自設定，工具存在不代表已由當前聊天載入。

錯誤與測試失敗依[除錯流程](debugging.md)處理。工具按任務使用，不強制每次完整執行 Graphify、Repomix 或所有 MCP。

## Assignment contract

同一聊天也需記錄目標、檔案範圍、依賴、必須保留的行為及驗收條件。跨聊天交接使用[六欄交接](handoff.md)，不增加發布或部署權限。

[開發流程索引](./)
