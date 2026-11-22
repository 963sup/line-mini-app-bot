import { listMemberGroups } from "@/business/group/memberGroups";
import React, { useEffect, useState, useRef } from "react";
import { type Request } from "@/identity/Session";
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
import { rejectedAttempt, errorMessage, useLatestRequest } from "@/shared/http/lifecycle";
import { WorkContext, useWorkOptions, type WorkOptionsLoader } from "./WorkContext";
import type { components } from "@repo/api-client";
import { collect } from "./pagination";
type Draft = components["schemas"]["JournalCreation"];
const roles = { management: "管理職", control: "管制人員", workers: "施作人員" } as const;
const workDate = () => new Intl.DateTimeFormat("sv-SE", { timeZone: "Asia/Taipei" }).format(new Date());
type Option = {id:string;name:string};
type Target = { type: "repository"; id: string; name: string };
function Rows({title,rows,change,targets,request,workOptions,onPending}:{title:string;rows:string[];change:(v:string[])=>void;targets:Target[];request:Request;workOptions:WorkOptionsLoader;onPending:(id:string,value:boolean)=>void}) {
  const options = React.useMemo(() => [
    {label: "無關聯", value: ""},
    ...targets.map(t => ({label: `儲存庫 · ${t.name}`, value: t.type + t.id})),
  ], [targets]);
  return <div className="grid min-w-0 grid-cols-1 gap-3">
    {rows.map((id, i) => <WorkRow key={id} id={id} index={i} options={options} targets={targets} request={request} workOptions={workOptions} onPending={onPending} remove={() => change(rows.filter(k => k !== id))}/>)}
    <Button type="button" variant="outline" disabled={rows.length >= 100} onClick={() => change([...rows, crypto.randomUUID()])}>新增{title}</Button>
  </div>;
}
function WorkRow({id,index:i,options,targets,request,workOptions,remove,onPending}:{id:string;index:number;options:{label:string;value:string}[];targets:Target[];request:Request;workOptions:WorkOptionsLoader;remove:()=>void;onPending:(id:string,value:boolean)=>void}) {
  const [reference,setReference]=useState<(typeof options)[number] | null>(null), [locked,setLocked]=useState(false);
  const pending = useRef(false);
  const lock = (value:boolean) => {pending.current=value;setLocked(value);onPending(id,value);};
  const repository = targets.find(t => t.type === "repository" && t.type+t.id === reference?.value);
  return <FieldSet className="min-w-0 rounded-lg border p-3">
      <FieldLegend>第 {i + 1} 項</FieldLegend>
      <Field><FieldLabel htmlFor={id+"name"}>工作名稱</FieldLabel>
        <Input id={id+"name"} name={id+"name"} required maxLength={120} /></Field>
      <Field><FieldLabel htmlFor={id+"percent"}>累計完成度（%）</FieldLabel>
        <Input id={id+"percent"} name={id+"percent"} required type="number" min={0} max={100} step={1} /></Field>
      <Field><FieldLabel htmlFor={id+"ref"}>工作關聯</FieldLabel>
        <Combobox<(typeof options)[number]> items={options} name={id+"ref"} value={reference} onValueChange={value=>{if(!pending.current)setReference(value);}} disabled={locked} itemToStringValue={option => option.value}>
          <ComboboxInput disabled={locked} id={id+"ref"} placeholder="選擇目前儲存庫（可無關聯）" />
          <ComboboxContent>
            <ComboboxEmpty>沒有符合的工作關聯</ComboboxEmpty>
            <ComboboxList>{(option: (typeof options)[number]) =>
              <ComboboxItem key={option.value} value={option}>{option.label}</ComboboxItem>}
            </ComboboxList>
          </ComboboxContent>
        </Combobox>
      </Field>
      {repository && <WorkContext key={repository.id} id={id} repositoryId={repository.id} request={request} options={workOptions} lock={lock}/>}
      <Button type="button" variant="outline" disabled={locked} onClick={()=>{if(!pending.current)remove();}}>移除第 {i+1} 項</Button>
    </FieldSet>;
}
function AttendanceReference({request,date}: {request:Request;date:string}) {
  const lookup = useLatestRequest();
  const groupsRead = useLatestRequest();
  const [groups,setGroups] = useState<Option[] | null>(null);
  const [group,setGroup] = useState("");
  const [notice,setNotice] = useState("");
  const [groupNotice,setGroupNotice] = useState("");
  function open() {
    setGroupNotice("正在讀取群組……");
    void groupsRead.run(signal=>collect(cursor=>listMemberGroups(request,signal,cursor||undefined)),
      items=>{setGroups(items);setGroupNotice(items.length ? "" : "目前沒有可查詢的群組。");},
      ()=>setGroupNotice("暫時無法載入群組，可重試或繼續填寫。"));
  }
  useEffect(()=>{
    setNotice("");
    if (!group || !date) return;
    setNotice("正在查詢出勤參考……");
    void lookup.run(
      signal=>request((c,s)=>c.GET("/api/v1/attendance/daily-headcount",{signal:s,params:{query:{group,date}}}),{signal}),
      data=>setNotice(`當日開始班次：${data.count} 人（僅供參考，更新於 ${new Date(data.asOf).toLocaleString("zh-TW",{timeZone:"Asia/Taipei"})}）`),
      ()=>setNotice("暫時無法取得出勤資料，仍可繼續填寫。"),
    );
    return lookup.cancel;
  },[request,date,group,lookup.run,lookup.cancel]);
  return <div className="grid gap-2">
    {groups === null ? <Button type="button" variant="outline" onClick={open}>查看出勤參考（選填）</Button>
      : groups.length > 0 && <FormField>出勤參考群組 <NativeSelect value={group} onChange={event=>setGroup(event.target.value)}>
        <NativeSelectOption value="">不顯示出勤參考</NativeSelectOption>
        {groups.map(item=><NativeSelectOption key={item.id} value={item.id}>{item.name||item.id}</NativeSelectOption>)}
      </NativeSelect></FormField>}
    {(groupNotice || notice) && <Status role="status">{groupNotice || notice}</Status>}
  </div>;
}

