import dynamic from "next/dynamic";
import React, { useEffect, useState, useRef } from "react";
import { useSession, type Request } from "@/identity/Session";
import { PagePanel } from "@/shared/PagePanel";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";
import type { components } from "@repo/api-client";
const GroupCharts = dynamic(() => import("./GroupCharts"), {
  loading: () => <p role="status">正在載入群組圖表……</p>,
});
type Journal = components["schemas"]["Journal"];
type Draft = components["schemas"]["JournalCreation"];
type Reference = components["schemas"]["JournalReference"];
const roles = { management: "管理職", control: "管制人員", workers: "施作人員" } as const;
const workDate = () => new Intl.DateTimeFormat("sv-SE", { timeZone: "Asia/Taipei" }).format(new Date());

type Option = {id:string;name:string};
type Target = Reference & {name:string};
async function collect<T>(fetch:(cursor:string)=>Promise<{items:T[];nextCursor:string}>):Promise<T[]> {
  const all:T[]=[];let cursor="";
  do {const page=await fetch(cursor);all.push(...page.items);cursor=page.nextCursor;} while(cursor);
  return all;
}
function Rows({title,rows,change,targets}:{title:string;rows:string[];change:(v:string[])=>void;targets:Target[]}) {
  return <fieldset className="grid gap-2"><legend>{title}</legend>{rows.map((id,i)=><fieldset key={id}><legend>第 {i+1} 項</legend>
    <label>工作名稱<Input name={id+"name"} required maxLength={120}/></label>
    <label>累計完成度（%）<Input name={id+"percent"} required type="number" min={0} max={100} step={1}/></label>
    <label>工作關聯 <select name={id+"ref"}><option value="">無關聯</option>{targets.map(t=><option key={t.type+t.id} value={t.type+t.id}>{t.type==="project"?"專案":"儲存庫"} · {t.name}</option>)}</select></label>
    <Button type="button" variant="outline" onClick={()=>change(rows.filter(k=>k!==id))}>移除第 {i+1} 項</Button>
  </fieldset>)}<Button type="button" variant="outline" disabled={rows.length>=100} onClick={()=>change([...rows,crypto.randomUUID()])}>新增{title}</Button></fieldset>;
}
function Composer({request,group,saved}:{request:Request;group:string;saved:()=>void}) {
  const [today,setToday]=useState(["first"]),[tomorrow,setTomorrow]=useState<string[]>([]);
  const [targets,setTargets]=useState<Target[]>([]),[notice,setNotice]=useState(""),[error,setError]=useState(""),[busy,setBusy]=useState(false);
  const publishAttempt=useRef({serializedBody:"",idempotencyKey:""}),submitting=useRef(false);
  const write=useLatestRequest();
  useEffect(()=>{
    const controller=new AbortController(), options={signal:controller.signal};
    Promise.all([
      collect(cursor=>request((c,signal)=>c.GET("/api/v1/repositories",{signal,params:{query:{cursor:cursor||undefined}}}),options)),
      collect(cursor=>request((c,signal)=>c.GET("/api/v1/owners/{ownerId}/projects",{signal,params:{path:{ownerId:group},query:{cursor:cursor||undefined}}}),options))
    ]).then(([repos,projects])=>{
      if(!controller.signal.aborted)setTargets([...repos.filter(r=>r.ownerId===group).map(r=>({type:"repository" as const,id:r.id,name:r.name})),...projects.map(p=>({type:"project" as const,id:p.id,name:p.title}))]);
    }).catch(e=>{if(!controller.signal.aborted)setNotice("關聯清單無法載入："+errorMessage(e));});
    return ()=>controller.abort();
  },[group,request]);
  return <PagePanel heading="寫日誌"><form onSubmit={event=>{
    event.preventDefault();if(submitting.current)return;
    const form=new FormData(event.currentTarget),get=(k:string)=>String(form.get(k)??"").trim();
    const people=(k:string)=>get(k).split(/[、，,\n]/).map(s=>s.trim()).filter(Boolean);
    const row=(id:string)=>{const t=targets.find(t=>t.type+t.id===get(id+"ref"));return {name:get(id+"name"),reference:t?{type:t.type,id:t.id}:undefined};};
    const body:Draft={groupId:group,workDate:get("date"), personnel:{management:people("management"),control:people("control"),workers:people("workers")},
      todayItems:today.map(id=>({...row(id),completionPercent:Number(get(id+"percent"))})),
      tomorrowItems:tomorrow.map(id=>({...row(id),expectedCompletionPercent:Number(get(id+"percent"))}))};
    if(!today.length&&!tomorrow.length){setError("請新增至少一項工作。");return;}
    const json=JSON.stringify(body);if(publishAttempt.current.serializedBody!==json)publishAttempt.current={serializedBody:json,idempotencyKey:crypto.randomUUID()};
    submitting.current=true;setBusy(true);setError("");
    void write.run(signal=>request((c,s)=>c.POST("/api/v1/journals",{signal:s,params:{header:{"Idempotency-Key":publishAttempt.current.idempotencyKey}},body}),{signal}),
      saved,e=>{submitting.current=false;setBusy(false);setError(errorMessage(e));});
  }}><fieldset disabled={busy} className="grid gap-3">
    <label>工作日期<Input required type="date" name="date" defaultValue={workDate()}/></label>
    {Object.entries(roles).map(([key,label])=><label key={key}>{label}<Input name={key} placeholder="多位姓名以頓號分隔"/></label>)}
    {notice&&<p role="status">{notice}</p>}
    <Rows title="今日工作" rows={today} change={setToday} targets={targets}/>
    <Rows title="明日計畫" rows={tomorrow} change={setTomorrow} targets={targets}/>
    {error&&<p role="alert">{error}</p>}<Button type="submit">{busy?"發布中":"發布日誌"}</Button>
  </fieldset></form></PagePanel>;
}

