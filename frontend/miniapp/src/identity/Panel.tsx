import React from "react";
import { Button } from "@/shared/ui/button";
import { useSession } from "./Session";
export function IdentityPanel() {
  const s = useSession();
  return (
    <section className="panel identity-panel" aria-labelledby="identity-heading">
      <div className="panel-heading">
        <h2 id="identity-heading">身分</h2>
        <div className="actions identity-actions">
          <Button
            id="login"
            hidden={!s.ready || !!s.user || s.inClient || s.busy}
            onClick={s.login}
            title="使用 LINE 登入"
            aria-label="使用 LINE 登入"
          >
            登入
          </Button>
          <Button
            id="verify"
            hidden={!s.ready && !s.busy}
            disabled={s.busy}
            onClick={s.verify}
            title="重新驗證身分"
            aria-label="重新驗證身分"
          >
            驗證
          </Button>
          <Button id="logout" hidden={!s.canLogout} disabled={s.busy} onClick={s.logout}>
            登出
          </Button>
          {!s.ready && !s.busy && <Button onClick={s.verify} title="重試初始化" aria-label="重試初始化">重試</Button>}
        </div>
      </div>
      <p id="mini-status" role="status" aria-live="polite">
        {s.status}
      </p>
      <dl id="identity" hidden={!s.user}>
        <div>
          <dt>LINE 名稱</dt>
          <dd id="display-name">{s.user?.displayName || "未提供名稱"}</dd>
        </div>
        <div>
          <dt>本人 ID</dt>
          <dd id="user-id">{s.user?.userId}</dd>
        </div>
      </dl>
    </section>
  );
}
