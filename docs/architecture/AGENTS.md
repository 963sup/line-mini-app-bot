# 共用系統基礎文件指引

## 用途與範圍

保存系統架構、契約、整合與執行邊界的說明。

## 架構與責任邊界

跨模組設計放在此處，單一領域的規則由領域文件維護。

## 閱讀與修改入口

先查閱、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[文件指引](../AGENTS.md) · [目錄索引](README.md)


遵循[文件指引](../AGENTS.md)與[根指引](../../AGENTS.md)。依[目錄索引](README.md)選讀主題。純文件修改驗證本機路徑、標題錨點及 `git diff --check`；設定變更另依父層規定檢查。

## architecture

適用於 runtime、identity、storage 與 contracts。

跨模組邊界、依賴方向、HTTP／資料契約與架構決策。
局部命令放模組 README；功能規則放 domains，部署操作放 guides/release。已接受決策以 decisions 中的 ADR 為準，未來規劃需明確標示。

依[基線決策](../decisions/0002-baseline.md)與[基礎開發規格](../guides/local-development/environment-verification.md)判斷階段範圍：先基礎建設，再 LINE／Firestore，接著基本出勤。平台擴充文件屬延後參考，不將 Worker、Garr、定位或新測試基礎設施設為此階段的必要條件。

## integrations

適用於 integrations 子目錄。

本目錄維護 LINE 與瀏覽器定位整合。Firestore 的儲存規則見 [Firestore 整合](storage/firestore.md)。
外部服務契約與專案功能規則分開；遵循根層查證政策，不在文件存放 token、憑證或機密。

## operations

發布與驗收依[操作指引](../guides/AGENTS.md)。
