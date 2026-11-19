import React, { useEffect, useId, useRef, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";
import { PagePanel } from "@/shared/PagePanel";
import { FormField } from "@/shared/FormField";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { RadioGroup, RadioGroupItem } from "@/shared/ui/radio-group";
import { FieldGroup, FieldSet, FieldLegend, Field, FieldLabel } from "@/shared/ui/field";
import { extractPdf } from "./pdf";

type Run = components["schemas"]["ObjectImportRun"];
type Creation = components["schemas"]["ObjectImportRunCreation"];
const reasons: Record<string, string> = { model_unavailable: "解析服務暫時無法使用", output_rejected: "無法安全辨識明細內容", creation_unavailable: "部分物件尚未建立", storage_unavailable: "進度暫時無法保存" };

export function Import({ repositoryId, onBlocked, canWrite, extract = extractPdf }: { repositoryId: string; onBlocked: (blocked: boolean) => void; canWrite?: () => boolean; extract?: typeof extractPdf }) {
  const typeID = useId();
  const { user, request, epoch } = useSession(), rpc = useLatestRequest();
  const [busy, setBusy] = useState(false), [blocked, setBlocked] = useState(false), [message, setMessage] = useState(""), [job, setJob] = useState<Run | null>(null);
  const lock = useRef(false), unresolved = useRef(false), generation = useRef(0), extraction = useRef<AbortController | null>(null);
  const [type, setType] = useState<"work" | "non_work">("work");
  const selectedType = useRef<"work" | "non_work">("work");
  const everUnknown = useRef(false), fileInput = useRef<HTMLInputElement>(null);
  const attempt = useRef<{ body: Creation; key: string } | null>(null);
  function block(value: boolean) { unresolved.current = value; setBlocked(value); onBlocked(value); }
  useEffect(() => () => { generation.current++; extraction.current?.abort(); rpc.cancel(); attempt.current = null; onBlocked(false); }, [repositoryId, user?.userId, epoch, onBlocked, rpc.cancel]);
  useEffect(() => {
    if (!busy && !blocked) return;
    const warn = (event: BeforeUnloadEvent) => { event.preventDefault(); event.returnValue = ""; };
    window.addEventListener("beforeunload", warn); return () => window.removeEventListener("beforeunload", warn);
  }, [busy, blocked]);
  function send() {
    if (lock.current || canWrite?.() === false || !user || !attempt.current) return;
    const pending = attempt.current;
    lock.current = true; setBusy(true); block(true); setMessage("正在解析明細並建立物件……");
    void rpc.run(signal => request((client, s) => client.POST("/api/v1/repositories/{repositoryId}/object-import-runs", { signal: s, params: { path: { repositoryId }, header: { "Idempotency-Key": pending.key } }, body: pending.body }), { signal, timeoutMs: 90000 }), value => {
      lock.current = false; setBusy(false); setJob(value);
      const count = value.rows.filter(row => row.resultObjectId).length;
      block(value.status !== "completed");
      setMessage(value.status === "completed" ? `已建立 ${count} 筆物件。可返回清單查看及修改。` : `${reasons[value.reason] || "匯入尚未完成"}；已建立 ${count}/${value.rows.length} 筆物件。請重試繼續原匯入。`);
    }, error => {
      lock.current = false; setBusy(false); everUnknown.current = true;
      setMessage(`${errorMessage(error)} 結果尚未確認，請重試原匯入；將沿用相同內容，避免重複建立。`);
    });
  }
  async function choose(file: File | undefined) {
    if (!file || lock.current || unresolved.current || canWrite?.() === false || !user) return;
    const batchType = selectedType.current;
    extraction.current?.abort(); const controller = new AbortController(); extraction.current = controller;
    const current = ++generation.current; attempt.current = null; everUnknown.current = false; setJob(null); lock.current = true; setBusy(true); block(true); setMessage("正在本機擷取支援的明細表……");
    try {
      const result = await extract(file, controller.signal);
      if (current !== generation.current) return;
      if (canWrite?.() === false) { block(false); setMessage("頁面正在切換，請重新選擇PDF。"); return; }
      attempt.current = { key: crypto.randomUUID(), body: { type: batchType, sourceDocument: `pdf:${result.fileHash}`, sourceRevision: result.sourceRevision || "", pages: result.pages.map(page => ({ ...page })) } };
      lock.current = false; send();
    } catch (error) { if (current === generation.current) { block(false); setMessage(errorMessage(error)); } }
    finally { if (current === generation.current && !attempt.current) { lock.current = false; setBusy(false); } }
  }
  return <PagePanel heading="PDF匯入物件" className="w-full min-w-0"><FieldGroup>
    <p>選擇PDF即會在本機擷取支援的明細表，將明細文字經伺服器傳送至OpenAI解析並自動建立物件。原始PDF留在本機。僅支援文字型訂單表格，不支援掃描或密碼PDF。</p>
    <FieldSet>
      <FieldLegend>整批類型</FieldLegend>
      <RadioGroup aria-label="整批類型" className="grid-cols-2" value={type} disabled={busy || blocked}
        onValueChange={value => {
          if (lock.current || unresolved.current || (value !== "work" && value !== "non_work")) return;
          selectedType.current = value; setType(value);
        }}>
        <Field orientation="horizontal" className="gap-2"><RadioGroupItem id={typeID + "-work"} value="work" aria-label="施作" /><FieldLabel htmlFor={typeID + "-work"} className="min-h-11 flex-1">施作</FieldLabel></Field>
        <Field orientation="horizontal" className="gap-2"><RadioGroupItem id={typeID + "-non-work"} value="non_work" aria-label="非施作" /><FieldLabel htmlFor={typeID + "-non-work"} className="min-h-11 flex-1">非施作</FieldLabel></Field>
      </RadioGroup>
    </FieldSet>

    <p>整批物件皆使用此類型，匯入後可在個別物件中修改。</p>
    <FormField>選擇PDF並匯入<Input ref={fileInput} type="file" accept="application/pdf,.pdf" disabled={busy || blocked} onChange={event => void choose(event.target.files?.[0])} /></FormField>
    <Status role="status">{message}</Status>
    {busy && !attempt.current && <Button variant="outline" onClick={() => extraction.current?.abort()}>取消擷取</Button>}
    {attempt.current && blocked && <Button disabled={busy} onClick={send}>重試原匯入</Button>}
    {job?.status === "pending" && job.rows.length === 0 && !!job.reason && !everUnknown.current && <Button variant="outline" disabled={busy} onClick={() => {
      if (lock.current) return;
      attempt.current = null; setJob(null); block(false); if (fileInput.current) fileInput.current.value = ""; setMessage("尚未建立物件，請重新選擇PDF。");
    }}>更換PDF</Button>}
    {job && <><p>已建立 {job.rows.filter(row => row.resultObjectId).length}/{job.rows.length} 筆物件。</p>{job.rows.map(row => <p key={row.id} className="min-w-0 whitespace-normal [overflow-wrap:anywhere]">{row.name} · {row.quantity} {row.unit}</p>)}</>}
  </FieldGroup></PagePanel>;
}
