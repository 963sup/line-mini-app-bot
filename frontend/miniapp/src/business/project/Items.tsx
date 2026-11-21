import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from "@/shared/ui/empty";
import React, { useCallback, useEffect, useId, useRef, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { useLatestRequest, errorMessage } from "@/shared/http/lifecycle";
import { PagePanel } from "@/shared/PagePanel";
import { FormField } from "@/shared/FormField";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";
import { Item, ItemGroup, ItemContent, ItemTitle } from "@/shared/ui/item";
import { RadioGroup, RadioGroupItem } from "@/shared/ui/radio-group";
import { FieldGroup, FieldSet, FieldLegend, Field, FieldLabel } from "@/shared/ui/field";

import { Inspections } from "./Inspections";
import { SourcePicker } from "./SourcePicker";

type Project = components["schemas"]["Project"];
type Reference = components["schemas"]["ProjectReference"];
type Addition = components["schemas"]["ProjectReferenceAddition"];
type ItemProgress = NonNullable<components["schemas"]["ProjectUpdate"]["itemProgress"]>;
const progressLabels = {todo:"待辦",in_progress:"進行中",done:"完成"};
type Kind = Addition["sourceType"] | "subobject";
type Option = { id: string; title: string };
const kinds: Record<Kind, string> = { object: "物件", subobject: "子物件", issue: "問題", discussion: "討論" };
const states: Record<string, string> = { open: "待處理", in_progress: "進行中", completed: "已完成", not_started: "未開始", closed: "已結束" };
type Props = { ownerId: string; sourceRepositoryId: string; project: Project; disabled: boolean; canWrite: () => boolean; onBlocked: (blocked: boolean) => void; onChange: (project: Project) => void; onRefresh: () => void };
export function Items({ ownerId, sourceRepositoryId: initialRepositoryId, project, disabled, canWrite, onBlocked, onChange, onRefresh }: Props) {
  const kindID = useId();
  const [sourceRepositoryId, setSourceRepositoryId] = useState(initialRepositoryId);
  const { user, request } = useSession(), read = useLatestRequest(), catalog = useLatestRequest(), parents = useLatestRequest(), write = useLatestRequest();
  const [items, setItems] = useState<Reference[]>([]), [options, setOptions] = useState<Option[]>([]), [objects, setObjects] = useState<Option[]>([]);
  const [kind, setKind] = useState<Kind>("issue"), [parent, setParent] = useState(""), [choice, setChoice] = useState("");
  const [cursor, setCursor] = useState(""), [catalogCursor, setCatalogCursor] = useState(""), [parentCursor, setParentCursor] = useState("");
  const [status, setStatus] = useState(""), [catalogStatus, setCatalogStatus] = useState(""), [parentStatus, setParentStatus] = useState("");
  const [busy, setBusy] = useState(false), [uncertain, setUncertain] = useState(false), [readBusy, setReadBusy] = useState(false), [catalogBusy, setCatalogBusy] = useState(false), [parentBusy, setParentBusy] = useState(false), [loaded, setLoaded] = useState(false), [needsRefresh, setNeedsRefresh] = useState(false);
  const lock = useRef(false), unknown = useRef(false), selecting = useRef(false);
  const pending = useRef<{ body: Addition } | { itemId: string; version: number } | null>(null);
  const linked = (project.linkedRepositoryIds ?? [initialRepositoryId]).includes(sourceRepositoryId);
  const canEdit = project.canEdit, sourceAvailable = linked && !!sourceRepositoryId && (sourceRepositoryId !== initialRepositoryId || project.sourceAvailable !== false);
  useEffect(() => { selecting.current = false; });
  useEffect(() => () => onBlocked(false), [onBlocked]);
  function block(value: boolean) { unknown.current = value; setUncertain(value); onBlocked(value); }
  useEffect(() => {
    if (!busy && !uncertain) return;
    const warn = (event: BeforeUnloadEvent) => { event.preventDefault(); event.returnValue = ""; };
    window.addEventListener("beforeunload", warn); return () => window.removeEventListener("beforeunload", warn);
  }, [busy, uncertain]);
  const load = useCallback((next = "") => {
    if (!user || lock.current || unknown.current) return;
    setReadBusy(true); setLoaded(false); if (!next) setItems([]); setCursor(""); setStatus("正在讀取專案引用……");
    void read.run(signal => request((client, s) => client.GET("/api/v1/owners/{ownerId}/projects/{id}/references", { signal: s, params: { path: { ownerId, id: project.id }, query: { cursor: next || undefined } } }), { signal }), page => {
      setItems(previous => [...new Map([...(next ? previous : []), ...page.items].map(item => [item.id, item])).values()]); setCursor(page.nextCursor); setReadBusy(false); setLoaded(true); setStatus("");
    }, error => {
      const status = (error as {httpStatus?: number}).httpStatus;
      if (next && (!status || status >= 500 || status === 408 || status === 429)) {
        setCursor(next);
      } else {
        setItems([]);
      }
      setReadBusy(false); setStatus(errorMessage(error));
    });
  }, [user, request, read.run, ownerId, project.id]);
  const loadParents = useCallback((next = "") => {
    if (!user || !canEdit || !sourceAvailable || kind !== "subobject") return;
    setParentBusy(true); if (!next) {setObjects([]);setParent("");setChoice("");setOptions([]);catalog.cancel();} setParentCursor(""); setParentStatus("正在讀取父物件……");
    void parents.run(signal => request((client,s) => client.GET("/api/v1/repositories/{repositoryId}/objects", {signal:s,params:{path:{repositoryId:sourceRepositoryId},query:{cursor:next || undefined}}}),{signal}), page => {
      setObjects(previous => [...new Map([...(next ? previous : []), ...page.items.map(item => ({id:item.id,title:`${item.name} · ${item.quantity} ${item.unit}`}))].map(item => [item.id,item])).values()]); setParentCursor(page.nextCursor); setParentBusy(false); setParentStatus("");
    }, error => {setObjects([]);setParent("");setChoice("");setOptions([]);catalog.cancel();setParentBusy(false);setParentStatus(errorMessage(error));});
  }, [user,canEdit,sourceAvailable,kind,parents.run,request,sourceRepositoryId,catalog.cancel]);
  const loadCatalog = useCallback((next = "") => {
    if (!user || !canEdit || !sourceAvailable || (kind === "subobject" && !parent)) return;
    setCatalogBusy(true); if (!next) { setOptions([]); setChoice(""); } setCatalogCursor(""); setCatalogStatus(`正在讀取可收錄${kinds[kind]}……`);
    void catalog.run(async signal => {
      const query = {cursor:next || undefined}, repositoryId=sourceRepositoryId;
      if (kind === "object") {
        const page=await request((client,s)=>client.GET("/api/v1/repositories/{repositoryId}/objects",{signal:s,params:{path:{repositoryId},query}}),{signal});
        return {items:page.items.map(item=>({id:item.id,title:`${item.name} · ${item.quantity} ${item.unit}`})),nextCursor:page.nextCursor};
      }
      if (kind === "subobject") {
        const page=await request((client,s)=>client.GET("/api/v1/repositories/{repositoryId}/objects",{signal:s,params:{path:{repositoryId},query:{...query,parentId:parent}}}),{signal});
        return {items:page.items.filter(item => item.parentId === parent).map(item=>({id:item.id,title:`${item.name} · ${item.quantity} ${item.unit}`})),nextCursor:page.nextCursor};
      }
      if (kind === "discussion") {
        const page=await request((client,s)=>client.GET("/api/v1/repositories/{repositoryId}/discussions",{signal:s,params:{path:{repositoryId},query}}),{signal});
        return {items:page.items.map(item=>({id:item.id,title:item.title})),nextCursor:page.nextCursor};
      }
      const page=await request((client,s)=>client.GET("/api/v1/repositories/{repositoryId}/issues",{signal:s,params:{path:{repositoryId},query}}),{signal});
      return {items:page.items.map(item=>({id:item.id,title:`#${item.number} ${item.title}`})),nextCursor:page.nextCursor};
    }, page => {setOptions(previous=>[...new Map([...(next?previous:[]),...page.items].map(item=>[item.id,item])).values()]);setCatalogCursor(page.nextCursor);setCatalogBusy(false);setCatalogStatus("");}, error=>{setOptions([]);setChoice("");setCatalogBusy(false);setCatalogStatus(errorMessage(error));});
  }, [user,canEdit,sourceAvailable,kind,parent,catalog.run,request,sourceRepositoryId]);
  useEffect(()=>{load();return read.cancel;},[load,read.cancel]);
  useEffect(()=>{loadParents();return parents.cancel;},[loadParents,parents.cancel]);
  useEffect(()=>{loadCatalog();return catalog.cancel;},[loadCatalog,catalog.cancel]);
  function change(itemId?: string) {
    if (!user || !canEdit || lock.current || selecting.current || needsRefresh || !canWrite() || (disabled && !unknown.current)) return;
    if (!pending.current) {
      if (!itemId && (!choice || !sourceAvailable || (kind === "subobject" && !parent))) return;
      pending.current = itemId ? {itemId,version:project.version} : {body:{sourceRepositoryId,sourceType:kind === "subobject" ? "object" : kind,sourceId:choice,expectedVersion:project.version}};
    }
    const attempt=pending.current, wasUnknown=unknown.current;
    lock.current=true;setBusy(true);block(true);setStatus("正在更新專案引用……");
    void write.run(signal=>request((client,s)=>"itemId" in attempt ? client.DELETE("/api/v1/owners/{ownerId}/projects/{id}/references/{itemId}",{signal:s,params:{path:{ownerId,id:project.id,itemId:attempt.itemId},query:{expectedVersion:attempt.version}}}) : client.POST("/api/v1/owners/{ownerId}/projects/{id}/references",{signal:s,params:{path:{ownerId,id:project.id}},body:attempt.body}),{signal}), value=>{
      lock.current=false;setBusy(false);pending.current=null;block(false);onChange(value);
    },error=>{
      lock.current=false;setBusy(false);
      const rejected=["invalid_input","forbidden","not_found","version_conflict","conflict"].includes((error as {code?:string}).code || "");
      if(rejected && !wasUnknown){pending.current=null;block(false);setNeedsRefresh(true);setStatus(errorMessage(error)+" 請重新讀取專案後再管理引用。");}
      else setStatus(errorMessage(error)+" 結果尚未確認，請重試原引用操作；不要更換來源。");
    });
  }
  function updateProject(itemProgress?: ItemProgress) {
    if (!canEdit || (!itemProgress && !sourceRepositoryId) || !canWrite() || disabled || lock.current || unknown.current) return;
    lock.current = true; setBusy(true); onBlocked(true);
    void write.run(signal => request((client, s) => client.PATCH("/api/v1/owners/{ownerId}/projects/{id}", {
      signal: s, params: { path: { ownerId, id: project.id } },
      body: { title: project.title, body: project.body, expectedVersion: project.version,
        ...(itemProgress ? {itemProgress} : linked ? { unlinkRepositoryId: sourceRepositoryId } : { linkRepositoryId: sourceRepositoryId }) },
    }), {signal}), updated => {
      lock.current = false; setBusy(false); onBlocked(false); onChange(updated);
    }, error => {
      lock.current = false; setBusy(false); onBlocked(false); setNeedsRefresh(true);
      setStatus(errorMessage(error) + " 請重新讀取專案以確認更新結果。");
    });
  }
  function verify() {
    if (lock.current || !unknown.current || !pending.current || !canWrite()) return;
    const attempt = pending.current;
    lock.current = true; setBusy(true); setStatus("正在核對引用結果……");
    read.cancel(); catalog.cancel(); parents.cancel(); setOptions([]); setObjects([]);
    setReadBusy(false); setCatalogBusy(false); setParentBusy(false);
    void write.run(signal => request((client, s) => client.GET("/api/v1/owners/{ownerId}/projects/{id}/reference-state", {
      signal: s, params: { path: { ownerId, id: project.id }, query: "itemId" in attempt
        ? { itemId: attempt.itemId, desiredPresent: false }
        : { sourceRepositoryId: attempt.body.sourceRepositoryId || initialRepositoryId, sourceType: attempt.body.sourceType,
            sourceId: attempt.body.sourceId, desiredPresent: true } },
    }), {signal}), result => {
      lock.current = false; setBusy(false);
      if (!result.matches) { setStatus("目前引用狀態尚未符合原操作，請保留原操作並稍後重試或核對。"); return; }
      pending.current = null; setNeedsRefresh(false); setChoice(""); block(false);
      setStatus("目前引用狀態已符合預期，正在重新讀取專案。");
      onRefresh();
    }, error => {lock.current=false;setBusy(false);setStatus(errorMessage(error)+" 原操作仍保留，請稍後再次核對。");});
  }
  return <PagePanel heading="專案引用"><p>顯示來源最新內容。加入或移出專案只變更引用，不會修改來源、授予來源編輯權或加總不同來源的進度。</p>
    <Status role="status">{status}</Status>
    <Button type="button" variant="outline" disabled={busy||uncertain||disabled} onClick={onRefresh}>重新讀取專案與引用</Button>
    {uncertain && <><Button type="button" disabled={busy} onClick={()=>change()}>重試原引用操作</Button><Button type="button" variant="outline" disabled={busy} onClick={verify}>核對引用結果</Button></>}
    {(["工程對象","待處理問題","決策討論"] as const).map(group=>{
      const entries=items.filter(item=>(item.sourceType==="issue"?"待處理問題":item.sourceType==="discussion"?"決策討論":"工程對象")===group);
      if(!entries.length)return null;
      return <section key={group}><h3>{group}</h3><ItemGroup className="min-w-0">{entries.map(item=>{
      const title=item.sourceUnavailable ? `不可讀取的${kinds[item.sourceType]}` : item.sourceTitle;
      return <Item key={item.id} variant="outline" className="min-w-0"><ItemContent className="min-w-0"><ItemTitle className="w-full min-w-0 whitespace-normal [overflow-wrap:anywhere]">{kinds[item.sourceType]}：{title}</ItemTitle>
        {item.sourceUnavailable ? <p>來源目前不可讀取；不顯示先前內容。</p> : <>
          {(item.type === "non_work" || item.sourceState) && <p>來源狀態：{item.type === "non_work" ? "不適用" : states[item.sourceState] || item.sourceState}</p>}
          {item.quantity !== undefined && <p className="whitespace-normal [overflow-wrap:anywhere]">數量：{item.quantity} {item.unit}</p>}
          {item.type === "non_work" ? <p>進度：不適用</p> : item.completedQuantity !== undefined && <p className="whitespace-normal [overflow-wrap:anywhere]">進度：{item.completedQuantity}/{item.quantity} {item.unit}</p>}
        </>}
        <FormField>本專案進度<NativeSelect value={item.itemStatus ?? "todo"} disabled={!canEdit || disabled || needsRefresh || busy || uncertain} onChange={e=>updateProject({itemId:item.id,status:e.target.value as ItemProgress["status"]})}>{Object.entries(progressLabels).map(([value,label])=><NativeSelectOption key={value} value={value}>{label}</NativeSelectOption>)}</NativeSelect></FormField>
        {(item.inspectionMode === "history_only" || item.inspectionMode === "submit_if_empty") && <Inspections key={item.inspectionMode} canSubmit={item.inspectionMode === "submit_if_empty"} ownerId={ownerId} projectId={project.id} itemId={item.id}/>}
        {canEdit&&<Button type="button" variant="outline" aria-label={`移出專案：${title}`} disabled={disabled||needsRefresh||busy||uncertain} onClick={()=>change(item.id)}>移出專案</Button>}
      </ItemContent></Item>;
    })}</ItemGroup></section>;})}
    {loaded&&!items.length&&<Empty><EmptyHeader><EmptyTitle>尚未收錄來源。</EmptyTitle>{canEdit && <EmptyDescription>請從下方選擇來源儲存庫與項目。</EmptyDescription>}</EmptyHeader></Empty>}
    {cursor&&<Button type="button" variant="outline" disabled={readBusy||busy||uncertain} onClick={()=>load(cursor)}>載入更多專案引用</Button>}
    {canEdit && <SourcePicker ownerId="" initialSource="" value={sourceRepositoryId} disabled={disabled||needsRefresh||busy||uncertain} onChange={id => { catalog.cancel(); parents.cancel(); setChoice(""); setParent(""); setOptions([]); setObjects([]); setSourceRepositoryId(id); }} />}
    {canEdit && sourceRepositoryId && <Button type="button" variant="outline" disabled={disabled||needsRefresh||busy||uncertain} onClick={()=>updateProject()}>{linked ? "解除儲存庫連結" : "連結此儲存庫"}</Button>}
    {canEdit && (sourceAvailable ? <FieldGroup>
      <FieldSet><FieldLegend>來源種類</FieldLegend>
        <RadioGroup aria-label="來源種類" className="grid-cols-2 sm:grid-cols-4" value={kind}
          disabled={disabled || needsRefresh || busy || uncertain}
          onValueChange={value => {
            if (lock.current || unknown.current || value === kind) return;
            selecting.current = true;
            catalog.cancel(); parents.cancel();
            setKind(value as Kind); setParent(""); setChoice(""); setOptions([]); setObjects([]);
            setCatalogCursor(""); setParentCursor(""); setCatalogStatus(""); setParentStatus("");
          }}>
          {Object.entries(kinds).map(([value, label]) => (
            <Field key={value} orientation="horizontal" className="gap-2">
              <RadioGroupItem id={kindID + "-" + value} value={value} aria-label={label} /><FieldLabel htmlFor={kindID + "-" + value} className="min-h-11 flex-1">{label}</FieldLabel>
            </Field>
          ))}
        </RadioGroup>
      </FieldSet>
      {kind === "subobject" && <>
        <Status role="status">{parentStatus}</Status>
        <FormField>父物件<NativeSelect value={parent} disabled={disabled||needsRefresh||busy||uncertain||parentBusy} onChange={event=>{if(lock.current||unknown.current||event.target.value===parent)return;selecting.current=true;catalog.cancel();setParent(event.target.value);setChoice("");setOptions([]);setCatalogCursor("");}}><NativeSelectOption value="">請先選擇父物件</NativeSelectOption>{objects.map(item=><NativeSelectOption key={item.id} value={item.id}>{item.title}</NativeSelectOption>)}</NativeSelect></FormField>
        <Button type="button" variant="outline" disabled={disabled||needsRefresh||busy||uncertain||parentBusy} onClick={()=>loadParents()}>重新讀取父物件</Button>
        {parentCursor&&<Button type="button" variant="outline" disabled={disabled||needsRefresh||busy||uncertain||parentBusy} onClick={()=>loadParents(parentCursor)}>載入更多父物件</Button>}
      </>}
      <Status role="status">{catalogStatus}</Status>
      <FormField>選擇既有{kinds[kind]}<NativeSelect value={choice} disabled={disabled||needsRefresh||busy||uncertain||catalogBusy||(kind === "subobject"&&!parent)} onChange={event=>{if(event.target.value===choice)return;selecting.current=true;setChoice(event.target.value);}}><NativeSelectOption value="">請選擇{kinds[kind]}</NativeSelectOption>{options.map(item=><NativeSelectOption key={item.id} value={item.id}>{item.title}</NativeSelectOption>)}</NativeSelect></FormField>
      <Button type="button" disabled={disabled||needsRefresh||busy||uncertain||catalogBusy||!choice} onClick={()=>change()}>加入專案</Button>
      <Button type="button" variant="outline" disabled={disabled||catalogBusy||busy||uncertain||(kind === "subobject"&&!parent)} onClick={()=>loadCatalog()}>重新讀取可收錄來源</Button>
      {catalogCursor&&<Button type="button" variant="outline" disabled={disabled||catalogBusy||busy||uncertain} onClick={()=>loadCatalog(catalogCursor)}>載入更多可收錄來源</Button>}
    </FieldGroup> : <p>{linked ? "來源儲存庫目前不可讀取。" : "請先連結儲存庫，再加入引用。"}</p>)}
  </PagePanel>;
}
