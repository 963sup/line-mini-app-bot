import React,{useCallback,useEffect,useRef,useState} from "react";
import type {components} from "@repo/api-client";
import {useSession} from "@/identity/Session";
import {useLatestRequest,errorMessage} from "@/shared/http/lifecycle";
import {PagePanel} from "@/shared/PagePanel";
import {FormField} from "@/shared/FormField";
import {Status} from "@/shared/Status";
import {Button} from "@/shared/ui/button";
import {NativeSelect,NativeSelectOption} from "@/shared/ui/native-select";
import {Item,ItemGroup,ItemContent,ItemTitle} from "@/shared/ui/item";
import {FieldGroup} from "@/shared/ui/field";
type Project=components["schemas"]["Project"];
const states={open:"待處理",in_progress:"進行中",completed:"已完成"};
export function Items({spaceId,sourceRepositoryId,project,disabled,onChange,onRefresh}:{spaceId:string;sourceRepositoryId:string;project:Project;disabled:boolean;onChange:(v:Project)=>void;onRefresh:()=>void}){
 const {user,request}=useSession(),read=useLatestRequest(),catalog=useLatestRequest(),write=useLatestRequest();
 const [items,setItems]=useState<components["schemas"]["ProjectItem"][]>([]),[issues,setIssues]=useState<components["schemas"]["Issue"][]>([]);
 const [cursor,setCursor]=useState(""),[issueCursor,setIssueCursor]=useState(""),[choice,setChoice]=useState(""),[status,setStatus]=useState(""),[catalogStatus,setCatalogStatus]=useState("");
 const [busy,setBusy]=useState(false),[readBusy,setReadBusy]=useState(false),[catalogBusy,setCatalogBusy]=useState(false),[needsRefresh,setNeedsRefresh]=useState(false),[loaded,setLoaded]=useState(false);
 const lock=useRef(false),mine=project.creatorId===user?.userId;
 const load=useCallback((next="")=>{if(!user)return;setReadBusy(true);setLoaded(false);if(!next)setItems([]);setCursor("");setStatus("正在讀取專案事項……");void read.run(signal=>request((c,s)=>c.GET("/api/v1/spaces/{spaceId}/projects/{id}/items",{signal:s,params:{path:{spaceId,id:project.id},query:{cursor:next||undefined}}}),{signal}),page=>{setItems(prev=>[...new Map([...(next?prev:[]),...page.items].map(i=>[i.id,i])).values()]);setCursor(page.nextCursor);setReadBusy(false);setLoaded(true);setStatus("");},e=>{setItems([]);setReadBusy(false);setStatus(errorMessage(e));});},[user,request,read.run,spaceId,project.id]);
 const loadIssues=useCallback((next="")=>{if(!user||!mine)return;setCatalogBusy(true);if(!next)setIssues([]);setIssueCursor("");setCatalogStatus("正在讀取可收錄事項……");void catalog.run(signal=>request((c,s)=>c.GET("/api/v1/repositories/{repositoryId}/issues",{signal:s,params:{path:{repositoryId:sourceRepositoryId},query:{cursor:next||undefined}}}),{signal}),page=>{setIssues(prev=>[...new Map([...(next?prev:[]),...page.items].map(i=>[i.id,i])).values()]);setIssueCursor(page.nextCursor);setCatalogBusy(false);setCatalogStatus("");},e=>{setIssues([]);setCatalogBusy(false);setCatalogStatus(errorMessage(e));});},[user,mine,request,catalog.run,sourceRepositoryId]);
 useEffect(()=>{load();loadIssues();return()=>{read.cancel();catalog.cancel();};},[load,loadIssues,read.cancel,catalog.cancel]);
 function change(itemId?:string){if(!user||!mine||disabled||lock.current||needsRefresh||(!itemId&&!choice))return;lock.current=true;setBusy(true);setStatus("正在更新專案事項……");void write.run(signal=>request((c,s)=>itemId?c.DELETE("/api/v1/spaces/{spaceId}/projects/{id}/items/{itemId}",{signal:s,params:{path:{spaceId,id:project.id,itemId},query:{expectedVersion:project.version}}}):c.POST("/api/v1/spaces/{spaceId}/projects/{id}/items",{signal:s,params:{path:{spaceId,id:project.id}},body:{issueId:choice,expectedVersion:project.version}}),{signal}),v=>{lock.current=false;setBusy(false);onChange(v);},e=>{lock.current=false;setBusy(false);setNeedsRefresh(true);setStatus(errorMessage(e)+" 請重新讀取專案後再管理事項。");});}
 return <PagePanel heading="專案事項"><p>沿用事項原有狀態；加入專案不會變更事項或授予編輯權。</p><Status role="status">{status}</Status>
 <Button type="button" variant="outline" disabled={busy||disabled} onClick={onRefresh}>重新讀取專案與事項</Button>
 <ItemGroup>{items.map(i=><Item key={i.id} variant="outline"><ItemContent><ItemTitle>{i.title}</ItemTitle><p>{states[i.state]}</p>{mine&&<Button type="button" variant="outline" disabled={disabled||busy||needsRefresh} onClick={()=>change(i.id)}>移出專案：{i.title}</Button>}</ItemContent></Item>)}</ItemGroup>
 {loaded&&!items.length&&<p>尚未收錄事項。</p>}{cursor&&<Button type="button" variant="outline" disabled={readBusy||busy} onClick={()=>load(cursor)}>載入更多專案事項</Button>}
 {mine&&<FieldGroup><Status role="status">{catalogStatus}</Status><FormField>選擇既有事項<NativeSelect value={choice} disabled={disabled||busy||needsRefresh||catalogBusy} onChange={e=>setChoice(e.target.value)}><NativeSelectOption value="">請選擇事項</NativeSelectOption>{issues.map(i=><NativeSelectOption key={i.id} value={i.id}>#{i.number} {i.title}</NativeSelectOption>)}</NativeSelect></FormField>
 <Button type="button" disabled={disabled||busy||needsRefresh||!choice} onClick={()=>change()}>加入專案</Button><Button type="button" variant="outline" disabled={catalogBusy||busy} onClick={()=>loadIssues()}>重新載入可收錄事項</Button>{issueCursor&&<Button type="button" variant="outline" disabled={catalogBusy||busy} onClick={()=>loadIssues(issueCursor)}>載入更多可收錄事項</Button>}</FieldGroup>}
 </PagePanel>;
}
