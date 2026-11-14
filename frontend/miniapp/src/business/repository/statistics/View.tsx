import React,{useCallback,useEffect,useState} from "react";
import type {components} from "@repo/api-client";
import {useSession} from "@/identity/Session";
import {useLatestRequest,errorMessage} from "@/shared/http/lifecycle";
import {PagePanel} from "@/shared/PagePanel";
import {Status} from "@/shared/Status";
import {Button} from "@/shared/ui/button";
import {displayTime} from "@/shared/time/time";
export function View({repositoryId}:{repositoryId:string}){
 const {user,request}=useSession(),{run,cancel}=useLatestRequest();const [data,setData]=useState<components["schemas"]["RepositoryStatistics"]|null>(null),[busy,setBusy]=useState(false),[status,setStatus]=useState("");
 const load=useCallback(()=>{if(!user)return;setData(null);setBusy(true);setStatus("正在讀取統計……");void run(signal=>request((c,s)=>c.GET("/api/v1/repositories/{repositoryId}/statistics",{signal:s,params:{path:{repositoryId}}}),{signal}),v=>{setData(v);setBusy(false);setStatus("");},e=>{setData(null);setBusy(false);setStatus(errorMessage(e));});},[user,request,repositoryId,run]);useEffect(()=>{load();return cancel;},[load,cancel]);
 return <PagePanel heading="空間統計"><Button variant="outline" disabled={!user||busy} onClick={load}>重新載入統計</Button><Status role="status">{status}</Status>{data&&<><dl className="grid grid-cols-2 gap-2"><dt>待處理問題</dt><dd>{data.openIssues}</dd><dt>進行中問題</dt><dd>{data.inProgressIssues}</dd><dt>已完成問題</dt><dd>{data.completedIssues}</dd><dt>討論</dt><dd>{data.discussions}</dd><dt>里程碑</dt><dd>{data.milestones}</dd><dt>專案</dt><dd>{data.projects}</dd></dl><p className="text-sm text-muted-foreground">統計時間：{displayTime(data.observedAt)}</p></>}</PagePanel>;
}
