import React, { useEffect, useRef, useState } from "react";
import type { components } from "@repo/api-client";
import type { Request } from "@/identity/Session";
import { FormField } from "@/shared/FormField";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Field, FieldLabel } from "@/shared/ui/field";
import { Combobox, ComboboxChip, ComboboxChips, ComboboxChipsInput, ComboboxContent, ComboboxEmpty, ComboboxInput, ComboboxItem, ComboboxList, ComboboxValue } from "@/shared/ui/combobox";
import { errorMessage } from "@/shared/http/lifecycle";

type Space = components["schemas"]["RepositorySpace"];
type Object = components["schemas"]["Object"];
type Choice = { id: string; objectId: string; subobjectId?: string; name: string; spaceId?: string };
export async function collect<T>(fetch: (cursor: string) => Promise<{items:T[];nextCursor:string}>): Promise<T[]> {
  const all:T[]=[]; let cursor="";
  do { const page=await fetch(cursor); all.push(...page.items); cursor=page.nextCursor; } while(cursor);
  return all;
}

export function WorkContext({id, repositoryId, request, lock}: {id:string;repositoryId:string;request:Request;lock:(value:boolean)=>void}) {
  const [choices, setChoices] = useState<Choice[]>([]), [spaces, setSpaces] = useState<Space[]>([]);
  const [target, setTarget] = useState<Choice | null>(null), [selected, setSelected] = useState<Space[]>([]);
  const [error, setError] = useState(""), [busy, setBusy] = useState(true), [revision, reload] = useState(0);
  const [name, setName] = useState(""), [creating, setCreating] = useState(false), [createError, setCreateError] = useState("");
  const pending = useRef<{key:string;name:string;unknown:boolean} | null>(null), writing = useRef(false);
  const alive = useRef(true);
  useEffect(() => { alive.current=true; return () => { alive.current=false; }; }, []);
  useEffect(() => {
    const controller = new AbortController(), options = {signal:controller.signal};
    setBusy(true); setError("");
    void Promise.all([
      collect(cursor => request((c, signal) => c.GET("/api/v1/repositories/{repositoryId}/objects", {signal, params:{path:{repositoryId},query:{cursor:cursor || undefined}}}), options)),
      collect(cursor => request((c, signal) => c.GET("/api/v1/repositories/{repositoryId}/spaces", {signal, params:{path:{repositoryId},query:{cursor:cursor || undefined}}}), options)),
    ]).then(async ([objects, locations]) => {
      const nested = await Promise.all(objects.map(async (object: Object) => {
        const children = await collect(cursor => request((c, signal) => c.GET("/api/v1/repositories/{repositoryId}/objects/{objectId}/subobjects", {signal,params:{path:{repositoryId,objectId:object.id},query:{cursor:cursor || undefined}}}), options));
        return [{id:object.id,objectId:object.id,name:object.name,spaceId:object.spaceId}, ...children.map(child => ({id:child.id,objectId:object.id,subobjectId:child.id,name:object.name+" › "+child.name,spaceId:object.spaceId}))];
      }));
      if (!controller.signal.aborted) { setChoices(nested.flat()); setSpaces(locations); setBusy(false); }
    }).catch(reason => { if (!controller.signal.aborted) { setError(errorMessage(reason)); setBusy(false); } });
    return () => controller.abort();
  }, [repositoryId, request, revision]);
  async function create() {
    if (writing.current || !name.trim() || (!pending.current && selected.length >= 20)) return;
    const attempt = pending.current ?? {key:crypto.randomUUID(),name:name.trim(),unknown:false};
    pending.current=attempt; writing.current=true; lock(true); setCreating(true); setCreateError("");
    try {
      const space = await request((c, signal) => c.POST("/api/v1/repositories/{repositoryId}/spaces", {signal,params:{path:{repositoryId},header:{"Idempotency-Key":attempt.key}},body:{name:attempt.name}}));
      if (!alive.current) return;
      setSpaces(old => [...old.filter(s => s.id !== space.id), space]);
      setSelected(old => [...old.filter(s => s.id !== space.id), space]);
      pending.current=null; lock(false); setName("");
    } catch (reason) {
      if (!alive.current) return;
      if (!attempt.unknown && ["invalid_input", "forbidden", "not_found", "idempotency_conflict"].includes((reason as {code?:string}).code || "")) { pending.current=null; lock(false); }
      else attempt.unknown=true;
      setCreateError(errorMessage(reason));
    } finally { writing.current=false; if (alive.current) setCreating(false); }
  }
  return <div className="grid min-w-0 grid-cols-1 gap-3" data-pending-space={pending.current ? "" : undefined}>
    <Status>{busy ? "正在讀取工作對象與空間……" : error}</Status>
    {error && <Button type="button" variant="outline" onClick={() => reload(v => v+1)}>重試讀取工作對象與空間</Button>}
    <Field><FieldLabel htmlFor={id+"target"}>工作對象（選填）</FieldLabel>
      <Combobox<Choice> items={choices} value={target} onValueChange={value => {
        if (pending.current) return;
        setTarget(value); setSelected(value?.spaceId ? spaces.filter(s => s.id === value.spaceId) : []);
      }} itemToStringLabel={item => item.name} disabled={busy || !!error || !!pending.current}>
        <ComboboxInput disabled={busy || !!error || !!pending.current} id={id+"target"} placeholder="搜尋物件或子物件" showClear />
        <ComboboxContent><ComboboxEmpty>沒有符合的工作對象</ComboboxEmpty><ComboboxList>{(item:Choice) => <ComboboxItem key={item.id} value={item}><span className="truncate" title={item.name}>{item.name}</span></ComboboxItem>}</ComboboxList></ComboboxContent>
      </Combobox>
      <input type="hidden" name={id+"object"} value={target?.objectId || ""}/>
      <input type="hidden" name={id+"subobject"} value={target?.subobjectId || ""}/>
    </Field>
    <Field><FieldLabel htmlFor={id+"spaces"}>實際作業空間（選填，可複選）</FieldLabel>
      <Combobox<Space, true> multiple items={spaces} value={selected} onValueChange={value => { if (!pending.current && value.length <= 20) setSelected(value); }} itemToStringLabel={item => item.name} isItemEqualToValue={(a,b) => a.id === b.id} disabled={busy || !!error || !!pending.current}>
        <ComboboxChips><ComboboxValue>{selected.map(space => <ComboboxChip key={space.id} className="max-w-full"><span className="truncate" title={space.name}>{space.name}</span></ComboboxChip>)}</ComboboxValue><ComboboxChipsInput id={id+"spaces"} placeholder="搜尋空間"/></ComboboxChips>
        <ComboboxContent><ComboboxEmpty>沒有符合的空間</ComboboxEmpty><ComboboxList>{(space:Space) => <ComboboxItem key={space.id} value={space}><span className="truncate" title={space.name}>{space.name}</span></ComboboxItem>}</ComboboxList></ComboboxContent>
      </Combobox>
      {selected.map(space => <input key={space.id} type="hidden" name={id+"space"} value={space.id}/>)}
    </Field>
    <FormField>新增空間（選填）<Input maxLength={120} value={name} disabled={!!pending.current} onChange={e => {if(!pending.current)setName(e.target.value);}}/></FormField>
    <p>空間會立即建立到此儲存庫，即使未發布日誌也會保留。</p>
    <Button type="button" variant="outline" disabled={creating || busy || !!error || !name.trim() || selected.length >= 20} onClick={() => void create()}>{creating ? "建立中" : pending.current ? "重試原空間建立請求" : "建立並選用"}</Button>
    <Status>{createError}</Status>
  </div>;
}
