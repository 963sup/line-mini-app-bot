import React, { useEffect, useRef, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";
import { PagePanel } from "@/shared/PagePanel";
import { FormField } from "@/shared/FormField";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { Message, MessageContent } from "@/shared/ui/message";
import { Item, ItemActions, ItemContent, ItemDescription, ItemTitle } from "@/shared/ui/item";
import { Input } from "@/shared/ui/input";
import { FieldGroup } from "@/shared/ui/field";


type Run = components["schemas"]["ObjectImportRun"];
type Creation = components["schemas"]["ObjectImportRunCreation"];
const reasons: Record<string, string> = { model_unavailable: "解析服務暫時無法使用", output_rejected: "無法安全辨識明細內容", creation_unavailable: "部分物件尚未建立", storage_unavailable: "進度暫時無法保存" };

async function prepareImport(file: File, signal: AbortSignal): Promise<Creation> {
  const format = /\.(pdf|xlsx|csv|tsv|txt|md)$/i.exec(file.name)?.[1].toLowerCase();
  if (format) {
    if (format === "pdf" && (!file.size || file.size > 10 * 1024 * 1024)) throw Error("PDF 上限為 10 MiB。");
    if (format !== "pdf" && (!file.size || file.size > 300 * 1024)) throw Error("表格檔案上限為 300 KiB。");
    const bytes = new Uint8Array(await file.arrayBuffer());
    if (format === "xlsx" && (bytes[0] !== 0x50 || bytes[1] !== 0x4b || bytes[2] !== 3 || bytes[3] !== 4)) throw Error("檔案內容不是 XLSX 活頁簿。");
    const digest = await crypto.subtle.digest("SHA-256", bytes);
    signal.throwIfAborted();
    let binary = "";
    for (let offset = 0; offset < bytes.length; offset += 32768) binary += String.fromCharCode(...bytes.subarray(offset, offset + 32768));
    const payload = btoa(binary);
    return { classification: "per_item", sourceDocument: format + ":" + Array.from(new Uint8Array(digest), b => b.toString(16).padStart(2, "0")).join(""), sourceRevision: "", ...(format === "pdf" ? {pdfFile: payload} : format === "xlsx" ? {workbook: payload} : {textFile: payload}) };
  }
  throw Error("不支援的檔案格式。");
}

export function Import({ repositoryId, onBlocked, canWrite, prepare = prepareImport }: { repositoryId: string; onBlocked: (blocked: boolean) => void; canWrite?: () => boolean; prepare?: typeof prepareImport }) {
  const { user, request, epoch } = useSession(), rpc = useLatestRequest();
  const [busy, setBusy] = useState(false), [blocked, setBlocked] = useState(false), [message, setMessage] = useState(""), [job, setJob] = useState<Run | null>(null);
  const lock = useRef(false), unresolved = useRef(false), generation = useRef(0), extraction = useRef<AbortController | null>(null);
  const everUnknown = useRef(false), fileInput = useRef<HTMLInputElement>(null);
  const attempt = useRef<{ body: Creation; key: string } | null>(null);
  const resumeSlot = `pdf-import:per-item:${repositoryId}:${user?.userId}`;
  const savedSlot = resumeSlot + ":saved";
  type Fields = { name: string; quantity: string; unit: string };
  const [edits, setEdits] = useState<Record<string, Fields>>({});
  const [saved, setSaved] = useState<{ key: string; document: string } | null>(null);
  useEffect(() => {
    const value = window.sessionStorage.getItem(savedSlot);
    setSaved(value ? JSON.parse(value) : null);
  }, [savedSlot]);
  function forget(documentId: string, key: string) {
    const slot = resumeSlot + ":" + documentId;
    if (window.sessionStorage.getItem(slot) === key) window.sessionStorage.removeItem(slot);
  }
  function block(value: boolean) { unresolved.current = value; setBlocked(value); onBlocked(value); }
  useEffect(() => () => { generation.current++; extraction.current?.abort(); rpc.cancel(); attempt.current = null; onBlocked(false); }, [repositoryId, user?.userId, epoch, onBlocked, rpc.cancel]);
  useEffect(() => {
    if (!busy && !blocked) return;
    const warn = (event: BeforeUnloadEvent) => { event.preventDefault(); event.returnValue = ""; };
    window.addEventListener("beforeunload", warn); return () => window.removeEventListener("beforeunload", warn);
  }, [busy, blocked]);
  const retryOnly = useRef(false);
  function send() {
    if (lock.current || canWrite?.() === false || !user || !attempt.current) return;
    const pending = attempt.current;
    lock.current = true; setBusy(true); block(true); setMessage("正在解析明細並建立物件……");
    void rpc.run(signal => request((client, s) => client.POST("/api/v1/repositories/{repositoryId}/object-import-runs", { signal: s, params: { path: { repositoryId }, header: { "Idempotency-Key": pending.key } }, body: pending.body }), { signal, timeoutMs: 90000 }), value => {
      retryOnly.current = false; lock.current = false; setBusy(false); setJob(value);
      const count = value.createdCount;
      const fileBacked = !!(pending.body.pdfFile || pending.body.workbook || pending.body.textFile || pending.body.resume);
      if (fileBacked && value.status === "processing") {
        const next = { key: pending.key, document: pending.body.sourceDocument };
        window.sessionStorage.setItem(savedSlot, JSON.stringify(next));
        setSaved(next);
        attempt.current = { key: pending.key, body: { sourceDocument: next.document, sourceRevision: "", classification: "per_item", resume: true } };
      }
      if (value.status === "completed") {
        if (saved?.key === pending.key) {
          window.sessionStorage.removeItem(savedSlot);
          setSaved(null);
        }
      }
      block(value.status !== "completed");
      if (fileBacked && value.status === "processing") onBlocked(false);
      if (value.status === "completed") forget(pending.body.sourceDocument, pending.key);
      const ambiguous = value.needsReviewCount;
      if (ambiguous && !value.reason) { setMessage(`已建立 ${count} 筆物件，另有 ${ambiguous} 筆需要確認類型。`); return; }
      setMessage(value.status === "completed" ? `已建立 ${count} 筆物件。可返回清單查看及修改。` : `${reasons[value.reason] || "匯入尚未完成"}；已建立 ${count}/${value.rows.length} 筆物件。請重試繼續原匯入。`);
    }, error => {
      lock.current = false; setBusy(false);
      if ((pending.body.pdfFile || pending.body.workbook || pending.body.textFile) && !everUnknown.current && !job && error instanceof Error && "httpStatus" in error && error.httpStatus === 400) {
        attempt.current = null; block(false);
        if (fileInput.current) fileInput.current.value = "";
        setMessage(pending.body.pdfFile ? "伺服器無法辨識 PDF 明細。請確認為文字型訂單，且可辨識項次、品名、數量和單位欄及明細範圍。" : "無法讀取明細。請核對表頭、項次與數量；文字檔須為 UTF-8，Excel 明細不可含公式。"); return;
      }
      if (pending.body.resume && !job && error instanceof Error && "httpStatus" in error && error.httpStatus === 404) {
        window.sessionStorage.removeItem(savedSlot);
        setSaved(null); attempt.current = null; block(false);
        setMessage("找不到上次的匯入作業。請重新選擇原檔。");
        return;
      }
      if (pending.body.resume && error instanceof Error && "httpStatus" in error && error.httpStatus === 409) {
        const { sourceDocument } = pending.body;
        pending.body = { sourceDocument, sourceRevision: "", classification: "per_item", resume: true };
        retryOnly.current = true;
        onBlocked(false);
        setMessage("匯入進度已變更。請重試以讀取最新結果。");
        return;
      }
      if (pending.body.resume && job && error instanceof Error && "httpStatus" in error && error.httpStatus === 400) {
        pending.body = { sourceDocument: pending.body.sourceDocument, sourceRevision: "", classification: "per_item", resume: true };
        onBlocked(false);
        setMessage("請核對品名、數量與單位後再確認。");
        return;
      }
      everUnknown.current = true; retryOnly.current = true;
      if (pending.body.resume) onBlocked(false);
      setMessage(`${errorMessage(error)} 結果尚未確認，請重試原匯入；將沿用相同內容，避免重複建立。`);
    });
  }
  async function choose(file: File | undefined) {
    if (!file || saved || lock.current || unresolved.current || canWrite?.() === false || !user) return;
    extraction.current?.abort(); const controller = new AbortController(); extraction.current = controller;
    const current = ++generation.current; attempt.current = null; everUnknown.current = false; retryOnly.current = false; setJob(null); lock.current = true; setBusy(true); block(true); setMessage("正在準備上傳文件……");
    try {
      const body = await prepare(file, controller.signal);
      if (current !== generation.current) return;
      if (canWrite?.() === false) { block(false); setMessage("頁面正在切換，請重新選擇檔案。"); return; }
      const documentId = body.sourceDocument;
      const slot = resumeSlot + ":" + documentId;
      const key = window.sessionStorage.getItem(slot) || crypto.randomUUID();
      window.sessionStorage.setItem(slot, key);
      attempt.current = { key, body };
      lock.current = false; send();
    } catch (error) { if (current === generation.current) { block(false); setMessage(errorMessage(error)); } }
    finally { if (current === generation.current && !attempt.current) { lock.current = false; setBusy(false); } }
  }
  return <PagePanel heading="文件匯入物件" className="w-full min-w-0"><FieldGroup>
    <p>文字型 PDF 會上傳至伺服器並使用 OpenAI 解析。若無法辨識表格版型，可能傳送整份擷取文字，包括價格與條款，結果須逐列確認。Excel（.xlsx）會將整份檔案送至伺服器讀取；啟用來源留存的服務會保存原檔供補查，不送模型；請使用單一工作表，首列含項次、品名、契約數量或數量、單位。CSV、Tab 分欄 TXT／TSV 與 Markdown 表格須為 UTF-8；整份傳至伺服器依表頭讀取，不送模型。確認類型後建立物件。</p>
    {saved && !job && !busy && !attempt.current && <Button onClick={() => {
      attempt.current = { key: saved.key, body: { sourceDocument: saved.document, sourceRevision: "", classification: "per_item", resume: true } };
      send();
    }}>繼續上次匯入</Button>}
    <FormField>選擇明細文件<Input ref={fileInput} type="file" accept="application/pdf,.pdf,.xlsx,.csv,.tsv,.txt,.md" disabled={busy || blocked || !!saved} onChange={event => void choose(event.target.files?.[0])} /></FormField>
    <Status role="status">{message}</Status>
    {busy && !attempt.current && <Button variant="outline" onClick={() => extraction.current?.abort()}>取消上傳</Button>}
    {attempt.current && blocked && <Button disabled={busy} onClick={send}>重試原匯入</Button>}
    {job?.canReplaceSource && !everUnknown.current && <Button variant="outline" disabled={busy} onClick={() => {
      if (lock.current) return;
      if (attempt.current) forget(attempt.current.body.sourceDocument, attempt.current.key);
      attempt.current = null; setJob(null); block(false); if (fileInput.current) fileInput.current.value = ""; setMessage("尚未建立物件，請重新選擇檔案。");
    }}>更換檔案</Button>}
    {job && <><p>已建立 {job.createdCount}/{job.rows.length} 筆物件。</p>{job.rows.map(row => {
      const location = /^(xlsx|csv|tsv|txt|md):/.test(attempt.current?.body.sourceDocument ?? "") ? row.sourcePages : `第 ${row.sourcePages} 頁`;
      const fields = edits[row.id] ?? { name: row.name, quantity: row.quantity, unit: row.unit };
      return <Item key={row.id} variant="outline" className="min-w-0 flex-wrap" aria-label={`項次 ${row.sourceItemNo} ${row.name}`}>
        <ItemContent>
          <ItemTitle className="break-words">項次 {row.sourceItemNo} · {row.name} · {row.quantity} {row.unit}</ItemTitle>
          <ItemDescription>{location}{row.type === "work" ? " · 施作" : row.type === "non_work" ? " · 非施作" : " · 待確認類型"}</ItemDescription>
          {row.description && <Message><MessageContent><blockquote className="border-l-2 pl-3 [overflow-wrap:anywhere]">「{row.description}」</blockquote></MessageContent></Message>}
        </ItemContent>
        {row.type === "needs_review" && attempt.current?.body.resume && <FieldGroup>
          {([["name", "品名"], ["quantity", "數量"], ["unit", "單位"]] as const).map(([field, label]) =>
            <FormField key={field}>{label}<Input aria-label={label} value={fields[field]} disabled={busy || retryOnly.current}
              onChange={event => setEdits(previous => ({ ...previous, [row.id]: { ...fields, [field]: event.target.value } }))} /></FormField>
          )}
        </FieldGroup>}
        {row.type === "needs_review" && <ItemActions className="flex-wrap">{(["work","non_work"] as const).map(t => <Button key={t} disabled={busy || retryOnly.current} onClick={() => {
          const pending = attempt.current;
          if (!pending || lock.current || retryOnly.current) return;
          const corrections = pending.body.resume && edits[row.id] ? { [row.id]: fields } : undefined;
          pending.body = {...pending.body, decisions: {...pending.body.decisions, [row.id]: t}, ...(corrections && { corrections })};
          send();
        }}>{t === "work" ? "確認為施作" : "確認為非施作"}</Button>)}</ItemActions>}
      </Item>;
    })}</>}
  </FieldGroup></PagePanel>;
}
