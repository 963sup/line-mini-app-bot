import { TeamSummary } from "../team/TeamSummary";
import { AuditTab, AuditPanel } from "@/shared/AuditTab";
import { EnterpriseTeamAssignment, TeamManagement } from "@/business/team/TeamActions";
import { useEffect, useRef, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { useLatestRequest, errorMessage, rejectedAttempt } from "@/shared/http/lifecycle";
import { Input } from "@/shared/ui/input";
import { Button } from "@/shared/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/shared/ui/tabs";
import { PagePanel } from "@/shared/PagePanel";

type EnterpriseAction = "create" | "owner" | "organization";

export function Enterprise({
  id,
  canCreate,
  eligibilityError,
  onSaved,
}: {
  id: string;
  canCreate?: boolean;
  eligibilityError?: string;
  onSaved: (enterpriseId: string) => void;
}) {
  const { request, user } = useSession(),
    { run, cancel } = useLatestRequest();
  const write = useLatestRequest();
  const [view, setView] = useState<components["schemas"]["Enterprise"]>(),
    [busy, setBusy] = useState(false),
    [pending, setPending] = useState<{ kind: EnterpriseAction; value: string }>(),
    [revision, setRevision] = useState(0),
    [readFailed, setReadFailed] = useState(false),
    [message, setMessage] = useState("");
  const [checkingTeamAccess, setCheckingTeamAccess] = useState(false);
  const writing = useRef(false);
  const create = id === "new";
  const eligibilityMessage =
    create &&
    (canCreate === undefined ? eligibilityError : !canCreate && "帳號尚無可用的企業開通資格。");
  useEffect(() => {
    if (!user || create) return;
    setReadFailed(false);
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
      (value) => {
        setView(value);
        setCheckingTeamAccess(false);
      },
      (e) => {
        setView(undefined);
        setReadFailed(true);
        setMessage(errorMessage(e));
      },
    );
    return cancel;
  }, [id, create, user, request, run, cancel, revision]);
  const refreshTeams = () => setRevision((value) => value + 1);
  const teamDenied = (error: unknown) => {
    setMessage(errorMessage(error));
    setCheckingTeamAccess(true);
    refreshTeams();
  };
  function form(kind: EnterpriseAction, label: string, action: string) {
    const retry = pending?.kind === kind;
    const blocked = busy || (!!pending && !retry);
    return (
      <form
        onSubmit={(event) => {
          event.preventDefault();
          if (writing.current || blocked) return;
          const value =
            pending?.value ?? String(new FormData(event.currentTarget).get("value") ?? "").trim();
          if (!value) return;
          writing.current = true;
          setBusy(true);
          setMessage("");
          const params = { path: { enterpriseId: id } };
          void write.run(
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
                      : api.POST("/api/v1/enterprises/{enterpriseId}/organizations", {
                          params,
                          body: { organizationId: value },
                          signal: s,
                        }),
                { signal },
              ),
            (v) => {
              setPending(undefined);
              onSaved(v.id);
            },
            (e) => {
              writing.current = false;
              setBusy(false);
              const unknown = retry || !rejectedAttempt(e, ["conflict", "not_configured"]);
              if (unknown) setPending({ kind, value });
              const status = (e as { httpStatus?: number })?.httpStatus;
              if (kind !== "create" && (status === 401 || status === 403 || status === 404)) {
                setView(undefined);
                setRevision((previous) => previous + 1);
              }
              setMessage(errorMessage(e) + (unknown ? " 結果待確認，請沿用原資料重試。" : ""));
            },
          );
        }}
        className="flex flex-wrap items-end gap-2"
      >
        <label>
          {label}
          <Input
            name="value"
            required
            defaultValue={retry ? pending?.value : undefined}
            readOnly={retry}
            disabled={blocked}
          />
        </label>
        <Button disabled={blocked}>
          {retry ? (kind === "create" ? "重試原企業名稱" : "重試原操作") : action}
        </Button>
      </form>
    );
  }
  if (!user) return null;
  return (
    <section className="grid gap-4">
      {(message || eligibilityMessage) && <p role="status">{message || eligibilityMessage}</p>}
      {readFailed && (
        <Button
          onClick={() => {
            setMessage("");
            setRevision((previous) => previous + 1);
          }}
        >
          重新讀取企業
        </Button>
      )}
      {create && canCreate && form("create", "企業名稱", "建立企業")}
      {view && (
        <>
          <h2>{view.name}</h2>
          <Tabs key={`${id}:${view.canManage}`} defaultValue="organizations">
            <TabsList aria-label="企業功能" className="h-auto w-full flex-wrap">
              <TabsTrigger value="organizations">旗下組織</TabsTrigger>
              <TabsTrigger value="teams">企業團隊</TabsTrigger>
              {view.canManage && <TabsTrigger value="management">企業管理</TabsTrigger>}
              {view.auditHistoryStatus === "unavailable" && <AuditTab />}
            </TabsList>
            <TabsContent value="organizations">
              <PagePanel heading="旗下組織">
                {view.organizations.length ? (
                  <ul>
                    {view.organizations.map((o) => (
                      <li key={o.id}>{o.name}</li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-sm text-muted-foreground">尚無可見的旗下組織。</p>
                )}
              </PagePanel>
            </TabsContent>
            <TabsContent value="teams">
              <PagePanel heading="企業團隊">
                {view.teams.length ? (
                  <ul>
                    {view.teams.map((t) => (
                      <li key={t.id}>
                        <TeamSummary
                          name={t.name || t.id}
                          management="企業團隊"
                          auditStatus={t.auditHistoryStatus}
                        />
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-sm text-muted-foreground">尚無可見的企業團隊。</p>
                )}
                {view.canManageTeams && !checkingTeamAccess && (
                  <>
                    <TeamManagement
                      scope="enterprise"
                      managementId={id}
                      onChanged={refreshTeams}
                      onDenied={teamDenied}
                    />
                    <EnterpriseTeamAssignment
                      enterprise={view}
                      onChanged={refreshTeams}
                      onDenied={teamDenied}
                    />
                  </>
                )}
              </PagePanel>
            </TabsContent>
            {view.canManage && (
              <TabsContent value="management">
                <PagePanel heading="企業管理">
                  {form("organization", "組織 ID", "加入組織")}
                  {form("owner", "LINE 使用者 ID", "新增企業 Owner")}
                </PagePanel>
              </TabsContent>
            )}
            {view.auditHistoryStatus === "unavailable" && <AuditPanel scope={view.name} />}
          </Tabs>
        </>
      )}
    </section>
  );
}
