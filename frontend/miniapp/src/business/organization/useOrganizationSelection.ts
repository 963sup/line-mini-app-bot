import { useEffect, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";

export function useOrganizationSelection(onSelected: (organizationId: string) => void) {
  const { request } = useSession();
  const { run, cancel } = useLatestRequest();
  const [context, setContext] =
    useState<components["schemas"]["OrganizationContext"]>();
  const [error, setError] = useState(""),
    [saving, setSaving] = useState(false);
  function load() {
    return run(
      (signal) =>
        request(
          (api, s) => api.GET("/api/v1/me/organization-context", { signal: s }),
          { signal },
        ),
      setContext,
      (e) => setError(errorMessage(e)),
    );
  }
  useEffect(() => {
    void load();
    return cancel;
  }, [request, run, cancel]);
  function select(organizationId: string) {
    if (saving || context?.selected?.organizationId === organizationId) return;
    setSaving(true);
    setError("");
    void run(
      (signal) =>
        request(
          (api, s) =>
            api.PUT("/api/v1/me/organization-context", {
              signal: s,
              body: { organizationId },
            }),
          { signal },
        ),
      (next) => {
        setContext(next);
        onSelected(organizationId);
      },
      (e) => {
        setSaving(false);
        setError(errorMessage(e));
        void load();
      },
    );
  }
  return { context, error, saving, select };
}
