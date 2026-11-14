import React, { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { useLatestRequest, errorMessage } from "@/shared/http/lifecycle";
import { PagePanel } from "@/shared/PagePanel";
import { FormField } from "@/shared/FormField";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { FieldGroup } from "@/shared/ui/field";
import { Empty, EmptyHeader, EmptyTitle } from "@/shared/ui/empty";
import { Item, ItemGroup, ItemContent, ItemTitle } from "@/shared/ui/item";

type Label = components["schemas"]["Label"];
export function View({ repositoryId }: { repositoryId: string }) {
  const { user, request } = useSession();
  const { run: readRun, cancel: cancelRead } = useLatestRequest(), write = useLatestRequest();
  const [items, setItems] = useState<Label[]>([]), [cursor, setCursor] = useState("");
  const [name, setName] = useState(""), [status, setStatus] = useState(""), [listStatus, setListStatus] = useState("");
  const [loaded, setLoaded] = useState(false), [readBusy, setReadBusy] = useState(false), [busy, setBusy] = useState(false), [uncertain, setUncertain] = useState(false);
  const lock = useRef(false), attempt = useRef<{ key: string; name: string } | null>(null);
  const load = useCallback((nextCursor = "") => {
    if (!user) return;
    setReadBusy(true); setLoaded(false); setCursor("");
    if (!nextCursor) setItems([]);
    setListStatus("正在讀取標籤……");
    void readRun(
      (signal) => request((client, s) => client.GET("/api/v1/repositories/{repositoryId}/labels", {
        signal: s, params: { path: { repositoryId }, query: { cursor: nextCursor || undefined } },
      }), { signal }),
      (page) => {
        setItems((previous) => [...new Map([...(nextCursor ? previous : []), ...page.items].map((item) => [item.id, item])).values()]);
        setCursor(page.nextCursor); setReadBusy(false); setLoaded(true); setListStatus(page.nextCursor ? "可載入更多標籤。" : "已讀取所有標籤。");
      },
      (error) => { setItems([]); setReadBusy(false); setLoaded(false); setListStatus(errorMessage(error)); },
    );
  }, [user, request, repositoryId, readRun]);
  useEffect(() => { load(); return cancelRead; }, [load, cancelRead]);
  function submit(event: FormEvent) {
    event.preventDefault(); if (!user || lock.current) return;
    if (!uncertain || !attempt.current) {
      const value = name.trim(); if (!value) { setStatus("請填寫標籤名稱。"); return; }
      attempt.current = { key: crypto.randomUUID(), name: value };
    }
    const pending = attempt.current; lock.current = true; setBusy(true); setUncertain(true); setStatus("正在建立標籤……");
    void write.run(
      (signal) => request((client, s) => client.POST("/api/v1/repositories/{repositoryId}/labels", {
        signal: s, params: { path: { repositoryId }, header: { "Idempotency-Key": pending.key } }, body: { name: pending.name },
      }), { signal }),
      () => { lock.current = false; setBusy(false); setUncertain(false); attempt.current = null; setName(""); setStatus("已建立標籤。"); load(); },
      (error) => {
        lock.current = false; setBusy(false); const code = (error as { code?: string })?.code;
        if (["invalid_input", "not_found", "forbidden", "name_conflict", "idempotency_conflict"].includes(code || "")) setUncertain(false);
        if (code === "name_conflict") load();
        setStatus(errorMessage(error) + " 結果未確認時請沿用原名稱重試，離頁或重新整理前請先確認目錄。");
      },
    );
  }
  return <PagePanel heading="標籤">
    <p className="mb-4 text-sm text-muted-foreground">可建立共用標籤，再到自己的問題或討論詳情選取並儲存分類。</p>
    <form onSubmit={submit}><FieldGroup>
      <FormField>標籤名稱<Input required maxLength={40} value={name} disabled={!user || busy || uncertain} onChange={(e) => setName(e.target.value)} /></FormField>
      <Button type="submit" disabled={!user || busy}>{uncertain ? "重試原建立請求" : "建立標籤"}</Button>
    </FieldGroup></form>
    <Status role="status">{status}</Status>
    <Button variant="outline" disabled={!user || readBusy || busy} onClick={() => load()}>重新載入標籤</Button>
    <Status role="status">{listStatus}</Status>
    <ItemGroup className="gap-3">
      {items.map((item) => <Item key={item.id} variant="outline" role="listitem"><ItemContent><ItemTitle>{item.name}</ItemTitle></ItemContent></Item>)}
    </ItemGroup>
    {loaded && !items.length && <Empty><EmptyHeader><EmptyTitle>目前沒有標籤</EmptyTitle></EmptyHeader></Empty>}
    {cursor && <Button variant="outline" disabled={readBusy || busy} onClick={() => load(cursor)}>載入更多標籤</Button>}
  </PagePanel>;
}
