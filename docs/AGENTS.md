# docs 工作指引

## 用途與範圍

集中維護架構、領域規格、開發流程、操作指南與提案。

## 架構與責任邊界

實作現況由程式與契約驗證，文件按主題維護。

## 閱讀與修改入口

先查閱[architecture](architecture)、[domains](domains)、[development](development)、[guides](guides)、[proposals](proposals)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[README.md](README.md#目錄責任) · [文件索引](README.md)


適用於本目錄及其子目錄；先遵循 [根目錄 AGENTS.md](../AGENTS.md)，再套用較近的子目錄補充。使用者及主機指令優先；本檔不增加外部寫入授權。

## 職責與邊界

文件分類與各主題入口見 [README.md](README.md#目錄責任)。歷史變更透過 Git 查閱。

以已接受 ADR 與實際來源為準，區分已實作、待驗收與未來規劃。不在多處複製版本、連線設定、機密或完整政策。

文件維持主題分類；docs/README.md 提供 Backend、API、Frontend、Scripts、Deploy 閱讀入口。

## 驗證

驗證本機連結與來源路徑；工具圖譜不證明語意正確。純文件變更執行 git diff --check，相關設定變更再執行設定檢查。

命令除另有標示外均在儲存庫根目錄執行。閱讀 [文件索引](README.md)；回報實際執行結果與限制。

## 主題子目錄

各主題目錄 README 列出完整文件入口，AGENTS.md 補充歸屬與維護規則；入口見 [文件歸屬](README.md#文件歸屬與維護)。新增或搬移單篇文件時，更新所屬分類索引及受影響的引用。分類入口變動時，才更新引用該入口的上層索引。

業務操作模型維持於 [單一來源](architecture/contracts/business-operations.md)，功能文件保存具體操作契約；引用模型不代表既有功能已通過驗收。

[現行實作圖索引](diagrams.md#現行實作) · [未來設計圖](diagrams.md#未來設計)。
