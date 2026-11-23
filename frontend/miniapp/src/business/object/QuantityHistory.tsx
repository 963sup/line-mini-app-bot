import { useCallback, useEffect, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { useLatestRequest, errorMessage, isTemporaryReadError } from "@/shared/http/lifecycle";
import { PagePanel } from "@/shared/PagePanel";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { Button } from "@/shared/ui/button";
import { displayTime } from "@/shared/time/time";

const amount = (quantity: string, unit: string) =>
  quantity ? [quantity, unit].filter(Boolean).join(" ") : "未填";

export function QuantityHistory({ repositoryId }: { repositoryId: string }) {
  const { request, user, epoch } = useSession();
  const { run, cancel } = useLatestRequest();
  const [items, setItems] = useState<components["schemas"]["ObjectQuantityChange"][]>([]);
  const [cursor, setCursor] = useState("");
  const [busy, setBusy] = useState(true);
  const [error, setError] = useState("");
  const load = useCallback(
    (next: string) => {
      setBusy(true);
      setError("");
      if (!next) {
        setItems([]);
        setCursor("");
      }
      void run(
        (signal) =>
          request(
            (client, s) =>
              client.GET("/api/v1/repositories/{repositoryId}/object-quantity-changes", {
                signal: s,
                params: { path: { repositoryId }, query: { cursor: next || undefined } },
              }),
            { signal },
          ),
        (page) => {
          setItems((previous) => (next ? [...previous, ...page.items] : page.items));
          setCursor(page.nextCursor);
          setBusy(false);
        },
        (failure) => {
          if (!isTemporaryReadError(failure)) {
            setItems([]);
            setCursor("");
          }
          setError(errorMessage(failure));
          setBusy(false);
        },
      );
    },
    [repositoryId, run, request],
  );
  useEffect(() => {
    if (user) load("");
    return cancel;
  }, [load, user, epoch, cancel]);
  return (
    <PagePanel heading="物件數量修改歷史">
      <ReadFeedback
        busy={busy}
        message={error}
        label="正在讀取修改歷史……"
        onRetry={error ? () => load(cursor) : undefined}
      />
      {!busy && !error && !items.length && !cursor && <p>尚無數量修改紀錄。</p>}
      <ul className="grid gap-3">
        {items.map((item) => (
          <li key={item.objectId + ":" + item.version}>
            <strong>{item.name}</strong> · {amount(item.beforeQuantity, item.beforeUnit)} →{" "}
            {amount(item.afterQuantity, item.afterUnit)}
            <p className="text-sm text-muted-foreground">
              {item.actor} · <time dateTime={item.at}>{displayTime(item.at)}</time>
            </p>
          </li>
        ))}
      </ul>
      {cursor && !error && (
        <Button variant="outline" disabled={busy} onClick={() => load(cursor)}>
          載入更早紀錄
        </Button>
      )}
    </PagePanel>
  );
}
