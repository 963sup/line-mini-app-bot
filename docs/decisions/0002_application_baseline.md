# 0002：應用架構基線

- 狀態：Accepted
- 取代：早期要求常駐 worker 的假設

## 背景

專案需要同時支援 LINE Webhook、MINI App、Firestore 與多個業務能力，但尚無獨立部署或分散式系統的必要。

## 決策

- 使用 Go 模組化單體，一個容器部署至 Cloud Run。
- 採輕量 Hexagonal；Domain 不依賴 LINE、HTTP 或 Firestore SDK。
- 不同入口共用應用用例，依賴在 composition root 手動組裝。
- Firestore 是主要持久化方向；後端執行授權。
- 不預設微服務、完整 CQRS、Event Sourcing 或通用 Repository framework。

## 後果

模組可在同一程序內低成本合作，但仍必須維持 bounded context 與明確 port。只有部署、擴展、資料或可靠性需求證明單體不足時才新增 ADR 調整。
