import React, { useCallback, useEffect, useRef, useState } from "react";
import { useRouter } from "next/router";
import type { components } from "@repo/api-client";
import { OrganizationTeams } from "@/business/team/OrganizationTeams";
import { useSession } from "@/identity/Session";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";
import { useOrganizationSelection } from "./OrganizationSelection";

const memberLabel = (id: string) => `成員 ${id.slice(0, 12)}`;

export function OrganizationPage({ organizationId }: { organizationId: string }) {
  const router = useRouter();
  const { epoch, user, request } = useSession();
  const { refresh } = useOrganizationSelection();
  const { run, cancel } = useLatestRequest();
  const [snapshot, setSnapshot] = useState<{ epoch: number; roster: components["schemas"]["OrganizationMembers"] }>();
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const keys = useRef(new Map<string, string>());
  const roster = user && snapshot?.epoch === epoch ? snapshot.roster : undefined;

  const load = useCallback((clearOnFailure: boolean) => run(
    signal => request((api, s) => api.GET("/api/v1/organizations/{organizationId}/members", {
      signal: s, params: { path: { organizationId } },
    }), { signal }),
    roster => setSnapshot({ epoch, roster }),
    cause => {
      if (clearOnFailure) setSnapshot(undefined);
      setError(errorMessage(cause));
    },
  ), [epoch, organizationId, request, run]);

  useEffect(() => {
    keys.current.clear();
    setBusy(false);
    setError("");
    if (user) void load(true);
    return cancel;
  }, [cancel, load, user]);

  function depart(memberId?: string) {
    if (busy || !roster) return;
    const label = memberId ? memberLabel(memberId) : roster.organizationName;
    if (!window.confirm(memberId ? `要移除「${label}」嗎？\n請先與對方核對成員代碼。` : `要退出「${label}」嗎？`)) return;
    const target = memberId ?? "";
    const key = keys.current.get(target) ?? crypto.randomUUID();
    keys.current.set(target, key);
    setBusy(true);
    setError("");
    void run(
      signal => request((api, s) => memberId
        ? api.DELETE("/api/v1/organizations/{organizationId}/members/{memberId}", {
            signal: s, params: { path: { organizationId, memberId }, header: { "Idempotency-Key": key } },
          })
        : api.DELETE("/api/v1/organizations/{organizationId}/membership", {
            signal: s, params: { path: { organizationId }, header: { "Idempotency-Key": key } },
          }), { signal }),
      () => {
        keys.current.delete(target);
        refresh();
        if (memberId) void load(true).finally(() => setBusy(false));
        else {
          setBusy(false);
          setSnapshot(undefined);
          void router.replace("/miniapp");
        }
      },
      cause => {
        setError(errorMessage(cause));
        void load(true).finally(() => setBusy(false));
      },
    );
  }

  return <>
    <section>
      <h2>概況</h2>
      {error && <p role="alert">{error}</p>}
      {roster ? <>
        <p>{roster.organizationName} · {roster.canRemove ? "擁有者" : "成員"}</p>
        {roster.canLeave
          ? <button type="button" disabled={busy} onClick={() => depart()}>退出組織</button>
          : <p>請先移轉所有權，再退出組織。</p>}
      </> : !error && user && <p role="status">載入組織中…</p>}
    </section>
    <section>
      <h2>成員</h2>
      {roster && <>
        <ul>{roster.members.map(member => <li key={member.memberId}>
          {memberLabel(member.memberId)}{member.self ? "（你）" : ""}{member.owner ? " · 擁有者" : ""}
          {roster.canRemove && !member.owner && !member.self &&
            <button type="button" disabled={busy} onClick={() => depart(member.memberId)}>移除</button>}
        </li>)}</ul>
        <p>成員可分享自己的代碼供核對。</p>
        {roster.canRemove && <p>邀請新成員請使用 LINE 官方帳號。</p>}
      </>}
    </section>
    <OrganizationTeams organizationId={organizationId} />
    <section><h2>儲存庫</h2><p>尚未支援組織儲存庫。</p></section>
  </>;
}
