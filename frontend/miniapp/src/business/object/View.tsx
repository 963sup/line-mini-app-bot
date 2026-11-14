import React, { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import type { extractPdf } from "./pdf";
import { Import } from "./Import";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";
import { PagePanel } from "@/shared/PagePanel";
import { FormField } from "@/shared/FormField";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Textarea } from "@/shared/ui/textarea";
import { FieldGroup } from "@/shared/ui/field";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";
import { Badge } from "@/shared/ui/badge";
import { Item, ItemGroup, ItemContent, ItemTitle } from "@/shared/ui/item";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from "@/shared/ui/empty";

type Resource = components["schemas"]["Object"];
type Creation = components["schemas"]["ObjectCreation"];
type Props = { repositoryId: string; onBlocked?: (blocked: boolean) => void; canWrite?: () => boolean; extract?: typeof extractPdf };
const states = { not_started: "未開始", in_progress: "進行中", completed: "已完成" };
const empty: Creation = { name: "", description: "", type: "未分類", quantity: "1", unit: "SET", sourceDocument: "", sourceRevision: "", sourceItemNo: "", sourcePages: "", materialCode: "" };
const fields = [
  ["name", "工程項目名稱", 120], ["type", "類型", 80], ["quantity", "數量", 25], ["unit", "單位", 32],
  ["sourceDocument", "來源文件識別", 200], ["sourceRevision", "來源版本", 80], ["sourceItemNo", "來源項次", 80], ["sourcePages", "來源頁碼", 80], ["materialCode", "料號", 120],
] as const;

export function View(props: Props) {
  const { epoch, user } = useSession();
  return <Content key={`${props.repositoryId}:${epoch}:${user?.userId || ""}`} {...props} />;
}

