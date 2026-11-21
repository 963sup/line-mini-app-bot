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
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";
import { WorkContext, useWorkOptions, type WorkOptionsLoader } from "./WorkContext";
import type { components } from "@repo/api-client";
import { collect } from "./pagination";
type Draft = components["schemas"]["JournalCreation"];
const roles = { management: "管理職", control: "管制人員", workers: "施作人員" } as const;
const workDate = () => new Intl.DateTimeFormat("sv-SE", { timeZone: "Asia/Taipei" }).format(new Date());
type Option = {id:string;name:string};
type Target = components["schemas"]["JournalReferenceTarget"];
function Rows({title,rows,change,targets,request,workOptions,onPending}:{title:string;rows:string[];change:(v:string[])=>void;targets:Target[];request:Request;workOptions:WorkOptionsLoader;onPending:(id:string,value:boolean)=>void}) {
  const options = React.useMemo(() => [
    {label: "無關聯", value: ""},
    ...targets.map(t => ({label: `${t.type === "project" ? "專案" : "儲存庫"} · ${t.name}`, value: t.type + t.id})),
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
          <ComboboxInput disabled={locked} id={id+"ref"} placeholder="搜尋專案或儲存庫（可無關聯）" />
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
function Composer({request,group,saved,onPending,isBlocked}:{request:Request;group:string;saved:()=>void;onPending:(id:string,value:boolean)=>void;isBlocked:()=>boolean}) {
  const workOptions = useWorkOptions(request);
  const [step, setStep] = useState("people");
  const [today,setToday]=useState(["first"]),[tomorrow,setTomorrow]=useState<string[]>([]);
  const [targets,setTargets]=useState<Target[]>([]),[notice,setNotice]=useState(""),[error,setError]=useState(""),[busy,setBusy]=useState(false);
  const publishAttempt=useRef({serializedBody:"",idempotencyKey:""}),submitting=useRef(false);
  const write=useLatestRequest();
  useEffect(()=>{
    const controller=new AbortController(), options={signal:controller.signal};
    collect(cursor=>request((client,signal)=>client.GET("/api/v1/groups/{groupId}/journal-reference-targets", {
      signal,params:{path:{groupId:group},query:{cursor:cursor||undefined}}
    }),options)).then(items=>{
      if(!controller.signal.aborted)setTargets(items);
    }).catch(e=>{if(!controller.signal.aborted)setNotice("關聯清單無法載入："+errorMessage(e));});
    return ()=>controller.abort();
  },[group,request]);
  return <PagePanel heading="寫日誌"><Questionnaire item={step} onItemChange={setStep} onSubmitCapture={event=>{
    // Each step contains several native fields. Validate the active group together.
    event.preventDefault();
    event.stopPropagation();
    if (submitting.current) return;
    if (isBlocked()) {setError("請先完成或重試空間建立，再發布日誌。");return;}
    const invalid = event.currentTarget.querySelector<HTMLInputElement>("fieldset[data-active] input:invalid");
    if (invalid) { invalid.reportValidity(); invalid.focus(); return; }
    if (step === "people") { setStep("today"); return; }
    if (step === "today") { setStep("tomorrow"); return; }
    const form=new FormData(event.currentTarget),get=(k:string)=>String(form.get(k)??"").trim();
    const people=(k:string)=>get(k).split(/[、，,\n]/).map(s=>s.trim()).filter(Boolean);
    const row=(id:string)=>{const t=targets.find(t=>t.type+t.id===get(id+"ref"));return {name:get(id+"name"),reference:t?{type:t.type,id:t.id}:undefined, ...(t?.type === "repository" ? {objectId:get(id+"object") || undefined,subobjectId:get(id+"subobject") || undefined,spaceIds:form.getAll(id+"space").map(String)} : {})};};
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
    <FieldSet disabled={busy} className="grid min-w-0 grid-cols-1 gap-3">
    <QuestionnaireItem name="people"><QuestionnaireTitle>日期與人員</QuestionnaireTitle>
    <FormField>工作日期<Input required type="date" name="date" defaultValue={workDate()}/></FormField>
    {Object.entries(roles).map(([key,label])=><FormField key={key}>{label}<Input name={key} placeholder="多位姓名以頓號分隔"/></FormField>)}
    </QuestionnaireItem>
    <QuestionnaireItem name="today"><QuestionnaireTitle>今日工作</QuestionnaireTitle>
    <Rows title="今日工作" rows={today} change={setToday} targets={targets} request={request} workOptions={workOptions} onPending={onPending}/>
    </QuestionnaireItem>
    <QuestionnaireItem name="tomorrow"><QuestionnaireTitle>明日計畫與發布</QuestionnaireTitle>
    <Rows title="明日計畫" rows={tomorrow} change={setTomorrow} targets={targets} request={request} workOptions={workOptions} onPending={onPending}/>
    </QuestionnaireItem>
    {notice && step !== "people" && <Status>{notice}</Status>}
    {error && <Alert variant="destructive"><AlertDescription>{error}</AlertDescription></Alert>}
    <QuestionnaireActions>
      <QuestionnairePrevious>上一步</QuestionnairePrevious>
      <Button type="submit" className="col-start-3">{step === "tomorrow" ? (busy ? "發布中" : "發布日誌") : "下一步"}</Button>
    </QuestionnaireActions>
  </FieldSet></Questionnaire></PagePanel>;
}

export function ComposeView({request, close}: {request: Request; close: () => void}) {
  const pendingRows = useRef(new Set<string>());
  const [blocked,setBlocked] = useState(false);
  const onPending = React.useCallback((id:string,value:boolean)=>{
    if(value)pendingRows.current.add(id);else pendingRows.current.delete(id);
    setBlocked(pendingRows.current.size>0);
  },[]);
  const isBlocked = React.useCallback(()=>pendingRows.current.size>0,[]);
  const [groups, setGroups] = useState<Option[] | null>(null);
  const [publishGroup, setPublishGroup] = useState("");
  const [error, setError] = useState("");
  const cancelButton = useRef<HTMLButtonElement>(null);
  const discard = useDiscardChanges(!!publishGroup, () => cancelButton.current?.focus());
  const [revision, retry] = useState(0);

  useEffect(() => {
    const controller = new AbortController();
    setError("");
    collect(cursor => listMemberGroups(request, controller.signal, cursor || undefined)).then(
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
        <Button ref={cancelButton} type="button" variant="outline" disabled={blocked} onClick={() => {if(!isBlocked())discard.navigate(()=>{if(!isBlocked())close();});}}>取消填寫</Button>
        {discard.dialog}
      </div>
    </PagePanel>
    {publishGroup && <Composer request={request} group={publishGroup} saved={close} onPending={onPending} isBlocked={isBlocked} />}
  </>;
}

