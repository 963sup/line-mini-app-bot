# 開發與發布操作

## 用途與範圍

保存設定、部署與營運操作指引。

## 架構與責任邊界

操作步驟按場景分工，產品規則由領域規格維護。

## 閱讀與修改入口

先查閱[line-setup](line-setup)、[release](release)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[工具設定](local-development/setup.md) · [環境驗證](local-development/environment-verification.md)


| 要完成的工作 | 入口 |
| --- | --- |
| 建立開發環境 | [工具設定](local-development/setup.md)、[環境驗證](local-development/environment-verification.md) |
| 確認 Firestore 可讀 | [唯讀檢查](local-development/firestore-read-check.md) |
| 接入 LINE | [啟動](line-setup/README.md)、[Webhook](line-setup/webhook.md)、[Session](line-setup/session.md)、[Rich Menu](line-setup/rich-menu.md) |
| 部署與回復 | [Cloud Run](release/cloud-run.md)、[發布驗收](release/release-acceptance.md)、[故障排查](release/troubleshooting.md) |
