# 領域服務（Domain Service）

Domain Service 表達有業務意義、但不自然屬於單一 Entity 或 Value Object 的運算或決策。它以通用語言命名，不以「所有規則都放 Service」取代模型行為。

## 與其他責任的差異

Entity 擁有自身合法轉換；Factory 負責建立；Application 協調讀取、授權、交易與外部行為。Domain Service 負責領域決策，不處理 HTTP 格式、SDK 錯誤或依賴組裝。

優先讓 Domain Service 取得已準備好的模型與事實，回傳決策或值，不直接連接外部技術。若決策需要新的事實，由 Application 透過 Port 取得，並說明其時效與一致性；不是把 I/O 隱藏進純函數。

## 設計例子

未來有效成員資格計算可能需要目前關係與上層 generation。若這項判斷跨多個模型而不屬於單一物件，可用 Domain Service；先由 Application 載入事實，服務再判斷有效性。邀請自身的撤銷規則仍留在 Invitation。

## 設計產出與核對

說明為何該行為不適合既有模型、輸入事實、輸出決策與規則擁有者。名稱若只有 Manager、Helper 或 Validator，應再確認是否缺少清楚的業務語意。

概念參考：[Microsoft Learn：戰術 DDD](https://learn.microsoft.com/en-us/azure/architecture/microservices/model/tactical-domain-driven-design)。僅採概念，不採其微服務部署前提。

[Entity](entity.md) · [Application Service](application-service.md) · [DDD 設計索引](design-guide.md)
