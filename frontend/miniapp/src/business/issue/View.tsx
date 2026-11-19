import React, { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import { useRepositoryGuard } from "../repository/NavigationGuard";
import { useDiscardChanges } from "@/shared/useDiscardChanges";
import { ReadFeedback } from "@/shared/ReadFeedback";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { useLatestRequest, errorMessage } from "@/shared/http/lifecycle";
import { Picker as MilestonePicker } from "../repository/milestones/Picker";
import { Picker as LabelPicker } from "../repository/labels/Picker";
import { PagePanel } from "@/shared/PagePanel";
import { FormField } from "@/shared/FormField";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Textarea } from "@/shared/ui/textarea";
import { FieldGroup } from "@/shared/ui/field";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";
import { Badge } from "@/shared/ui/badge";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from "@/shared/ui/empty";
import { Item, ItemGroup, ItemContent, ItemTitle, ItemActions, ItemFooter } from "@/shared/ui/item";

type Issue = components["schemas"]["Issue"];
type Creation = components["schemas"]["IssueCreation"];
const states = { open: "開啟", in_progress: "開啟（原進行中）", completed: "已關閉" };
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
  const [createOutcomeUnknown, setCreateOutcomeUnknown] = useState(false), [needsRefresh, setNeedsRefresh] = useState(false);
  const [status, setStatus] = useState("");
  const lock = useRef(false);
  const createAttempt = useRef<{ idempotencyKey: string; body: Creation } | null>(null);
  const mine = !!selected && selected.creatorId === user?.userId;
  const [activeId, setActiveId] = useState("");
  const baseline = useRef(JSON.stringify(["", "", "", []]));
  const dirty = !!mode && JSON.stringify([title, body, milestoneID, labelIDs]) !== baseline.current;
  const titleInput = useRef<HTMLInputElement>(null), editorPanel = useRef<HTMLElement>(null), opener = useRef<HTMLElement | null>(null);
  const navigating = useRef(false);
  function focusEditor() {
    if (titleInput.current && !titleInput.current.matches(":disabled")) titleInput.current.focus();
    else editorPanel.current?.focus();
  }
  const discard = useDiscardChanges(dirty, focusEditor);
  const guard = useRepositoryGuard(dirty, focusEditor);
  useEffect(() => { guard.block(busy || createOutcomeUnknown || needsRefresh); }, [busy, createOutcomeUnknown, needsRefresh]);
  useEffect(() => { navigating.current = false; });
  useEffect(() => {
    if (!mode) opener.current?.focus();
    else if (!detailBusy && !busy) focusEditor();
  }, [mode, detailBusy, busy]);
  function clearEditor() {
    detailRequest.cancel(); setDetailBusy(false); setSelected(null); setMode(null); setActiveId(""); setNeedsRefresh(false); setStatus("");
  }
  function close() {
    if (lock.current || createOutcomeUnknown || needsRefresh) return;
    discard.navigate(() => { navigating.current = true; clearEditor(); });
  }
  function saved(item: Issue, message: string) {
    guard.block(false); setCreateOutcomeUnknown(false); createAttempt.current = null;
    clearEditor();
    setItems(previous => previous.some(value => value.id === item.id) ? previous.map(value => value.id === item.id ? item : value) : [item, ...previous]);
    setListStatus(message);
    if (filter.query || filter.state) load();
  }

  const load = useCallback((nextCursor = "") => {
    if (!user) return;
    setListBusy(true); setLoaded(false); setCursor("");
    if (!nextCursor) setItems([]);
    setListStatus("");
    void readRun(
      (signal) => request((client, s) => client.GET("/api/v1/repositories/{repositoryId}/issues", {
        signal: s, params: { path: { repositoryId }, query: { cursor: nextCursor || undefined, state: filter.state || undefined, query: filter.query || undefined } },
      }), { signal }),
      (page) => {
        setItems((previous) => [...new Map([...(nextCursor ? previous : []), ...page.items].map((item) => [item.id, item])).values()]);
        setCursor(page.nextCursor); setListBusy(false); setLoaded(true);
        setListStatus("");
      },
      (error) => { setItems([]); setListBusy(false); setLoaded(false); setListStatus(errorMessage(error)); },
    );
  }, [user, request, repositoryId, filter, readRun]);
  useEffect(() => { load(); return cancelRead; }, [load, cancelRead]);

  function show(item: Issue) {
    baseline.current = JSON.stringify([item.title, item.body, item.milestoneId || "", item.labelIds || []]); setActiveId(item.id); guard.block(false);
    setSelected(item); setTitle(item.title); setBody(item.body); setLabelIDs(item.labelIds || []);setMilestoneID(item.milestoneId || ""); setMode("detail");
    setNeedsRefresh(false); setCreateOutcomeUnknown(false); createAttempt.current = null;
  }
  function open(id: string, trigger?: HTMLElement) {
    if (!user || lock.current || createOutcomeUnknown || (needsRefresh && id !== activeId) || !guard.canWrite()) return;
    discard.navigate(() => {
      if (trigger) opener.current = trigger;
      navigating.current = true; cancelRead(); setListBusy(false);
      baseline.current = JSON.stringify(["", "", "", []]); setTitle(""); setBody(""); setMilestoneID(""); setLabelIDs([]);
      setActiveId(id); setSelected(null); setMode("detail"); setDetailBusy(true); setStatus("");
      void detailRequest.run(
        signal => request((client, s) => client.GET("/api/v1/repositories/{repositoryId}/issues/{id}", { signal: s, params: { path: { repositoryId, id } } }), { signal }),
        item => { show(item); setDetailBusy(false); },
        error => { setDetailBusy(false); setStatus(errorMessage(error)); },
      );
    });
  }
  function openCreateForm(trigger: HTMLElement) {
    if (lock.current || createOutcomeUnknown || needsRefresh || !guard.canWrite()) return;
    discard.navigate(() => {
      opener.current = trigger; navigating.current = true; cancelRead(); setListBusy(false); clearEditor();
      baseline.current = JSON.stringify(["", "", "", []]);
      setMode("create"); setTitle(""); setBody(""); setMilestoneID(""); setLabelIDs([]); createAttempt.current = null;
    });
  }
  function submit(event: FormEvent) {
    event.preventDefault();
    if (!user || lock.current || navigating.current || needsRefresh || discard.isConfirming() || !guard.canWrite()) return;
    if (mode === "detail") { if (selected) save(selected.state); return; }
    const content = { title: title.trim(), body: body.trim() };
    if (!content.title) { setStatus("請填寫問題標題。"); return; }
    if (!createOutcomeUnknown || !createAttempt.current) createAttempt.current = { idempotencyKey: crypto.randomUUID(), body: content };
    const pending = createAttempt.current;
    const previousCreateOutcomeUnknown = createOutcomeUnknown;
    guard.block(true);
    lock.current = true; setBusy(true); setCreateOutcomeUnknown(true); setStatus("正在建立問題……");
    void write.run(
      (signal) => request((client, s) => client.POST("/api/v1/repositories/{repositoryId}/issues", {
        signal: s, params: { path: { repositoryId }, header: { "Idempotency-Key": pending.idempotencyKey } }, body: pending.body,
      }), { signal }),
      (item) => { lock.current = false; setBusy(false); saved(item, "已建立問題。"); },
      (error) => {
        lock.current = false; setBusy(false);
        if (!previousCreateOutcomeUnknown && ["invalid_input", "forbidden", "not_found", "idempotency_conflict"].includes((error as { code?: string }).code || "")) { setCreateOutcomeUnknown(false); guard.block(false); }
        setStatus(errorMessage(error) + " 結果未確認時請沿用原內容重試，或先查詢問題；重新開頁前請先確認紀錄。");
      },
    );
  }
  function save(state: Issue["state"]) {
    if (!user || !selected || lock.current || navigating.current || needsRefresh || discard.isConfirming() || !guard.canWrite()) return;
    const id = selected.id;
    guard.block(true);
    lock.current = true; setBusy(true); setStatus("正在更新問題……");
    void write.run(
      (signal) => request((client, s) => client.PATCH("/api/v1/repositories/{repositoryId}/issues/{id}", {
        signal: s, params: { path: { repositoryId, id } }, body: { title: title.trim(), body: body.trim(), state, expectedVersion: selected.version, labelIds:labelIDs,milestoneId:milestoneID||null },
      }), { signal }),
      (item) => { lock.current = false; setBusy(false); saved(item, "已更新問題。"); },
      (error) => { lock.current = false; setBusy(false); setNeedsRefresh(true); setStatus(errorMessage(error) + " 請重新讀取問題後再修改。"); },
    );
  }
  const editor = <section ref={editorPanel} tabIndex={-1} aria-label="問題編輯" className="flex w-full min-w-0 flex-col gap-3">
      <ReadFeedback busy={detailBusy} message={status} label="正在讀取問題……" />
      {mode === "detail" && <div className="mb-4 flex flex-wrap items-center gap-2">
        <Button variant="outline" disabled={busy || detailBusy} onClick={() => open(activeId)}>重新讀取問題</Button>
      </div>}
      {(mode === "create" || selected) && <form onSubmit={submit}>
        <FieldGroup>
          <fieldset className="grid min-w-0 gap-4" disabled={busy || detailBusy || createOutcomeUnknown || needsRefresh || (mode === "detail" && !mine)}>
            <FormField>標題<Input ref={titleInput} required maxLength={120} value={title} onChange={(e) => setTitle(e.target.value)} /></FormField>
            <FormField>內容<Textarea maxLength={8000} value={body} onChange={(e) => setBody(e.target.value)} /></FormField>
          </fieldset>
          {selected && <MilestonePicker repositoryId={repositoryId} value={milestoneID} onChange={setMilestoneID} disabled={busy || detailBusy || createOutcomeUnknown || needsRefresh} />}
          {selected && <LabelPicker repositoryId={repositoryId} value={labelIDs} onChange={setLabelIDs} disabled={busy || detailBusy || createOutcomeUnknown || needsRefresh} />}
          {(mode === "create" || selected) && <div className="flex flex-wrap gap-2">
            <Button type="submit" disabled={busy || needsRefresh}>{mode === "create" ? createOutcomeUnknown ? "重試原建立請求" : "建立問題" : mine ? "儲存內容" : "儲存分類"}</Button>
            {selected && mine && <Button type="button" variant="outline" disabled={busy || needsRefresh} onClick={() => save(selected.state === "completed" ? "open" : "completed")}>{selected.state === "completed" ? "重新開啟" : "關閉問題"}</Button>}
          </div>}
        </FieldGroup>
      </form>}
      <Button type="button" variant="outline" disabled={busy || createOutcomeUnknown || needsRefresh} onClick={close}>{mine || mode === "create" ? "取消" : "關閉"}</Button>
    </section>;
  return <>
    <PagePanel heading="問題">
      <p className="mb-4 text-sm text-muted-foreground">空間成員可整理標籤及里程碑；問題內容與狀態由建立者修改。</p>
      <div className="mb-4 flex flex-wrap gap-2">
        <Button disabled={!user || busy || createOutcomeUnknown || needsRefresh} onClick={event => openCreateForm(event.currentTarget)}>新增問題</Button>
        <Button variant="outline" disabled={!user || listBusy || busy || !!mode || createOutcomeUnknown || needsRefresh} onClick={() => load()}>重新載入問題</Button>
      </div>
      <form onSubmit={(event) => { event.preventDefault(); if (!lock.current && !createOutcomeUnknown && !needsRefresh) discard.navigate(() => { navigating.current = true; clearEditor(); setFilter({ query: query.trim(), state }); }); }}>
        <FieldGroup className="mb-4">
          <FormField>搜尋標題<Input maxLength={120} value={query} onChange={(e) => setQuery(e.target.value)} /></FormField>
          <FormField>狀態<NativeSelect value={state} onChange={(e) => setState(e.target.value as typeof state)}>
            <NativeSelectOption value="">全部</NativeSelectOption>
            {Object.entries(states).map(([value, label]) => <NativeSelectOption key={value} value={value}>{label}</NativeSelectOption>)}
          </NativeSelect></FormField>
          <Button type="submit" disabled={!user || listBusy || busy}>套用篩選</Button>
        </FieldGroup>
      </form>
      <ReadFeedback busy={listBusy} message={listStatus} label="正在讀取問題……" />
      {mode === "create" && <PagePanel heading="新增問題">{editor}</PagePanel>}
      <ItemGroup className="gap-3">
        {items.map((item) => <Item key={item.id} variant="outline" role="listitem" className="min-w-0">
          <ItemContent className="min-w-0">
            <ItemTitle className="min-w-0 max-w-full"><Button variant="link" className="h-auto min-w-0 max-w-full justify-start whitespace-normal text-left [overflow-wrap:anywhere]" disabled={busy || detailBusy || createOutcomeUnknown || needsRefresh} onClick={event => open(item.id, event.currentTarget)}>#{item.number} {item.title}</Button></ItemTitle>
            <Badge variant="outline">{states[item.state]}</Badge>
          </ItemContent>
          {item.creatorId === user?.userId && <ItemActions><Button variant="outline" aria-label={`編輯問題：${item.title}`} disabled={busy || detailBusy || createOutcomeUnknown || needsRefresh} onClick={event => open(item.id, event.currentTarget)}>編輯</Button></ItemActions>}
          {mode === "detail" && activeId === item.id && <ItemFooter>{editor}</ItemFooter>}
        </Item>)}
      </ItemGroup>
      {loaded && !items.length && <Empty><EmptyHeader><EmptyTitle>{cursor ? "本頁沒有符合的問題" : "沒有符合的問題"}</EmptyTitle><EmptyDescription>{cursor ? "請繼續載入下一頁。" : "可調整篩選條件，或建立新的問題。"}</EmptyDescription></EmptyHeader></Empty>}
      {cursor && <Button variant="outline" disabled={listBusy || busy || !!mode || createOutcomeUnknown || needsRefresh} onClick={() => load(cursor)}>載入更多問題</Button>}
    </PagePanel>

    {discard.dialog}
  </>;
}
