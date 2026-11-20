# Attendance 套件分工

管理打卡、出勤紀錄、補正審核、請假與出勤查詢。依賴由適配器及應用層指向領域層，外部實作由 `internal/bootstrap` 組裝。

## 現行結構

```text
attendance/
├── domain/
│   ├── entity/
│   ├── valueobject/
│   ├── service/
│   ├── event/
│   └── policy/
├── application/
│   ├── command/
│   ├── query/
│   ├── port/
│   └── dto/
└── adapter/
    ├── http/
    ├── line/
    ├── firestore/
    ├── messaging/
    └── scheduler/
```

## 閱讀入口

[領域模型](domain/README.md) 區分實體、業務值、運算、事件與政策；[應用層](application/README.md) 分開寫入、查詢、輸出介面與 DTO；[適配器](adapter/README.md) 處理協定及持久化。`messaging` 與 `scheduler` 是保留入口，目前沒有佇列或排程實作。

補正審核由 [command](application/command/review.go) 協調，[交易 port](application/port/review.go) 取得最新事實。Firestore 在同一交易內執行用例、核對版本及保存核准證據；交易重試時重新判斷。HTTP 與 LINE 共用 [月報 query](application/query/monthly.go)。

## 規格與驗證

[出勤規格](../../../../docs/domains/attendance/README.md) · [工作指引](AGENTS.md) · [驗收矩陣](../../../../docs/development/acceptance.md)。Go 檢查在根目錄透過 `bash scripts/tooling/tool-exec.sh mise run check` 執行；Firestore 整合測試另需 `FIRESTORE_EMULATOR_HOST`。
