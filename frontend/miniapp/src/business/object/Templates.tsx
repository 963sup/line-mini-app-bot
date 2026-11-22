import { ChevronDown } from "lucide-react";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/shared/ui/collapsible";
import { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import { useRepositories } from "../repository/RepositoryContext";
import { Evidence, TemplatePreparation, type Review } from "./TemplatePreparation";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { rejectedAttempt, useLatestRequest, errorMessage } from "@/shared/http/lifecycle";
import { useRepositoryGuard } from "../repository/NavigationGuard";
import { useDiscardChanges } from "@/shared/useDiscardChanges";
import { FormField } from "@/shared/FormField";
import { PagePanel } from "@/shared/PagePanel";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Textarea } from "@/shared/ui/textarea";
import { FieldGroup, FieldSet, FieldLegend } from "@/shared/ui/field";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from "@/shared/ui/empty";
import { Item, ItemContent, ItemGroup, ItemTitle } from "@/shared/ui/item";
import { Badge } from "@/shared/ui/badge";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";

type Snapshot = components["schemas"]["ObjectTemplate"];
type Creation = components["schemas"]["ObjectTemplateCreation"];
type Props = { standalone?: boolean; initiallyOpen?: boolean; repositoryId: string; disabled: boolean; onBlocked: (blocked: boolean) => void; onApplied: (result: components["schemas"]["ObjectTemplateInstance"], repositoryId: string) => void };
type Attempt = { key: string } & ({ kind: "create"; body: Creation } | { kind: "apply"; id: string; repositoryId: string; body: components["schemas"]["ObjectTemplateInstantiation"] });
const initial = (): Creation => ({ name: "", nodes: [{ id: "root", name: "" }] });

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
  const { request, user } = useSession();
  const read = useLatestRequest(), write = useLatestRequest();
  const [open, setOpen] = useState(false), [creating, setCreating] = useState(false);
  const [draft, setDraft] = useState<Creation>(initial), [items, setItems] = useState<Snapshot[]>([]);
  const [method, setMethod] = useState<"choose" | Review["source"]>();
  const [review, setReview] = useState<Review>();
  const [folded, setFolded] = useState<string[]>([]);
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
    if (!user || !initiallyOpen || disabled || openedFromEntry.current) return;
    openedFromEntry.current = true;
    setOpen(true);
    load();
  }, [user, initiallyOpen, disabled, load]);
  function close() {
    if (lock.current || uncertain) return;
    discard.navigate(() => { read.cancel(); onBlocked(false); setOpen(false); setMethod(undefined); setPlan(undefined); setExcluded([]); setCreating(false); setDraft(initial()); setReview(undefined); setFolded([]); setMessage(""); trigger.current?.focus(); });
  }
  function send(pending: Attempt) {
    if (lock.current || discard.isConfirming() || !guard.canWrite()) return;
    attempt.current = pending;
    const wasUnknown = uncertain;
    lock.current = true; setBusy(true); setUncertain(true); guard.block(true); setMessage("正在處理範本……");
    const failed = (error: unknown) => {
      lock.current = false; setBusy(false);
      if (!wasUnknown && rejectedAttempt(error)) { setUncertain(false); guard.block(false); attempt.current = null; }
      setMessage(errorMessage(error) + " 結果未確認時，請重試原請求。");
    };
    const settled = () => { lock.current = false; setBusy(false); setUncertain(false); guard.block(false); attempt.current = null; };
    if (pending.kind === "create") {
      void write.run(signal => request((client, s) => standalone ? client.POST("/api/v1/object-templates", { signal: s, params: { header: { "Idempotency-Key": pending.key } }, body: pending.body }) : client.POST("/api/v1/repositories/{repositoryId}/object-templates", { signal: s, params: { path: { repositoryId }, query: { scope: scope === "personal" ? "personal" : undefined }, header: { "Idempotency-Key": pending.key } }, body: pending.body }), { signal }), snapshot => {
        settled(); setItems(previous => [snapshot, ...previous.filter(item => item.id !== snapshot.id)]); setSelected(snapshot.id); setCreating(false); setDraft(initial()); setReview(undefined); setFolded([]); setMessage(pending.body.templateId ? `已儲存為 v${snapshot.version}。` : "已建立範本，可直接套用。");
      }, failed);
    } else {
      void write.run(signal => request((client, s) => client.POST("/api/v1/repositories/{repositoryId}/object-templates/{id}/instances", { signal: s, params: { path: { repositoryId: pending.repositoryId, id: pending.id }, query: { scope: scope === "personal" ? "personal" : undefined }, header: { "Idempotency-Key": pending.key } }, body: pending.body }), { signal }), result => {
        settled(); read.cancel(); setMessage(`已建立 ${result.objectIds.length} 個物件。`); setOpen(standalone); setPlan(undefined); setExcluded([]); setCreating(false); setDraft(initial()); setReview(undefined); setFolded([]); onBlocked(false); onApplied(result, pending.repositoryId);
      }, failed);
    }
  }
  function create(event: FormEvent) {
    event.preventDefault();
    const object = draft.nodes.find(node => !node.parentId);
    if (!object?.name.trim()) return;
    send({ kind: "create", key: crypto.randomUUID(), body: structuredClone({ ...draft, name: object.name }) });
  }
  const nameTooLong = (name: string) => Array.from(name.trim()).length > 120;
  const draftTooLong = draft.nodes.some(node => nameTooLong(node.name));
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
  const parents = new Set(review?.result.draft.nodes.map(node => node.parentId));
  const leaves = review?.result.draft.nodes.filter(node => node.parentId && !parents.has(node.id)) ?? [];
  const evidence = new Map(leaves.map((node, i) => [node.id, review?.result.candidates[i]]));
  const draftById = new Map(draft.nodes.map(node => [node.id, node]));
  return <>
    {!standalone && <Button ref={trigger} variant="outline" disabled={disabled || open} onClick={() => { if (!guard.canWrite()) return; onBlocked(true); setOpen(true); load(); }}>物件範本</Button>}
    {open && <PagePanel heading={creating ? "編輯範本" : plan ? "確認本次建立的物件" : method === "choose" ? "如何建立範本？" : method === "document" ? "從文件建立範本" : method === "procedure" ? "從工項推導工序" : standalone ? "我的物件範本" : "物件範本"} description={plan ? snapshot?.name : undefined}><section ref={panel} tabIndex={-1} aria-label="物件範本操作" className="flex min-w-0 flex-col gap-4 outline-none">
      <ReadFeedback busy={reading || busy} message={message} label="正在讀取範本……" />
      {method === "choose" && <div className="grid gap-3">
        <Button onClick={() => { setMethod(undefined); setCreating(true); }}>從空白建立</Button>
        <Button variant="outline" onClick={() => setMethod("document")}>從文件建立</Button>
        <Button variant="outline" onClick={() => setMethod("procedure")}>從工項推導工序</Button>
      </div>}
      {(method === "procedure" || method === "document") && <TemplatePreparation key={method} source={method} disabled={busy || uncertain} onDraft={review => { setReview(review); setDraft(review.result.draft); setMethod(undefined); setCreating(true); }} />}
      {method && <Button variant="outline" onClick={() => setMethod(method === "choose" ? undefined : "choose")}>{method === "choose" ? "返回範本庫" : "返回建立方式"}</Button>}
      {!method && !creating && !plan && <>
        {!standalone && <FormField>範本庫<NativeSelect value={scope} disabled={busy || uncertain} onChange={event => {
          const library = event.target.value === "personal" ? "personal" : "repository";
          setScope(library); load("", library);
        }}>
          <NativeSelectOption value="repository">本庫</NativeSelectOption>
          <NativeSelectOption value="personal">我的範本</NativeSelectOption>
        </NativeSelect></FormField>}
        {items.length > 0 && <FormField>選擇範本<NativeSelect value={selected} disabled={busy || uncertain} onChange={event => setSelected(event.target.value)}>
          <NativeSelectOption value="">請選擇</NativeSelectOption>
          {items.map(item => <NativeSelectOption key={item.id} value={item.id}>{item.name} · v{item.version}</NativeSelectOption>)}
        </NativeSelect></FormField>}
        {!reading && !busy && !message && !items.length && <Empty><EmptyHeader><EmptyTitle>尚無範本，請先建立。</EmptyTitle><EmptyDescription>選擇「新增範本」，整理可重複使用的物件與工序。</EmptyDescription></EmptyHeader></Empty>}
        <div className="flex flex-wrap gap-2">
          {snapshot && <Button disabled={busy || uncertain} onClick={() => { if (snapshot) setPlan(structuredClone(snapshot.nodes)); }}>使用範本</Button>}
          {standalone && snapshot && <Button variant="outline" disabled={busy || uncertain || reading} onClick={() => {
            setDraft({ templateId: snapshot.id, baseVersion: snapshot.version, name: snapshot.name, nodes: structuredClone(snapshot.nodes) });
            setCreating(true);
          }}>編輯範本</Button>}
          <Button variant="outline" disabled={busy || uncertain || reading} onClick={() => scope === "personal" ? setMethod("choose") : setCreating(true)}>新增範本</Button>
          {cursor && <Button variant="outline" disabled={busy || uncertain || reading} onClick={() => load(cursor)}>載入更多範本</Button>}
        </div>
        {snapshot && <Collapsible><CollapsibleTrigger asChild><Button variant="ghost">預覽 {snapshot.nodes.length} 個物件 <ChevronDown /></Button></CollapsibleTrigger><CollapsibleContent><ItemGroup aria-label="範本節點" className="gap-3">{snapshot.nodes.map(node => <Item key={node.id} role="listitem" variant="outline">
          <ItemContent className="min-w-0">
            <ItemTitle className="max-w-full flex-wrap [overflow-wrap:anywhere]">{node.name}{!node.parentId && <Badge variant="secondary">物件</Badge>}</ItemTitle>
            {node.parentId && <p className="text-sm text-muted-foreground [overflow-wrap:anywhere]">父節點：{snapshot.nodes.find(parent => parent.id === node.parentId)?.name || node.parentId}</p>}
          </ItemContent>
        </Item>)}</ItemGroup></CollapsibleContent></Collapsible>}
      </>}
      {plan && <form onSubmit={event => {
        event.preventDefault();
        if (!targetRepository) return;
        send({ kind: "apply", repositoryId: targetRepository, key: crypto.randomUUID(), id: selected, body: { nodes: retained, ...(snapshot && snapshot.version > 1 ? { templateVersion: snapshot.version } : {}) } });
      }}><FieldGroup><FieldSet disabled={busy || uncertain}>
        {standalone && <TemplateTarget value={targetRepository} onChange={setTargetRepository} disabled={busy || uncertain} />}
        {plan.map(node => {
          const removed = isRemoved(node.id);
          const descendants = descendantsOf(node.id);
          return <Collapsible key={node.id} className="border-b">
            <CollapsibleTrigger asChild><Button type="button" variant="ghost" className="h-auto min-h-11 w-full justify-between gap-3 whitespace-normal text-start">
              <span>{node.name || "未命名物件"}</span><span className="flex shrink-0 items-center gap-2 text-muted-foreground">{removed ? "本次不建立" : [node.quantity, node.unit].filter(Boolean).join(" ")}<ChevronDown aria-hidden="true" /></span>
            </Button></CollapsibleTrigger>
            <CollapsibleContent><FieldSet className="px-2 pb-4">
            <FieldLegend>{node.parentId ? "工序" : "物件"}</FieldLegend>
            <FormField>物件名稱<NodeName required maxLength={120} disabled={removed} value={node.name} onChange={event => editNode(node.id, { name: event.target.value })} /></FormField>
            {node.parentId && <p>父物件：{byID.get(node.parentId)?.name}</p>}
            <div className="grid grid-cols-2 gap-2">
              <FormField>數量<Input inputMode="decimal" disabled={removed} value={node.quantity ?? ""} onChange={event => editNode(node.id, { quantity: event.target.value })} /></FormField>
              <FormField>單位<Input maxLength={32} disabled={removed} value={node.unit ?? ""} onChange={event => editNode(node.id, { unit: event.target.value })} /></FormField>
            </div>
            {node.parentId && !removed && descendants.length > 0 && <p>移除時一併移除：{descendants.map(child => child.name).join("、")}</p>}
            {node.parentId && (!removed || excluded.includes(node.id)) && <Button type="button" variant="outline" onClick={() => setExcluded(previous => removed ? previous.filter(id => id !== node.id) : [...previous, node.id])}>{removed ? "復原工序" : "移除工序（含子工序）"}</Button>}
            {removed && <p>本次不建立</p>}
          </FieldSet></CollapsibleContent></Collapsible>;
        })}
        <div className="sticky bottom-0 z-10 flex flex-wrap items-center gap-2 border-t bg-background py-3"><Button type="submit" disabled={!targetRepository || retained.some(node => !node.name.trim())}>建立 {retained.length} 個物件</Button>
        <Button type="button" variant="outline" onClick={() => discard.navigate(() => { setPlan(undefined); setExcluded([]); if (standalone) load(); })}>返回範本</Button></div>
      </FieldSet></FieldGroup></form>}
      {creating && draft.nodes.length > 100 && <p role="alert">範本最多 100 個節點（含根物件），請移除不需要的工序後再儲存。</p>}
      {creating && <form onSubmit={create}><FieldGroup><FieldSet disabled={busy || uncertain}>
        {review?.result.warnings.map((warning, i) => <p role="status" key={i}>{warning}</p>)}
        {draft.nodes.map((node, index) => {
          let depth = 0;
          for (let parent = node.parentId; parent; parent = draftById.get(parent)?.parentId) {
            if (folded.includes(parent)) return null;
            depth++;
          }
          const candidate = evidence.get(node.id);
          return <FieldSet key={node.id} style={{ marginInlineStart: Math.min(depth, 4) * 12 }}>
          {draft.nodes.some(child => child.parentId === node.id) && <Button type="button" variant="outline" aria-expanded={!folded.includes(node.id)} onClick={() => setFolded(current => current.includes(node.id) ? current.filter(id => id !== node.id) : [...current, node.id])}>{folded.includes(node.id) ? "展開子工序" : "收合子工序"}</Button>}
          {candidate && review && <Evidence candidate={candidate} source={review.source} />}
          <FieldLegend>{!node.parentId ? "物件" : `節點 ${index + 1}`}</FieldLegend>
          <FormField>{!node.parentId ? "物件名稱" : `節點 ${index + 1} 名稱`}<NodeName required maxLength={120} value={node.name} onChange={event => setDraft(previous => ({ ...previous, nodes: previous.nodes.map(n => n.id === node.id ? { ...n, name: event.target.value } : n) }))} /></FormField>
          {nameTooLong(node.name) && <p role="alert">此工項名稱超過 120 字，請先修正。</p>}
          <Collapsible><CollapsibleTrigger asChild><Button type="button" variant="ghost" size="sm">工項設定 <ChevronDown aria-hidden="true" /></Button></CollapsibleTrigger><CollapsibleContent className="grid gap-3 py-2">
          {scope === "personal" && <FormField>{`節點 ${index + 1} 單位`}<Input maxLength={32} value={node.unit ?? ""} onChange={event => setDraft(previous => ({ ...previous, nodes: previous.nodes.map(n => n.id === node.id ? { ...n, unit: event.target.value } : n) }))} /></FormField>}
          {node.parentId && <FormField>{`節點 ${index + 1} 的父節點`}<NativeSelect value={node.parentId} onChange={event => setDraft(previous => ({ ...previous, nodes: previous.nodes.map(n => n.id === node.id ? { ...n, parentId: event.target.value } : n) }))}>
            {removeDraftBranch(draft.nodes, node.id).map(n => <NativeSelectOption key={n.id} value={n.id}>{n.name || "未命名節點"}</NativeSelectOption>)}
          </NativeSelect></FormField>}
          {scope === "personal" && node.parentId && <Button type="button" variant="outline" onClick={() => setDraft(previous => ({ ...previous, nodes: removeDraftBranch(previous.nodes, node.id) }))}>移除工序（含子工序）</Button>}
          </CollapsibleContent></Collapsible>
        </FieldSet>; })}
        <Button type="button" variant="outline" disabled={draft.nodes.length >= 100} onClick={() => setDraft(previous => ({ ...previous, nodes: [...previous.nodes, { id: crypto.randomUUID(), parentId: previous.nodes.find(node => !node.parentId)!.id, name: "" }] }))}>新增節點</Button>
        <div className="sticky bottom-0 z-10 flex flex-wrap items-center gap-2 border-t bg-background py-3"><Button type="submit" disabled={draft.nodes.length > 100 || draftTooLong || draft.nodes.some(node => !node.name.trim())}>{draft.baseVersion ? `儲存為 v${draft.baseVersion + 1}` : "建立範本"}</Button>
        <Button type="button" variant="outline" onClick={() => discard.navigate(() => { setCreating(false); setDraft(initial()); setReview(undefined); setFolded([]); if (draft.templateId) load(); })}>返回範本庫</Button></div>
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
