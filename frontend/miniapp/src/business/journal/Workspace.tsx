import dynamic from "next/dynamic";
import { useEffect, useState } from "react";
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
import { Status } from "@/shared/Status";
import { collect } from "./pagination";
import type { components } from "@repo/api-client";
const RepositoryChart = dynamic(() => import("./GroupCharts"), {
  loading: () => <ReadFeedback busy label="正在載入儲存庫圖表……" />,
});
type Journal = components["schemas"]["Journal"];
const roles = { management: "管理職", control: "管制人員", workers: "施作人員" } as const;
const workDate = () => new Intl.DateTimeFormat("sv-SE", { timeZone: "Asia/Taipei" }).format(new Date());

type Option = {id:string;name:string};
const ComposeView = dynamic(() => import("./Composer").then(module => module.ComposeView), { loading: () => <ReadFeedback busy label="正在載入日誌表單……" /> });

type WorkspaceProps = {
  dashboard: boolean;
  journalId?: string;
  repositoryId?: string;
  onSelectRepository?: (id: string) => void;
  onCloseComposer?: () => void;
};

export function JournalWorkspace({dashboard, journalId = "", repositoryId = "", onSelectRepository, onCloseComposer}: WorkspaceProps) {
  const {user,epoch,request}=useSession();
  if(!user)return <p role="status">{onCloseComposer ? "請先登入以撰寫日誌。" : "請先登入以查閱儲存庫日誌。"}</p>;
  const sessionKey = epoch + ":" + user.userId;
  return onCloseComposer
    ? <ComposeView key={sessionKey + ":" + repositoryId} request={request} repositoryId={repositoryId} onSelectRepository={onSelectRepository} close={onCloseComposer} />
    : <View key={sessionKey + ":" + journalId + ":" + repositoryId} dashboard={dashboard} request={request} detail={journalId} repositoryId={repositoryId} />;
}

function CopyJournalLink({id, repositoryId}: {id:string; repositoryId:string}) {
  const [status,setStatus]=useState("");
  async function copy() {
    const url = new URL("/miniapp/feed", location.origin);
    url.searchParams.set("journalId", id);
    url.searchParams.set("repositoryId", repositoryId);
    try {
      await navigator.clipboard.writeText(url.toString());
      setStatus("已複製日誌連結，可貼到要分享的群組。");
    } catch {
      setStatus("無法複製，請重試或從網址列複製連結。");
    }
  }
  return <div className="flex flex-wrap items-center gap-2">
    <Button variant="outline" onClick={() => void copy()}>複製日誌連結</Button>
    <Status role="status">{status}</Status>
  </div>;
}

