import { useCallback, useEffect, useRef, useState, type ReactNode } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";

import { PagePanel } from "@/shared/PagePanel";
import { FormField } from "@/shared/FormField";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { Alert, AlertDescription } from "@/shared/ui/alert";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Item, ItemActions, ItemContent, ItemGroup, ItemTitle } from "@/shared/ui/item";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";

const memberLabel = (id: string) => `成員 ${id.slice(0, 12)}`;

export function OrganizationPage({ organizationId, children, onChanged, onLeave }: {
  organizationId: string;
  children: ReactNode;
  onChanged: () => void;
  onLeave: () => void;
}) {
  const { epoch, user, request } = useSession();
  const { run, cancel } = useLatestRequest();
  const [snapshot, setSnapshot] = useState<{ epoch: number; roster: components["schemas"]["OrganizationMembers"] }>();
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [successor, setSuccessor] = useState("");
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
    setSuccessor("");
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
        onChanged();
        if (memberId) void load(true).finally(() => setBusy(false));
        else {
          setBusy(false);
          setSnapshot(undefined);
          onLeave();
        }
      },
      cause => {
        setError(errorMessage(cause));
        void load(true).finally(() => setBusy(false));
      },
    );
  }

  const candidates = roster?.members.filter(member => !member.owner && !member.self) ?? [];
  const selectedSuccessor = candidates.find(member => member.memberId === successor);

  function transferOwnership() {
    if (busy || !roster?.canRemove || !selectedSuccessor) return;
    const memberId = selectedSuccessor.memberId;
    if (!window.confirm(`要將「${roster.organizationName}」移轉給「${memberLabel(memberId)}」嗎？你將成為一般成員。請先與對方核對成員代碼。`)) return;
    setBusy(true);
    setError("");
    void run(
      signal => request((api, s) => api.PUT("/api/v1/organizations/{organizationId}/owner", {
        signal: s, params: { path: { organizationId } }, body: { memberId },
      }), { signal }),
      () => {
        setSuccessor("");
        onChanged();
        void load(true).finally(() => setBusy(false));
      },
      cause => {
        setError(errorMessage(cause));
        void load(true).finally(() => setBusy(false));
      },
    );
  }

  if (!user) return null;

  return <>
    {error && <Alert variant="destructive" role="alert">
      <AlertDescription>{error}</AlertDescription>
    </Alert>}
    {!roster ? <ReadFeedback
      busy={!error}
      label="載入組織中…"
      onRetry={error ? () => { setError(""); void load(true); } : undefined}
    /> : <>
      <header className="flex min-w-0 flex-wrap items-center gap-2">
        <h2 className="min-w-0 break-words text-2xl font-semibold tracking-tight">{roster.organizationName}</h2>
        <Badge variant="secondary">{roster.canRemove ? "擁有者" : "成員"}</Badge>
      </header>
      <PagePanel heading="成員">
        <ItemGroup className="divide-y">
          {roster.members.map(member => <Item key={member.memberId} role="listitem" size="sm">
            <ItemContent className="min-w-0">
              <ItemTitle className="block max-w-full break-words">
                {memberLabel(member.memberId)}{member.self ? "（你）" : ""}
              </ItemTitle>
            </ItemContent>
            {member.owner && !member.self && <Badge variant="secondary">擁有者</Badge>}
            {roster.canRemove && !member.owner && !member.self && <ItemActions>
              <Button type="button" variant="outline" disabled={busy}
                aria-label={`移除${memberLabel(member.memberId)}`}
                onClick={() => depart(member.memberId)}>移除</Button>
            </ItemActions>}
          </Item>)}
        </ItemGroup>
        <p className="mt-3 text-sm text-muted-foreground">成員可分享自己的代碼供核對。</p>
        {roster.canRemove && <p className="text-sm text-muted-foreground">請透過 LINE 官方帳號邀請新成員。</p>}
      </PagePanel>
      {children}
      <PagePanel heading="組織管理">
        <div className="grid gap-4">
          {roster.canRemove && (candidates.length ? <>
            <FormField>
              新擁有者
              <NativeSelect value={successor} disabled={busy} onChange={event => setSuccessor(event.target.value)}>
                <NativeSelectOption value="">選擇接任成員</NativeSelectOption>
                {candidates.map(member => <NativeSelectOption key={member.memberId} value={member.memberId}>
                  {memberLabel(member.memberId)}
                </NativeSelectOption>)}
              </NativeSelect>
            </FormField>
            <Button type="button" className="justify-self-start" disabled={busy || !selectedSuccessor}
              onClick={transferOwnership}>移轉所有權</Button>
          </> : <p className="text-sm text-muted-foreground">
            請先邀請其他成員，再移轉所有權。移轉後即可退出組織。
          </p>)}
          {roster.canLeave
            ? <Button type="button" variant="destructive" className="justify-self-start"
                disabled={busy} onClick={() => depart()}>退出組織</Button>
            : candidates.length > 0 && <p className="text-sm text-muted-foreground">請先移轉所有權，再退出組織。</p>}
        </div>
      </PagePanel>
    </>}
  </>;
}
