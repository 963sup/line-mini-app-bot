import { ReadFeedback } from "@/shared/ReadFeedback";
import { useCallback, useEffect, useId, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { useLatestRequest, errorMessage, isTemporaryReadError } from "@/shared/http/lifecycle";
import { Button } from "@/shared/ui/button";
import { Checkbox } from "@/shared/ui/checkbox";
import { Field, FieldGroup, FieldLabel } from "@/shared/ui/field";

type Label = components["schemas"]["Label"];
export function Picker({
  repositoryId,
  value,
  onChange,
  disabled,
}: {
  repositoryId: string;
  value: string[];
  onChange: (ids: string[]) => void;
  disabled: boolean;
}) {
  const { user, request } = useSession();
  const { run, cancel } = useLatestRequest(),
    group = useId();
  const [items, setItems] = useState<Label[]>([]),
    [cursor, setCursor] = useState("");
  const [busy, setBusy] = useState(false),
    [status, setStatus] = useState("");
  const [failed, setFailed] = useState(false);
  const load = useCallback(
    (nextCursor = "") => {
      if (!user) return;
      setBusy(true);
      setCursor("");
      if (!nextCursor) setItems([]);
      setFailed(false);
      setStatus("");
      void run(
        (signal) =>
          request(
            (client, s) =>
              client.GET("/api/v1/repositories/{repositoryId}/labels", {
                signal: s,
                params: { path: { repositoryId }, query: { cursor: nextCursor || undefined } },
              }),
            { signal },
          ),
        (page) => {
          setItems((previous) => [
            ...new Map(
              [...(nextCursor ? previous : []), ...page.items].map((item) => [item.id, item]),
            ).values(),
          ]);
          setCursor(page.nextCursor);
          setBusy(false);
          setStatus(
            page.items.length || page.nextCursor || nextCursor
              ? ""
              : "目前沒有標籤，請先到「標籤」頁建立。",
          );
        },
        (error) => {
          if (nextCursor && isTemporaryReadError(error)) setCursor(nextCursor);
          else setItems([]);
          setFailed(true);
          setBusy(false);
          setStatus(errorMessage(error));
        },
      );
    },
    [user, request, repositoryId, run],
  );
  useEffect(() => {
    load();
    return cancel;
  }, [load, cancel]);
  return (
    <FieldGroup aria-label="分類標籤">
      <p className="text-sm">分類標籤（已選 {value.length}/10）</p>
      <ReadFeedback
        busy={busy}
        message={status}
        label="正在讀取可用標籤……"
        onRetry={failed ? () => load(cursor) : undefined}
      />
      {items.map((item) => (
        <Field key={item.id} orientation="horizontal">
          <Checkbox
            id={`${group}-${item.id}`}
            checked={value.includes(item.id)}
            disabled={disabled || busy || (!value.includes(item.id) && value.length >= 10)}
            onCheckedChange={(checked) =>
              onChange(
                checked === true ? [...value, item.id] : value.filter((id) => id !== item.id),
              )
            }
          />
          <FieldLabel htmlFor={`${group}-${item.id}`}>{item.name}</FieldLabel>
        </Field>
      ))}
      <div className="flex flex-wrap gap-2">
        {!failed && cursor && (
          <Button type="button" variant="outline" disabled={busy} onClick={() => load(cursor)}>
            載入更多分類標籤
          </Button>
        )}
      </div>
    </FieldGroup>
  );
}
