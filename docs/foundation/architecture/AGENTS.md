# architecture 工作指引

適用於本目錄及其子目錄。先遵循 [父層指引](../../AGENTS.md) 與 [根指引](../../../AGENTS.md)，較近的子目錄指引再補充；使用者與主機指令優先。

## 職責與邊界

跨模組邊界、依賴方向、HTTP／資料契約與架構決策。
局部命令放模組 README；功能規則放 features，部署操作放 operations。已接受決策以 decisions 中的 ADR 為準，未來規劃需明確標示。

## 維護與驗證

先讀 [本目錄索引](README.md) 與 [文件總索引](../../README.md)。索引更新依[文件維護指引](../../AGENTS.md#主題子目錄)。
純文件修改驗證本機路徑、標題錨點及 `git diff --check`；設定變更另依父層規定檢查。不把未執行的應用測試或外部驗收宣稱為通過。

[現行實作圖索引](../../diagrams.md#現行實作) · [未來設計圖](../../diagrams.md#未來設計)。

## 階段範圍

依[基線決策](decisions/0002-baseline.md)與[基礎開發規格](../../development/foundation.md)判斷階段範圍：先基礎建設，再 LINE／Firestore，接著基本出勤。平台擴充文件屬延後參考，不將 Worker、Garr、定位或新測試基礎設施設為此階段的必要條件。
