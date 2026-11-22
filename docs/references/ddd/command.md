# 命令（Command）

應用層接收的變更意圖，可能被接受或拒絕。

## 設計時要確認

具名輸入與拒絕語意；重送需要設計冪等範圍與結果。

## 示例

AcceptInvitation 指定目標邀請，InvitationAccepted 表示已發生的事實。

## 延伸閱讀

[domain-command](domain-command.md) · [query](query.md) · [idempotency](idempotency.md) · [DDD 設計索引](README.md)
