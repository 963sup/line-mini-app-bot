import { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { rejectedAttempt, useLatestRequest, errorMessage } from "@/shared/http/lifecycle";
import { PagePanel } from "@/shared/PagePanel";
import { FormField } from "@/shared/FormField";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Textarea } from "@/shared/ui/textarea";
import { FieldGroup, FieldSet } from "@/shared/ui/field";
import { SettingsEditor } from "./SettingsEditor";
import { Items } from "./Items";
import { SourcePicker } from "./SourcePicker";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from "@/shared/ui/empty";
import { Item, ItemGroup, ItemContent, ItemTitle } from "@/shared/ui/item";

type Project = components["schemas"]["Project"];
type Creation = components["schemas"]["OwnerProjectCreation"];

export function View({ ownerId, projectId, initialSource, copyFrom, destinations, onCopy, onSelect, onBlocked, canWrite }: { ownerId: string; projectId: string; initialSource: string; copyFrom?: Creation["copyFrom"]; destinations: {id:string;name:string}[]; onCopy:(destination:string)=>void; onSelect: (id: string) => void; onBlocked: (value: boolean) => void; canWrite: () => boolean }) {
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
  const [settingsDraft, setSettingsDraft] = useState<Project["settings"]>();
  const [sourceRepositoryId, setSourceRepositoryId] = useState("");
  const [busy, setBusy] = useState(false), [detailBusy, setDetailBusy] = useState(false);
  const [uncertain, setUncertain] = useState(false), [needsRefresh, setNeedsRefresh] = useState(false);
  const [status, setStatus] = useState("");
  const lock = useRef(false), unknown = useRef(false), referenceLock = useRef(false), navigation = useRef(false);
  const [referencesBlocked, setReferencesBlocked] = useState(false);
  const referenceBlock = useCallback((value: boolean) => { referenceLock.current = value; setReferencesBlocked(value); onBlocked(value || lock.current || unknown.current); }, [onBlocked]);
  const block = useCallback((value: boolean) => { unknown.current = value; onBlocked(value || referenceLock.current); }, [onBlocked]);
  useEffect(() => { navigation.current = false; });
  useEffect(() => () => onBlocked(false), [onBlocked]);
  useEffect(() => {
    if (!busy && !uncertain && !referencesBlocked) return;
    const warn = (event: BeforeUnloadEvent) => { event.preventDefault(); event.returnValue = ""; };
    window.addEventListener("beforeunload", warn); return () => window.removeEventListener("beforeunload", warn);
  }, [busy, uncertain, referencesBlocked]);
  const attempt = useRef<{ key: string; body: Creation } | null>(null);
  const canEdit = selected?.canEdit === true;

  const load = useCallback((nextCursor = "") => {
    if (!user || referenceLock.current) return;
    setListBusy(true); setLoaded(false); setCursor("");
    if (!nextCursor) setItems([]);
    setListStatus("正在讀取專案……");
    void readRun(
      (signal) => request((client, s) => client.GET("/api/v1/owners/{ownerId}/projects", {
        signal: s, params: { path: { ownerId }, query: { cursor: nextCursor || undefined } },
      }), { signal }),
      (page) => {
        setItems((previous) => [...new Map([...(nextCursor ? previous : []), ...page.items].map((item) => [item.id, item])).values()]);
        setCursor(page.nextCursor); setListBusy(false); setLoaded(true);
        setListStatus(page.nextCursor ? "可繼續載入更多專案。" : "已讀取完畢。");
      },
      (error) => { setItems([]); setListBusy(false); setLoaded(false); setListStatus(errorMessage(error)); },
    );
  }, [user, request, ownerId, readRun]);
  useEffect(() => { load(); return cancelRead; }, [load, cancelRead]);

  const show = useCallback((item: Project) => {
    setSettingsDraft(undefined); setSelected(item); setTitle(item.title); setBody(item.body); setMode("detail");
    setNeedsRefresh(false); setUncertain(false); attempt.current = null; block(false);
  }, [block]);
  const open = useCallback((id: string) => {
    if (!user || lock.current || referenceLock.current) return;
    navigation.current = true;
    setSelected(null); setMode("detail"); setDetailBusy(true); setStatus("正在核驗並讀取專案……");
    void detailRequest.run(
      (signal) => request((client, s) => client.GET("/api/v1/owners/{ownerId}/projects/{id}", { signal: s, params: { path: { ownerId, id } } }), { signal }),
      (item) => { show(item); setDetailBusy(false); setStatus("已讀取專案。"); },
      (error) => { setDetailBusy(false); setStatus(errorMessage(error)); },
    );
  }, [user, request, ownerId, detailRequest.run, show, canWrite]);
  useEffect(() => {
    detailRequest.cancel(); write.cancel(); lock.current = false; block(false);
    setBusy(false); setDetailBusy(false); setSelected(null); setMode(null); setUncertain(false); setNeedsRefresh(false); setStatus(""); attempt.current = null;
    setSourceRepositoryId("");
    if (projectId === "new") { setMode("create"); setTitle(""); setBody(""); }
    else if (projectId) {
      if (/^[0-9a-f]{32}$/.test(projectId)) open(projectId);
      else setStatus("專案連結無效。");
    }
    return () => { detailRequest.cancel(); write.cancel(); };
  }, [projectId, open, detailRequest.cancel, write.cancel, block]);
  function openCreateForm() {
    if (lock.current || unknown.current || referenceLock.current || !canWrite()) return;
    navigation.current = true;
    onSelect("new"); detailRequest.cancel(); setDetailBusy(false); setSelected(null); setMode("create");
    setTitle(""); setBody(""); setUncertain(false); setNeedsRefresh(false); setStatus(""); attempt.current = null;
  }
  function submit(event: FormEvent) {
    event.preventDefault();
    if (!user || lock.current || referenceLock.current || navigation.current || needsRefresh || !canWrite()) return;
    if (mode === "detail") { if (selected) save(); return; }
    const content = { title: title.trim(), body: body.trim(), defaultRepositoryId: copyFrom ? undefined : sourceRepositoryId || undefined, copyFrom };
    if (!content.title) { setStatus("請填寫專案標題。"); return; }
    if (!uncertain || !attempt.current) attempt.current = { key: crypto.randomUUID(), body: content };
    const pending = attempt.current;
    const wasUnknown = unknown.current;
    lock.current = true; block(true); setBusy(true); setUncertain(true); setStatus("正在建立專案……");
    void write.run(
      (signal) => request((client, s) => client.POST("/api/v1/owners/{ownerId}/projects", {
        signal: s, params: { path: { ownerId }, header: { "Idempotency-Key": pending.key } }, body: pending.body,
      }), { signal }),
      (item) => { lock.current = false; setBusy(false); show(item); onSelect(item.id); setStatus("已建立專案。"); load(); },
      (error) => {
        lock.current = false; setBusy(false);
        if (!wasUnknown && rejectedAttempt(error)) { setUncertain(false); block(false); }
        setStatus(errorMessage(error) + " 結果未確認時請沿用原內容重試，或先查詢專案；重新開頁前請先確認紀錄。");
      },
    );
  }
  function save() {
    if (!user || !selected || !canEdit || lock.current || unknown.current || referenceLock.current || navigation.current || needsRefresh || !canWrite()) return;
    const id = selected.id;
    lock.current = true; block(true); setBusy(true); setStatus("正在更新專案……");
    void write.run(
      (signal) => request((client, s) => client.PATCH("/api/v1/owners/{ownerId}/projects/{id}", {
        signal: s, params: { path: { ownerId, id } }, body: { title: title.trim(), body: body.trim(), expectedVersion: selected.version, settings: settingsDraft },
      }), { signal }),
      (item) => { lock.current = false; setBusy(false); show(item); setStatus("已更新專案。"); load(); },
      (error) => {
        lock.current = false; setBusy(false); setNeedsRefresh(true);
        const rejected = rejectedAttempt(error, ["version_conflict", "conflict"]);
        if (rejected) block(false); else { block(true); setUncertain(true); }
        setStatus(errorMessage(error) + " 請重新讀取專案以確認更新結果，再修改。");
      },
    );
  }
  return <div className="flex min-w-0 flex-col gap-4">
    <PagePanel heading="專案" description="建立個人或組織專案，並整理相關的物件、問題與討論。">
      <div className="flex flex-wrap gap-2">
        <Button disabled={!user || busy || uncertain || referencesBlocked} onClick={openCreateForm}>新增專案</Button>
        <Button variant="outline" disabled={!user || listBusy || busy || uncertain || referencesBlocked} onClick={() => load()}>重新載入專案</Button>
      </div>
      <Status role="status">{listStatus}</Status>
      {!mode && <Status role="status">{status}</Status>}
      <ItemGroup className="gap-3">
        {items.map((item) => <Item key={item.id} variant="outline" role="listitem">
          <ItemContent className="min-w-0">
            <ItemTitle className="min-w-0 max-w-full"><Button variant="link" className="h-auto min-w-0 max-w-full justify-start whitespace-normal text-left [overflow-wrap:anywhere]" disabled={busy || detailBusy || uncertain || referencesBlocked} onClick={() => { if (lock.current || unknown.current || referenceLock.current || !canWrite()) return; navigation.current = true; onSelect(item.id); }}>{item.title}</Button></ItemTitle>

          </ItemContent>
        </Item>)}
      </ItemGroup>
      {loaded && !items.length && <Empty><EmptyHeader><EmptyTitle>{cursor ? "本頁沒有專案" : "目前沒有專案"}</EmptyTitle><EmptyDescription>{cursor ? "請繼續載入下一頁。" : "可建立新的專案。"}</EmptyDescription></EmptyHeader></Empty>}
      {cursor && <Button variant="outline" disabled={listBusy || busy || uncertain || referencesBlocked} onClick={() => load(cursor)}>載入更多專案</Button>}
    </PagePanel>
    {mode && <PagePanel heading={mode === "create" ? copyFrom ? "建立專案副本" : "新增專案" : "專案詳情"} description={selected && <span className="block [overflow-wrap:anywhere]">{selected.title}</span>}>
      <Status role="status">{status}</Status>
      {mode === "detail" && /^[a-f0-9]{32}$/.test(projectId) && <div className="flex flex-wrap items-center gap-2">

        {selected && <FormField>複製到另一個擁有者<select value="" disabled={busy || detailBusy || uncertain || referencesBlocked} onChange={event => { if(event.target.value) onCopy(event.target.value); }}>
          <option value="">選擇副本目的地</option>
          {destinations.filter(owner => owner.id !== ownerId).map(owner => <option key={owner.id} value={owner.id}>{owner.name}</option>)}
        </select></FormField>}
        <Button variant="outline" disabled={busy || detailBusy || referencesBlocked} onClick={() => open(projectId)}>重新讀取專案</Button>
      </div>}
      {(mode === "create" || selected) && <form onSubmit={submit}>
        <FieldGroup>
          {mode === "create" && !copyFrom && <SourcePicker key={ownerId} ownerId={ownerId} initialSource={initialSource} value={sourceRepositoryId} onChange={setSourceRepositoryId} disabled={busy || uncertain} />}
          <FieldSet className="grid min-w-0 gap-4" disabled={busy || detailBusy || referencesBlocked || uncertain || needsRefresh || (mode === "detail" && !canEdit)}>
            <FormField>標題<Input required maxLength={120} value={title} onChange={(e) => setTitle(e.target.value)} /></FormField>
            <FormField>內容<Textarea className="min-h-40" maxLength={8000} value={body} onChange={(e) => setBody(e.target.value)} /></FormField>
            {selected && <SettingsEditor editable={canEdit} value={settingsDraft ?? selected.settings ?? {fields:[], views:[]}} onChange={setSettingsDraft} />}
          </FieldSet>

          {(mode === "create" || canEdit) && <div className="flex flex-wrap gap-2">
            <Button type="submit" disabled={busy || needsRefresh || referencesBlocked}>{mode === "create" ? uncertain ? "重試原建立請求" : "建立專案" : "儲存內容"}</Button>
          </div>}
        </FieldGroup>
      </form>}
    </PagePanel>}
    {selected && <Items key={selected.id+":"+selected.version} ownerId={ownerId} sourceRepositoryId={selected.defaultRepositoryId ?? ""} project={selected} disabled={busy||detailBusy||needsRefresh||!!settingsDraft||title!==selected.title||body!==selected.body} canWrite={() => !navigation.current && !lock.current && !unknown.current && !detailBusy && !needsRefresh && title === selected.title && body === selected.body && canWrite()} onBlocked={referenceBlock} onChange={(item)=>{show(item);load();}} onRefresh={()=>open(selected.id)} />}
  </div>;
}
