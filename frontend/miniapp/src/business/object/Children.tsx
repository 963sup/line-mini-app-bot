import { isPositiveQuantity, isValidCompletedQuantity } from "@/business/object/quantity";
import { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import { useRepositoryGuard } from "../repository/NavigationGuard";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { useDiscardChanges } from "@/shared/useDiscardChanges";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";
import { FormField } from "@/shared/FormField";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { Item, ItemGroup, ItemContent, ItemTitle, ItemActions, ItemFooter } from "@/shared/ui/item";
import { Input } from "@/shared/ui/input";
import { FieldGroup, FieldSet } from "@/shared/ui/field";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";

type Resource = components["schemas"]["Object"];
type Creation = { name: string; type: "work" | "non_work"; quantity: string; unit: string; completedQuantity?: string };
const initial: Creation = { name: "", type: "work", quantity: "1", unit: "SET", completedQuantity: "0" };
const states = { not_started: "未開始", in_progress: "進行中", completed: "已完成" };
type Props = { repositoryId: string; objectId: string; canEdit: boolean; parentEditing: boolean; onEdit: () => void; onBlocked: (value: boolean) => void; canWrite: () => boolean };

export function View(props: Props) {
  const { user, epoch } = useSession();
  return <Content key={`${props.repositoryId}:${props.objectId}:${user?.userId}:${epoch}`} {...props} />;
}
function Content({ repositoryId, objectId, canEdit, parentEditing, onEdit, onBlocked, canWrite }: Props) {
  const { request } = useSession(), read = useLatestRequest(), write = useLatestRequest();
  const [items, setItems] = useState<Resource[]>([]), [cursor, setCursor] = useState("");
  const [loading, setLoading] = useState(false), [readStatus, setReadStatus] = useState("");
  const [selected, setSelected] = useState<Resource | null>(null), [editing, setEditing] = useState(false), [content, setContent] = useState<Creation>(initial);
  const [busy, setBusy] = useState(false), [uncertain, setUncertain] = useState(false), [refresh, setRefresh] = useState(false), [status, setStatus] = useState("");
  const lock = useRef(false), unknown = useRef(false), navigation = useRef(false), pending = useRef<{ key: string; body: Creation } | null>(null);
  const baseline = useRef<Creation>(initial);
  const dirty = editing && JSON.stringify(content) !== JSON.stringify(baseline.current);
  const nameInput = useRef<HTMLInputElement>(null), editorPanel = useRef<HTMLFormElement>(null), opener = useRef<HTMLElement | null>(null);
  function focusEditor() {
    if (nameInput.current && !nameInput.current.matches(":disabled")) nameInput.current.focus();
    else editorPanel.current?.focus();
  }
  const discard = useDiscardChanges(dirty, focusEditor);
  const repositoryGuard = useRepositoryGuard(dirty, focusEditor);
  useEffect(() => { repositoryGuard.block(busy || uncertain); }, [busy, uncertain]);
  useEffect(() => {
    if (!editing) { opener.current?.focus(); return; }
    if (!busy) focusEditor();
  }, [editing, busy]);
  useEffect(() => {
    if (parentEditing) { opener.current = null; setEditing(false); setSelected(null); baseline.current = initial; setContent(initial); }
  }, [parentEditing]);
  const block = (value: boolean) => { repositoryGuard.block(value); unknown.current = value; setUncertain(value); onBlocked(value); };
  useEffect(() => { navigation.current = false; });
  useEffect(() => () => onBlocked(false), [onBlocked]);
  useEffect(() => { onBlocked(busy || uncertain || dirty); }, [busy, uncertain, dirty, onBlocked]);
  const load = useCallback((next = "") => {
    if (lock.current || unknown.current) return;
    setLoading(true); setReadStatus("");
    void read.run(signal => request((client, s) => client.GET("/api/v1/repositories/{repositoryId}/objects", {signal: s, params: {path: {repositoryId}, query: {parentId: objectId, cursor: next || undefined}}}), {signal}), page => {
      const children = page.items.filter(item => item.parentId === objectId);
      setItems(previous => next ? [...previous, ...children] : children); setCursor(page.nextCursor); setLoading(false); setReadStatus(!next && !page.nextCursor && !children.length ? "尚無子物件。" : "");
    }, error => { setLoading(false); setReadStatus(errorMessage(error)); });
  }, [repositoryId, objectId, request, read.run]);
  useEffect(() => { load(); return read.cancel; }, [load, read.cancel]);
  function close() {
    if (lock.current || unknown.current) return;
    discard.navigate(() => {
      navigation.current = true; setEditing(false); setSelected(null); baseline.current = initial; setContent(initial); setStatus("");
    });
  }
  function edit(item: Resource | null, trigger: HTMLElement) {
    if (!canEdit || (item && !item.canEdit) || lock.current || unknown.current || !repositoryGuard.canWrite() || !canWrite()) return;
    discard.navigate(() => {
      opener.current = trigger; onEdit(); read.cancel(); setLoading(false);
      navigation.current = true; setSelected(item); setEditing(true); setRefresh(false); setStatus(""); pending.current = null;
      baseline.current = item ? {name: item.name, type: item.type === "non_work" ? "non_work" : "work", quantity: item.quantity || "", unit: item.unit || "", completedQuantity: item.completedQuantity} : {...initial};
      setContent(baseline.current);
      if (item) reloadDetail(item);
    });
  }
  function reloadDetail(target = selected) {
    if (!target || lock.current || !repositoryGuard.canWrite() || !canWrite()) return;
    lock.current = true; setBusy(true); repositoryGuard.block(true); onBlocked(true);
    void write.run(signal => request((client,s) => client.GET("/api/v1/repositories/{repositoryId}/objects/{id}", {signal:s, params:{path:{repositoryId,id:target.id}}}), {signal}), item => {
      lock.current = false; setBusy(false); block(false); setSelected(item); baseline.current = {name:item.name,type:item.type === "non_work" ? "non_work" : "work",quantity:item.quantity || "",unit:item.unit || "",completedQuantity:item.completedQuantity}; setContent(baseline.current); setRefresh(false); setStatus(""); setItems(previous => previous.map(value => value.id === item.id ? item : value));
    }, error => {lock.current=false;setBusy(false);onBlocked(unknown.current);setRefresh(true);setStatus(errorMessage(error));});
  }
  function submit(event: FormEvent) {
    event.preventDefault();
    if (discard.isConfirming() || !canEdit || lock.current || navigation.current || refresh || !repositoryGuard.canWrite() || !canWrite()) return;
    let body = pending.current && unknown.current ? pending.current.body : {...content, name:content.name.trim(), quantity:content.quantity.trim(), unit:content.unit.trim(), completedQuantity:content.completedQuantity?.trim()};
    if (!body.name || !body.unit || !isPositiveQuantity(body.quantity)) {setStatus("請填寫名稱、單位及大於零的數量（最多18位整數及6位小數）。");return;}
    if (body.type === "work") {
      if (!isValidCompletedQuantity(body.completedQuantity, body.quantity)) {setStatus("完成數須介於0與子物件數量之間（最多18位整數及6位小數）。");return;}
    } else {body = {...body}; delete body.completedQuantity;}
    if (selected?.type === "work" && Number(selected.completedQuantity) > 0 && body.type === "non_work") {setStatus("請先將完成數改為0並儲存，再改為非施作。");return;}
    const wasUnknown = unknown.current;
    lock.current = true; setBusy(true); block(true); setStatus(selected ? "正在更新子物件……" : "正在建立子物件……");
    if (!selected && !pending.current) pending.current = {key:crypto.randomUUID(),body};
    void write.run(signal => selected ? request((client,s) => client.PATCH("/api/v1/repositories/{repositoryId}/objects/{id}", {signal:s,params:{path:{repositoryId,id:selected.id}},body:{...body,description:selected.description || "",expectedVersion:selected.version}}),{signal}) : request((client,s) => client.POST("/api/v1/repositories/{repositoryId}/objects", {signal:s,params:{path:{repositoryId},header:{"Idempotency-Key":pending.current!.key}},body:{...pending.current!.body,parentId:objectId}}),{signal}), item => {
      lock.current=false;setBusy(false);block(false);pending.current=null;setEditing(false);setSelected(null);baseline.current=initial;setContent(initial);setItems(previous=>previous.some(value=>value.id===item.id)?previous.map(value=>value.id===item.id?item:value):[...previous,item]);setReadStatus("已儲存子物件；父物件進度不變。");
    }, error => {
      lock.current=false;setBusy(false);
      const rejected = ["invalid_input","forbidden","not_found","idempotency_conflict","version_conflict"].includes((error as {code?:string}).code || "");
      if (selected) {setRefresh(true); if(rejected) block(false); setStatus(errorMessage(error)+" 請重新讀取子物件以確認更新結果。");}
      else {if(rejected && !wasUnknown) {block(false);pending.current=null;} setStatus(errorMessage(error)+" 結果未確認時，請重試原子物件建立請求。");}
    });
  }
  const editor = <form ref={editorPanel} tabIndex={-1} className="w-full min-w-0" aria-label="子物件編輯" onSubmit={submit}><FieldGroup>
      <Status role="status">{status}</Status>
      {selected && <Button type="button" variant="outline" disabled={busy} onClick={() => discard.navigate(reloadDetail)}>重新讀取子物件</Button>}
      <FieldSet className="grid min-w-0 gap-3" disabled={busy || uncertain || refresh}>
        <FormField>子物件名稱<Input ref={nameInput} required maxLength={120} value={content.name} onChange={event => setContent(previous=>({...previous,name:event.target.value}))}/></FormField>
        <FormField>子物件類型<NativeSelect value={content.type} onChange={event=>setContent(previous=>({...previous,type:event.target.value as Creation["type"],completedQuantity:previous.completedQuantity ?? (event.target.value === "work" ? "0" : undefined)}))}><NativeSelectOption value="work">施作</NativeSelectOption><NativeSelectOption value="non_work">非施作</NativeSelectOption></NativeSelect></FormField>
        <FormField>子物件數量<Input required maxLength={25} inputMode="decimal" value={content.quantity} onChange={event=>setContent(previous=>({...previous,quantity:event.target.value}))}/></FormField>
        <FormField>子物件單位<Input required maxLength={32} value={content.unit} onChange={event=>setContent(previous=>({...previous,unit:event.target.value}))}/></FormField>
        {content.type === "work" ? <FormField>子物件完成數<Input required maxLength={25} inputMode="decimal" value={content.completedQuantity || ""} onChange={event=>setContent(previous=>({...previous,completedQuantity:event.target.value}))}/></FormField> : <p>狀態與進度：不適用</p>}
      </FieldSet>
      <Button type="submit" disabled={busy || refresh}>{selected ? "儲存子物件" : uncertain ? "重試原子物件建立請求" : "建立子物件"}</Button>
      <Button type="button" variant="outline" disabled={busy || uncertain} onClick={close}>取消</Button>
    </FieldGroup></form>;
  return <section aria-label="子物件" className="grid min-w-0 gap-3">
    <p>子物件數量與單位獨立；父物件進度仍由人工維護，不會加總子物件完成數。</p>
    <ReadFeedback busy={loading} message={readStatus} label="正在讀取子物件……" />
    <Button variant="outline" disabled={loading || busy || uncertain || editing} onClick={() => load()}>重新讀取子物件清單</Button>
    <ItemGroup className="min-w-0 gap-3">{items.map(item => <Item role="listitem" variant="outline" key={item.id} className="min-w-0">
      <ItemContent className="min-w-0">
        <ItemTitle className="min-w-0 max-w-full">{item.canEdit && canEdit ? <Button variant="link" aria-label={item.name} className="h-auto min-w-0 max-w-full whitespace-normal text-left [overflow-wrap:anywhere]" disabled={busy || uncertain} onClick={event => edit(item, event.currentTarget)}>{item.name}</Button> : item.name}</ItemTitle>
      <p className="whitespace-normal [overflow-wrap:anywhere]">{item.quantity} {item.unit} · {item.type === "work" ? `施作 · ${states[item.status]} · ${item.completedQuantity}/${item.quantity} ${item.unit}` : "非施作 · 狀態與進度不適用"}</p>
      </ItemContent>
      {item.canEdit && canEdit && <ItemActions><Button variant="outline" aria-label={`編輯子物件：${item.name}`} disabled={busy || uncertain} onClick={event => edit(item, event.currentTarget)}>編輯</Button></ItemActions>}
      {editing && selected?.id === item.id && <ItemFooter>{editor}</ItemFooter>}
    </Item>)}</ItemGroup>
    {cursor && <Button disabled={loading || busy || uncertain || editing} onClick={() => load(cursor)}>載入更多子物件</Button>}
    {canEdit && <Button disabled={busy || uncertain} onClick={event => edit(null, event.currentTarget)}>新增子物件</Button>}
    {editing && !selected && editor}
    {discard.dialog}

  </section>;
}
