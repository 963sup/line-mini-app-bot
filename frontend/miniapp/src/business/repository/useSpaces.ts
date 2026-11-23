import { useCallback, useEffect, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";

export function useSpaces(repositoryId: string, enabled = true) {
  const { request, user } = useSession();
  const spaceRead = useLatestRequest();
  const [spaces, setSpaces] = useState<components["schemas"]["RepositorySpace"][]>([]);
  const [spaceCursor, setSpaceCursor] = useState(""),
    [spacesBusy, setSpacesBusy] = useState(false),
    [spaceStatus, setSpaceStatus] = useState("");
  const loadSpaces = useCallback(
    (next = "") => {
      if (!next) {
        setSpaces([]);
        setSpaceCursor("");
      }
      setSpacesBusy(true);
      setSpaceStatus("");
      void spaceRead.run(
        (signal) =>
          request(
            (client, s) =>
              client.GET("/api/v1/repositories/{repositoryId}/spaces", {
                signal: s,
                params: { path: { repositoryId }, query: { cursor: next || undefined } },
              }),
            { signal },
          ),
        (page) => {
          setSpaces((previous) => [
            ...new Map(
              [...(next ? previous : []), ...page.items].map((item) => [item.id, item]),
            ).values(),
          ]);
          setSpaceCursor(page.nextCursor);
          setSpacesBusy(false);
        },
        (error) => {
          const status = (error as { httpStatus?: number }).httpStatus;
          if (status === 401 || status === 403 || status === 404) {
            setSpaces([]);
            setSpaceCursor("");
          }
          setSpacesBusy(false);
          setSpaceStatus(errorMessage(error));
        },
      );
    },
    [request, repositoryId, spaceRead.run],
  );
  useEffect(() => {
    if (enabled && user) loadSpaces();
    return spaceRead.cancel;
  }, [enabled, user, loadSpaces, spaceRead.cancel]);
  return { spaces, setSpaces, spaceCursor, spacesBusy, spaceStatus, loadSpaces };
}
