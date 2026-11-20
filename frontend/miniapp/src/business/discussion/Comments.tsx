import React, { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import { useDiscardChanges } from "@/shared/useDiscardChanges";
import { useRepositoryGuard } from "../repository/NavigationGuard";
import { ReadFeedback } from "@/shared/ReadFeedback";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { useLatestRequest, errorMessage } from "@/shared/http/lifecycle";
import { displayTime } from "@/shared/time/time";
import { PagePanel } from "@/shared/PagePanel";
import { FormField } from "@/shared/FormField";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { Textarea } from "@/shared/ui/textarea";
import { FieldGroup } from "@/shared/ui/field";
import { Empty, EmptyHeader, EmptyTitle } from "@/shared/ui/empty";
import { Item, ItemGroup, ItemContent, ItemTitle } from "@/shared/ui/item";

type Comment = components["schemas"]["DiscussionComment"];
export function Comments({repositoryId,discussionId,onWriteStateChange}:{repositoryId:string;discussionId:string;onWriteStateChange:(blocked:boolean)=>boolean}) {
  const {user,request}=useSession();
  const {run:readRun,cancel:cancelRead}=useLatestRequest(),write=useLatestRequest();
  const [items,setItems]=useState<Comment[]>([]),[cursor,setCursor]=useState("");
  const [loaded,setLoaded]=useState(false),[readBusy,setReadBusy]=useState(false),[busy,setBusy]=useState(false);
  const [listStatus,setListStatus]=useState(""),[status,setStatus]=useState("");
  const [body,setBody]=useState(""),[editing,setEditing]=useState<Comment|null>(null),[editBody,setEditBody]=useState("");
  const [uncertain,setUncertain]=useState(false),[needsRefresh,setNeedsRefresh]=useState(false);
  const navigating = useRef(false);
  useEffect(() => { navigating.current = false; });
  const lock=useRef(false),attempt=useRef<{key:string;body:string}|null>(null);
  const recoveryTarget = useRef<Comment | null>(null);
  const input = useRef<HTMLTextAreaElement>(null);
  const dirty = !!body || (!!editing && editBody !== editing.body);
  const focusEditor = () => input.current?.focus();
  const guard = useRepositoryGuard(dirty, focusEditor);
  const discard = useDiscardChanges(dirty, focusEditor);
  useEffect(() => { guard.block(busy || uncertain || needsRefresh); }, [busy, uncertain, needsRefresh]);
  useEffect(() => { onWriteStateChange(busy || uncertain || needsRefresh || dirty); }, [busy, uncertain, needsRefresh, dirty, onWriteStateChange]);
  useEffect(() => () => { onWriteStateChange(false); }, [onWriteStateChange]);
  const load=useCallback((nextCursor="")=>{
    if(!user || lock.current)return;
    navigating.current = true;
    setReadBusy(true);setLoaded(false);setCursor("");
    if(!nextCursor)setItems([]);
    setListStatus("");
    void readRun(
      signal=>request((client,s)=>client.GET("/api/v1/repositories/{repositoryId}/discussions/{id}/comments",{signal:s,params:{path:{repositoryId,id:discussionId},query:{cursor:nextCursor||undefined}}}),{signal}),
      page=>{
        setItems(previous=>[...new Map([...(nextCursor?previous:[]),...page.items].map(item=>[item.id,item])).values()]);
        setCursor(page.nextCursor);setLoaded(true);setReadBusy(false);setListStatus("");
        // Only a fresh read containing the original target resolves its update.
        const target = recoveryTarget.current;
        const fresh = target && page.items.find(item => item.id === target.id);
        if (fresh) {
          recoveryTarget.current = null;
          setEditing(fresh);setEditBody(fresh.body);setNeedsRefresh(false);
          guard.block(false);onWriteStateChange(false);setStatus("已重新讀取原留言，請核對內容後再編輯。");
        } else if (target) {
          setStatus("尚未讀到原留言，請繼續載入更多留言以確認更新結果。");
        }
      },
      error=>{if(!nextCursor)setItems([]);setCursor(nextCursor);setReadBusy(false);setLoaded(false);setListStatus(errorMessage(error));},
    );
  },[user,request,repositoryId,discussionId,readRun,guard.block,onWriteStateChange]);
  useEffect(()=>{load();return cancelRead;},[load,cancelRead]);
  function submit(event:FormEvent){
    event.preventDefault();if(!user||lock.current||navigating.current||needsRefresh||discard.isConfirming()||!guard.canWrite())return;
    if(!uncertain||!attempt.current){const text=body.trim();if(!text){setStatus("請填寫留言。");return;}attempt.current={key:crypto.randomUUID(),body:text};}
    if(discard.isConfirming()||!guard.canWrite()||!onWriteStateChange(true))return;
    guard.block(true);
    const wasUnknown=uncertain;const pending=attempt.current;lock.current=true;setBusy(true);setUncertain(true);setStatus("正在新增留言……");
    void write.run(
      signal=>request((client,s)=>client.POST("/api/v1/repositories/{repositoryId}/discussions/{id}/comments",{signal:s,params:{path:{repositoryId,id:discussionId},header:{"Idempotency-Key":pending.key}},body:{body:pending.body}}),{signal}),
      ()=>{lock.current=false;guard.block(false);onWriteStateChange(false);setBusy(false);setUncertain(false);attempt.current=null;setBody("");setStatus("已新增留言；留言依建立時間顯示，可載入更多。");load();},
      error=>{lock.current=false;setBusy(false);if(!wasUnknown&&["invalid_input","forbidden","not_found","idempotency_conflict"].includes((error as {code?:string}).code||"")){setUncertain(false);guard.block(false);onWriteStateChange(false);}setStatus(errorMessage(error)+" 結果不明時請沿用原留言重試；確認前請勿切換主題或重新開頁。");},
    );
  }
  function save(event:FormEvent){
    event.preventDefault();if(!user||!editing||!editing.canEdit||lock.current||navigating.current||needsRefresh)return;
    if(discard.isConfirming()||!guard.canWrite()||!onWriteStateChange(true))return;
    guard.block(true);
    const item=editing;lock.current=true;setBusy(true);setStatus("正在更新留言……");
    void write.run(
      signal=>request((client,s)=>client.PATCH("/api/v1/repositories/{repositoryId}/discussions/{id}/comments/{commentId}",{signal:s,params:{path:{repositoryId,id:discussionId,commentId:item.id}},body:{body:editBody.trim(),expectedVersion:item.version}}),{signal}),
      updated => {
        lock.current = false; guard.block(false); onWriteStateChange(false);
        setBusy(false); setEditing(null); setStatus("已更新留言。");
        // A page requested before the edit could still return the old version.
        if (readBusy) load();
        else setItems(rows => rows.map(row => row.id === updated.id ? updated : row));
      },
      error=>{lock.current=false;recoveryTarget.current=item;onWriteStateChange(true);setBusy(false);setNeedsRefresh(true);setStatus(errorMessage(error)+" 請重新載入留言後再編輯。");},
    );
  }
  return <PagePanel heading="留言">
    <p className="mb-4 text-sm text-muted-foreground">空間成員可留言，每則留言由自己的作者編輯；目前提供單層留言。</p>
    <Button variant="outline" disabled={!user||readBusy||busy||uncertain} onClick={()=>{if(recoveryTarget.current){load();return;}discard.navigate(()=>{setBody("");setEditBody("");setEditing(null);load();});}}>重新載入留言</Button>
    <ReadFeedback busy={readBusy} message={listStatus} label="正在讀取留言……" />
    <ItemGroup className="gap-3">
      {items.map(item=><Item key={item.id} variant="outline" role="listitem"><ItemContent>
        <ItemTitle>{item.authorId===user?.userId?"本人":"空間成員"} · {displayTime(item.createdAt)}</ItemTitle>
        <p className="whitespace-pre-wrap break-words">{item.body}</p>
        {item.canEdit&&<Button variant="outline" disabled={busy||uncertain||needsRefresh} onClick={()=>discard.navigate(()=>{navigating.current=true;setBody("");setEditing(item);setEditBody(item.body);})}>編輯留言</Button>}
      </ItemContent></Item>)}
    </ItemGroup>
    {loaded&&!items.length&&<Empty><EmptyHeader><EmptyTitle>目前沒有留言</EmptyTitle></EmptyHeader></Empty>}
    {cursor&&<Button variant="outline" disabled={readBusy||busy||uncertain} onClick={()=>load(cursor)}>載入更多留言</Button>}
    <Status role="status">{status}</Status>
    {editing?<form onSubmit={save}><FieldGroup>
      <FormField>編輯自己的留言<Textarea ref={input} required maxLength={4000} value={editBody} disabled={busy||needsRefresh} onChange={e=>setEditBody(e.target.value)}/></FormField>
      <div className="flex flex-wrap gap-2"><Button type="submit" disabled={busy||needsRefresh}>儲存留言</Button><Button type="button" variant="outline" disabled={busy||needsRefresh} onClick={()=>discard.navigate(()=>setEditing(null))}>取消編輯</Button></div>
    </FieldGroup></form>:<form onSubmit={submit}><FieldGroup>
      <FormField>新增留言<Textarea ref={input} required maxLength={4000} value={body} disabled={busy||uncertain||needsRefresh} onChange={e=>setBody(e.target.value)}/></FormField>
      <Button type="submit" disabled={!user||busy||needsRefresh}>{uncertain?"重試原留言":"送出留言"}</Button>
    </FieldGroup></form>}
    {discard.dialog}
  </PagePanel>;
}
