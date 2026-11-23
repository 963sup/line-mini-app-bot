# Mermaid

## CLI 安裝與發布信任

Mermaid v13 的 `mermaid` 套件內建 `mermaid render`，與舊版 `@mermaid-js/mermaid-cli` 的 `mmdc` 分開。版本與安裝任務以 [mise.toml](../../../../mise.toml) 為準。`bootstrap.sh core` 安裝 Mermaid；`cloud` 只安裝應用程式工具，日常命令由 wrapper 按需準備已宣告的工具。單項安裝可執行 `mise run install-mermaid`、`install-shadcn`，均透過 `bash scripts/tooling/tool-exec.sh`。

2026-11-20 核對 npm `mermaid@13.0.0-alpha.3` 的 tarball SHA-512 與 registry integrity 一致，套件版本、CLI 入口及 scripts 符合官方 [release revision](https://github.com/mermaid-js/mermaid-v13/tree/beaa8010fc218d7db123ec4084ca987964acf9c2/packages/mermaid)。該 alpha 發布者為現有 maintainer `ashishjain0512`，有 registry 簽章但未提供 provenance；mise/aube 因較早版本有較強信任證明而拒絕。依 [mise npm 指引](https://mise.jdx.dev/dev-tools/backends/npm.html)，僅對這個確切版本設定 `trust_policy_excludes`，不關閉其他套件的信任檢查。未來換版須重新查證。

共用鎖檔與安裝失敗判定見 [mise 安裝限制](mise.md#鎖檔與安裝結果)。

## 專案適用範圍

| 圖表目的         | 必讀來源與邊界                                                                                                                     |
| ---------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| 儲存庫與專案關係 | [Project](../../../domains/project/project.md)：個人或組織擁有 Project，來源可為空或跨庫，來源讀權逐庫核驗                         |
| 物件與子工序     | [Object 與子物件](../../../domains/object/design.md)：單層子項，數量／單位／進度獨立                                               |
| 問題與討論       | [Issue](../../../domains/issue/design.md)、[Discussion](../../../domains/discussion/design.md)：各自狀態與資料，不自動推進物件進度 |
| 授權與寫入流程   | [操作模型](../../../contracts/business-operations.md)：區分主體、資源、權限、領域規則、交易與結果                                  |

以 `GroupID` 識別 LINE 群組，以 `SpaceID` 識別 Repository 內的工區。[庫內工區](../../../domains/repository/design.md#spaces)

圖表須註明來源版本、相關未提交變更及已實作／提案／待確認狀態，不以舊圖譜或規劃文件代替現行程式證據。

## 操作與驗收

1. 核對來源並依[文件圖表規範](../../../standards/documentation.md#獨立圖檔)在主題 Markdown 維護 Mermaid 區塊。
2. 使用專案鎖定的本機 Mermaid v13 CLI。需要安裝時執行 `bash scripts/tooling/tool-exec.sh mise run install-mermaid`。
3. 將區塊內容複製到暫存目錄，用 `bash scripts/tooling/tool-exec.sh mise run render-mermaid -- render /tmp/diagram.mmd --format svg > /tmp/diagram.svg` 驗證，核對圖中文字與箭頭符合來源。
