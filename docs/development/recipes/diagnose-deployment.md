# 診斷 Cloud Build 與 Cloud Run

從失敗通知、build ID、revision 或已知提交開始。先核對目標專案、區域、服務和實際 trigger。

依[失敗階段對照表](../../operations/troubleshooting.md#cloud-buildcloud-run-與部署後同步)選擇要查的建置、部署、執行或同步證據。

1. 依[故障排查](../../operations/troubleshooting.md)取得原始錯誤，找出第一個失敗點。從[部署設定說明](../../operations/README.md)分辨實際 trigger 設定、Dockerfile 與手動 YAML。
2. 對照同一提交的程式、映像和入口。若問題出在步驟環境，以同一映像或同一執行身分的證據支持原因。
3. 涉及修復時，執行對應的[設定或容器驗收](../tooling/testing/commands.md)，正式環境結果另按[發布驗收](../../operations/release-acceptance.md)記錄。

完成診斷時，交付失敗步驟與原始證據、受影響的提交、目前承接流量的 revision、服務影響與下一個可執行動作。缺少雲端讀取權限時，留下具體缺項。[Cloud Run 證據要求](../verification/evidence.md#cloud-run-and-browser-location)

[操作配方](README.md)
