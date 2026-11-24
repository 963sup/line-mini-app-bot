# 依賴規則（Dependency Rule）

程式碼依賴由外部機制指向核心穩定契約。

## 設計時要確認

檢查 import 圖與公開型別，禁止 SDK 型別透入核心；執行時可經注入的介面呼叫外部實作。

## 示例

核心呼叫保存 Port 時，資料庫 Adapter 在編譯時仍依賴核心契約。

## 延伸閱讀

[dependency-inversion](dependency_inversion.md) · [architecture-testing](architecture_testing.md) · [DDD 設計索引](design_guide.md)
