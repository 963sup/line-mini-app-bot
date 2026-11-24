import { useRef, useState, type ReactNode } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { errorMessage, rejectedAttempt, useLatestRequest } from "@/shared/http/lifecycle";
import { FormField } from "@/shared/FormField";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native_select";

async function teamWrite(operation: Promise<{ response: Response; error?: unknown }>) {
  const result = await operation;
  return { ...result, data: result.response.status === 204 ? null : undefined };
}

// A later rejection only describes that attempt; it cannot resolve a prior unknown result.
function TeamAction({
  label,
  children,
  execute,
  onChanged,
  onDenied,
}: {
  label: string;
  children: ReactNode;
  execute: (data: FormData, signal: AbortSignal) => Promise<unknown>;
  onChanged: () => void;
  onDenied?: (error: unknown) => void;
}) {
  const { run } = useLatestRequest();
  const pending = useRef<((signal: AbortSignal) => Promise<unknown>) | null>(null);
  const writing = useRef(false);
  const [busy, setBusy] = useState(false),
    [retry, setRetry] = useState(false),
    [message, setMessage] = useState("");
  return (
    <form
      className="grid gap-2"
      onSubmit={(event) => {
        event.preventDefault();
        if (writing.current) return;
        const data = new FormData(event.currentTarget);
        pending.current ??= (signal) => execute(data, signal);
        writing.current = true;
        setBusy(true);
        setMessage("");
        void run(
          pending.current,
          () => {
            pending.current = null;
            writing.current = false;
            setBusy(false);
            setRetry(false);
            setMessage("已完成。");
            onChanged();
          },
          (error) => {
            writing.current = false;
            setBusy(false);
            const response = error as { httpStatus?: number } | undefined;
            const uncertain = retry || !rejectedAttempt(error, ["conflict", "not_configured"]);
            if (!uncertain) pending.current = null;
            setRetry(uncertain);
            setMessage(errorMessage(error) + (uncertain ? " 結果待確認，請重試原操作。" : ""));
            if ([401, 403, 404].includes(response?.httpStatus ?? 0)) onDenied?.(error);
          },
        );
      }}
    >
      <fieldset disabled={busy || retry} className="grid min-w-0 gap-2">
        {children}
      </fieldset>
      <Button className="justify-self-start" disabled={busy}>
        {retry ? "重試原操作" : label}
      </Button>
      {message && <p role="status">{message}</p>}
    </form>
  );
}

export function TeamManagement({
  scope,
  managementId,
  onChanged,
  onDenied,
}: {
  scope: components["schemas"]["TeamManagement"]["managementScope"];
  managementId: string;
  onChanged: () => void;
  onDenied?: (error: unknown) => void;
}) {
  const { request } = useSession();
  return (
    <TeamAction
      key={`${scope}:${managementId}`}
      label="綁定團隊"
      onChanged={onChanged}
      onDenied={onDenied}
      execute={(data, signal) =>
        request(
          (api, s) =>
            teamWrite(
              api.PUT("/api/v1/teams/{teamId}/management", {
                params: { path: { teamId: String(data.get("teamId")).trim() } },
                body: { managementScope: scope, managementId },
                signal: s,
              }),
            ),
          { signal },
        )
      }
    >
      <FormField>
        LINE 群組 ID
        <Input name="teamId" required />
      </FormField>
    </TeamAction>
  );
}

export function EnterpriseTeamAssignment({
  enterprise,
  onChanged,
  onDenied,
}: {
  enterprise: components["schemas"]["Enterprise"];
  onChanged: () => void;
  onDenied?: (error: unknown) => void;
}) {
  const { request } = useSession();
  if (!enterprise.teams.length || !enterprise.assignableOrganizations?.length) return null;
  return (
    <TeamAction
      key={enterprise.id}
      label="送出指派設定"
      onChanged={onChanged}
      onDenied={onDenied}
      execute={(data, signal) =>
        request(
          (api, s) => {
            const options = {
              params: {
                path: {
                  organizationId: String(data.get("organizationId")),
                  teamId: String(data.get("teamId")),
                },
              },
              signal: s,
            };
            return teamWrite(
              data.get("action") === "remove"
                ? api.DELETE(
                    "/api/v1/organizations/{organizationId}/enterprise-teams/{teamId}",
                    options,
                  )
                : api.PUT(
                    "/api/v1/organizations/{organizationId}/enterprise-teams/{teamId}",
                    options,
                  ),
            );
          },
          { signal },
        )
      }
    >
      <FormField>
        企業團隊
        <NativeSelect name="teamId" required>
          {enterprise.teams.map((team) => (
            <NativeSelectOption key={team.id} value={team.id}>
              {team.name || team.id}
            </NativeSelectOption>
          ))}
        </NativeSelect>
      </FormField>
      <FormField>
        旗下組織
        <NativeSelect name="organizationId" required>
          {enterprise.assignableOrganizations.map((org) => (
            <NativeSelectOption key={org.id} value={org.id}>
              {org.name || org.id}
            </NativeSelectOption>
          ))}
        </NativeSelect>
      </FormField>
      <FormField>
        操作
        <NativeSelect name="action">
          <NativeSelectOption value="assign">指派至組織</NativeSelectOption>
          <NativeSelectOption value="remove">移除指派</NativeSelectOption>
        </NativeSelect>
      </FormField>
    </TeamAction>
  );
}
