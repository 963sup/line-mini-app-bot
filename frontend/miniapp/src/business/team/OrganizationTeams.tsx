import { TeamSummary } from "./TeamSummary";
import { TeamManagement } from "./TeamActions";
import { useEffect, useState } from "react";
import { useSession } from "@/identity/Session";
import { useLatestRequest, errorMessage } from "@/shared/http/lifecycle";
import { PagePanel } from "@/shared/PagePanel";
import { ReadFeedback } from "@/shared/ReadFeedback";
import type { components } from "@repo/api-client";

export function OrganizationTeams({
  organizationId,
}: {
  organizationId: string;
}) {
  const { request, user } = useSession(),
    { run, cancel } = useLatestRequest();
  const [teams, setTeams] = useState<
    components["schemas"]["OrganizationTeamOption"][]
  >([]);
  const [canBindTeam, setCanBindTeam] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [attempt, setAttempt] = useState(0);
  useEffect(() => {
    if (!user) return;
    setTeams([]);
    setLoading(true);
    setError("");
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
      ({ teams, canBindTeam }) => {
        setCanBindTeam(canBindTeam);
        setTeams(teams);
        setLoading(false);
      },
      (e) => {
        setError(errorMessage(e));
        setLoading(false);
      },
    );
    return cancel;
  }, [request, user, organizationId, run, cancel, attempt]);
  if (!user) return null;
  return (
    <PagePanel heading="組織團隊">
      <ReadFeedback busy={loading} label="載入團隊中…" message={error}
        onRetry={error ? () => setAttempt(value => value + 1) : undefined} />
      {!loading && !error && (teams.length ? (
        <ul className="divide-y">
          {teams.map(team => <li key={team.teamId}>
            <TeamSummary name={team.name || team.teamId} management={team.managementScope === "enterprise" ? "企業團隊" : "組織團隊"} auditStatus={team.auditHistoryStatus} />
          </li>)}
        </ul>
      ) : <p className="text-sm text-muted-foreground">尚無可見團隊。</p>)}
      {!loading && !error && canBindTeam && <TeamManagement key={organizationId} scope="organization" managementId={organizationId} onChanged={() => setAttempt(value => value + 1)} />}
    </PagePanel>
  );
}
