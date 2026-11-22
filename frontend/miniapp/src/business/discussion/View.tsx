import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from "@/shared/ui/input-group";
import { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import { useRepositoryGuard } from "../repository/NavigationGuard";
import { useDiscardChanges } from "@/shared/useDiscardChanges";
import { ReadFeedback } from "@/shared/ReadFeedback";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { rejectedAttempt, useLatestRequest, errorMessage } from "@/shared/http/lifecycle";
import { Picker as LabelPicker } from "../repository/labels/Picker";
import { PagePanel } from "@/shared/PagePanel";
import { FormField } from "@/shared/FormField";
import { Comments } from "./Comments";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Textarea } from "@/shared/ui/textarea";
import { FieldGroup, FieldSet, Field, FieldLabel } from "@/shared/ui/field";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from "@/shared/ui/empty";
import { Item, ItemGroup, ItemContent, ItemTitle } from "@/shared/ui/item";

type Discussion = components["schemas"]["Discussion"];
type Creation = components["schemas"]["DiscussionCreation"];

export function View({ repositoryId }: { repositoryId: string }) {
  const { user, request } = useSession();
  const { run: readRun, cancel: cancelRead } = useLatestRequest();
  const detailRequest = useLatestRequest(), write = useLatestRequest();
  const [items, setItems] = useState<Discussion[]>([]), [cursor, setCursor] = useState("");
  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState({ query: "" });
  const [listBusy, setListBusy] = useState(false), [loaded, setLoaded] = useState(false);
  const [listStatus, setListStatus] = useState("");
  const [mode, setMode] = useState<"create" | "detail" | null>(null);
  const [selected, setSelected] = useState<Discussion | null>(null);
  const [title, setTitle] = useState(""), [body, setBody] = useState("");
  const [labelIDs,setLabelIDs]=useState<string[]>([]);
  const [busy, setBusy] = useState(false), [detailBusy, setDetailBusy] = useState(false);
  const [createOutcomeUnknown, setCreateOutcomeUnknown] = useState(false), [needsRefresh, setNeedsRefresh] = useState(false);
  const [status, setStatus] = useState("");
  const lock = useRef(false);
  const commentLock = useRef(false);
  const navigating = useRef(false);
  const [commentsBlocked, setCommentsBlocked] = useState(false);
  const onCommentWriteStateChange = useCallback((blocked: boolean) => {
    // Guard same-batch events before React commits the disabled buttons.
    if (blocked && navigating.current) return false;
    commentLock.current = blocked;
    setCommentsBlocked(blocked);
    return true;
  }, []);
  const createAttempt = useRef<{ idempotencyKey: string; body: Creation } | null>(null);
  const canEdit = selected?.canEdit === true;
  const [activeId, setActiveId] = useState("");
  const baseline = useRef(JSON.stringify(["", "", []]));
  const dirty = !!mode && JSON.stringify([title, body, labelIDs]) !== baseline.current;
  const titleInput = useRef<HTMLInputElement>(null);
  const focusEditor = () => titleInput.current?.focus();
  const discard = useDiscardChanges(dirty, focusEditor);
  const guard = useRepositoryGuard(dirty, focusEditor);
  useEffect(() => { guard.block(busy || createOutcomeUnknown || needsRefresh); }, [busy, createOutcomeUnknown, needsRefresh]);
  useEffect(() => { if (!detailBusy) navigating.current = false; }, [mode, detailBusy]);

  const load = useCallback((nextCursor = "") => {
    if (!user) return;
    setListBusy(true); setLoaded(false); setCursor("");
    if (!nextCursor) setItems([]);
    setListStatus("");
    void readRun(
      (signal) => request((client, s) => client.GET("/api/v1/repositories/{repositoryId}/discussions", {
        signal: s, params: { path: { repositoryId }, query: { cursor: nextCursor || undefined, query: filter.query || undefined } },
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

  function show(item: Discussion) {
    setActiveId(item.id); baseline.current = JSON.stringify([item.title, item.body, item.labelIds || []]); guard.block(false);
    navigating.current = false;
    setSelected(item); setTitle(item.title); setBody(item.body); setLabelIDs(item.labelIds || []); setMode("detail");
    setNeedsRefresh(false); setCreateOutcomeUnknown(false); createAttempt.current = null;
  }
  function open(id: string) {
    if (!user || lock.current || commentLock.current || createOutcomeUnknown || (needsRefresh && id !== activeId) || !guard.canWrite()) return;
    discard.navigate(() => {
      navigating.current = true;
      setActiveId(id); baseline.current = JSON.stringify(["", "", []]); setTitle(""); setBody(""); setLabelIDs([]);
      setSelected(null); setMode("detail"); setDetailBusy(true); setStatus("");
      void detailRequest.run(
        (signal) => request((client, s) => client.GET("/api/v1/repositories/{repositoryId}/discussions/{id}", { signal: s, params: { path: { repositoryId, id } } }), { signal }),
        (item) => { show(item); setDetailBusy(false); setStatus(""); },
        (error) => { navigating.current = false; setDetailBusy(false); setStatus(errorMessage(error)); },
      );
    });
  }
  function openCreateForm() {
    if (lock.current || commentLock.current || createOutcomeUnknown || needsRefresh || !guard.canWrite()) return;
    discard.navigate(() => {
      setActiveId(""); baseline.current = JSON.stringify(["", "", []]); setLabelIDs([]);
      navigating.current = true;
      detailRequest.cancel(); setDetailBusy(false); setSelected(null); setMode("create");
      setTitle(""); setBody(""); setCreateOutcomeUnknown(false); setNeedsRefresh(false); setStatus(""); createAttempt.current = null;
    });
  }

  function submit(event: FormEvent) {
    event.preventDefault();
    if (!user || lock.current || navigating.current || needsRefresh || discard.isConfirming() || !guard.canWrite()) return;
    if (mode === "detail") { if (selected) save(); return; }
    const content = { title: title.trim(), body: body.trim() };
    if (!content.title) { setStatus("請填寫討論標題。"); return; }
    if (!createOutcomeUnknown || !createAttempt.current) createAttempt.current = { idempotencyKey: crypto.randomUUID(), body: content };
    const pending = createAttempt.current;
    const previousCreateOutcomeUnknown = createOutcomeUnknown; guard.block(true);
    lock.current = true; setBusy(true); setCreateOutcomeUnknown(true); setStatus("正在建立討論……");
    void write.run(
      (signal) => request((client, s) => client.POST("/api/v1/repositories/{repositoryId}/discussions", {
        signal: s, params: { path: { repositoryId }, header: { "Idempotency-Key": pending.idempotencyKey } }, body: pending.body,
      }), { signal }),
      (item) => { lock.current = false; setBusy(false); show(item); setStatus("已建立討論。"); load(); },
      (error) => {
        lock.current = false; setBusy(false);
        if (!previousCreateOutcomeUnknown && rejectedAttempt(error)) { setCreateOutcomeUnknown(false); guard.block(false); }
        setStatus(errorMessage(error) + " 結果未確認時請沿用原內容重試，或先查詢討論；重新開頁前請先確認紀錄。");
      },
    );
  }
  function save() {
    if (!user || !selected || lock.current || navigating.current || needsRefresh || discard.isConfirming() || !guard.canWrite()) return;
    const id = selected.id; guard.block(true);
    lock.current = true; setBusy(true); setStatus("正在更新討論……");
    void write.run(
      (signal) => request((client, s) => client.PATCH("/api/v1/repositories/{repositoryId}/discussions/{id}", {
        signal: s, params: { path: { repositoryId, id } }, body: { title: title.trim(), body: body.trim(), expectedVersion: selected.version, labelIds:labelIDs },
      }), { signal }),
      (item) => { lock.current = false; setBusy(false); show(item); setStatus("已更新討論。"); load(); },
      (error) => { lock.current = false; setBusy(false); setNeedsRefresh(true); setStatus(errorMessage(error) + " 請重新讀取討論後再修改。"); },
    );
  }
  return <>
    <PagePanel heading="討論">
      <p className="mb-4 text-sm text-muted-foreground">空間成員可整理討論標籤；主題正文與留言各由作者編輯。</p>
      <div className="mb-4 flex flex-wrap gap-2">
        <Button disabled={!user || busy || createOutcomeUnknown || needsRefresh || commentsBlocked} onClick={openCreateForm}>發起討論</Button>
      </div>
      <form onSubmit={(event) => { event.preventDefault(); if (!lock.current && !createOutcomeUnknown && !needsRefresh && !commentLock.current) discard.navigate(() => { navigating.current = true; detailRequest.cancel(); setDetailBusy(false); setMode(null); setSelected(null); setStatus(""); setFilter({ query: query.trim() }); }); }}>
        <Field className="mb-4">
        <FieldLabel htmlFor="discussion-search">搜尋標題</FieldLabel>
        <InputGroup>
          <InputGroupInput id="discussion-search" maxLength={120} value={query} onChange={e => setQuery(e.target.value)} />
          <InputGroupAddon align="inline-end">
            <InputGroupButton type="submit" disabled={!user || listBusy || busy}>套用篩選</InputGroupButton>
          </InputGroupAddon>
        </InputGroup>
        </Field>
      </form>
      <ReadFeedback busy={listBusy} message={listStatus} label="正在讀取討論……" onRetry={user && listStatus && !loaded && !busy ? () => load() : undefined} />
      <ItemGroup className="gap-3">
        {items.map((item) => <Item key={item.id} variant="outline" role="listitem">
          <ItemContent>
            <ItemTitle><Button variant="link" className="h-auto whitespace-normal text-left" disabled={busy || detailBusy || createOutcomeUnknown || needsRefresh || commentsBlocked} onClick={() => open(item.id)}>#{item.number} {item.title}</Button></ItemTitle>
          </ItemContent>
        </Item>)}
      </ItemGroup>
      {loaded && !items.length && <Empty><EmptyHeader><EmptyTitle>{cursor ? "本頁沒有符合的討論" : "沒有符合的討論"}</EmptyTitle><EmptyDescription>{cursor ? "請繼續載入下一頁。" : "可調整篩選條件，或建立新的討論。"}</EmptyDescription></EmptyHeader></Empty>}
      {cursor && <Button variant="outline" disabled={listBusy || busy} onClick={() => load(cursor)}>載入更多討論</Button>}
    </PagePanel>
    {mode && <PagePanel heading={mode === "create" ? "發起討論" : selected ? `#${selected.number} ${selected.title}` : "討論詳情"}>
      <ReadFeedback busy={detailBusy} message={status} label="正在讀取討論……" />
      {mode === "detail" && <div className="mb-4 flex flex-wrap items-center gap-2">
        <Button variant="outline" disabled={busy || detailBusy || commentsBlocked} onClick={() => open(activeId)}>重新讀取討論</Button>
      </div>}
      {(mode === "create" || selected) && <form onSubmit={submit}>
        <FieldGroup>
          <FieldSet className="grid min-w-0 gap-4" disabled={busy || detailBusy || createOutcomeUnknown || needsRefresh || (mode === "detail" && !canEdit)}>
            <FormField>標題<Input ref={titleInput} required maxLength={120} value={title} onChange={(e) => setTitle(e.target.value)} /></FormField>
            <FormField>內容<Textarea maxLength={8000} value={body} onChange={(e) => setBody(e.target.value)} /></FormField>
          </FieldSet>
          {selected && <LabelPicker repositoryId={repositoryId} value={labelIDs} onChange={setLabelIDs} disabled={busy || detailBusy || createOutcomeUnknown || needsRefresh} />}
          {(mode === "create" || selected) && <div className="flex flex-wrap gap-2">
            <Button type="submit" disabled={busy || needsRefresh}>{mode === "create" ? createOutcomeUnknown ? "重試原建立請求" : "建立討論" : canEdit ? "儲存內容" : "儲存分類"}</Button>
          </div>}
        </FieldGroup>
      </form>}
      <Button type="button" variant="outline" disabled={busy || createOutcomeUnknown || needsRefresh || commentsBlocked} onClick={() => discard.navigate(() => { navigating.current = true; detailRequest.cancel(); setMode(null); setSelected(null); })}>關閉</Button>
    </PagePanel>}
    {selected && <Comments key={selected.id} repositoryId={repositoryId} discussionId={selected.id} onWriteStateChange={onCommentWriteStateChange} />}
    {discard.dialog}
  </>;
}