function Content({ repositoryId, onBlocked, canWrite, extract }: Props) {
  const { user, request } = useSession();
  const read = useLatestRequest(), detail = useLatestRequest(), write = useLatestRequest();
  const [items, setItems] = useState<Resource[]>([]), [cursor, setCursor] = useState("");
  const [listBusy, setListBusy] = useState(false), [loaded, setLoaded] = useState(false), [listStatus, setListStatus] = useState("");
  const [mode, setMode] = useState<"create" | "detail" | null>(null), [selected, setSelected] = useState<Resource | null>(null);
  const [content, setContent] = useState<Creation>(empty), [state, setState] = useState<Resource["status"]>("not_started");
  const [busy, setBusy] = useState(false), [detailBusy, setDetailBusy] = useState(false), [uncertain, setUncertain] = useState(false), [needsRefresh, setNeedsRefresh] = useState(false), [status, setStatus] = useState("");
  const [importing, setImporting] = useState(false), [importBlocked, setImportBlocked] = useState(false);
  const importLock = useRef(false);
  const importBlock = useCallback((value: boolean) => { importLock.current = value; setImportBlocked(value); onBlocked?.(value); }, [onBlocked]);
  const lock = useRef(false), formNavigation = useRef(false), pending = useRef<{ key: string; body: Creation } | null>(null);
  useEffect(() => { formNavigation.current = false; });
  const canImportWrite = useCallback(() => !formNavigation.current && canWrite?.() !== false, [canWrite]);
  const mine = !!selected && selected.creatorId === user?.userId;
  const block = (value: boolean) => { onBlocked?.(value); };
  useEffect(() => () => { onBlocked?.(false); }, [onBlocked]);
  useEffect(() => {
    if (!busy && !uncertain) return;
    const warn = (event: BeforeUnloadEvent) => { event.preventDefault(); event.returnValue = ""; };
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [busy, uncertain]);

  const load = useCallback((next = "") => {
    if (!user) return;
    setListBusy(true); setLoaded(false); setCursor(""); if (!next) setItems([]);
    setListStatus("正在讀取工程項目……");
    void read.run(signal => request((client, s) => client.GET("/api/v1/repositories/{repositoryId}/objects", { signal: s, params: { path: { repositoryId }, query: { cursor: next || undefined } } }), { signal }), page => {
      setItems(previous => [...new Map([...(next ? previous : []), ...page.items].map(item => [item.id, item])).values()]);
      setCursor(page.nextCursor); setListBusy(false); setLoaded(true); setListStatus("已讀取工程項目。");
    }, error => { setItems([]); setListBusy(false); setListStatus(errorMessage(error)); });
  }, [user, request, repositoryId, read.run]);
  useEffect(() => { load(); return read.cancel; }, [load, read.cancel]);
  function show(item: Resource) {
    setSelected(item); setContent({ name: item.name, description: item.description, type: item.type, quantity: item.quantity, unit: item.unit,
      sourceDocument: item.sourceDocument || "", sourceRevision: item.sourceRevision || "", sourceItemNo: item.sourceItemNo || "", sourcePages: item.sourcePages || "", materialCode: item.materialCode || "" });
    setState(item.status); setMode("detail"); setNeedsRefresh(false); setUncertain(false); pending.current = null; block(false);
  }
  function open(id: string) {
    if (!user || lock.current || uncertain || formNavigation.current) return;
    formNavigation.current = true;
    setMode("detail"); setSelected(null); setDetailBusy(true); setStatus("正在讀取工程項目……");
    void detail.run(signal => request((client, s) => client.GET("/api/v1/repositories/{repositoryId}/objects/{id}", { signal: s, params: { path: { repositoryId, id } } }), { signal }), item => {
      show(item); setDetailBusy(false); setStatus("已讀取工程項目。");
    }, error => { setDetailBusy(false); setStatus(errorMessage(error)); });
  }
  function create() {
    if (lock.current || uncertain || formNavigation.current) return;
    formNavigation.current = true;
    detail.cancel(); setDetailBusy(false); setMode("create"); setSelected(null); setContent({ ...empty }); setState("not_started"); setStatus(""); setNeedsRefresh(false); pending.current = null;
  }
  function submit(event: FormEvent) {
    event.preventDefault();
    if (!user || lock.current || formNavigation.current || needsRefresh || canWrite?.() === false || (mode === "detail" && !mine)) return;
    const body = Object.fromEntries(Object.entries(content).map(([key, value]) => [key, value?.trim() || ""])) as Creation;
    if (!body.name || !body.type || !body.unit || !/^(?:0|[1-9][0-9]{0,17})(?:\.[0-9]{1,6})?$/.test(body.quantity) || !/[1-9]/.test(body.quantity)) {
      setStatus("請填寫名稱、類型、單位及大於零的數量（最多18位整數及6位小數）。"); return;
    }
    if (mode === "detail" && selected) {
      lock.current = true; setBusy(true); block(true); setStatus("正在更新工程項目……");
      void write.run(signal => request((client, s) => client.PATCH("/api/v1/repositories/{repositoryId}/objects/{id}", { signal: s, params: { path: { repositoryId, id: selected.id } }, body: { ...body, status: state, expectedVersion: selected.version } }), { signal }), item => {
        lock.current = false; setBusy(false); show(item); setStatus("已更新工程項目。"); load();
      }, error => { lock.current = false; setBusy(false); block(false); setNeedsRefresh(true); setStatus(errorMessage(error) + " 請重新讀取工程項目後再修改。"); });
      return;
    }
    if (!uncertain || !pending.current) pending.current = { key: crypto.randomUUID(), body };
    const attempt = pending.current;
    lock.current = true; setBusy(true); setUncertain(true); block(true); setStatus("正在建立工程項目……");
    void write.run(signal => request((client, s) => client.POST("/api/v1/repositories/{repositoryId}/objects", { signal: s, params: { path: { repositoryId }, header: { "Idempotency-Key": attempt.key } }, body: attempt.body }), { signal }), item => {
      lock.current = false; setBusy(false); show(item); setStatus("已建立工程項目。"); load();
    }, error => {
      lock.current = false; setBusy(false);
      if (["invalid_input", "forbidden", "not_found", "idempotency_conflict"].includes((error as { code?: string }).code || "")) { setUncertain(false); block(false); }
      setStatus(errorMessage(error) + " 結果未確認時請重試原建立請求，避免重複建立；離開頁面前請先確認紀錄。");
    });
  }
  if (importing) return <>
    <Button variant="outline" disabled={importBlocked} onClick={() => { if (!importLock.current) { formNavigation.current = true; setImporting(false); load(); } }}>返回工程項目清單</Button>
    <Import repositoryId={repositoryId} onBlocked={importBlock} canWrite={canImportWrite} extract={extract} />
  </>;
  return <>
    <PagePanel heading="Objects／工程項目">
      <p className="mb-4 text-sm text-muted-foreground">每個清單項次是一筆工程項目，數量不拆成單件。成員可新增，僅建立者可編輯；完成不代表驗收或付款。</p>
      <div className="mb-4 flex flex-wrap gap-2">
        <Button disabled={!user || busy || uncertain} onClick={create}>新增工程項目</Button>
        <Button variant="outline" disabled={!user || busy || uncertain} onClick={() => { if (!lock.current && !uncertain) { formNavigation.current = true; detail.cancel(); setImporting(true); } }}>PDF文字匯入</Button>
        <Button variant="outline" disabled={!user || busy || listBusy} onClick={() => load()}>重新載入工程項目</Button>
      </div>
      <Status role="status">{listStatus}</Status>
      <ItemGroup className="gap-3">{items.map(item => <Item key={item.id} variant="outline"><ItemContent>
        <ItemTitle><Button variant="link" className="h-auto whitespace-normal text-left" disabled={busy || detailBusy || uncertain} onClick={() => open(item.id)}>{item.name}</Button></ItemTitle>
        <p>{item.type} · {item.quantity} {item.unit}{item.sourceItemNo ? ` · 來源項次 ${item.sourceItemNo}` : ""}</p>
        <Badge variant="outline">{states[item.status]}</Badge>
      </ItemContent></Item>)}</ItemGroup>
      {loaded && !items.length && <Empty><EmptyHeader><EmptyTitle>尚無工程項目</EmptyTitle><EmptyDescription>可手動建立清單項目；PDF文字匯入需先人工核對；不提供CSV匯入。</EmptyDescription></EmptyHeader></Empty>}
      {cursor && <Button variant="outline" disabled={busy || listBusy} onClick={() => load(cursor)}>載入更多工程項目</Button>}
    </PagePanel>
    {mode && <PagePanel heading={mode === "create" ? "新增工程項目" : selected?.name || "工程項目詳情"}>
      <Status role="status">{status}</Status>
      {selected && <Button variant="outline" disabled={busy || detailBusy} onClick={() => open(selected.id)}>重新讀取工程項目</Button>}
      {(mode === "create" || selected) && <form onSubmit={submit}><FieldGroup>
        <fieldset className="grid min-w-0 gap-4" disabled={busy || detailBusy || uncertain || needsRefresh || (mode === "detail" && !mine)}>
          {fields.map(([key, label, max]) => <FormField key={key}>{label}<Input required={["name", "type", "quantity", "unit"].includes(key)} maxLength={max} inputMode={key === "quantity" ? "decimal" : undefined} value={content[key] || ""} onChange={event => setContent(previous => ({ ...previous, [key]: event.target.value }))} /></FormField>)}
          <FormField>完整說明<Textarea maxLength={8000} value={content.description} onChange={event => setContent(previous => ({ ...previous, description: event.target.value }))} /></FormField>
          {selected && <FormField>追蹤狀態<NativeSelect value={state} onChange={event => setState(event.target.value as Resource["status"])}>{Object.entries(states).map(([value, label]) => <NativeSelectOption key={value} value={value}>{label}</NativeSelectOption>)}</NativeSelect></FormField>}
        </fieldset>
        <p className="text-sm text-muted-foreground">來源欄位可留白，僅供追溯；不同文件版本不會自動覆寫內容或狀態。</p>
        {(mode === "create" || mine) && <Button type="submit" disabled={busy || detailBusy || needsRefresh}>{mode === "create" ? uncertain ? "重試原工程項目建立請求" : "建立工程項目" : "儲存工程項目"}</Button>}
      </FieldGroup></form>}
    </PagePanel>}
  </>;
}
