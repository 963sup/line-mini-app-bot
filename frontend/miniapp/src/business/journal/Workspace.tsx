import dynamic from "next/dynamic";
import React, { useEffect, useState, useRef } from "react";
import { useSession, type Request } from "@/identity/Session";
import { Alert, AlertDescription } from "@/shared/ui/alert";
import { Combobox, ComboboxContent, ComboboxEmpty, ComboboxInput, ComboboxItem, ComboboxList } from "@/shared/ui/combobox";
import { Field, FieldLabel, FieldLegend, FieldSet } from "@/shared/ui/field";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";
import { Empty, EmptyHeader, EmptyTitle } from "@/shared/ui/empty";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { FormField } from "@/shared/FormField";
import { Status } from "@/shared/Status";
import { useDiscardChanges } from "@/shared/useDiscardChanges";
import { PagePanel } from "@/shared/PagePanel";
import { Button } from "@/shared/ui/button";
import { Questionnaire, QuestionnaireActions, QuestionnaireItem, QuestionnairePrevious, QuestionnaireProgress, QuestionnaireTitle } from "@/shared/ui/questionnaire";
import { Input } from "@/shared/ui/input";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";
import type { components } from "@repo/api-client";
const GroupCharts = dynamic(() => import("./GroupCharts"), {
  loading: () => <ReadFeedback busy label="正在載入群組圖表……" />,
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
  const options = React.useMemo(() => [
    {label: "無關聯", value: ""},
    ...targets.map(t => ({label: `${t.type === "project" ? "專案" : "儲存庫"} · ${t.name}`, value: t.type + t.id})),
  ], [targets]);
  return <div className="grid gap-3">
    {rows.map((id, i) => <FieldSet key={id} className="rounded-lg border p-3">
      <FieldLegend>第 {i + 1} 項</FieldLegend>
      <Field><FieldLabel htmlFor={id+"name"}>工作名稱</FieldLabel>
        <Input id={id+"name"} name={id+"name"} required maxLength={120} /></Field>
      <Field><FieldLabel htmlFor={id+"percent"}>累計完成度（%）</FieldLabel>
        <Input id={id+"percent"} name={id+"percent"} required type="number" min={0} max={100} step={1} /></Field>
      <Field><FieldLabel htmlFor={id+"ref"}>工作關聯</FieldLabel>
        <Combobox<(typeof options)[number]> items={options} name={id+"ref"} itemToStringValue={option => option.value}>
          <ComboboxInput id={id+"ref"} placeholder="搜尋專案或儲存庫（可無關聯）" />
          <ComboboxContent>
            <ComboboxEmpty>沒有符合的工作關聯</ComboboxEmpty>
            <ComboboxList>{(option: (typeof options)[number]) =>
              <ComboboxItem key={option.value} value={option}>{option.label}</ComboboxItem>}
            </ComboboxList>
          </ComboboxContent>
        </Combobox>
      </Field>
      <Button type="button" variant="outline" onClick={() => change(rows.filter(k => k !== id))}>移除第 {i+1} 項</Button>
    </FieldSet>)}
    <Button type="button" variant="outline" disabled={rows.length >= 100} onClick={() => change([...rows, crypto.randomUUID()])}>新增{title}</Button>
  </div>;
}
function Composer({request,group,saved}:{request:Request;group:string;saved:()=>void}) {
  const [step, setStep] = useState("people");
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
  return <PagePanel heading="寫日誌"><Questionnaire item={step} onItemChange={setStep} onSubmitCapture={event=>{
    // Each step contains several native fields. Validate the active group together.
    event.preventDefault();
    event.stopPropagation();
    if (submitting.current) return;
    const invalid = event.currentTarget.querySelector<HTMLInputElement>("fieldset[data-active] input:invalid");
    if (invalid) { invalid.reportValidity(); invalid.focus(); return; }
    if (step === "people") { setStep("today"); return; }
    if (step === "today") { setStep("tomorrow"); return; }
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
  }}>
    <QuestionnaireProgress aria-label="日誌填寫進度" aria-valuetext={step === "people" ? "日期與人員" : step === "today" ? "今日工作" : "明日計畫與發布"}>第 {step === "people" ? 1 : step === "today" ? 2 : 3} 步，共 3 步</QuestionnaireProgress>
    <FieldSet disabled={busy} className="grid gap-3">
    <QuestionnaireItem name="people"><QuestionnaireTitle>日期與人員</QuestionnaireTitle>
    <FormField>工作日期<Input required type="date" name="date" defaultValue={workDate()}/></FormField>
    {Object.entries(roles).map(([key,label])=><FormField key={key}>{label}<Input name={key} placeholder="多位姓名以頓號分隔"/></FormField>)}
    </QuestionnaireItem>
    <QuestionnaireItem name="today"><QuestionnaireTitle>今日工作</QuestionnaireTitle>
    <Rows title="今日工作" rows={today} change={setToday} targets={targets}/>
    </QuestionnaireItem>
    <QuestionnaireItem name="tomorrow"><QuestionnaireTitle>明日計畫與發布</QuestionnaireTitle>
    <Rows title="明日計畫" rows={tomorrow} change={setTomorrow} targets={targets}/>
    </QuestionnaireItem>
    {notice && step !== "people" && <Status>{notice}</Status>}
    {error && <Alert variant="destructive"><AlertDescription>{error}</AlertDescription></Alert>}
    <QuestionnaireActions>
      <QuestionnairePrevious>上一步</QuestionnairePrevious>
      <Button type="submit" className="col-start-3">{step === "tomorrow" ? (busy ? "發布中" : "發布日誌") : "下一步"}</Button>
    </QuestionnaireActions>
  </FieldSet></Questionnaire></PagePanel>;
}

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

function ComposeView({request, close}: {request: Request; close: () => void}) {
  const [groups, setGroups] = useState<Option[] | null>(null);
  const [publishGroup, setPublishGroup] = useState("");
  const [error, setError] = useState("");
  const cancelButton = useRef<HTMLButtonElement>(null);
  const discard = useDiscardChanges(!!publishGroup, () => cancelButton.current?.focus());
  const [revision, retry] = useState(0);

  useEffect(() => {
    const controller = new AbortController();
    setError("");
    collect(cursor => request((client, signal) => client.GET("/api/v1/groups", {
      signal, params: {query: {cursor: cursor || undefined}},
    }), {signal: controller.signal})).then(
      items => { if (!controller.signal.aborted) setGroups(items); },
      reason => { if (!controller.signal.aborted) setError(errorMessage(reason)); },
    );
    return () => controller.abort();
  }, [request, revision]);

  return <>
    <PagePanel heading="發布群組">
      <div className="flex flex-wrap items-center gap-3">
        {error ? <Alert variant="destructive"><AlertDescription>
          {error}
          <Button type="button" variant="outline" onClick={() => retry(value => value + 1)}>重新讀取</Button>
        </AlertDescription></Alert> : !groups ? <ReadFeedback busy label="正在讀取群組……" />
          : !groups.length ? <Empty><EmptyHeader><EmptyTitle>目前沒有可選擇的群組。</EmptyTitle></EmptyHeader></Empty>
          : publishGroup ? <p>{groups.find(group => group.id === publishGroup)?.name || publishGroup}</p>
          : <FormField>選擇發布群組 <NativeSelect value="" onChange={event => setPublishGroup(event.target.value)}>
            <NativeSelectOption value="">請選擇群組</NativeSelectOption>
            {groups.map(group => <NativeSelectOption key={group.id} value={group.id}>{group.name || group.id}</NativeSelectOption>)}
          </NativeSelect></FormField>}
        <Button ref={cancelButton} type="button" variant="outline" onClick={() => discard.navigate(close)}>取消填寫</Button>
        {discard.dialog}
      </div>
    </PagePanel>
    {publishGroup && <Composer request={request} group={publishGroup} saved={close} />}
  </>;
}

function View({dashboard,request,detail}:{dashboard:boolean;request:Request;detail:string}){
  const [groups,setGroups]=useState<Option[]>([]),[group,setGroup]=useState(""),[date,setDate]=useState(workDate);
  const [entries,setEntries]=useState<Journal[]>([]),[cursor,setCursor]=useState("");
  const [revision,reload]=useState(0),[error,setError]=useState(""),[groupError,setGroupError]=useState(""),[busy,setBusy]=useState(true);
  const list=useLatestRequest();
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
  useEffect(()=>{load();return list.cancel;},[group,date,revision,detail,dashboard,request]);
  const renderEntry = (j:Journal) => <PagePanel key={j.id} heading={j.workDate+" · "+(groups.find(g=>g.id===j.groupId)?.name||"群組日誌")}>
    {Object.entries(roles).map(([key,label])=><p key={key}>{label}：{j.personnel[key as keyof typeof roles].join("、")||"未填"}</p>)}
    {([["今日工作",j.todayItems],["明日計畫",j.tomorrowItems]] as const).map(([title,items])=><section key={title}><h3>{title}</h3>{items.length?<ul>{items.map((i,k)=><li key={k}>{i.name} · {"completionPercent" in i?i.completionPercent:i.expectedCompletionPercent}%
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
