import { listMemberGroups } from "@/business/group/memberGroups";
import dynamic from "next/dynamic";
import React, { useEffect, useState } from "react";
import { useSession, type Request } from "@/identity/Session";
import { Alert, AlertDescription } from "@/shared/ui/alert";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";
import { Empty, EmptyHeader, EmptyTitle } from "@/shared/ui/empty";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { FormField } from "@/shared/FormField";
import { PagePanel } from "@/shared/PagePanel";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";
import { collect } from "./pagination";
import type { components } from "@repo/api-client";
const GroupCharts = dynamic(() => import("./GroupCharts"), {
  loading: () => <ReadFeedback busy label="正在載入群組圖表……" />,
});
type Journal = components["schemas"]["Journal"];
const roles = { management: "管理職", control: "管制人員", workers: "施作人員" } as const;
const workDate = () => new Intl.DateTimeFormat("sv-SE", { timeZone: "Asia/Taipei" }).format(new Date());

type Option = {id:string;name:string};
const ComposeView = dynamic(() => import("./Composer").then(module => module.ComposeView), { loading: () => <ReadFeedback busy label="正在載入日誌表單……" /> });

type WorkspaceProps = {
  dashboard: boolean;
  journalId?: string;
  onCloseComposer?: () => void;
};

export function JournalWorkspace({dashboard, journalId = "", onCloseComposer}: WorkspaceProps) {
  const {user,epoch,request}=useSession();
  if(!user)return <p role="status">{onCloseComposer ? "請先登入以撰寫日誌。" : "請先登入以查閱群組日誌。"}</p>;
  const sessionKey = epoch + ":" + user.userId;
  return onCloseComposer
    ? <ComposeView key={sessionKey} request={request} close={onCloseComposer} />
    : <View key={sessionKey + ":" + journalId} dashboard={dashboard} request={request} detail={journalId} />;
}

function View({dashboard,request,detail}:{dashboard:boolean;request:Request;detail:string}){
  const [groups,setGroups]=useState<Option[]>([]),[group,setGroup]=useState(""),[date,setDate]=useState(workDate);
  const [entries,setEntries]=useState<Journal[]>([]),[cursor,setCursor]=useState("");
  const [revision,reload]=useState(0),[error,setError]=useState(""),[groupError,setGroupError]=useState(""),[busy,setBusy]=useState(true);
  const list=useLatestRequest();
  useEffect(()=>{
    const controller=new AbortController();setGroupError("");
    collect(cursor => listMemberGroups(request, controller.signal, cursor || undefined))
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
  useEffect(()=>{load();return list.cancel;},[group,date,revision,detail,dashboard,request]);
  const renderEntry = (j:Journal) => <PagePanel key={j.id} heading={j.workDate+" · "+(groups.find(g=>g.id===j.groupId)?.name||"群組日誌")}>
    {Object.entries(roles).map(([key,label])=><p key={key}>{label}：{j.personnel[key as keyof typeof roles].join("、")||"未填"}</p>)}
    {([["今日工作",j.todayItems],["明日計畫",j.tomorrowItems]] as const).map(([title,items])=><section key={title}><h3>{title}</h3>{items.length?<ul>{items.map((i,k)=><li key={k}>{i.name} · {"completionPercent" in i?i.completionPercent:i.expectedCompletionPercent}%
      {i.objectName && <> · {i.objectName}</>}{!!i.spaceNames?.length && <> · {i.spaceNames.join("、")}</>}
      {i.reference&&<a href={i.reference.type==="repository"?"/miniapp/repositories?selected="+i.reference.id:"/miniapp/projects?owner="+j.groupId+"&project="+i.reference.id}> · {i.reference.type==="repository"?"儲存庫":"專案"}</a>}</li>)}</ul>:<p>未填</p>}</section>)}
    {!detail&&<a href={"/miniapp/feed?journalId="+j.id}>查看原文</a>}
  </PagePanel>;
  return <><PagePanel heading={dashboard?"工作概況":"工作牆"}>
    {detail?<a href="/miniapp/feed">返回 Feed</a>:<FormField>群組 <NativeSelect value={group} onChange={e=>setGroup(e.target.value)}><NativeSelectOption value="">全部可讀群組</NativeSelectOption>{groups.map(g=><NativeSelectOption key={g.id} value={g.id}>{g.name||g.id}</NativeSelectOption>)}</NativeSelect></FormField>}
    {dashboard&&<FormField>工作日期<Input type="date" value={date} onChange={e=>setDate(e.target.value)}/></FormField>}
    {(error||groupError)&&<Alert variant="destructive"><AlertDescription>{error||groupError}<Button onClick={()=>reload(v=>v+1)}>重新讀取</Button></AlertDescription></Alert>}
  </PagePanel>
  <ReadFeedback busy={busy} label="正在讀取日誌……" />
  {!busy&&!error&&!groupError&&!entries.length&&<Empty><EmptyHeader><EmptyTitle>目前條件下沒有日誌。</EmptyTitle></EmptyHeader></Empty>}
  {dashboard?<GroupCharts entries={entries} groups={groups} renderEntry={renderEntry}/>:entries.map(renderEntry)}
  {cursor&&<Button disabled={busy} onClick={()=>load(cursor)}>載入更多</Button>}</>;
}
