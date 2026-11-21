# 值物件（Value Object）

Value Object 以內容判等，沒有需要獨立追蹤的生命週期。領域上相等的兩個值可以互換；變更時以新值取代舊值。

## 適用判斷與責任

以需要一起解釋的值封裝語意，例如時間區間必須同時包含起訖及其合法關係。不要只因幾個欄位常一起出現就包成沒有規則的型別。

建立入口驗證值本身的合法性；是否可以用在目前操作中，由模型或用例依上下文判斷。不可變性包含避免暴露可直接修改的 slice、map 或指標；語言的 struct 宣告本身不保證不可變。

內容判等前須定義單位、時區、大小寫及正規化規則，不能任意清洗識別碼。識別碼也可用值物件表達，但被它識別的對象仍可能是 Entity。

## 設計例子

未來的 ResourceRef 可由渠道、資源種類與 ID 組成。三者一致才代表同一資源引用；合法 ResourceRef 不代表操作者有權存取資源。

## 設計產出與核對

列出欄位、相等性、正規化、非法值、不可變性及建立失敗結果。核對替換值不會偷偷改變其他持有者的狀態。

概念參考：[Microsoft Learn：戰術 DDD](https://learn.microsoft.com/en-us/azure/architecture/microservices/model/tactical-domain-driven-design)。僅採概念，不採其微服務部署前提。

[Entity](entity.md) · [Invariant](invariant.md) · [DDD 設計索引](README.md)
