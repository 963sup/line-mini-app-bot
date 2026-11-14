import React, { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { useLatestRequest, errorMessage } from "@/shared/http/lifecycle";
import { Picker as MilestonePicker } from "../repository/milestones/Picker";
import { Picker as LabelPicker } from "../repository/labels/Picker";
import { PagePanel } from "@/shared/PagePanel";
import { FormField } from "@/shared/FormField";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Textarea } from "@/shared/ui/textarea";
import { FieldGroup } from "@/shared/ui/field";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";
import { Badge } from "@/shared/ui/badge";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from "@/shared/ui/empty";
import { Item, ItemGroup, ItemContent, ItemTitle } from "@/shared/ui/item";

type Issue = components["schemas"]["Issue"];
type Creation = components["schemas"]["IssueCreation"];
const states = { open: "待處理", in_progress: "進行中", completed: "已完成" };
const nextState = { open: "in_progress", in_progress: "completed", completed: "open" } as const;
const action = { open: "開始處理", in_progress: "標記完成", completed: "重新開啟" };

export function View({ repositoryId }: { repositoryId: string }) {
  const { user, request } = useSession();
  const { run: readRun, cancel: cancelRead } = useLatestRequest();
  const detailRequest = useLatestRequest(), write = useLatestRequest();
  const [items, setItems] = useState<Issue[]>([]), [cursor, setCursor] = useState("");
  const [query, setQuery] = useState(""), [state, setState] = useState<Issue["state"] | "">("");
  const [filter, setFilter] = useState({ query: "", state: "" as Issue["state"] | "" });
  const [listBusy, setListBusy] = useState(false), [loaded, setLoaded] = useState(false);
  const [listStatus, setListStatus] = useState("");
  const [mode, setMode] = useState<"create" | "detail" | null>(null);
  const [selected, setSelected] = useState<Issue | null>(null);
  const [title, setTitle] = useState(""), [body, setBody] = useState("");
  const [milestoneID,setMilestoneID]=useState("");
  const [labelIDs,setLabelIDs]=useState<string[]>([]);
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
    setListStatus("正在讀取問題……");
    void readRun(
      (signal) => request((client, s) => client.GET("/api/v1/repositories/{repositoryId}/issues", {
        signal: s, params: { path: { repositoryId }, query: { cursor: nextCursor || undefined, state: filter.state || undefined, query: filter.query || undefined } },
      }), { signal }),
      (page) => {
        setItems((previous) => [...new Map([...(nextCursor ? previous : []), ...page.items].map((item) => [item.id, item])).values()]);
        setCursor(page.nextCursor); setListBusy(false); setLoaded(true);
        setListStatus(page.nextCursor ? "可繼續載入更多問題；篩選後可能出現空頁。" : "已讀取完畢。");
      },
      (error) => { setItems([]); setListBusy(false); setLoaded(false); setListStatus(errorMessage(error)); },
    );
  }, [user, request, repositoryId, filter, readRun]);
  useEffect(() => { load(); return cancelRead; }, [load, cancelRead]);

  function show(item: Issue) {
    setSelected(item); setTitle(item.title); setBody(item.body); setLabelIDs(item.labelIds || []);setMilestoneID(item.milestoneId || ""); setMode("detail");
    setNeedsRefresh(false); setUncertain(false); attempt.current = null;
  }
  function open(id: string) {
    if (!user || lock.current) return;
    setSelected(null); setMode("detail"); setDetailBusy(true); setStatus("正在核驗並讀取問題……");
    void detailRequest.run(
      (signal) => request((client, s) => client.GET("/api/v1/repositories/{repositoryId}/issues/{id}", { signal: s, params: { path: { repositoryId, id } } }), { signal }),
      (item) => { show(item); setDetailBusy(false); setStatus("已讀取問題。"); },
      (error) => { setDetailBusy(false); setStatus(errorMessage(error)); },
    );
  }
  function create() {
    if (lock.current) return;
    detailRequest.cancel(); setDetailBusy(false); setSelected(null); setMode("create");
    setTitle(""); setBody(""); setUncertain(false); setNeedsRefresh(false); setStatus(""); attempt.current = null;
  }
  function submit(event: FormEvent) {
    event.preventDefault();
    if (!user || lock.current || needsRefresh) return;
    if (mode === "detail") { if (selected) save(selected.state); return; }
    const content = { title: title.trim(), body: body.trim() };
    if (!content.title) { setStatus("請填寫問題標題。"); return; }
    if (!uncertain || !attempt.current) attempt.current = { key: crypto.randomUUID(), body: content };
    const pending = attempt.current;
    lock.current = true; setBusy(true); setUncertain(true); setStatus("正在建立問題……");
    void write.run(
      (signal) => request((client, s) => client.POST("/api/v1/repositories/{repositoryId}/issues", {
        signal: s, params: { path: { repositoryId }, header: { "Idempotency-Key": pending.key } }, body: pending.body,
      }), { signal }),
      (item) => { lock.current = false; setBusy(false); show(item); setStatus("已建立問題。"); load(); },
      (error) => {
        lock.current = false; setBusy(false);
        if (["invalid_input", "forbidden", "not_found", "idempotency_conflict"].includes((error as { code?: string }).code || "")) setUncertain(false);
        setStatus(errorMessage(error) + " 結果未確認時請沿用原內容重試，或先查詢問題；重新開頁前請先確認紀錄。");
      },
    );
  }
  function save(state: Issue["state"]) {
    if (!user || !selected || !mine || lock.current || needsRefresh) return;
    const id = selected.id;
    lock.current = true; setBusy(true); setStatus("正在更新問題……");
    void write.run(
      (signal) => request((client, s) => client.PATCH("/api/v1/repositories/{repositoryId}/issues/{id}", {
        signal: s, params: { path: { repositoryId, id } }, body: { title: title.trim(), body: body.trim(), state, expectedVersion: selected.version, labelIds:labelIDs,milestoneId:milestoneID||null },
      }), { signal }),
      (item) => { lock.current = false; setBusy(false); show(item); setStatus("已更新問題。"); load(); },
      (error) => { lock.current = false; setBusy(false); setNeedsRefresh(true); setStatus(errorMessage(error) + " 請重新讀取問題後再修改。"); },
    );
  }
  return <>
    <PagePanel heading="問題">
      <p className="mb-4 text-sm text-muted-foreground">空間成員可建立問題；只有建立者可以修改或變更狀態。</p>
      <div className="mb-4 flex flex-wrap gap-2">
        <Button disabled={!user || busy || uncertain} onClick={create}>新增問題</Button>
        <Button variant="outline" disabled={!user || listBusy || busy} onClick={() => load()}>重新載入問題</Button>
      </div>
      <form onSubmit={(event) => { event.preventDefault(); setFilter({ query: query.trim(), state }); }}>
        <FieldGroup className="mb-4">
          <FormField>搜尋標題<Input maxLength={120} value={query} onChange={(e) => setQuery(e.target.value)} /></FormField>
          <FormField>狀態<NativeSelect value={state} onChange={(e) => setState(e.target.value as typeof state)}>
            <NativeSelectOption value="">全部</NativeSelectOption>
            {Object.entries(states).map(([value, label]) => <NativeSelectOption key={value} value={value}>{label}</NativeSelectOption>)}
          </NativeSelect></FormField>
          <Button type="submit" disabled={!user || listBusy || busy}>套用篩選</Button>
        </FieldGroup>
      </form>
      <Status role="status">{listStatus}</Status>
      <ItemGroup className="gap-3">
        {items.map((item) => <Item key={item.id} variant="outline" role="listitem">
          <ItemContent>
            <ItemTitle><Button variant="link" className="h-auto whitespace-normal text-left" disabled={busy || detailBusy || uncertain} onClick={() => open(item.id)}>#{item.number} {item.title}</Button></ItemTitle>
            <Badge variant="outline">{states[item.state]}</Badge>
          </ItemContent>
        </Item>)}
      </ItemGroup>
      {loaded && !items.length && <Empty><EmptyHeader><EmptyTitle>{cursor ? "本頁沒有符合的問題" : "沒有符合的問題"}</EmptyTitle><EmptyDescription>{cursor ? "請繼續載入下一頁。" : "可調整篩選條件，或建立新的問題。"}</EmptyDescription></EmptyHeader></Empty>}
      {cursor && <Button variant="outline" disabled={listBusy || busy} onClick={() => load(cursor)}>載入更多問題</Button>}
    </PagePanel>
    {mode && <PagePanel heading={mode === "create" ? "新增問題" : selected ? `#${selected.number} ${selected.title}` : "問題詳情"}>
      <Status role="status">{status}</Status>
      {selected && <div className="mb-4 flex flex-wrap items-center gap-2">
        <Badge variant="outline">{states[selected.state]}</Badge>
        <Button variant="outline" disabled={busy || detailBusy} onClick={() => open(selected.id)}>重新讀取問題</Button>
      </div>}
      {(mode === "create" || selected) && <form onSubmit={submit}>
        <FieldGroup>
          <fieldset className="grid min-w-0 gap-4" disabled={busy || detailBusy || uncertain || needsRefresh || (mode === "detail" && !mine)}>
            <FormField>標題<Input required maxLength={120} value={title} onChange={(e) => setTitle(e.target.value)} /></FormField>
            <FormField>內容<Textarea maxLength={8000} value={body} onChange={(e) => setBody(e.target.value)} /></FormField>
          </fieldset>
          {selected && <MilestonePicker repositoryId={repositoryId} value={milestoneID} onChange={setMilestoneID} disabled={busy || detailBusy || needsRefresh || !mine} />}
          {selected && <LabelPicker repositoryId={repositoryId} value={labelIDs} onChange={setLabelIDs} disabled={busy || detailBusy || needsRefresh || !mine} />}
          {(mode === "create" || mine) && <div className="flex flex-wrap gap-2">
            <Button type="submit" disabled={busy || needsRefresh}>{mode === "create" ? uncertain ? "重試原建立請求" : "建立問題" : "儲存內容"}</Button>
            {selected && <Button type="button" variant="outline" disabled={busy || needsRefresh} onClick={() => save(nextState[selected.state])}>{action[selected.state]}</Button>}
          </div>}
        </FieldGroup>
      </form>}
    </PagePanel>}
  </>;
}
