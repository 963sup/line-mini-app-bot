# 共用系統基礎文件指引

沿用[文件指引與驗證方式](../AGENTS.md)，依[目錄索引](README.md)選取跨模組契約、整合或架構決策。

## architecture

適用於 runtime、identity 與 storage；共用契約見 [docs/contracts](../contracts/README.md)。

跨模組邊界、依賴方向、HTTP／資料契約與架構決策。
局部命令放模組 README；功能規則放 domains，部署操作放 operations。已接受決策以 decisions 中的 ADR 為準，未來規劃需明確標示。

依[基線決策](../decisions/0002-baseline.md)與[基礎開發規格](../development/environment/verification.md)判斷階段範圍：先基礎建設，再 LINE／Firestore，接著基本出勤。平台擴充文件屬延後參考，不將 Worker、Garr、定位或新測試基礎設施設為此階段的必要條件。

## integrations

適用於 integrations 子目錄。

LINE、瀏覽器定位、Firestore 與 Cloud Storage 由[外部整合索引](../integrations/README.md)導覽。Firestore 的儲存規則見 [Firestore 整合](../integrations/firestore/connection.md)。
外部服務契約與專案功能規則分開；遵循根層查證政策，不在文件存放 token、憑證或機密。

## operations

發布與驗收依[維運指引](../operations/AGENTS.md)。
