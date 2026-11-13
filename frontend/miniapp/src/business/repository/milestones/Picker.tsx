import React,{useCallback,useEffect,useState} from "react";
import type {components} from "@repo/api-client";
import {useSession} from "@/identity/Session";
import {useLatestRequest,errorMessage} from "@/shared/http/lifecycle";
import {FormField} from "@/shared/FormField";
import {Status} from "@/shared/Status";
import {Button} from "@/shared/ui/button";
import {FieldGroup} from "@/shared/ui/field";
import {NativeSelect,NativeSelectOption} from "@/shared/ui/native-select";
export function Picker({repositoryId,value,onChange,disabled}:{repositoryId:string;value:string;onChange:(v:string)=>void;disabled:boolean}){
 const {user,request}=useSession(),{run,cancel}=useLatestRequest();const [items,setItems]=useState<components["schemas"]["Milestone"][]>([]),[cursor,setCursor]=useState(""),[busy,setBusy]=useState(false),[status,setStatus]=useState("");
 const load=useCallback((next="")=>{if(!user)return;setBusy(true);setCursor("");if(!next)setItems([]);setStatus("正在讀取里程碑……");void run(signal=>request((c,s)=>c.GET("/api/v1/repositories/{repositoryId}/milestones",{signal:s,params:{path:{repositoryId},query:{cursor:next||undefined}}}),{signal}),page=>{setItems(prev=>[...new Map([...(next?prev:[]),...page.items].map(i=>[i.id,i])).values()]);setCursor(page.nextCursor);setBusy(false);setStatus("");},e=>{setItems([]);setBusy(false);setStatus(errorMessage(e));});},[user,request,repositoryId,run]);
 useEffect(()=>{load();return cancel;},[load,cancel]);
 return <FieldGroup><Status role="status">{status}</Status><FormField>所屬里程碑<NativeSelect value={value} disabled={disabled||busy} onChange={e=>onChange(e.target.value)}><NativeSelectOption value="">不指定里程碑</NativeSelectOption>{value&&!items.some(i=>i.id===value)&&<NativeSelectOption value={value}>目前里程碑（尚未載入名稱）</NativeSelectOption>}{items.map(i=><NativeSelectOption key={i.id} value={i.id}>{i.title}{i.state==="completed"?"（已完成）":""}</NativeSelectOption>)}</NativeSelect></FormField><Button type="button" variant="outline" disabled={busy} onClick={()=>load()}>重新載入可選里程碑</Button>{cursor&&<Button type="button" variant="outline" disabled={busy} onClick={()=>load(cursor)}>載入更多可選里程碑</Button>}</FieldGroup>;
}
