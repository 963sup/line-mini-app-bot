import { TeamSummary } from "./TeamSummary";
import { TeamManagement } from "./TeamActions";
import { useEffect, useState } from "react";
import { useSession } from "@/identity/Session";
import { useLatestRequest, errorMessage } from "@/shared/http/lifecycle";
import { PagePanel } from "@/shared/PagePanel";
import { ReadFeedback } from "@/shared/ReadFeedback";
import type { components } from "@repo/api-client";

export function OrganizationTeams({ organizationId }: { organizationId: string }) {
  const { request, user } = useSession(),
    { run, cancel } = useLatestRequest();
  const [result, setResult] = useState<
    components["schemas"]["OrganizationTeamsResponse"] | string | null
  >(null);
  const [attempt, setAttempt] = useState(0);
  useEffect(() => {
    if (!user) return;
    setResult(null);
    void run(
      (signal) =>
        request(
          (api, s) =>
            api.GET("/api/v1/organizations/{organizationId}/teams", {
              signal: s,
              params: { path: { organizationId } },
            }),
          { signal },
        ),
      setResult,
      (e) => setResult(errorMessage(e) || "讀取團隊失敗，請重試。"),
    );
    return cancel;
  }, [request, user, organizationId, run, cancel, attempt]);
  if (!user) return null;
  const error = typeof result === "string" ? result : undefined;
  const data = typeof result === "string" ? null : result;
  return (
    <PagePanel heading="團隊">
      <ReadFeedback
        busy={result === null}
        label="載入團隊中…"
        message={error}
        onRetry={error ? () => setAttempt((value) => value + 1) : undefined}
      />
      {data &&
        (data.teams.length ? (
          <ul className="divide-y">
            {data.teams.map((team) => (
              <li key={team.teamId}>
                <TeamSummary
                  name={team.name || team.teamId}
                  management={team.managementScope === "enterprise" ? "企業團隊" : "團隊"}
                  auditStatus={team.auditHistoryStatus}
                />
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-sm text-muted-foreground">尚無可見團隊。</p>
        ))}
      {data?.canBindTeam && (
        <TeamManagement
          key={organizationId}
          scope="organization"
          managementId={organizationId}
          onChanged={() => setAttempt((value) => value + 1)}
        />
      )}
    </PagePanel>
  );
}
