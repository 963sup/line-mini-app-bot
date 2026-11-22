# LINE MINI App 身分交換

MINI App 初始化 LIFF 後，將原始 ID token 放在 Authorization Bearer，透過 POST /api/v1/session 交換 Session，JSON body 必須是空物件。Go 向 LINE 驗證 issuer、audience、subject、exp，不接受前端自行提供的 user ID 或 decoded profile。成功回 SessionResponse（user、csrfToken、expiresAt）與 Cookie。

Messaging API Webhook 使用 LINE 簽章；MINI App 登入使用環境所選 Channel 的 token。公開設定、靜態資產及登入頁不授予業務權限。

瀏覽器 Cookie、CSRF、登出、有效期及多頁籤行為以[應用 Session 契約](../../contracts/identity.md)為準。設定與真機驗收見 [LINE 接入指南](mini-app-setup.md)。
