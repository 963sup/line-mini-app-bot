# reCAPTCHA 簡要規格

**狀態：公開操作出現濫用風險時導入。** 本頁規定本專案的使用邊界；雲端是否已啟用、費率、配額及實際部署另行驗證。

## 用途與最小流程

保護公開表單或高風險操作，提供自動化濫用的風險訊號。

前端為指定 action 取得 token → 隨操作送交 Go → 後端建立 assessment → 核對有效性、預期 action、網站來源及風險結果 → 依政策執行或拒絕。

## 資料、授權與失敗邊界

reCAPTCHA 不取代 LINE 身分、Session、CSRF、資源授權與速率限制；瀏覽器不能自行提交可信分數。LINE Webhook 及內部 worker 不要求瀏覽器挑戰。

## 導入條件

選定 key 類型、受保護 action、網域、門檻及無法評估時的策略；不預設全站挑戰。官方現將 reCAPTCHA 納入 Google Cloud Fraud Defense，採用時核對對應方案及 assessment 契約。

## 驗收

無效／過期／重用 token、action 或來源不符、低分、評估服務不可用及 LINE WebView／無障礙使用情境有預期結果；token 不寫日誌。

## 來源與相關規格

[官方文件](https://docs.cloud.google.com/recaptcha/docs/create-assessment-website)，查證範圍見[整合總覽](README.md#sources)。[身分與 Session](../../identity/browser-session.md) · [領域授權](../../../references/ddd/domain-authorization.md)。

[Google 服務索引](README.md)
