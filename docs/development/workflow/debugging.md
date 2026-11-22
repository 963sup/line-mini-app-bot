# 系統化除錯與根因分析

目前代理負責診斷、最小修正與驗收；依[實作能力](capabilities/implementer.md)及[測試能力](capabilities/tester.md)按需查閱，最後進行唯讀自我審查。

## Assignment and collaboration

記錄預期與實際行為、去識別錯誤、來源版本、執行環境、重現輸入、檔案範圍及驗收命令。區分唯讀診斷與授權修正；保留其他聊天的變更，跨檔案時先確認所有權。以重現和回歸證據判定完成，不以信心代替驗證。

## Diagnostic loop

**開發與維運指引**

```mermaid
flowchart TB
  BASE["記錄 revision、環境、預期與實際結果；以最小輸入重現"] --> REPRO{"原始觸發條件可重現？"}
  REPRO -->|否| PARTIAL["記錄頻率、條件與阻塞；僅交付暫定診斷或部分修正"]
  REPRO -->|是| SOURCE["Graphify 定位；以限定 Repomix 與當前來源核對"]
  SOURCE --> DOCS["依根指引查官方／版本相符文件，記錄來源與不確定性"]
  DOCS --> HYP["提出可反駁假設；每次改一個變因，記錄區辨檢查與結果"]
  HYP --> CAUSE{"證據確認因果鏈？"}
  CAUSE -->|否| HYP
  CAUSE -->|是| FIX["在授權範圍內完整修正根因；跨檔所有權先協調"]
  FIX --> CHECK["重跑原始重現與相關既有驗收檢查"]
  CHECK --> PASS{"原始錯誤消失，且相關檢查通過？"}
  PASS -->|否| NEXT{"檢查可繼續？"}
  NEXT -->|是| HYP
  NEXT -->|否| PARTIAL
  PASS -->|是| VERIFIED["交付已驗證修正；通過範圍以實際檢查為限"]
```

來源與限制：[根查證路由](../../../AGENTS.md#mandatory-documentation-verification) · [驗收矩陣](../tooling/testing/commands.md) · [證據重用](handoff.md#minimal-handoff-and-evidence-reuse)。本圖定義除錯流程，不代表已執行某次診斷。

- Context7 先 resolve ID 再查適用版本，與安裝依賴比對；Codex、LINE 與 Google 採根指引的特定官方來源。查詢不可用時明列限制。
- 區分觀察、推論、排除的假設與確認根因。AST／bundle 不證明執行結果；不收錄憑證或私人原始日誌。
- 不吞錯、不削弱驗證或 assertions、不關檢查、不加無界重試；無關重構不納入修正。環境或依賴修復須明確協調。
- 修正回歸時優先提供修正前失敗、修正後通過的證據。缺少憑證、重現或整合驗收時，明列缺口。

## Completion and handoff

使用[六欄交接](handoff.md#minimal-handoff-and-evidence-reuse)，保留症狀與影響、版本／環境、重現命令、已確認因果鏈與檔案／行號或去識別日誌、排除假設、修改必要性、修正後重現結果、回歸命令／exit status／skips／涵蓋限制、文件來源與版本。發布另依任務授權及主機邊界。

## Configuration and activation

目前代理在既有權限與授權範圍內採用實作能力，不啟動子代理。

[開發流程索引](README.md)
