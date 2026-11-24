# 依賴反轉（Dependency Inversion）

高階用例與低階技術透過由需求界定的抽象合作。

## 設計時要確認

以核心需要的能力定義小介面，由出站實作滿足；組裝處選擇具體實作。

## 示例

用例只需要按 ID 載入及帶版本保存邀請，介面無須暴露 SDK 查詢建構器。

## 延伸閱讀

[dependency-rule](dependency_rule.md) · [driven-port](driven_port.md) · [composition-root](composition_root.md) · [DDD 設計索引](design_guide.md)
