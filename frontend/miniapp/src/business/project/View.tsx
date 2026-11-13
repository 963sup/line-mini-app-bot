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
import { Items } from "./Items";
import { SourcePicker } from "./SourcePicker";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from "@/shared/ui/empty";
import { Item, ItemGroup, ItemContent, ItemTitle } from "@/shared/ui/item";

type Project = components["schemas"]["Project"];
type Creation = components["schemas"]["SpaceProjectCreation"];

export function View({ spaceId, projectId, initialSource, onSelect }: { spaceId: string; projectId: string; initialSource: string; onSelect: (id: string) => void }) {
  const { user, request } = useSession();
  const { run: readRun, cancel: cancelRead } = useLatestRequest();
  const detailRequest = useLatestRequest(), write = useLatestRequest();
  const [items, setItems] = useState<Project[]>([]), [cursor, setCursor] = useState("");
  const [listBusy, setListBusy] = useState(false), [loaded, setLoaded] = useState(false);
  const [listStatus, setListStatus] = useState("");
  const [mode, setMode] = useState<"create" | "detail" | null>(null);
  const [detail, setSelected] = useState<Project | null>(null);
  const selected = detail?.id === projectId ? detail : null;
  const [title, setTitle] = useState(""), [body, setBody] = useState("");
  const [sourceRepositoryId, setSourceRepositoryId] = useState("");
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
    setListStatus("正在讀取專案……");
    void readRun(
      (signal) => request((client, s) => client.GET("/api/v1/spaces/{spaceId}/projects", {
        signal: s, params: { path: { spaceId }, query: { cursor: nextCursor || undefined } },
      }), { signal }),
      (page) => {
        setItems((previous) => [...new Map([...(nextCursor ? previous : []), ...page.items].map((item) => [item.id, item])).values()]);
        setCursor(page.nextCursor); setListBusy(false); setLoaded(true);
        setListStatus(page.nextCursor ? "可繼續載入更多專案。" : "已讀取完畢。");
      },
      (error) => { setItems([]); setListBusy(false); setLoaded(false); setListStatus(errorMessage(error)); },
    );
  }, [user, request, spaceId, readRun]);
  useEffect(() => { load(); return cancelRead; }, [load, cancelRead]);

  const show = useCallback((item: Project) => {
    setSelected(item); setTitle(item.title); setBody(item.body); setMode("detail");
    setNeedsRefresh(false); setUncertain(false); attempt.current = null;
  }, []);
  const open = useCallback((id: string) => {
    if (!user || lock.current) return;
    setSelected(null); setMode("detail"); setDetailBusy(true); setStatus("正在核驗並讀取專案……");
    void detailRequest.run(
      (signal) => request((client, s) => client.GET("/api/v1/spaces/{spaceId}/projects/{id}", { signal: s, params: { path: { spaceId, id } } }), { signal }),
      (item) => { show(item); setDetailBusy(false); setStatus("已讀取專案。"); },
      (error) => { setDetailBusy(false); setStatus(errorMessage(error)); },
    );
  }, [user, request, spaceId, detailRequest.run, show]);
  useEffect(() => {
    detailRequest.cancel(); write.cancel(); lock.current = false;
    setBusy(false); setDetailBusy(false); setSelected(null); setMode(null); setUncertain(false); setNeedsRefresh(false); setStatus(""); attempt.current = null;
    setSourceRepositoryId("");
    if (projectId === "new") { setMode("create"); setTitle(""); setBody(""); }
    else if (projectId) {
      if (/^[0-9a-f]{32}$/.test(projectId)) open(projectId);
      else setStatus("專案連結無效。");
    }
    return () => { detailRequest.cancel(); write.cancel(); };
  }, [projectId, open, detailRequest.cancel, write.cancel]);
  function create() {
    if (lock.current) return;
    onSelect("new"); detailRequest.cancel(); setDetailBusy(false); setSelected(null); setMode("create");
    setTitle(""); setBody(""); setUncertain(false); setNeedsRefresh(false); setStatus(""); attempt.current = null;
  }
  function submit(event: FormEvent) {
    event.preventDefault();
    if (!user || lock.current || needsRefresh) return;
    if (mode === "detail") { if (selected) save(); return; }
    const content = { title: title.trim(), body: body.trim(), sourceRepositoryId };
    if (!content.title || !sourceRepositoryId) { setStatus("請填寫專案標題並選擇來源儲存庫。"); return; }
    if (!uncertain || !attempt.current) attempt.current = { key: crypto.randomUUID(), body: content };
    const pending = attempt.current;
    lock.current = true; setBusy(true); setUncertain(true); setStatus("正在建立專案……");
    void write.run(
      (signal) => request((client, s) => client.POST("/api/v1/spaces/{spaceId}/projects", {
        signal: s, params: { path: { spaceId }, header: { "Idempotency-Key": pending.key } }, body: pending.body,
      }), { signal }),
      (item) => { lock.current = false; setBusy(false); show(item); onSelect(item.id); setStatus("已建立專案。"); load(); },
      (error) => {
        lock.current = false; setBusy(false);
        if (["invalid_input", "forbidden", "not_found", "idempotency_conflict"].includes((error as { code?: string }).code || "")) setUncertain(false);
        setStatus(errorMessage(error) + " 結果未確認時請沿用原內容重試，或先查詢專案；重新開頁前請先確認紀錄。");
      },
    );
  }
  function save() {
    if (!user || !selected || !mine || lock.current || needsRefresh) return;
    const id = selected.id;
    lock.current = true; setBusy(true); setStatus("正在更新專案……");
    void write.run(
      (signal) => request((client, s) => client.PATCH("/api/v1/spaces/{spaceId}/projects/{id}", {
        signal: s, params: { path: { spaceId, id } }, body: { title: title.trim(), body: body.trim(), expectedVersion: selected.version },
      }), { signal }),
      (item) => { lock.current = false; setBusy(false); show(item); setStatus("已更新專案。"); load(); },
      (error) => { lock.current = false; setBusy(false); setNeedsRefresh(true); setStatus(errorMessage(error) + " 請重新讀取專案後再修改。"); },
    );
  }
  return <>
    <PagePanel heading="專案">
      <p className="mb-4 text-sm text-muted-foreground">空間成員可建立專案；只有建立者可以修改及管理收錄事項。</p>
      <div className="mb-4 flex flex-wrap gap-2">
        <Button disabled={!user || busy || uncertain} onClick={create}>新增專案</Button>
        <Button variant="outline" disabled={!user || listBusy || busy} onClick={() => load()}>重新載入專案</Button>
      </div>
      <Status role="status">{listStatus}</Status>
      {!mode && <Status role="status">{status}</Status>}
      <ItemGroup className="gap-3">
        {items.map((item) => <Item key={item.id} variant="outline" role="listitem">
          <ItemContent>
            <ItemTitle><Button variant="link" className="h-auto whitespace-normal text-left" disabled={busy || detailBusy || uncertain} onClick={() => onSelect(item.id)}>{item.title}</Button></ItemTitle>

          </ItemContent>
        </Item>)}
      </ItemGroup>
      {loaded && !items.length && <Empty><EmptyHeader><EmptyTitle>{cursor ? "本頁沒有專案" : "目前沒有專案"}</EmptyTitle><EmptyDescription>{cursor ? "請繼續載入下一頁。" : "可建立新的專案。"}</EmptyDescription></EmptyHeader></Empty>}
      {cursor && <Button variant="outline" disabled={listBusy || busy} onClick={() => load(cursor)}>載入更多專案</Button>}
    </PagePanel>
    {mode && <PagePanel heading={mode === "create" ? "新增專案" : selected ? selected.title : "專案詳情"}>
      <Status role="status">{status}</Status>
      {selected && <p>{selected.sourceAvailable === false ? "來源儲存庫目前不可讀取，暫不提供事項與進度。" : selected.issueCount ? `已完成 ${selected.completedIssueCount}/${selected.issueCount} 項（${Math.round(selected.completedIssueCount/selected.issueCount*100)}%）` : "尚無事項"}</p>}
      {selected && <div className="mb-4 flex flex-wrap items-center gap-2">

        <Button variant="outline" disabled={busy || detailBusy} onClick={() => open(selected.id)}>重新讀取專案</Button>
      </div>}
      {(mode === "create" || selected) && <form onSubmit={submit}>
        <FieldGroup>
          {mode === "create" && <SourcePicker spaceId={spaceId} initialSource={initialSource} value={sourceRepositoryId} onChange={setSourceRepositoryId} disabled={busy || uncertain} />}
          <fieldset className="grid min-w-0 gap-4" disabled={busy || detailBusy || uncertain || needsRefresh || (mode === "detail" && !mine)}>
            <FormField>標題<Input required maxLength={120} value={title} onChange={(e) => setTitle(e.target.value)} /></FormField>
            <FormField>內容<Textarea maxLength={8000} value={body} onChange={(e) => setBody(e.target.value)} /></FormField>
          </fieldset>

          {(mode === "create" || mine) && <div className="flex flex-wrap gap-2">
            <Button type="submit" disabled={busy || needsRefresh}>{mode === "create" ? uncertain ? "重試原建立請求" : "建立專案" : "儲存內容"}</Button>
          </div>}
        </FieldGroup>
      </form>}
    </PagePanel>}
    {selected && selected.sourceAvailable !== false && <Items key={selected.id+":"+selected.version} spaceId={spaceId} sourceRepositoryId={selected.sourceRepositoryId || selected.repositoryId} project={selected} disabled={busy||detailBusy||needsRefresh||title!==selected.title||body!==selected.body} onChange={(item)=>{show(item);load();}} onRefresh={()=>open(selected.id)} />}
  </>;
}
