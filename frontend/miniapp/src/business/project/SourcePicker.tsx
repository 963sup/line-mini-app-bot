import { useCallback, useEffect, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { errorMessage, isTemporaryReadError, useLatestRequest } from "@/shared/http/lifecycle";
import { FormField } from "@/shared/FormField";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";

type Repository = components["schemas"]["Repository"];
export function SourcePicker({ ownerId, initialSource, value, onChange, disabled }: {
  ownerId: string; initialSource: string; value: string; onChange: (id: string) => void; disabled: boolean;
}) {
  const { request, user } = useSession();
  const list = useLatestRequest(), detail = useLatestRequest();
  const [items, setItems] = useState<Repository[]>([]), [cursor, setCursor] = useState("");
  const [busy, setBusy] = useState(false), [status, setStatus] = useState("");
  const [listError, setListError] = useState("");
  const load = useCallback((next = "") => {
    if (!user) return;
    setBusy(true); setCursor(""); setListError("");
    void list.run(signal => request((client, s) => client.GET("/api/v1/repositories", {
      signal: s, params: { query: { cursor: next || undefined, ownerId: ownerId || undefined } },
    }), { signal }), page => {
      setItems(previous => [...new Map([...previous, ...page.items].map(item => [item.id, item])).values()]);
      setCursor(page.nextCursor); setBusy(false);
    }, error => {
      if (isTemporaryReadError(error)) setCursor(next);
      else { setItems([]); detail.cancel(); }
      setBusy(false); setListError(errorMessage(error));
    });
  }, [user, ownerId, request, list.run, detail.cancel]);
  useEffect(() => { load(); return list.cancel; }, [load, list.cancel]);
  useEffect(() => {
    if (!user || !initialSource || !/^[0-9a-f]{32}$/.test(initialSource)) return;
    void detail.run(signal => request((client, s) => client.GET("/api/v1/repositories/{id}", {
      signal: s, params: { path: { id: initialSource } },
    }), { signal }), item => {
      if (ownerId && item.ownerId !== ownerId) { setStatus("來源擁有者不符或無法讀取。"); return; }
      setItems(previous => [...new Map([...previous, item].map(value => [value.id, value])).values()]);
      onChange(item.id);
    }, error => setStatus(errorMessage(error)));
    return detail.cancel;
  }, [user, initialSource, ownerId, request, detail.run, detail.cancel, onChange]);
  return <>
    <FormField>來源儲存庫<NativeSelect value={value} disabled={disabled || busy} onChange={event => onChange(event.target.value)}>
      <NativeSelectOption value="">請選擇有權讀取的來源</NativeSelectOption>
      {items.map(item => <NativeSelectOption key={item.id} value={item.id}>{item.name || "群組儲存庫"}</NativeSelectOption>)}
    </NativeSelect></FormField>
    <ReadFeedback busy={busy} message={listError} label="正在讀取來源儲存庫……" onRetry={listError && !disabled ? () => load(cursor) : undefined} />
    <Status role="status">{status}</Status>
    {!listError && cursor && <Button type="button" variant="outline" disabled={disabled || busy} onClick={() => load(cursor)}>載入更多來源儲存庫</Button>}
  </>;
}
