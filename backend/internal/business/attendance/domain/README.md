# Attendance 領域層

領域層不引用 application、adapter、HTTP 或 Firestore。

## 目錄與責任

- [entity](entity/README.md)：業務物件、識別、狀態與不變條件。
- [valueobject](valueobject/README.md)：沒有識別的業務值與驗證。
- [service](service/README.md)：跨物件計算、補正與證據核對。
- [event](event/README.md)：已記錄的打卡事實。
- [policy](policy/README.md)：審核資格、重送與請假規則。

## 相關入口

[模組導覽](../README.md) · [工作指引](../AGENTS.md)
