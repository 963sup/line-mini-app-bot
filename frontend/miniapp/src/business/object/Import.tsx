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
import { extractPdf } from "./pdf";

type Run = components["schemas"]["ObjectImportRun"];
type Creation = components["schemas"]["ObjectImportRunCreation"];
const reasons: Record<string, string> = { model_unavailable: "解析服務暫時無法使用", output_rejected: "無法安全辨識明細內容", creation_unavailable: "部分物件尚未建立", storage_unavailable: "進度暫時無法保存" };

async function prepareImport(file: File, signal: AbortSignal, extract: typeof extractPdf): Promise<Creation> {
  const format = /\.(xlsx|csv|tsv|txt|md)$/i.exec(file.name)?.[1].toLowerCase();
  if (format) {
    if (!file.size || file.size > 300 * 1024) throw Error("表格檔案上限為 300 KiB。");
    const bytes = new Uint8Array(await file.arrayBuffer());
    if (format === "xlsx" && (bytes[0] !== 0x50 || bytes[1] !== 0x4b || bytes[2] !== 3 || bytes[3] !== 4)) throw Error("檔案內容不是 XLSX 活頁簿。");
    const digest = await crypto.subtle.digest("SHA-256", bytes);
    signal.throwIfAborted();
    const payload = btoa(Array.from(bytes, b => String.fromCharCode(b)).join(""));
    return { classification: "per_item", sourceDocument: format + ":" + Array.from(new Uint8Array(digest), b => b.toString(16).padStart(2, "0")).join(""), sourceRevision: "", ...(format === "xlsx" ? {workbook: payload} : {textFile: payload}) };
  }
  const result = await extract(file, signal);
  return { classification: "per_item", sourceDocument: "pdf:" + result.fileHash, sourceRevision: result.sourceRevision || "", pages: result.pages };
}

export function Import({ repositoryId, onBlocked, canWrite, extract = extractPdf }: { repositoryId: string; onBlocked: (blocked: boolean) => void; canWrite?: () => boolean; extract?: typeof extractPdf }) {
  const { user, request, epoch } = useSession(), rpc = useLatestRequest();
  const [busy, setBusy] = useState(false), [blocked, setBlocked] = useState(false), [message, setMessage] = useState(""), [job, setJob] = useState<Run | null>(null);
  const lock = useRef(false), unresolved = useRef(false), generation = useRef(0), extraction = useRef<AbortController | null>(null);
  const everUnknown = useRef(false), fileInput = useRef<HTMLInputElement>(null);
  const attempt = useRef<{ body: Creation; key: string } | null>(null);
  const resumeSlot = `pdf-import:per-item:${repositoryId}:${user?.userId}`;
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
      block(value.status !== "completed");
      if (value.status === "completed") forget(pending.body.sourceDocument, pending.key);
      const ambiguous = value.needsReviewCount;
      if (ambiguous && !value.reason) { setMessage(`已建立 ${count} 筆物件，另有 ${ambiguous} 筆需要確認類型。`); return; }
      setMessage(value.status === "completed" ? `已建立 ${count} 筆物件。可返回清單查看及修改。` : `${reasons[value.reason] || "匯入尚未完成"}；已建立 ${count}/${value.rows.length} 筆物件。請重試繼續原匯入。`);
    }, error => {
      lock.current = false; setBusy(false);
      if ((pending.body.workbook || pending.body.textFile) && !everUnknown.current && !job && error instanceof Error && "httpStatus" in error && error.httpStatus === 400) {
        forget(pending.body.sourceDocument, pending.key); attempt.current = null; block(false);
        if (fileInput.current) fileInput.current.value = "";
        setMessage("無法讀取明細。請核對表頭、項次與數量；文字檔須為 UTF-8，Excel 明細不可含公式。"); return;
      }
      everUnknown.current = true; retryOnly.current = true;
      setMessage(`${errorMessage(error)} 結果尚未確認，請重試原匯入；將沿用相同內容，避免重複建立。`);
    });
  }
  async function choose(file: File | undefined) {
    if (!file || lock.current || unresolved.current || canWrite?.() === false || !user) return;
    extraction.current?.abort(); const controller = new AbortController(); extraction.current = controller;
    const current = ++generation.current; attempt.current = null; everUnknown.current = false; retryOnly.current = false; setJob(null); lock.current = true; setBusy(true); block(true); setMessage("正在本機擷取支援的明細表……");
    try {
      const body = await prepareImport(file, controller.signal, extract);
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
    <p>文字型 PDF 在本機擷取，明細經伺服器送至 OpenAI 解析。Excel（.xlsx）會將整份檔案送至伺服器讀取；啟用來源留存的服務會保存原檔供補查，不送模型；請使用單一工作表，首列含項次、品名、契約數量或數量、單位。CSV、Tab 分欄 TXT／TSV 與 Markdown 表格須為 UTF-8；整份傳至伺服器依表頭讀取，不送模型。確認類型後建立物件。</p>
    <FormField>選擇明細文件<Input ref={fileInput} type="file" accept="application/pdf,.pdf,.xlsx,.csv,.tsv,.txt,.md" disabled={busy || blocked} onChange={event => void choose(event.target.files?.[0])} /></FormField>
    <Status role="status">{message}</Status>
    {busy && !attempt.current && <Button variant="outline" onClick={() => extraction.current?.abort()}>取消擷取</Button>}
    {attempt.current && blocked && <Button disabled={busy} onClick={send}>重試原匯入</Button>}
    {job?.canReplaceSource && !everUnknown.current && <Button variant="outline" disabled={busy} onClick={() => {
      if (lock.current) return;
      if (attempt.current) forget(attempt.current.body.sourceDocument, attempt.current.key);
      attempt.current = null; setJob(null); block(false); if (fileInput.current) fileInput.current.value = ""; setMessage("尚未建立物件，請重新選擇檔案。");
    }}>更換檔案</Button>}
    {job && <><p>已建立 {job.createdCount}/{job.rows.length} 筆物件。</p>{job.rows.map(row => {
      const location = attempt.current?.body.workbook || attempt.current?.body.textFile ? row.sourcePages : `第 ${row.sourcePages} 頁`;
      return <Item key={row.id} variant="outline" className="min-w-0 flex-wrap" aria-label={`項次 ${row.sourceItemNo} ${row.name}`}>
        <ItemContent>
          <ItemTitle className="break-words">項次 {row.sourceItemNo} · {row.name} · {row.quantity} {row.unit}</ItemTitle>
          <ItemDescription>{location}{row.type === "work" ? " · 施作" : row.type === "non_work" ? " · 非施作" : " · 待確認類型"}</ItemDescription>
          {row.description && <Message><MessageContent><blockquote className="border-l-2 pl-3 [overflow-wrap:anywhere]">「{row.description}」</blockquote></MessageContent></Message>}
        </ItemContent>
        {row.type === "needs_review" && <ItemActions className="flex-wrap">{(["work","non_work"] as const).map(t => <Button key={t} disabled={busy || retryOnly.current} onClick={() => {
          const pending = attempt.current;
          if (!pending || lock.current || retryOnly.current) return;
          pending.body = {...pending.body, decisions: {...pending.body.decisions, [row.id]: t}};
          send();
        }}>{t === "work" ? "確認為施作" : "確認為非施作"}</Button>)}</ItemActions>}
      </Item>;
    })}</>}
  </FieldGroup></PagePanel>;
}
