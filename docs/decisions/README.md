# 架構決策

ADR 記錄跨模組、難以逆轉或會長期影響維護成本的決策。現行規範與操作步驟仍由對應專題文件維護。

| ADR                                  | 狀態     | 決策                                         |
| ------------------------------------ | -------- | -------------------------------------------- |
| [0001](0001_example.md)              | Template | 新決策範本                                   |
| [0002](0002_application_baseline.md) | Accepted | 模組化單體與輕量 Hexagonal                   |
| [0003](0003_cloud_run_primary.md)    | Accepted | Cloud Run 是正式應用執行環境                 |
| [0004](0004_monorepo.md)             | Accepted | Backend、Frontend 與 API 契約分離的 monorepo |

被新 ADR 取代的決策不刪除，改標示 Superseded 並連到接替文件。
