import React, { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { useLatestRequest, errorMessage } from "@/shared/http/lifecycle";
import { useRepositoryGuard } from "../repository/NavigationGuard";
import { useDiscardChanges } from "@/shared/useDiscardChanges";
import { FormField } from "@/shared/FormField";
import { PagePanel } from "@/shared/PagePanel";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { FieldGroup, FieldSet } from "@/shared/ui/field";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";

type Snapshot = components["schemas"]["ObjectTemplate"];
type Creation = components["schemas"]["ObjectTemplateCreation"];
type Props = { repositoryId: string; disabled: boolean; onBlocked: (blocked: boolean) => void; onApplied: (result: components["schemas"]["ObjectTemplateInstance"]) => void };
type Attempt = { key: string } & ({ kind: "create"; body: Creation } | { kind: "apply"; id: string; body: components["schemas"]["ObjectTemplateInstantiation"] });
const initial = (): Creation => ({ name: "", nodes: [{ id: "root", name: "" }] });

// Only input, request lifecycle and navigation live here. Go validates and creates the tree.
export function Templates({ repositoryId, disabled, onBlocked, onApplied }: Props) {
  const { request } = useSession();
  const read = useLatestRequest(), write = useLatestRequest();
  const [open, setOpen] = useState(false), [creating, setCreating] = useState(false);
  const [draft, setDraft] = useState<Creation>(initial), [items, setItems] = useState<Snapshot[]>([]);
  const [selected, setSelected] = useState(""), [cursor, setCursor] = useState("");
  const [busy, setBusy] = useState(false), [reading, setReading] = useState(false), [message, setMessage] = useState("");
  const [uncertain, setUncertain] = useState(false);
  const attempt = useRef<Attempt | null>(null), lock = useRef(false), panel = useRef<HTMLElement>(null), trigger = useRef<HTMLButtonElement>(null);
  const focus = () => panel.current?.focus();
  const dirty = creating && JSON.stringify(draft) !== JSON.stringify(initial());
  const guard = useRepositoryGuard(dirty, focus), discard = useDiscardChanges(dirty, focus);
  useEffect(() => { onBlocked(open); return () => onBlocked(false); }, [open, onBlocked]);
  useEffect(() => { guard.block(busy || uncertain); }, [busy, uncertain, guard.block]);
  useEffect(() => { if (open) focus(); }, [open]);
  const load = useCallback((next = "") => {
    setReading(true); setMessage("");
    if (!next) { setItems([]); setSelected(""); }
    void read.run(signal => request((client, s) => client.GET("/api/v1/repositories/{repositoryId}/object-templates", { signal: s, params: { path: { repositoryId }, query: { cursor: next || undefined } } }), { signal }), page => {
      setItems(previous => next ? [...previous, ...page.items] : page.items); setCursor(page.nextCursor); setReading(false);
    }, error => { setReading(false); setMessage(errorMessage(error)); });
  }, [read.run, request, repositoryId]);
  function close() {
    if (lock.current || uncertain) return;
    discard.navigate(() => { read.cancel(); onBlocked(false); setOpen(false); setCreating(false); setDraft(initial()); setMessage(""); trigger.current?.focus(); });
  }
  function send(pending: Attempt) {
    if (lock.current || discard.isConfirming() || !guard.canWrite()) return;
    attempt.current = pending;
    const wasUnknown = uncertain;
    lock.current = true; setBusy(true); setUncertain(true); guard.block(true); setMessage("正在處理範本……");
    const failed = (error: unknown) => {
      lock.current = false; setBusy(false);
      if (!wasUnknown && ["invalid_input", "not_found", "forbidden", "idempotency_conflict"].includes((error as { code?: string }).code || "")) { setUncertain(false); guard.block(false); attempt.current = null; }
      setMessage(errorMessage(error) + " 結果未確認時，請重試原請求。");
    };
    const settled = () => { lock.current = false; setBusy(false); setUncertain(false); guard.block(false); attempt.current = null; };
    if (pending.kind === "create") {
      void write.run(signal => request((client, s) => client.POST("/api/v1/repositories/{repositoryId}/object-templates", { signal: s, params: { path: { repositoryId }, header: { "Idempotency-Key": pending.key } }, body: pending.body }), { signal }), snapshot => {
        settled(); setItems(previous => [snapshot, ...previous.filter(item => item.id !== snapshot.id)]); setSelected(snapshot.id); setCreating(false); setDraft(initial()); setMessage("已建立範本，可直接套用。");
      }, failed);
    } else {
      void write.run(signal => request((client, s) => client.POST("/api/v1/repositories/{repositoryId}/object-templates/{id}/instances", { signal: s, params: { path: { repositoryId, id: pending.id }, header: { "Idempotency-Key": pending.key } }, body: pending.body }), { signal }), result => {
        settled(); read.cancel(); setOpen(false); setCreating(false); setDraft(initial()); onBlocked(false); onApplied(result);
      }, failed);
    }
  }
  function create(event: FormEvent) { event.preventDefault(); send({ kind: "create", key: crypto.randomUUID(), body: structuredClone(draft) }); }
  const snapshot = items.find(item => item.id === selected);
  return <>
    <Button ref={trigger} variant="outline" disabled={disabled || open} onClick={() => { if (!guard.canWrite()) return; onBlocked(true); setOpen(true); load(); }}>物件範本</Button>
    {open && <PagePanel heading="物件範本"><section ref={panel} tabIndex={-1} aria-label="物件範本操作" className="flex min-w-0 flex-col gap-3">
      <ReadFeedback busy={reading || busy} message={message} label="正在讀取範本……" />
      {!creating && <>
        <FormField>選擇範本<NativeSelect value={selected} disabled={busy || uncertain} onChange={event => setSelected(event.target.value)}>
          <NativeSelectOption value="">請選擇</NativeSelectOption>
          {items.map(item => <NativeSelectOption key={item.id} value={item.id}>{item.name} · v{item.version}</NativeSelectOption>)}
        </NativeSelect></FormField>
        {snapshot && <ul aria-label="範本節點">{snapshot.nodes.map(node => <li key={node.id} className="[overflow-wrap:anywhere]">{node.name}{node.parentId ? `（父節點：${snapshot.nodes.find(parent => parent.id === node.parentId)?.name || node.parentId}）` : "（根節點）"}</li>)}</ul>}
        {!reading && !items.length && <p>尚無範本，請先建立。</p>}
        <div className="flex flex-wrap gap-2">
          <Button disabled={!snapshot || busy || uncertain} onClick={() => send({ kind: "apply", key: crypto.randomUUID(), id: selected, body: {} })}>套用範本</Button>
          <Button variant="outline" disabled={busy || uncertain || reading} onClick={() => setCreating(true)}>新增範本</Button>
          <Button variant="outline" disabled={busy || uncertain || reading} onClick={() => load()}>重新讀取範本</Button>
          {cursor && <Button variant="outline" disabled={busy || uncertain || reading} onClick={() => load(cursor)}>載入更多範本</Button>}
        </div>
      </>}
      {creating && <form onSubmit={create}><FieldGroup><FieldSet disabled={busy || uncertain}>
        <FormField>範本名稱<Input required maxLength={120} value={draft.name} onChange={event => setDraft(previous => ({ ...previous, name: event.target.value }))} /></FormField>
        {draft.nodes.map((node, index) => <FieldGroup key={node.id}>
          <FormField>{index === 0 ? "根物件名稱" : `節點 ${index + 1} 名稱`}<Input required maxLength={120} value={node.name} onChange={event => setDraft(previous => ({ ...previous, nodes: previous.nodes.map(n => n.id === node.id ? { ...n, name: event.target.value } : n) }))} /></FormField>
          {index > 0 && <FormField>{`節點 ${index + 1} 的父節點`}<NativeSelect value={node.parentId} onChange={event => setDraft(previous => ({ ...previous, nodes: previous.nodes.map(n => n.id === node.id ? { ...n, parentId: event.target.value } : n) }))}>
            {draft.nodes.filter(n => n.id !== node.id).map(n => <NativeSelectOption key={n.id} value={n.id}>{n.name || "未命名節點"}</NativeSelectOption>)}
          </NativeSelect></FormField>}
        </FieldGroup>)}
        <Button type="button" variant="outline" disabled={draft.nodes.length >= 100} onClick={() => setDraft(previous => ({ ...previous, nodes: [...previous.nodes, { id: crypto.randomUUID(), parentId: "root", name: "" }] }))}>新增節點</Button>
        <Button type="submit">建立範本</Button>
      </FieldSet></FieldGroup></form>}
      {uncertain && !busy && <Button onClick={() => { if (attempt.current) send(attempt.current); }}>重試原範本請求</Button>}
      <Button variant="outline" disabled={busy || uncertain} onClick={close}>關閉範本</Button>
    </section></PagePanel>}
    {discard.dialog}
  </>;
}
