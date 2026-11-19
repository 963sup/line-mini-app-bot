import React, { useEffect, useState } from "react";
import { useSession } from "@/identity/Session";
import { useLatestRequest, errorMessage } from "@/shared/http/lifecycle";
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
  const [status, setStatus] = useState("載入團隊中…");
  useEffect(() => {
    if (!user) return;
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
        setStatus(teams.length ? "" : "尚無可見團隊");
      },
      (e) => setStatus(errorMessage(e)),
    );
    return cancel;
  }, [request, user, organizationId, run, cancel]);
  return (
    <section>
      <h2>組織團隊</h2>
      {status && <p role="status">{status}</p>}
      <ul>
        {teams.map((team) => (
          <li key={team.teamId}>
            {team.name || team.teamId}（
            {team.managementScope === "enterprise" ? "企業" : "組織"}團隊）
          </li>
        ))}
      </ul>
    </section>
  );
}
