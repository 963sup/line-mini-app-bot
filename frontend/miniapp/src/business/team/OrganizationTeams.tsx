import { useEffect, useState } from "react";
import { useSession } from "@/identity/Session";
import { useLatestRequest, errorMessage } from "@/shared/http/lifecycle";
import { PagePanel } from "@/shared/PagePanel";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { Badge } from "@/shared/ui/badge";
import { Item, ItemContent, ItemGroup, ItemTitle } from "@/shared/ui/item";
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
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [attempt, setAttempt] = useState(0);
  useEffect(() => {
    if (!user) return;
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
      ({ teams }) => {
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
  return (
    <PagePanel heading="組織團隊">
      <ReadFeedback busy={loading} label="載入團隊中…" message={error}
        onRetry={error ? () => setAttempt(value => value + 1) : undefined} />
      {!loading && !error && (teams.length ? (
        <ItemGroup className="divide-y">
          {teams.map(team => (
            <Item key={team.teamId} role="listitem" size="sm">
              <ItemContent className="min-w-0">
                <ItemTitle className="block max-w-full break-words">{team.name || team.teamId}</ItemTitle>
              </ItemContent>
              <Badge variant="outline">{team.managementScope === "enterprise" ? "企業團隊" : "組織團隊"}</Badge>
            </Item>
          ))}
        </ItemGroup>
      ) : <p className="text-sm text-muted-foreground">尚無可見團隊。</p>)}
    </PagePanel>
  );
}
