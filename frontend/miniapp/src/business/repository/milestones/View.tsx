import React, { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { useLatestRequest, errorMessage } from "@/shared/http/lifecycle";
import { PagePanel } from "@/shared/PagePanel";
import { FormField } from "@/shared/FormField";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Textarea } from "@/shared/ui/textarea";
import { FieldGroup } from "@/shared/ui/field";
import { Badge } from "@/shared/ui/badge";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from "@/shared/ui/empty";
import { Item, ItemGroup, ItemContent, ItemTitle } from "@/shared/ui/item";

type Milestone = components["schemas"]["Milestone"];
type Creation = components["schemas"]["MilestoneCreation"];
const states = { open: "進行中", completed: "已完成" };
const nextState = { open: "completed", completed: "open" } as const;
const action = { open: "標記完成", completed: "重新開啟" };

export function View({ repositoryId }: { repositoryId: string }) {
  const { user, request } = useSession();
  const { run: readRun, cancel: cancelRead } = useLatestRequest();
  const detailRequest = useLatestRequest(), write = useLatestRequest();
  const [items, setItems] = useState<Milestone[]>([]), [cursor, setCursor] = useState("");
  const [listBusy, setListBusy] = useState(false), [loaded, setLoaded] = useState(false);
  const [listStatus, setListStatus] = useState("");
  const [mode, setMode] = useState<"create" | "detail" | null>(null);
  const [selected, setSelected] = useState<Milestone | null>(null);
  const [dueDate,setDueDate]=useState("");
  const [title, setTitle] = useState(""), [body, setBody] = useState("");
  const [busy, setBusy] = useState(false), [detailBusy, setDetailBusy] = useState(false);
  const [uncertain, setUncertain] = useState(false), [needsRefresh, setNeedsRefresh] = useState(false);
  const [status, setStatus] = useState("");
  const lock = useRef(false);
  const attempt = useRef<{ key: string; body: Creation } | null>(null);
  const mine = !!selected && selected.creatorId === user?.userId;

  const load = useCallback((nextCursor = "") => {
    if (!user) return;
    setListBusy(true); setLoaded(false); setCursor("");
    if (!nextCursor) setItems([]);
    setListStatus("正在讀取里程碑……");
    void readRun(
      (signal) => request((client, s) => client.GET("/api/v1/repositories/{repositoryId}/milestones", {
        signal: s, params: { path: { repositoryId }, query: { cursor: nextCursor || undefined } },
      }), { signal }),
      (page) => {
        setItems((previous) => [...new Map([...(nextCursor ? previous : []), ...page.items].map((item) => [item.id, item])).values()]);
        setCursor(page.nextCursor); setListBusy(false); setLoaded(true);
        setListStatus(page.nextCursor ? "可繼續載入更多里程碑。" : "已讀取完畢。");
      },
      (error) => { setItems([]); setListBusy(false); setLoaded(false); setListStatus(errorMessage(error)); },
    );
  }, [user, request, repositoryId, readRun]);
  useEffect(() => { load(); return cancelRead; }, [load, cancelRead]);

  function show(item: Milestone) {
    setSelected(item); setTitle(item.title); setBody(item.body);setDueDate(item.dueDate); setMode("detail");
    setNeedsRefresh(false); setUncertain(false); attempt.current = null;
  }
  function open(id: string) {
    if (!user || lock.current) return;
    setSelected(null); setMode("detail"); setDetailBusy(true); setStatus("正在核驗並讀取里程碑……");
    void detailRequest.run(
      (signal) => request((client, s) => client.GET("/api/v1/repositories/{repositoryId}/milestones/{id}", { signal: s, params: { path: { repositoryId, id } } }), { signal }),
      (item) => { show(item); setDetailBusy(false); setStatus("已讀取里程碑。"); },
      (error) => { setDetailBusy(false); setStatus(errorMessage(error)); },
    );
  }
  function create() {
    if (lock.current) return;
    detailRequest.cancel(); setDetailBusy(false); setSelected(null); setMode("create");
    setTitle(""); setBody("");setDueDate(""); setUncertain(false); setNeedsRefresh(false); setStatus(""); attempt.current = null;
  }
  function submit(event: FormEvent) {
    event.preventDefault();
    if (!user || lock.current || needsRefresh) return;
    if (mode === "detail") { if (selected) save(selected.state); return; }
    const content = { title: title.trim(), body: body.trim(),dueDate };
    if (!content.title) { setStatus("請填寫里程碑標題。"); return; }
    if (!uncertain || !attempt.current) attempt.current = { key: crypto.randomUUID(), body: content };
    const pending = attempt.current;
    lock.current = true; setBusy(true); setUncertain(true); setStatus("正在建立里程碑……");
    void write.run(
      (signal) => request((client, s) => client.POST("/api/v1/repositories/{repositoryId}/milestones", {
        signal: s, params: { path: { repositoryId }, header: { "Idempotency-Key": pending.key } }, body: pending.body,
      }), { signal }),
      (item) => { lock.current = false; setBusy(false); show(item); setStatus("已建立里程碑。"); load(); },
      (error) => {
        lock.current = false; setBusy(false);
        if (["invalid_input", "forbidden", "not_found", "idempotency_conflict"].includes((error as { code?: string }).code || "")) setUncertain(false);
        setStatus(errorMessage(error) + " 結果未確認時請沿用原內容重試，或先查詢里程碑；重新開頁前請先確認紀錄。");
      },
    );
  }
  function save(state: Milestone["state"]) {
    if (!user || !selected || !mine || lock.current || needsRefresh) return;
    const id = selected.id;
    lock.current = true; setBusy(true); setStatus("正在更新里程碑……");
    void write.run(
      (signal) => request((client, s) => client.PATCH("/api/v1/repositories/{repositoryId}/milestones/{id}", {
        signal: s, params: { path: { repositoryId, id } }, body: { title: title.trim(), body: body.trim(),dueDate, state, expectedVersion: selected.version },
      }), { signal }),
      (item) => { lock.current = false; setBusy(false); show(item); setStatus("已更新里程碑。"); load(); },
      (error) => { lock.current = false; setBusy(false); setNeedsRefresh(true); setStatus(errorMessage(error) + " 請重新讀取里程碑後再修改。"); },
    );
  }
  return <>
    <PagePanel heading="里程碑">
      <p className="mb-4 text-sm text-muted-foreground">空間成員可建立里程碑；只有建立者可以編輯、完成或重新開啟。</p>
      <div className="mb-4 flex flex-wrap gap-2">
        <Button disabled={!user || busy || uncertain} onClick={create}>新增里程碑</Button>
        <Button variant="outline" disabled={!user || listBusy || busy} onClick={() => load()}>重新載入里程碑</Button>
      </div>
      <Status role="status">{listStatus}</Status>
      <ItemGroup className="gap-3">
        {items.map((item) => <Item key={item.id} variant="outline" role="listitem">
          <ItemContent>
            <ItemTitle><Button variant="link" className="h-auto whitespace-normal text-left" disabled={busy || detailBusy || uncertain} onClick={() => open(item.id)}>{item.title}</Button></ItemTitle>
            <Badge variant="outline">{states[item.state]}</Badge>
          </ItemContent>
        </Item>)}
      </ItemGroup>
      {loaded && !items.length && <Empty><EmptyHeader><EmptyTitle>{cursor ? "本頁沒有里程碑" : "目前沒有里程碑"}</EmptyTitle><EmptyDescription>{cursor ? "請繼續載入下一頁。" : "可建立新的里程碑。"}</EmptyDescription></EmptyHeader></Empty>}
      {cursor && <Button variant="outline" disabled={listBusy || busy} onClick={() => load(cursor)}>載入更多里程碑</Button>}
    </PagePanel>
    {mode && <PagePanel heading={mode === "create" ? "新增里程碑" : selected ? selected.title : "里程碑詳情"}>
      <Status role="status">{status}</Status>
      {selected && <p>{selected.issueCount ? `已完成 ${selected.completedIssueCount}/${selected.issueCount} 項（${Math.round(selected.completedIssueCount/selected.issueCount*100)}%）` : "尚無事項"}</p>}
      {selected && <div className="mb-4 flex flex-wrap items-center gap-2">
        <Badge variant="outline">{states[selected.state]}</Badge>
        <Button variant="outline" disabled={busy || detailBusy} onClick={() => open(selected.id)}>重新讀取里程碑</Button>
      </div>}
      {(mode === "create" || selected) && <form onSubmit={submit}>
        <FieldGroup>
          <fieldset className="grid min-w-0 gap-4" disabled={busy || detailBusy || uncertain || needsRefresh || (mode === "detail" && !mine)}>
            <FormField>到期日<Input type="date" value={dueDate} onChange={(e)=>setDueDate(e.target.value)} /></FormField>
          <FormField>標題<Input required maxLength={120} value={title} onChange={(e) => setTitle(e.target.value)} /></FormField>
            <FormField>內容<Textarea maxLength={8000} value={body} onChange={(e) => setBody(e.target.value)} /></FormField>
          </fieldset>

          {(mode === "create" || mine) && <div className="flex flex-wrap gap-2">
            <Button type="submit" disabled={busy || needsRefresh}>{mode === "create" ? uncertain ? "重試原建立請求" : "建立里程碑" : "儲存內容"}</Button>
            {selected && <Button type="button" variant="outline" disabled={busy || needsRefresh} onClick={() => save(nextState[selected.state])}>{action[selected.state]}</Button>}
          </div>}
        </FieldGroup>
      </form>}
    </PagePanel>}
  </>;
}
