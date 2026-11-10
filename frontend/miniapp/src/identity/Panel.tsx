import React from "react";
import { useSession } from "./Session";
export function IdentityPanel() {
  const s = useSession();
  return (
    <section className="panel" aria-labelledby="identity-heading">
      <h2 id="identity-heading">確認身分</h2>
      <p id="mini-status" role="status" aria-live="polite">
        {s.status}
      </p>
      <div className="actions">
        <button
          id="login"
          hidden={!s.ready || !!s.user || s.inClient || s.busy}
          onClick={s.login}
        >
          使用 LINE 登入
        </button>
        <button
          id="verify"
          hidden={!s.ready && !s.busy}
          disabled={s.busy}
          onClick={s.verify}
        >
          重新驗證身分
        </button>
        <button id="logout" hidden={!s.user} onClick={s.logout}>
          登出此頁
        </button>
        {!s.ready && !s.busy && <button onClick={s.verify}>重試初始化</button>}
      </div>
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
