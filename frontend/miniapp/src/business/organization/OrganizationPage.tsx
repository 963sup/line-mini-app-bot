import { AuditTab, AuditPanel } from "@/shared/AuditTab";
import { useCallback, useEffect, useRef, useState, type ReactNode } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";

import { PagePanel } from "@/shared/PagePanel";
import { FormField } from "@/shared/FormField";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { Alert, AlertDescription } from "@/shared/ui/alert";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/shared/ui/alert_dialog";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Item, ItemActions, ItemContent, ItemGroup, ItemTitle } from "@/shared/ui/item";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native_select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/shared/ui/tabs";

const memberLabel = (id: string) => `成員 ${id.slice(0, 12)}`;

function ConfirmAction({
  trigger,
  title,
  description,
  action,
  onConfirm,
}: {
  trigger: ReactNode;
  title: string;
  description: string;
  action: string;
  onConfirm: () => void;
}) {
  return (
    <AlertDialog>
      <AlertDialogTrigger asChild>{trigger}</AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle className="[overflow-wrap:anywhere]">{title}</AlertDialogTitle>
          <AlertDialogDescription className="[overflow-wrap:anywhere]">
            {description}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction variant="destructive" onClick={onConfirm}>
            {action}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

export function OrganizationPage({
  organizationId,
  children,
  onChanged,
  onLeave,
}: {
  organizationId: string;
  children: ReactNode;
  onChanged: () => void;
  onLeave: () => void;
}) {
  const { epoch, user, request } = useSession();
  const { run, cancel } = useLatestRequest();
  const [snapshot, setSnapshot] = useState<{
    epoch: number;
    roster: components["schemas"]["OrganizationMembers"];
  }>();
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [successor, setSuccessor] = useState("");
  const keys = useRef(new Map<string, string>());
  const roster = user && snapshot?.epoch === epoch ? snapshot.roster : undefined;

  const load = useCallback(
    (clearOnFailure: boolean) =>
      run(
        (signal) =>
          request(
            (api, s) =>
              api.GET("/api/v1/organizations/{organizationId}/members", {
                signal: s,
                params: { path: { organizationId } },
              }),
            { signal },
          ),
        (roster) => setSnapshot({ epoch, roster }),
        (cause) => {
          if (clearOnFailure) setSnapshot(undefined);
          setError(errorMessage(cause));
        },
      ),
    [epoch, organizationId, request, run],
  );

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
    const target = memberId ?? "";
    const key = keys.current.get(target) ?? crypto.randomUUID();
    keys.current.set(target, key);
    setBusy(true);
    setError("");
    void run(
      (signal) =>
        request(
          (api, s) =>
            memberId
              ? api.DELETE("/api/v1/organizations/{organizationId}/members/{memberId}", {
                  signal: s,
                  params: {
                    path: { organizationId, memberId },
                    header: { "Idempotency-Key": key },
                  },
                })
              : api.DELETE("/api/v1/organizations/{organizationId}/membership", {
                  signal: s,
                  params: { path: { organizationId }, header: { "Idempotency-Key": key } },
                }),
          { signal },
        ),
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
      (cause) => {
        setError(errorMessage(cause));
        void load(true).finally(() => setBusy(false));
      },
    );
  }

  const candidates = roster?.members.filter((member) => member.canReceiveOwnership) ?? [];
  const selectedSuccessor = candidates.find((member) => member.memberId === successor);

  function transferOwnership() {
    if (busy || !roster?.canTransferOwnership || !selectedSuccessor) return;
    const memberId = selectedSuccessor.memberId;
    setBusy(true);
    setError("");
    void run(
      (signal) =>
        request(
          (api, s) =>
            api.PUT("/api/v1/organizations/{organizationId}/owner", {
              signal: s,
              params: { path: { organizationId } },
              body: { memberId },
            }),
          { signal },
        ),
      () => {
        setSuccessor("");
        onChanged();
        void load(true).finally(() => setBusy(false));
      },
      (cause) => {
        setError(errorMessage(cause));
        void load(true).finally(() => setBusy(false));
      },
    );
  }

  if (!user) return null;

  return (
    <>
      {error && (
        <Alert variant="destructive" role="alert">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
      {!roster ? (
        <ReadFeedback
          busy={!error}
          label="載入組織中…"
          onRetry={
            error
              ? () => {
                  setError("");
                  void load(true);
                }
              : undefined
          }
        />
      ) : (
        <>
          <header className="flex min-w-0 flex-wrap items-center gap-2">
            <h2 className="min-w-0 break-words text-2xl font-semibold tracking-tight">
              {roster.organizationName}
            </h2>
            <Badge variant="secondary">{roster.isOwner ? "擁有者" : "成員"}</Badge>
          </header>
          <Tabs key={organizationId} defaultValue="members">
            <TabsList aria-label="組織功能" className="h-auto w-full flex-wrap">
              <TabsTrigger value="members">成員</TabsTrigger>
              <TabsTrigger value="teams">組織團隊</TabsTrigger>
              <TabsTrigger value="management">組織管理</TabsTrigger>
              {roster.auditHistoryStatus === "unavailable" && <AuditTab />}
            </TabsList>
            <TabsContent value="members">
              <PagePanel heading="成員">
                <ItemGroup className="divide-y">
                  {roster.members.map((member) => (
                    <Item key={member.memberId} role="listitem" size="sm">
                      <ItemContent className="min-w-0">
                        <ItemTitle className="block max-w-full break-words">
                          {memberLabel(member.memberId)}
                          {member.self ? "（你）" : ""}
                        </ItemTitle>
                      </ItemContent>
                      {member.owner && !member.self && <Badge variant="secondary">擁有者</Badge>}
                      {member.canRemove && (
                        <ItemActions>
                          <ConfirmAction
                            title={`從「${roster.organizationName}」移除成員？`}
                            description={`請核對成員代碼：${member.memberId}。移除後，該成員將失去組織成員資格。`}
                            action="移除成員"
                            onConfirm={() => depart(member.memberId)}
                            trigger={
                              <Button
                                type="button"
                                variant="outline"
                                disabled={busy}
                                aria-label={`移除${memberLabel(member.memberId)}`}
                              >
                                移除
                              </Button>
                            }
                          />
                        </ItemActions>
                      )}
                    </Item>
                  ))}
                </ItemGroup>
                <p className="mt-3 text-sm text-muted-foreground">成員可分享自己的代碼供核對。</p>
                {roster.isOwner && (
                  <p className="text-sm text-muted-foreground">請透過 LINE 官方帳號邀請新成員。</p>
                )}
              </PagePanel>
            </TabsContent>
            <TabsContent value="teams">{children}</TabsContent>
            <TabsContent value="management">
              <PagePanel heading="組織管理">
                <div className="grid gap-4">
                  {roster.canTransferOwnership &&
                    (candidates.length ? (
                      <>
                        <FormField>
                          新擁有者
                          <NativeSelect
                            value={successor}
                            disabled={busy}
                            onChange={(event) => setSuccessor(event.target.value)}
                          >
                            <NativeSelectOption value="">選擇接任成員</NativeSelectOption>
                            {candidates.map((member) => (
                              <NativeSelectOption key={member.memberId} value={member.memberId}>
                                {memberLabel(member.memberId)}
                              </NativeSelectOption>
                            ))}
                          </NativeSelect>
                        </FormField>
                        <ConfirmAction
                          title={`移轉「${roster.organizationName}」的所有權？`}
                          description={`接任成員代碼：${successor}。移轉後，你將成為一般成員。`}
                          action="移轉所有權"
                          onConfirm={transferOwnership}
                          trigger={
                            <Button
                              type="button"
                              className="justify-self-start"
                              disabled={busy || !selectedSuccessor}
                            >
                              移轉所有權
                            </Button>
                          }
                        />
                      </>
                    ) : (
                      <p className="text-sm text-muted-foreground">
                        請先邀請其他成員，再移轉所有權。移轉後即可退出組織。
                      </p>
                    ))}
                  {roster.canLeave ? (
                    <ConfirmAction
                      title={`退出「${roster.organizationName}」？`}
                      description="退出後，你將失去這個組織的成員資格。"
                      action="退出組織"
                      onConfirm={() => depart()}
                      trigger={
                        <Button
                          type="button"
                          variant="destructive"
                          className="justify-self-start"
                          disabled={busy}
                        >
                          退出組織
                        </Button>
                      }
                    />
                  ) : (
                    candidates.length > 0 && (
                      <p className="text-sm text-muted-foreground">請先移轉所有權，再退出組織。</p>
                    )
                  )}
                </div>
              </PagePanel>
            </TabsContent>
            {roster.auditHistoryStatus === "unavailable" && (
              <AuditPanel scope={roster.organizationName} />
            )}
          </Tabs>
        </>
      )}
    </>
  );
}