function Composer({request,repository,saved,onPending,isBlocked}:{request:Request;repository:Option;saved:()=>void;onPending:(id:string,value:boolean)=>void;isBlocked:()=>boolean}) {
  const workOptions = useWorkOptions(request);
  const [step, setStep] = useState("people");
  const [date,setDate] = useState(workDate);
  const [today,setToday]=useState(["first"]),[tomorrow,setTomorrow]=useState<string[]>([]);
  const targets: Target[] = [{type:"repository", id:repository.id, name:repository.name}];
  const [error,setError]=useState(""),[busy,setBusy]=useState(false);
  const publishAttempt=useRef<{body:Draft;key:string} | null>(null),submitting=useRef(false);
  const write=useLatestRequest();
  const [uncertain,setUncertain]=useState(false);
  function publish(body: Draft) {
    const retry = !!publishAttempt.current;
    const attempt = publishAttempt.current ?? {body,key:crypto.randomUUID()};
    publishAttempt.current = attempt;
    submitting.current=true;setBusy(true);setError("");onPending("publish",true);
    void write.run(signal=>request((c,s)=>c.POST("/api/v1/journals",{signal:s,params:{header:{"Idempotency-Key":attempt.key}},body:attempt.body}),{signal}),
      () => {onPending("publish",false);saved();}, e=>{
        submitting.current=false;setBusy(false);
        const unknown=retry || !rejectedAttempt(e);
        if (!unknown) publishAttempt.current=null;
        setUncertain(unknown);onPending("publish",unknown);
        setError(errorMessage(e)+(unknown?" 請沿用原內容重試。":""));
      });
  }
  return <PagePanel heading={"寫日誌 · " + repository.name}><Questionnaire item={step} onItemChange={setStep} onSubmitCapture={event=>{
    // Each step contains several native fields. Validate the active group together.
    event.preventDefault();
    event.stopPropagation();
    if (submitting.current) return;
    if (publishAttempt.current) {publish(publishAttempt.current.body);return;}
    if (isBlocked()) {setError("請先完成或重試空間建立，再儲存日誌。");return;}
    const invalid = event.currentTarget.querySelector<HTMLInputElement>("fieldset[data-active] input:invalid");
    if (invalid) { invalid.reportValidity(); invalid.focus(); return; }
    if (step === "people") { setStep("today"); return; }
    if (step === "today") { setStep("tomorrow"); return; }
    const form=new FormData(event.currentTarget),get=(k:string)=>String(form.get(k)??"").trim();
    const people=(k:string)=>get(k).split(/[、，,\n]/).map(s=>s.trim()).filter(Boolean);
    const row=(id:string)=>{const t=targets.find(t=>t.type+t.id===get(id+"ref"));return {name:get(id+"name"),reference:t?{type:t.type,id:t.id}:undefined, ...(t?.type === "repository" ? {objectId:get(id+"object") || undefined,spaceIds:form.getAll(id+"space").map(String)} : {})};};
    const body:Draft={repositoryId:repository.id,workDate:get("date"), personnel:{management:people("management"),control:people("control"),workers:people("workers")},
      todayItems:today.map(id=>({...row(id),completionPercent:Number(get(id+"percent"))})),
      tomorrowItems:tomorrow.map(id=>({...row(id),expectedCompletionPercent:Number(get(id+"percent"))}))};
    if(!today.length&&!tomorrow.length){setError("請新增至少一項工作。");return;}
    publish(body);
  }}>
    <QuestionnaireProgress aria-label="日誌填寫進度" aria-valuetext={step === "people" ? "日期與人員" : step === "today" ? "今日工作" : "明日計畫與儲存"}>第 {step === "people" ? 1 : step === "today" ? 2 : 3} 步，共 3 步</QuestionnaireProgress>
    <FieldSet disabled={busy || uncertain} className="grid min-w-0 grid-cols-1 gap-3">
    <QuestionnaireItem name="people"><QuestionnaireTitle>日期與人員</QuestionnaireTitle>
    <FormField>工作日期<Input required type="date" name="date" value={date} onChange={event=>setDate(event.target.value)}/></FormField>
    <AttendanceReference request={request} date={date} />
    {Object.entries(roles).map(([key,label])=><FormField key={key}>{label}<Input name={key} placeholder="多位姓名以頓號分隔"/></FormField>)}
    </QuestionnaireItem>
    <QuestionnaireItem name="today"><QuestionnaireTitle>今日工作</QuestionnaireTitle>
    <Rows title="今日工作" rows={today} change={setToday} targets={targets} request={request} workOptions={workOptions} onPending={onPending}/>
    </QuestionnaireItem>
    <QuestionnaireItem name="tomorrow"><QuestionnaireTitle>明日計畫與儲存</QuestionnaireTitle>
    <Rows title="明日計畫" rows={tomorrow} change={setTomorrow} targets={targets} request={request} workOptions={workOptions} onPending={onPending}/>
    </QuestionnaireItem>
    {error && <Alert variant="destructive"><AlertDescription>{error}</AlertDescription></Alert>}
    <QuestionnaireActions>
      <QuestionnairePrevious>上一步</QuestionnairePrevious>
      {!uncertain && <Button type="submit" className="col-start-3">{step === "tomorrow" ? (busy ? "儲存中" : "儲存日誌") : "下一步"}</Button>}
    </QuestionnaireActions>
  </FieldSet>{uncertain && <Button type="submit" disabled={busy}>儲存日誌</Button>}</Questionnaire></PagePanel>;
}

