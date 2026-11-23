import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";

type Snapshot = {
  owner: string;
  context?: components["schemas"]["OrganizationContext"];
  error: string;
  saving: boolean;
};
type Selection = Omit<Snapshot, "owner"> & {
  refresh: () => void;
  select: (organizationId: string, onSelected: (id: string) => void) => void;
};
const SelectionContext = createContext<Selection | null>(null);
const empty: Snapshot = { owner: "", error: "", saving: false };

export function OrganizationSelectionProvider({ children }: { children: ReactNode }) {
  const { epoch, user, request } = useSession();
  const owner = user ? epoch + ":" + user.userId : "";
  const [snapshot, setSnapshot] = useState(empty);
  const current = owner && snapshot.owner === owner ? snapshot : empty;
  const { run, cancel } = useLatestRequest();
  const load = useCallback(
    (error = "") =>
      run(
        (signal) =>
          request((api, s) => api.GET("/api/v1/me/organization-context", { signal: s }), {
            signal,
          }),
        (context) => setSnapshot({ owner, context, error, saving: false }),
        (cause) => setSnapshot({ owner, error: errorMessage(cause), saving: false }),
      ),
    [request, run, owner],
  );

  useEffect(() => {
    if (owner) void load();
    return cancel;
  }, [load, cancel, owner]);

  function select(organizationId: string, onSelected: (id: string) => void) {
    if (!owner || current.saving || current.context?.selected?.organizationId === organizationId)
      return;
    setSnapshot({ ...current, owner, saving: true, error: "" });
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
      (context) => {
        setSnapshot({ owner, context, error: "", saving: false });
        onSelected(organizationId);
      },
      (cause) => {
        const error = errorMessage(cause);
        setSnapshot({ owner, error, saving: false });
        void load(error);
      },
    );
  }
  return (
    <SelectionContext.Provider
      value={{
        ...current,
        select,
        refresh: () => {
          void load();
        },
      }}
    >
      {children}
    </SelectionContext.Provider>
  );
}

export function useOrganizationSelection() {
  const selection = useContext(SelectionContext);
  if (!selection) throw new Error("OrganizationSelectionProvider is required");
  return selection;
}