function View({dashboard,request,detail,repositoryId: initialRepositoryId}:{dashboard:boolean;request:Request;detail:string;repositoryId:string}){
  const [repositoryId,setRepositoryId]=useState(initialRepositoryId);
  const [repositories,setRepositories]=useState<Option[] | null>(null);
  const {run:readCatalog,cancel:cancelCatalog}=useLatestRequest();
  const [date,setDate]=useState(workDate);
  const [entries,setEntries]=useState<Journal[]>([]),[cursor,setCursor]=useState("");
  const [revision,reload]=useState(0),[error,setError]=useState(""),[catalogError,setCatalogError]=useState(""),[busy,setBusy]=useState(true);
  const list=useLatestRequest();
  useEffect(()=>{
    if (detail) return;
    setCatalogError("");
    void readCatalog(signal=>collect(cursor=>request((c,s)=>c.GET("/api/v1/repositories",{signal:s,params:{query:{cursor:cursor||undefined}}}),{signal})),
      items=>{
        setRepositories(items);
        setRepositoryId(previous=>previous || items[0]?.id || "");
      },error=>setCatalogError(errorMessage(error)));
    return cancelCatalog;
  },[request,revision,detail,readCatalog,cancelCatalog]);
  function load(next=""){
    if (!detail && !repositoryId) {setBusy(false);return;}
    setBusy(true);setError("");setCursor("");if(!next)setEntries([]);
    void list.run(async signal=>{
      if(detail){const entry=await request((c,s)=>c.GET("/api/v1/journals/{journalId}",{signal:s,params:{path:{journalId:detail}}}),{signal});return {items:[entry],nextCursor:""};}
      return request((c,s)=>c.GET("/api/v1/journals",{signal:s,params:{query:{includeProgress:dashboard||undefined,repositoryId,workDate:date||undefined,cursor:next||undefined}}}),{signal});
    },page=>{setEntries(old=>next?[...old,...page.items]:page.items);setCursor(page.nextCursor);setBusy(false);},e=>{setEntries([]);setBusy(false);setError(errorMessage(e));});
  }
  useEffect(()=>{load();return list.cancel;},[date,revision,detail,dashboard,request,repositoryId]);
  const repositoryName = repositories?.find(repo=>repo.id===repositoryId)?.name;
  const renderEntry = (j:Journal) => <PagePanel key={j.id} heading={j.workDate+" · "+(repositoryName||"工程日誌")}>
    {Object.entries(roles).map(([key,label])=><p key={key}>{label}：{j.personnel[key as keyof typeof roles].join("、")||"未填"}</p>)}
    {([["今日工作",j.todayItems],["明日計畫",j.tomorrowItems]] as const).map(([title,items])=><section key={title}><h3>{title}</h3>{items.length?<ul>{items.map((i,k)=><li key={k}>{i.name} · {"completionPercent" in i?i.completionPercent:i.expectedCompletionPercent}%
      {i.objectName && <> · {i.objectName}</>}{!!i.spaceNames?.length && <> · {i.spaceNames.join("、")}</>}
      {i.reference&&(i.reference.type === "repository" || j.groupId)&&<a href={i.reference.type==="repository"?"/miniapp/repositories?selected="+i.reference.id:"/miniapp/projects?owner="+j.groupId+"&project="+i.reference.id}> · {i.reference.type==="repository"?"儲存庫":"專案"}</a>}</li>)}</ul>:<p>未填</p>}</section>)}
    {detail && j.repositoryId && <CopyJournalLink id={j.id} repositoryId={j.repositoryId} />}
    {!detail&&<a href={"/miniapp/feed?" + new URLSearchParams({journalId:j.id,...(repositoryId?{repositoryId}:{})})}>查看原文</a>}
  </PagePanel>;
  return <><PagePanel heading={repositoryId?"工程日誌":dashboard?"工作概況":"工作牆"}>
    {detail?<a href={"/miniapp/feed" + (repositoryId ? "?repositoryId=" + encodeURIComponent(repositoryId) : "")}>返回日誌</a>:repositoryId?<Button asChild><a href={"/miniapp/feed?" + new URLSearchParams({repositoryId,compose:"1"})}>寫日誌</a></Button>:null}
    {!detail && (repositories === null
      ? !catalogError && <ReadFeedback busy label="正在讀取儲存庫……" />
      : repositories.length ? <FormField>儲存庫 <NativeSelect value={repositoryId} onChange={event=>setRepositoryId(event.target.value)}>
          {repositoryId && !repositories.some(repo=>repo.id===repositoryId) && <NativeSelectOption value={repositoryId}>{repositoryId}</NativeSelectOption>}
          {repositories.map(repo=><NativeSelectOption key={repo.id} value={repo.id}>{repo.name||repo.id}</NativeSelectOption>)}
        </NativeSelect></FormField> : <p>目前沒有可用的儲存庫。</p>)}
    {(dashboard||!!repositoryId)&&!detail&&<FormField>工作日期<Input type="date" value={date} onChange={e=>setDate(e.target.value)}/></FormField>}
    {(error||catalogError)&&<Alert variant="destructive"><AlertDescription>{error||catalogError}<Button onClick={()=>reload(v=>v+1)}>重新讀取</Button></AlertDescription></Alert>}
  </PagePanel>
  <ReadFeedback busy={busy} label="正在讀取日誌……" />
  {(detail||repositoryId)&&!busy&&!error&&!catalogError&&!entries.length&&<Empty><EmptyHeader><EmptyTitle>目前條件下沒有日誌。</EmptyTitle></EmptyHeader></Empty>}
  {dashboard?<RepositoryChart entries={entries} name={repositoryName||repositoryId} renderEntry={renderEntry}/>:entries.map(renderEntry)}
  {cursor&&<Button disabled={busy} onClick={()=>load(cursor)}>載入更多</Button>}</>;
}