export function ComposeView({request, close, repositoryId = "", onSelectRepository}: {request: Request; close: () => void; repositoryId?: string; onSelectRepository?: (id: string) => void}) {
  const pendingRows = useRef(new Set<string>());
  const [blocked,setBlocked] = useState(false);
  const onPending = React.useCallback((id:string,value:boolean)=>{
    if(value)pendingRows.current.add(id);else pendingRows.current.delete(id);
    setBlocked(pendingRows.current.size>0);
  },[]);
  const isBlocked = React.useCallback(()=>pendingRows.current.size>0,[]);
  const [repositories, setRepositories] = useState<Option[] | null>(null);
  const [error, setError] = useState("");
  const cancelButton = useRef<HTMLButtonElement>(null);
  const discard = useDiscardChanges(!!repositoryId, () => cancelButton.current?.focus());
  const [revision, retry] = useState(0);

  useEffect(() => {
    const controller = new AbortController();
    setError("");
    const options = repositoryId
      ? request((client, signal) => client.GET("/api/v1/repositories/{id}", {
          signal, params: {path: {id: repositoryId}},
        }), {signal: controller.signal}).then(item => [item])
      : collect(cursor => request((client, signal) => client.GET("/api/v1/repositories", {
            signal, params: {query: {cursor: cursor || undefined}},
          }), {signal: controller.signal}));
    options.then(
      items => { if (!controller.signal.aborted) setRepositories(items); },
      reason => { if (!controller.signal.aborted) setError(errorMessage(reason)); },
    );
    return () => controller.abort();
  }, [request, revision, repositoryId]);

  return <>
    <PagePanel heading="施工儲存庫">
      <div className="flex flex-wrap items-center gap-3">
        {error ? <Alert variant="destructive"><AlertDescription>
          {error}
          {repositoryId && <Button asChild variant="outline"><a href="/miniapp/feed?compose=1">改選儲存庫</a></Button>}
          <Button type="button" variant="outline" onClick={() => retry(value => value + 1)}>重新讀取</Button>
        </AlertDescription></Alert> : !repositories ? <ReadFeedback busy label="正在讀取儲存庫……" />
          : !repositories.length ? <Empty><EmptyHeader><EmptyTitle>目前沒有可選擇的儲存庫。</EmptyTitle></EmptyHeader></Empty>
          : repositoryId ? <p>{repositories[0]?.name || repositoryId}</p>
          : <FormField>選擇儲存庫 <NativeSelect value="" onChange={event => {
              if (!event.target.value) return;
              onSelectRepository?.(event.target.value);
            }}>
            <NativeSelectOption value="">請選擇儲存庫</NativeSelectOption>
            {repositories.map(repo => <NativeSelectOption key={repo.id} value={repo.id}>{repo.name || repo.id}</NativeSelectOption>)}
          </NativeSelect></FormField>}
        <Button ref={cancelButton} type="button" variant="outline" disabled={blocked} onClick={() => {if(!isBlocked())discard.navigate(()=>{if(!isBlocked())close();});}}>取消填寫</Button>
        {discard.dialog}
      </div>
    </PagePanel>
    {repositoryId && repositories?.[0] ? <Composer request={request} repository={repositories[0]} saved={close} onPending={onPending} isBlocked={isBlocked} />
      : null}
  </>;
}
