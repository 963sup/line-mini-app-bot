# 契約來源

適用於本目錄及其子目錄；繼承 [上層指引](../AGENTS.md)，並遵循其引用的根目錄規則。此處只補充本範圍責任；使用者與主機指令優先，不增加外部寫入授權。

## 職責與邊界

- 公開 HTTP 由 `openapi.yaml` 組合，內部 worker HTTP 由 `internal-openapi.yaml` 組合；入口只維護 metadata 與 `$ref`，業務操作與 DTO 歸所屬能力。
- 依 [能力索引](README.md) 維護唯一歸屬；能力檔不得反向引用組合入口，跨能力引用遵循擁有者方向，不形成循環。
- 共通傳輸定義放 common；業務模型不因多處使用就移入 common。資料夾使用能力名稱，實體與 REST 資源命名遵循上層命名規範。
- Issue 的 types／comments、Discussion categories 與 Project fields／views 保留 scaffold，空 paths／schemas 不接入組合入口。Project／items 與 Repository milestones／statistics 已依 [基礎追蹤規格](../../docs/features/planning.md) 接入；其餘能力先完成操作規格再新增契約及實作。
- 新增操作同步維護 [能力清單](../capabilities.json)、對應 adapter、消費端及功能規格；事件入口不為目錄一致而新增 HTTP API。
- 契約變更由儲存庫根目錄執行 `bash scripts/tooling/tool-exec.sh pnpm api:check`；純指引變更檢查連結與 whitespace，不刷新相容性 baseline。
