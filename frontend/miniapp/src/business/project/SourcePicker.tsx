import React, { useCallback, useEffect, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";
import { FormField } from "@/shared/FormField";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";

type Repository = components["schemas"]["Repository"];
export function SourcePicker({ spaceId, initialSource, value, onChange, disabled }: {
  spaceId: string; initialSource: string; value: string; onChange: (id: string) => void; disabled: boolean;
}) {
  const { request } = useSession();
  const list = useLatestRequest(), detail = useLatestRequest();
  const [items, setItems] = useState<Repository[]>([]), [cursor, setCursor] = useState("");
  const [busy, setBusy] = useState(false), [status, setStatus] = useState("");
  const load = useCallback((next = "") => {
    setBusy(true); setCursor("");
    void list.run(signal => request((client, s) => client.GET("/api/v1/repositories", {
      signal: s, params: { query: { cursor: next || undefined } },
    }), { signal }), page => {
      const allowed = page.items.filter(item => item.groupId === spaceId && item.status === "ACTIVE");
      setItems(previous => [...new Map([...previous, ...allowed].map(item => [item.id, item])).values()]);
      setCursor(page.nextCursor); setBusy(false); setStatus(page.nextCursor ? "可繼續查詢有權讀取的來源。" : "已讀取可用來源。");
    }, error => { setBusy(false); setStatus(errorMessage(error)); });
  }, [spaceId, request, list.run]);
  useEffect(() => { load(); return list.cancel; }, [load, list.cancel]);
  useEffect(() => {
    if (!initialSource || !/^[0-9a-f]{32}$/.test(initialSource)) return;
    void detail.run(signal => request((client, s) => client.GET("/api/v1/repositories/{id}", {
      signal: s, params: { path: { id: initialSource } },
    }), { signal }), item => {
      if (item.groupId !== spaceId || item.status !== "ACTIVE") { setStatus("來源不屬於此空間或不可讀取。"); return; }
      setItems(previous => [...new Map([...previous, item].map(value => [value.id, value])).values()]);
      onChange(item.id);
    }, error => setStatus(errorMessage(error)));
    return detail.cancel;
  }, [initialSource, spaceId, request, detail.run, detail.cancel, onChange]);
  return <>
    <FormField>來源儲存庫<NativeSelect required value={value} disabled={disabled || busy} onChange={event => onChange(event.target.value)}>
      <NativeSelectOption value="">請選擇有權讀取的來源</NativeSelectOption>
      {items.map(item => <NativeSelectOption key={item.id} value={item.id}>{item.name || "群組儲存庫"}</NativeSelectOption>)}
    </NativeSelect></FormField>
    <Status role="status">{status}</Status>
    {cursor && <Button type="button" variant="outline" disabled={disabled || busy} onClick={() => load(cursor)}>載入更多來源儲存庫</Button>}
  </>;
}
