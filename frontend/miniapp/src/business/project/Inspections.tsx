import React, { useRef, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";
import { Badge } from "@/shared/ui/badge";
import { Item, ItemContent, ItemDescription } from "@/shared/ui/item";
import { Button } from "@/shared/ui/button";
import { Textarea } from "@/shared/ui/textarea";
import { FormField } from "@/shared/FormField";
import { Status } from "@/shared/Status";
import { Empty, EmptyHeader, EmptyTitle } from "@/shared/ui/empty";

type Submission = components["schemas"]["InspectionSubmission"];
const endpoint = "/api/v1/owners/{ownerId}/projects/{id}/references/{itemId}/inspection-submissions";

export function Inspections({canSubmit, ownerId, projectId, itemId}: {canSubmit:boolean;ownerId:string;projectId:string;itemId:string}) {
  const {request} = useSession();
  const {run} = useLatestRequest();
  const [rows,setRows] = useState<Submission[] | null>(null);
  const [note,setNote] = useState("");
  const [status,setStatus] = useState("");
  const [busy,setBusy] = useState(false);
  const lock = useRef(false);
  function perform(submit:boolean) {
    if(lock.current) return;
    lock.current=true;setBusy(true);setStatus("");
    const params={path:{ownerId,id:projectId,itemId}};
    void run(async signal => {
      if(submit) {
        const entry=await request((client,s)=>client.POST(endpoint,{signal:s,params,body:{note:note.trim()}}),{signal});
        return [entry];
      }
      return request((client,s)=>client.GET(endpoint,{signal:s,params}),{signal});
    },entries=>{lock.current=false;setBusy(false);setRows(entries);setNote("");},
    error=>{lock.current=false;setBusy(false);setRows(null);setStatus(errorMessage(error));});
  }
  return <div>
    <Button type="button" variant="outline" disabled={busy} onClick={()=>perform(false)}>查看驗收紀錄</Button>
    {rows?.map(row => <Item key={row.id} variant="outline" size="sm">
      <ItemContent className="min-w-0">
        <p className="whitespace-pre-wrap [overflow-wrap:anywhere]">{row.note}</p>
        <ItemDescription>{row.submittedBy} · <time dateTime={row.submittedAt}>{new Date(row.submittedAt).toLocaleString()}</time></ItemDescription>
      </ItemContent>
      <Badge variant="outline">待審</Badge>
    </Item>)}
    {rows?.length===0 && !canSubmit && <Empty><EmptyHeader><EmptyTitle>目前沒有驗收紀錄。</EmptyTitle></EmptyHeader></Empty>}
    {rows?.length===0 && canSubmit && <>
      <FormField>檢查說明<Textarea maxLength={2000} disabled={busy} value={note} onChange={e=>setNote(e.target.value)}/></FormField>
      <Button type="button" disabled={busy||!note.trim()} onClick={()=>perform(true)}>提交驗收</Button>
    </>}
    {status && <Status role="status">{status}</Status>}
  </div>;
}
