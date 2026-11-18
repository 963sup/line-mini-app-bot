// Isolated controls for Session lifecycle tests; production uses sidebar-07 NavUser.
import { Card } from "@/shared/ui/card";
import { Status } from "@/shared/Status";
import React from "react";
import { Button } from "@/shared/ui/button";
import { useSession } from "../src/identity/Session";
export function SessionControls() {
  const s = useSession();
  return (
    <Card
      className="mb-3 gap-4 rounded-2xl p-4 sm:p-6"
      aria-labelledby="identity-heading"
    >
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-xl font-semibold" id="identity-heading">
          身分
        </h2>
        <div className="flex flex-wrap justify-end gap-2">
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
          <Button
            id="logout"
            hidden={!s.canLogout}
            disabled={s.busy}
            onClick={s.logout}
          >
            登出
          </Button>
          {!s.ready && !s.busy && (
            <Button
              onClick={s.verify}
              title="重試初始化"
              aria-label="重試初始化"
            >
              重試
            </Button>
          )}
        </div>
      </div>
      <Status id="mini-status" role="status" aria-live="polite">
        {s.status}
      </Status>
      <dl
        className="grid gap-2 text-sm [&>div]:grid [&>div]:grid-cols-[5.5em_minmax(0,1fr)] [&>div]:gap-2 [&_dd]:break-all [&_dd]:text-muted-foreground"
        id="identity"
        hidden={!s.user}
      >
        <div>
          <dt>LINE 名稱</dt>
          <dd id="display-name">{s.user?.displayName || "未提供名稱"}</dd>
        </div>
        <div>
          <dt>本人 ID</dt>
          <dd id="user-id">{s.user?.userId}</dd>
        </div>
      </dl>
    </Card>
  );
}
