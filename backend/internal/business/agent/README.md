# agent

[領域文件](../../../../docs/domains/agent/README.md) · [全域索引](../../../../docs/domains/README.md#domain-matrix)

依任務目標與目前結果，從允許的操作中選擇下一步。

遇到缺少來源頁碼的候選項目時，可選擇重讀授權範圍內的頁面，或回報需要人工確認。

[補查與終止流程圖](../../../../docs/proposals/agent.md#design-diagram) · [維護指引](AGENTS.md)

[上層導覽](../README.md)

目前進度：已實作唯讀證據補查。見[Go 骨架](application/port/investigate.go)。

## 模型接線

呼叫端以 [NewInvestigator](../../bootstrap/agent.go) 注入會核驗身分、資源與來源版本的 PageReader，再呼叫 Investigate。設定 AGENT_PROVIDER 為 openai 或 gemini、AGENT_MODEL 為該供應商模型，並使用對應 API key。每一步最多讀取一個允許的頁面；只有原文中存在的引文可作為候選證據。業務欄位與單位仍由 呼叫端 驗證。
