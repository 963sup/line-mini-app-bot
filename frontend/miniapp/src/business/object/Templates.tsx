import React, { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import { useRepositories } from "../repository/RepositoryContext";
import { Candidates, type Candidate } from "./Candidates";
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
import { Textarea } from "@/shared/ui/textarea";
import { FieldGroup, FieldSet } from "@/shared/ui/field";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";

type Snapshot = components["schemas"]["ObjectTemplate"];
type Creation = components["schemas"]["ObjectTemplateCreation"];
type Props = { standalone?: boolean; initiallyOpen?: boolean; repositoryId: string; disabled: boolean; onBlocked: (blocked: boolean) => void; onApplied: (result: components["schemas"]["ObjectTemplateInstance"]) => void };
type Attempt = { key: string } & ({ kind: "create"; body: Creation } | { kind: "apply"; id: string; repositoryId: string; body: components["schemas"]["ObjectTemplateInstantiation"] });
const initial = (): Creation => ({ name: "", nodes: [{ id: "root", name: "" }] });

function documentDraft(filename: string, candidates: Candidate[]): Creation {
  const name = filename.replace(/\.[^.]+$/, "").trim() || filename;
  const nodes: Creation["nodes"] = [{ id: "root", name }];
  const headings = new Map<string, string>();
  for (const candidate of candidates) {
    let parentId = "root";
    for (const heading of candidate.categoryPath ?? []) {
      const name = heading.trim();
      if (!name) continue;
      const key = JSON.stringify([parentId, name]);
      let id = headings.get(key);
      if (!id) {
        id = crypto.randomUUID();
        headings.set(key, id);
        nodes.push({ id, parentId, name });
      }
      parentId = id;
    }
    nodes.push({ id: crypto.randomUUID(), parentId, name: candidate.name, unit: candidate.unit });
  }
  return { name, nodes };
}

function removeDraftBranch(nodes: Creation["nodes"], id: string) {
  const removed = new Set([id]);
  while (true) {
    const before = removed.size;
    for (const node of nodes) if (node.parentId && removed.has(node.parentId)) removed.add(node.id);
    if (before === removed.size) return nodes.filter(node => !removed.has(node.id));
  }
}

// Only input, request lifecycle and navigation live here. Go validates and creates the tree.
export function Templates({ repositoryId, disabled, onBlocked, onApplied, initiallyOpen = false, standalone = false }: Props) {
  const { request } = useSession();
  const read = useLatestRequest(), write = useLatestRequest();
  const [open, setOpen] = useState(false), [creating, setCreating] = useState(false);
  const [draft, setDraft] = useState<Creation>(initial), [items, setItems] = useState<Snapshot[]>([]);
  const [scope, setScope] = useState<"repository" | "personal">(standalone ? "personal" : "repository");
  const [targetRepository, setTargetRepository] = useState(repositoryId);
  const [plan, setPlan] = useState<Creation["nodes"]>();
  const [excluded, setExcluded] = useState<string[]>([]);
  const [selected, setSelected] = useState(""), [cursor, setCursor] = useState("");
  const [busy, setBusy] = useState(false), [reading, setReading] = useState(false), [message, setMessage] = useState("");
  const [uncertain, setUncertain] = useState(false);
  const attempt = useRef<Attempt | null>(null), lock = useRef(false), panel = useRef<HTMLElement>(null), trigger = useRef<HTMLButtonElement>(null);
  const focus = () => panel.current?.focus();
  const dirty = !!plan || (creating && JSON.stringify(draft) !== JSON.stringify(initial()));
  const guard = useRepositoryGuard(dirty, focus), discard = useDiscardChanges(dirty, focus);
  useEffect(() => { onBlocked(open); return () => onBlocked(false); }, [open, onBlocked]);
  useEffect(() => { guard.block(busy || uncertain); }, [busy, uncertain, guard.block]);
  useEffect(() => { if (open) focus(); }, [open]);
  const load = useCallback((next = "", library = scope) => {
    setReading(true); setMessage("");
    if (!next) { setItems([]); setSelected(""); }
    void read.run(signal => request((client, s) => standalone ? client.GET("/api/v1/object-templates", { signal: s, params: { query: { cursor: next || undefined } } }) : client.GET("/api/v1/repositories/{repositoryId}/object-templates", { signal: s, params: { path: { repositoryId }, query: { cursor: next || undefined, scope: library === "personal" ? "personal" : undefined } } }), { signal }), page => {
      setItems(previous => next ? [...previous, ...page.items] : page.items); setCursor(page.nextCursor); setReading(false);
    }, error => { setReading(false); setMessage(errorMessage(error)); });
  }, [read.run, request, repositoryId, scope, standalone]);
  const openedFromEntry = useRef(false);
  useEffect(() => {
    if (!initiallyOpen || disabled || openedFromEntry.current) return;
    openedFromEntry.current = true;
    setOpen(true);
    load();
  }, [initiallyOpen, disabled, load]);
  function close() {
    if (lock.current || uncertain) return;
    discard.navigate(() => { read.cancel(); onBlocked(false); setOpen(false); setPlan(undefined); setExcluded([]); setCreating(false); setDraft(initial()); setMessage(""); trigger.current?.focus(); });
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
      void write.run(signal => request((client, s) => standalone ? client.POST("/api/v1/object-templates", { signal: s, params: { header: { "Idempotency-Key": pending.key } }, body: pending.body }) : client.POST("/api/v1/repositories/{repositoryId}/object-templates", { signal: s, params: { path: { repositoryId }, query: { scope: scope === "personal" ? "personal" : undefined }, header: { "Idempotency-Key": pending.key } }, body: pending.body }), { signal }), snapshot => {
        settled(); setItems(previous => [snapshot, ...previous.filter(item => item.id !== snapshot.id)]); setSelected(snapshot.id); setCreating(false); setDraft(initial()); setMessage("已建立範本，可直接套用。");
      }, failed);
    } else {
      void write.run(signal => request((client, s) => client.POST("/api/v1/repositories/{repositoryId}/object-templates/{id}/instances", { signal: s, params: { path: { repositoryId: pending.repositoryId, id: pending.id }, query: { scope: scope === "personal" ? "personal" : undefined }, header: { "Idempotency-Key": pending.key } }, body: pending.body }), { signal }), result => {
        settled(); read.cancel(); setMessage(`已建立 ${result.objectIds.length} 個物件。`); setOpen(standalone); setPlan(undefined); setExcluded([]); setCreating(false); setDraft(initial()); onBlocked(false); onApplied(result);
      }, failed);
    }
  }
  function create(event: FormEvent) { event.preventDefault(); send({ kind: "create", key: crypto.randomUUID(), body: structuredClone(draft) }); }
  const nameTooLong = (name: string) => Array.from(name.trim()).length > 120;
  const draftTooLong = nameTooLong(draft.name) || draft.nodes.some(node => nameTooLong(node.name));
  const NodeName = scope === "personal" ? Textarea : Input;
  const snapshot = items.find(item => item.id === selected);
  const byID = new Map(plan?.map(node => [node.id, node]));
  function isRemoved(id: string): boolean {
    for (let node = byID.get(id); node; node = node.parentId ? byID.get(node.parentId) : undefined) {
      if (excluded.includes(node.id)) return true;
    }
    return false;
  }
  function descendantsOf(id: string) {
    return plan?.filter(node => {
      for (let parent = node.parentId; parent; parent = byID.get(parent)?.parentId) {
        if (parent === id) return true;
      }
      return false;
    }) ?? [];
  }
  const retained = plan?.filter(node => !isRemoved(node.id)) ?? [];
  function editNode(id: string, patch: Partial<Creation["nodes"][number]>) {
    setPlan(previous => previous?.map(node => node.id === id ? { ...node, ...patch } : node));
  }
  return <>
    {!standalone && <Button ref={trigger} variant="outline" disabled={disabled || open} onClick={() => { if (!guard.canWrite()) return; onBlocked(true); setOpen(true); load(); }}>物件範本</Button>}
    {open && <PagePanel heading="物件範本"><section ref={panel} tabIndex={-1} aria-label="物件範本操作" className="flex min-w-0 flex-col gap-3">
      <ReadFeedback busy={reading || busy} message={message} label="正在讀取範本……" />
      {!standalone && !plan && scope === "personal" && <Candidates repositoryId={repositoryId} disabled={busy || uncertain || creating} importDraft={(filename, candidates) => { setDraft(documentDraft(filename, candidates)); setCreating(true); }} />}
      {!creating && !plan && <>
        {!standalone && <FormField>範本庫<NativeSelect value={scope} disabled={busy || uncertain} onChange={event => {
          const library = event.target.value === "personal" ? "personal" : "repository";
          setScope(library); load("", library);
        }}>
          <NativeSelectOption value="repository">本庫</NativeSelectOption>
          <NativeSelectOption value="personal">我的範本</NativeSelectOption>
        </NativeSelect></FormField>}
        <FormField>選擇範本<NativeSelect value={selected} disabled={busy || uncertain} onChange={event => setSelected(event.target.value)}>
          <NativeSelectOption value="">請選擇</NativeSelectOption>
          {items.map(item => <NativeSelectOption key={item.id} value={item.id}>{item.name} · v{item.version}</NativeSelectOption>)}
        </NativeSelect></FormField>
        {snapshot && <ul aria-label="範本節點">{snapshot.nodes.map(node => <li key={node.id} className="[overflow-wrap:anywhere]">{node.name}{node.parentId ? `（父節點：${snapshot.nodes.find(parent => parent.id === node.parentId)?.name || node.parentId}）` : "（根節點）"}</li>)}</ul>}
        {!reading && !items.length && <p>尚無範本，請先建立。</p>}
        <div className="flex flex-wrap gap-2">
          <Button disabled={!snapshot || busy || uncertain} onClick={() => { if (snapshot) setPlan(structuredClone(snapshot.nodes)); }}>使用範本</Button>
          <Button variant="outline" disabled={busy || uncertain || reading} onClick={() => setCreating(true)}>新增範本</Button>
          <Button variant="outline" disabled={busy || uncertain || reading} onClick={() => load()}>重新讀取範本</Button>
          {cursor && <Button variant="outline" disabled={busy || uncertain || reading} onClick={() => load(cursor)}>載入更多範本</Button>}
        </div>
      </>}
      {plan && <form onSubmit={event => {
        event.preventDefault();
        if (!targetRepository) return;
        send({ kind: "apply", repositoryId: targetRepository, key: crypto.randomUUID(), id: selected, body: { nodes: retained } });
      }}><FieldGroup><FieldSet disabled={busy || uncertain}>
        {standalone && <TemplateTarget value={targetRepository} onChange={setTargetRepository} disabled={busy || uncertain} />}
        {plan.map(node => {
          const removed = isRemoved(node.id);
          const descendants = descendantsOf(node.id);
          return <FieldGroup key={node.id}>
            <FormField>物件名稱<NodeName required maxLength={120} disabled={removed} value={node.name} onChange={event => editNode(node.id, { name: event.target.value })} /></FormField>
            {node.parentId && <p>父物件：{byID.get(node.parentId)?.name}</p>}
            <div className="grid grid-cols-2 gap-2">
              <FormField>數量<Input inputMode="decimal" disabled={removed} value={node.quantity ?? ""} onChange={event => editNode(node.id, { quantity: event.target.value })} /></FormField>
              <FormField>單位<Input maxLength={32} disabled={removed} value={node.unit ?? ""} onChange={event => editNode(node.id, { unit: event.target.value })} /></FormField>
            </div>
            {node.parentId && !removed && descendants.length > 0 && <p>移除時一併移除：{descendants.map(child => child.name).join("、")}</p>}
            {node.parentId && (!removed || excluded.includes(node.id)) && <Button type="button" variant="outline" onClick={() => setExcluded(previous => removed ? previous.filter(id => id !== node.id) : [...previous, node.id])}>{removed ? "復原工序" : "移除工序（含子工序）"}</Button>}
            {removed && <p>本次不建立</p>}
          </FieldGroup>;
        })}
        <Button type="submit" disabled={!targetRepository}>建立 {retained.length} 個物件</Button>
        <Button type="button" variant="outline" onClick={() => discard.navigate(() => { setPlan(undefined); setExcluded([]); })}>返回範本</Button>
      </FieldSet></FieldGroup></form>}
      {creating && draft.nodes.length > 100 && <p role="alert">範本最多 100 個節點（含根物件），請移除不需要的工序後再儲存。</p>}
      {creating && <form onSubmit={create}><FieldGroup><FieldSet disabled={busy || uncertain}>
        <FormField>範本名稱<Input required maxLength={120} value={draft.name} onChange={event => setDraft(previous => ({ ...previous, name: event.target.value }))} /></FormField>
        {nameTooLong(draft.name) && <p role="alert">範本名稱超過 120 字。</p>}
        {draft.nodes.map((node, index) => <FieldGroup key={node.id}>
          <FormField>{index === 0 ? "根物件名稱" : `節點 ${index + 1} 名稱`}<NodeName required maxLength={120} value={node.name} onChange={event => setDraft(previous => ({ ...previous, nodes: previous.nodes.map(n => n.id === node.id ? { ...n, name: event.target.value } : n) }))} /></FormField>
          {nameTooLong(node.name) && <p role="alert">此工項名稱超過 120 字，請先修正。</p>}
          {scope === "personal" && <FormField>{`節點 ${index + 1} 單位`}<Input maxLength={32} value={node.unit ?? ""} onChange={event => setDraft(previous => ({ ...previous, nodes: previous.nodes.map(n => n.id === node.id ? { ...n, unit: event.target.value } : n) }))} /></FormField>}
          {index > 0 && <FormField>{`節點 ${index + 1} 的父節點`}<NativeSelect value={node.parentId} onChange={event => setDraft(previous => ({ ...previous, nodes: previous.nodes.map(n => n.id === node.id ? { ...n, parentId: event.target.value } : n) }))}>
            {draft.nodes.filter(n => n.id !== node.id).map(n => <NativeSelectOption key={n.id} value={n.id}>{n.name || "未命名節點"}</NativeSelectOption>)}
          </NativeSelect></FormField>}
          {scope === "personal" && index > 0 && <Button type="button" variant="outline" onClick={() => setDraft(previous => ({ ...previous, nodes: removeDraftBranch(previous.nodes, node.id) }))}>移除工序（含子工序）</Button>}
        </FieldGroup>)}
        <Button type="button" variant="outline" disabled={draft.nodes.length >= 100} onClick={() => setDraft(previous => ({ ...previous, nodes: [...previous.nodes, { id: crypto.randomUUID(), parentId: "root", name: "" }] }))}>新增節點</Button>
        <Button type="submit" disabled={draft.nodes.length > 100 || draftTooLong}>建立範本</Button>
      </FieldSet></FieldGroup></form>}
      {uncertain && !busy && <Button onClick={() => { if (attempt.current) send(attempt.current); }}>重試原範本請求</Button>}
      {!standalone && <Button variant="outline" disabled={busy || uncertain} onClick={close}>關閉範本</Button>}
    </section></PagePanel>}
    {discard.dialog}
  </>;
}

function TemplateTarget({ value, onChange, disabled }: { value: string; onChange: (id: string) => void; disabled: boolean }) {
  const repositories = useRepositories();
  useEffect(() => { repositories.ensure(); }, []);
  return <>
    <FormField>目標儲存庫<NativeSelect required value={value} disabled={disabled} onChange={event => onChange(event.target.value)}>
      <NativeSelectOption value="">請選擇要建立物件的儲存庫</NativeSelectOption>
      {repositories.items.map(item => <NativeSelectOption key={item.id} value={item.id}>{item.name || "群組空間"}</NativeSelectOption>)}
    </NativeSelect></FormField>
    <ReadFeedback busy={repositories.busy} message={repositories.status} label="正在讀取儲存庫……" />
    {repositories.catalogFailed && <Button type="button" variant="outline" disabled={disabled || repositories.busy} onClick={repositories.retryCatalog}>重試讀取儲存庫</Button>}
    {repositories.cursor && <Button type="button" variant="outline" disabled={disabled || repositories.busy} onClick={repositories.loadMore}>載入更多儲存庫</Button>}
  </>;
}
