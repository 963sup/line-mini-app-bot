import React, { useEffect, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { useLatestRequest, errorMessage } from "@/shared/http/lifecycle";
import { Input } from "@/shared/ui/input";
import { Button } from "@/shared/ui/button";

export function Enterprise({ id, canCreate, eligibilityError, onSaved }: {
  id: string;
  canCreate?: boolean;
  eligibilityError?: string;
  onSaved: (enterpriseId: string) => void;
}) {
  const { request, user } = useSession(),
    { run, cancel } = useLatestRequest();
  const [view, setView] = useState<components["schemas"]["Enterprise"]>(),
    [busy, setBusy] = useState(false),
    [message, setMessage] = useState("");
  const create = id === "new";
  const eligibilityMessage = create && (canCreate === undefined
    ? eligibilityError
    : !canCreate && "帳號尚無可用的企業開通資格。");
  useEffect(() => {
    if (!user || create) return;
    void run(
      (signal) =>
        request(
          (api, s) =>
            api.GET("/api/v1/enterprises/{enterpriseId}", {
              params: { path: { enterpriseId: id } },
              signal: s,
            }),
          { signal },
        ),
      setView,
      (e) => setMessage(errorMessage(e)),
    );
    return cancel;
  }, [id, create, user, request, run, cancel]);
  function form(
    kind: "create" | "owner" | "organization",
    label: string,
    action: string,
  ) {
    return (
      <form
        onSubmit={(event) => {
          event.preventDefault();
          if (busy) return;
          const value = String(
            new FormData(event.currentTarget).get("value") ?? "",
          ).trim();
          if (!value) return;
          setBusy(true);
          setMessage("");
          const params = { path: { enterpriseId: id } };
          void run(
            (signal) =>
              request(
                (api, s) =>
                  kind === "create"
                    ? api.POST("/api/v1/enterprises", {
                        body: { name: value },
                        signal: s,
                      })
                    : kind === "owner"
                      ? api.POST("/api/v1/enterprises/{enterpriseId}/owners", {
                          params,
                          body: { userId: value },
                          signal: s,
                        })
                      : api.POST(
                          "/api/v1/enterprises/{enterpriseId}/organizations",
                          {
                            params,
                            body: { organizationId: value },
                            signal: s,
                          },
                        ),
                { signal },
              ),
            (v) => onSaved(v.id),
            (e) => {
              setBusy(false);
              setMessage(errorMessage(e));
            },
          );
        }}
        className="flex flex-wrap items-end gap-2"
      >
        <label>
          {label}
          <Input name="value" required disabled={busy} />
        </label>
        <Button disabled={busy}>{action}</Button>
      </form>
    );
  }
  if (!user) return null;
  return (
    <section className="grid gap-4">
      {(message || eligibilityMessage) && <p role="status">{message || eligibilityMessage}</p>}
      {create && canCreate && form("create", "企業名稱", "建立企業")}
      {view && (
        <>
          <h2>{view.name}</h2>
          <h3>組織</h3>
          <ul>
            {view.organizations.map((o) => (
              <li key={o.id}>{o.name}</li>
            ))}
          </ul>
          <h3>企業團隊</h3>
          <ul>
            {view.teams.map((t) => (
              <li key={t.id}>{t.name || t.id}</li>
            ))}
          </ul>
          {view.canManage && (
            <>
              {form("organization", "組織 ID", "加入組織")}
              {form("owner", "LINE 使用者 ID", "新增企業 Owner")}
            </>
          )}
        </>
      )}
    </section>
  );
}