export function JournalWorkspace({dashboard}:{dashboard:boolean}){
  const {user,epoch,request}=useSession();
  if(!user)return <p role="status">請先登入以查閱群組日誌。</p>;
  return <View key={epoch+":"+user.userId} dashboard={dashboard} request={request}/>;
}
function View({dashboard,request}:{dashboard:boolean;request:Request}){
  const [groups,setGroups]=useState<Option[]>([]),[group,setGroup]=useState(""),[date,setDate]=useState(workDate);
  const [entries,setEntries]=useState<Journal[]>([]),[cursor,setCursor]=useState(""),[writing,setWriting]=useState(false);
  const [publishGroup,setPublishGroup]=useState("");
  const [revision,reload]=useState(0),[error,setError]=useState(""),[groupError,setGroupError]=useState(""),[busy,setBusy]=useState(true);
  const [detail,setDetail]=useState(""),[ready,setReady]=useState(false);
  const list=useLatestRequest();
  useEffect(()=>{
    const params=new URLSearchParams(location.search);
    const journalId=dashboard?"":params.get("journalId")||"";
    setDetail(journalId);setWriting(!dashboard && !journalId && params.get("compose")==="1");setReady(true);
  },[dashboard]);
  useEffect(()=>{
    const controller=new AbortController();setGroupError("");
    collect(cursor=>request((c,s)=>c.GET("/api/v1/groups",{signal:s,params:{query:{cursor:cursor||undefined}}}),{signal:controller.signal}))
      .then(gs=>{if(!controller.signal.aborted)setGroups(gs);}).catch(e=>{if(!controller.signal.aborted)setGroupError(errorMessage(e));});
    return ()=>controller.abort();
  },[request,revision]);
  function load(next=""){
    setBusy(true);setError("");setCursor("");if(!next)setEntries([]);
    void list.run(async signal=>{
      if(detail){const entry=await request((c,s)=>c.GET("/api/v1/journals/{journalId}",{signal:s,params:{path:{journalId:detail}}}),{signal});return {items:[entry],nextCursor:""};}
      return request((c,s)=>c.GET("/api/v1/journals",{signal:s,params:{query:{groupId:group||undefined,workDate:dashboard?date||undefined:undefined,cursor:next||undefined}}}),{signal});
    },page=>{setEntries(old=>next?[...old,...page.items]:page.items);setCursor(page.nextCursor);setBusy(false);},e=>{setEntries([]);setBusy(false);setError(errorMessage(e));});
  }
  useEffect(()=>{if(ready)load();return list.cancel;},[group,date,revision,detail,dashboard,ready,request]);
  function closeComposer(){
    const url=new URL(location.href);url.searchParams.delete("compose");window.history.replaceState(window.history.state,"",url);
    setWriting(false);
    setPublishGroup("");
  }
  const renderEntry = (j:Journal) => <PagePanel key={j.id} heading={j.workDate+" · "+(groups.find(g=>g.id===j.groupId)?.name||"群組日誌")}>
    {Object.entries(roles).map(([key,label])=><p key={key}>{label}：{j.personnel[key as keyof typeof roles].join("、")||"未填"}</p>)}
    {([["今日工作",j.todayItems],["明日計畫",j.tomorrowItems]] as const).map(([title,items])=><section key={title}><h3>{title}</h3>{items.length?<ul>{items.map((i,k)=><li key={k}>{i.name} · {"completionPercent" in i?i.completionPercent:i.expectedCompletionPercent}%
      {i.reference&&<a href={i.reference.type==="repository"?"/miniapp/repositories?selected="+i.reference.id:"/miniapp/projects?owner="+j.groupId+"&project="+i.reference.id}> · {i.reference.type==="repository"?"儲存庫":"專案"}</a>}</li>)}</ul>:<p>未填</p>}</section>)}
    {!detail&&<a href={"/miniapp/feed?journalId="+j.id}>查看原文</a>}
  </PagePanel>;
  return <><PagePanel heading={dashboard?"工作概況":"工作牆"}>
    {detail?<a href="/miniapp/feed">返回 Feed</a>:<label>群組 <select disabled={writing} value={group} onChange={e=>setGroup(e.target.value)}><option value="">全部可讀群組</option>{groups.map(g=><option key={g.id} value={g.id}>{g.name||g.id}</option>)}</select></label>}
    {dashboard&&<label>工作日期<Input type="date" value={date} onChange={e=>setDate(e.target.value)}/></label>}
    {(error||groupError)&&<><p role="alert">{error||groupError}</p><Button onClick={()=>reload(v=>v+1)}>重新讀取</Button></>}
  </PagePanel>
  {writing&&<PagePanel heading="發布群組">
    {publishGroup?<p>{groups.find(g=>g.id===publishGroup)?.name||publishGroup}</p>:
      <label>選擇發布群組 <select value="" onChange={e=>setPublishGroup(e.target.value)}>
        <option value="">請選擇群組</option>
        {groups.map(g=><option key={g.id} value={g.id}>{g.name||g.id}</option>)}
      </select></label>}
    <Button variant="outline" onClick={()=>{if(!publishGroup||window.confirm("放棄這份日誌？"))closeComposer();}}>取消填寫</Button>
  </PagePanel>}
  {writing&&publishGroup&&<Composer request={request} group={publishGroup} saved={()=>{closeComposer();reload(n=>n+1);}}/>}
  {busy&&<p role="status">正在讀取日誌……</p>}{!busy&&!error&&!entries.length&&<p>目前條件下沒有日誌。</p>}
  {dashboard?<GroupCharts entries={entries} groups={groups} renderEntry={renderEntry}/>:entries.map(renderEntry)}
  {cursor&&<Button disabled={busy} onClick={()=>load(cursor)}>載入更多</Button>}</>;
}
