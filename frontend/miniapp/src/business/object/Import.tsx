import React, { useEffect, useRef, useState } from "react";
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
import { Checkbox } from "@/shared/ui/checkbox";
import { extractPdf, PDF_LIMITS, type PdfPage } from "./pdf";

type Job = components["schemas"]["ObjectImport"];
type Creation = components["schemas"]["ObjectImportCreation"];
type Reviewed = components["schemas"]["ObjectImportReviewedRow"];
type Metadata = { id: string; fileHash: string };
const rowFields = [["name", "名稱", 120], ["description", "說明", 8000], ["type", "類型", 80], ["quantity", "數量", 25], ["unit", "單位", 32], ["materialCode", "料號", 120]] as const;
const reviewed = (job: Job): Reviewed[] => job.rows.map(({ id, name, description, type, quantity, unit, materialCode }) => ({ id, name, description, type, quantity, unit, materialCode }));

export function Import({ repositoryId, onBlocked, canWrite, extract = extractPdf }: { repositoryId: string; onBlocked: (blocked: boolean) => void; canWrite?: () => boolean; extract?: typeof extractPdf }) {
  const { user, request } = useSession(), rpc = useLatestRequest();
  const [pages, setPages] = useState<PdfPage[]>([]), [originalPages, setOriginalPages] = useState<PdfPage[]>([]), [chosen, setChosen] = useState<number[]>([]), [fileHash, setFileHash] = useState("");
  const [sourceDocument, setSourceDocument] = useState(""), [sourceRevision, setSourceRevision] = useState("");
  const [ack, setAck] = useState(false), [consent, setConsent] = useState(false), [busy, setBusy] = useState(false), [uncertain, setUncertain] = useState(false), [message, setMessage] = useState("");
  const [frozen, setFrozen] = useState<Creation | null>(null), [job, setJob] = useState<Job | null>(null), [rows, setRows] = useState<Reviewed[]>([]), [resume, setResume] = useState<Metadata | null>(null), [reviewDirty, setReviewDirty] = useState(false), [needsRead, setNeedsRead] = useState(false), [metadataSaved, setMetadataSaved] = useState<boolean | null>(null);
  const reviewChanged = useRef(false), uncertainRef = useRef(false), everUnknown = useRef(false);
  const markUncertain = (value: boolean) => { uncertainRef.current = value; setUncertain(value); };
  const lock = useRef(false), extraction = useRef<AbortController | null>(null), generation = useRef(0);
  const attempt = useRef<{ body: Creation; key: string } | null>(null), confirmation = useRef<number | null>(null), expectedSourceHash = useRef("");
  const storageKey = `object-import:${repositoryId}:${user?.userId || ""}`;
  const fileMatches = !!job && !!fileHash && resume?.id === job.id && resume.fileHash === fileHash;
  const changeBlock = (value: boolean) => { lock.current = value; setBusy(value); onBlocked(value || uncertainRef.current); };
  useEffect(() => {
    try { const value = JSON.parse(localStorage.getItem(storageKey) || "null") as Metadata | null; if (value && /^[a-f0-9]{32}$/.test(value.id) && /^[a-f0-9]{64}$/.test(value.fileHash)) setResume({ id: value.id, fileHash: value.fileHash }); } catch { setMetadataSaved(false); }
    return () => { generation.current++; extraction.current?.abort(); onBlocked(false); };
  }, [storageKey, onBlocked]);
  useEffect(() => {
    if (!busy && !uncertain) return;
    const warn = (event: BeforeUnloadEvent) => { event.preventDefault(); event.returnValue = ""; };
    window.addEventListener("beforeunload", warn); return () => window.removeEventListener("beforeunload", warn);
  }, [busy, uncertain]);
  function accept(value: Job) {
    setJob(value); setRows(reviewed(value)); setNeedsRead(false); reviewChanged.current = false; setReviewDirty(false); everUnknown.current = false; markUncertain(false); onBlocked(false);
    const metadata = { id: value.id, fileHash: resume?.id === value.id ? resume.fileHash : fileHash };
    setResume(metadata); try { localStorage.setItem(storageKey, JSON.stringify(metadata)); setMetadataSaved(true); } catch { setMetadataSaved(false); }
    if (value.confirmedFromReviewVersion) confirmation.current = value.confirmedFromReviewVersion;
  }
  async function choose(file: File | undefined) {
    if (!file || lock.current || uncertainRef.current) return;
    extraction.current?.abort(); const controller = new AbortController(); extraction.current = controller;
    const current = ++generation.current; setPages([]); setOriginalPages([]); setChosen([]); setFileHash(""); setAck(false); setConsent(false); setMessage("只在此瀏覽器擷取PDF文字……"); changeBlock(true);
    try {
      const result = await extract(file, controller.signal);
      if (current !== generation.current) return;
      setPages(result.pages.map(page => ({ ...page }))); setOriginalPages(result.pages); setFileHash(result.fileHash); setMessage("文字僅在本機。尚未選取任何頁面，請先移除個資及價格，再決定要傳送的文字。");
    } catch (error) { if (current === generation.current) setMessage(errorMessage(error)); }
    finally { if (current === generation.current) changeBlock(false); }
  }
  function parse() {
    if (lock.current || canWrite?.() === false || !user || !ack || !consent || frozen || job) return;
    const selected = pages.filter(page => chosen.includes(page.number)).map(page => ({ ...page }));
    if (!sourceDocument.trim() || !selected.length || selected.some(page => !page.text.trim()) || new TextEncoder().encode(selected.map(page => page.text).join("")).length > PDF_LIMITS.textBytes) { setMessage("請填來源識別、選取有文字的頁面，總文字不得超過128 KiB。"); return; }
    const body = { sourceDocument: sourceDocument.trim(), sourceRevision: sourceRevision.trim(), pages: selected, consent: true as const };
    changeBlock(true); setMessage("只將已選取並編輯的文字送至伺服器OpenAI解析；不傳送PDF。");
    void rpc.run(signal => request((client, s) => client.POST("/api/v1/repositories/{repositoryId}/object-imports/parse", { signal: s, params: { path: { repositoryId } }, body }), { signal, timeoutMs: 90000 }), result => {
      const creation = { sourceDocument: body.sourceDocument, sourceRevision: body.sourceRevision, pages: body.pages, rows: result.rows };
      attempt.current = { key: crypto.randomUUID(), body: creation }; expectedSourceHash.current = result.sourceHash; setFrozen(creation); changeBlock(false); setMessage("解析候選已凍結；尚未建立任何工程項目。請建立草稿後逐列核對。");
    }, error => { changeBlock(false); setMessage(errorMessage(error)); });
  }
  function createDraft() {
    if (lock.current || canWrite?.() === false || !attempt.current) return;
    const pending = attempt.current; changeBlock(true); markUncertain(true); onBlocked(true); setMessage("正在保存匯入草稿……");
    void rpc.run(signal => request((client, s) => client.POST("/api/v1/repositories/{repositoryId}/object-imports", { signal: s, params: { path: { repositoryId }, header: { "Idempotency-Key": pending.key } }, body: pending.body }), { signal }), value => {
      changeBlock(false);
      if (value.sourceHash !== expectedSourceHash.current) { setMessage("草稿來源不一致，請重新讀取核對，禁止確認。"); everUnknown.current = true; markUncertain(true); onBlocked(true); return; }
      accept(value); setMessage("草稿已保存，尚未建立工程項目；請核對原文與每列欄位後儲存審查。");
    }, error => {
      lock.current = false; setBusy(false);
      const rejected = ["invalid_input", "forbidden", "not_found", "idempotency_conflict"].includes((error as { code?: string }).code || "");
      if (rejected && !everUnknown.current) { markUncertain(false); onBlocked(false); setMessage(errorMessage(error) + " 草稿未建立，可放棄候選並修正文字後重新解析。"); }
      else { everUnknown.current = true; markUncertain(true); onBlocked(true); setMessage(errorMessage(error) + " 請重試原草稿建立請求；不要重新解析或更換內容。"); }
    });
  }
  function reload(id = job?.id || resume?.id) {
    if (!id || lock.current || uncertainRef.current) return;
    changeBlock(true);
    void rpc.run(signal => request((client, s) => client.GET("/api/v1/repositories/{repositoryId}/object-imports/{id}", { signal: s, params: { path: { repositoryId, id } } }), { signal }), value => { changeBlock(false); accept(value); setMessage("已讀取匯入進度；請重新選取同一份本機PDF比較原文。"); }, error => { changeBlock(false); setMessage(errorMessage(error)); });
  }
  function review() {
    if (!job || !fileMatches || lock.current || needsRead || canWrite?.() === false || job.status === "confirming" || job.status === "completed") return;
    changeBlock(true);
    void rpc.run(signal => request((client, s) => client.POST("/api/v1/repositories/{repositoryId}/object-imports/{id}/review", { signal: s, params: { path: { repositoryId, id: job.id } }, body: { expectedVersion: job.version, rows } }), { signal }), value => { changeBlock(false); accept(value); confirmation.current = null; setMessage("審查已保存。再次核對後明確確認，才會建立工程項目。"); }, error => { changeBlock(false); setMessage(errorMessage(error) + " 請重新讀取草稿，保留目前來源核對。"); });
  }
  function confirm() {
    if (!job || !fileMatches || lock.current || needsRead || reviewChanged.current || canWrite?.() === false || !["reviewed", "confirming"].includes(job.status)) return;
    const version = confirmation.current ?? job.version; confirmation.current = version; setNeedsRead(true); changeBlock(true);
    void rpc.run(signal => request((client, s) => client.POST("/api/v1/repositories/{repositoryId}/object-imports/{id}/confirm", { signal: s, params: { path: { repositoryId, id: job.id } }, body: { expectedVersion: version } }), { signal, timeoutMs: 90000 }), value => { changeBlock(false); accept(value); setMessage(value.status === "completed" ? "匯入完成，可返回工程項目清單。" : "已保存部分進度，請重新讀取並沿用原審查版本繼續。"); }, error => { changeBlock(false); setMessage(errorMessage(error) + " 可能已部分建立；請重新讀取進度後，沿用原審查版本重試確認。"); });
  }
  function editRow(index: number, field: (typeof rowFields)[number][0], value: string) {
    reviewChanged.current = true; setReviewDirty(true);
    setRows(previous => previous.map((row, i) => i === index ? { ...row, [field]: value } : row));
  }
  return <PagePanel heading="PDF文字匯入工程項目">
    <p>PDF只在瀏覽器擷取文字；不支援掃描或密碼PDF。解析文字會經伺服器傳給OpenAI，請先遮蔽個資、價格及不需分享的條款。AI結果必須人工核對。</p>
    <Status role="status">{message}</Status>
    <FieldGroup>
      <FormField>本機PDF<Input type="file" accept="application/pdf,.pdf" disabled={busy || uncertain || (!!frozen && !job)} onChange={event => void choose(event.target.files?.[0])} /></FormField>
      {busy && <Button variant="outline" onClick={() => { extraction.current?.abort(); rpc.cancel(); generation.current++; lock.current = false; setBusy(false); if (uncertainRef.current) everUnknown.current = true; onBlocked(uncertainRef.current); setMessage("操作已取消；若草稿建立結果不明，請沿用原請求重試。"); }}>取消目前操作</Button>}
      {resume && !job && <Button variant="outline" disabled={busy || uncertain} onClick={() => reload()}>繼續已保存的匯入</Button>}
      <fieldset className="grid min-w-0 gap-4" disabled={busy || !!frozen || !!job}>
        <FormField>來源文件／訂單號<Input maxLength={200} value={sourceDocument} onChange={event => setSourceDocument(event.target.value)} /></FormField>
        <FormField>文件版本<Input maxLength={80} value={sourceRevision} onChange={event => setSourceRevision(event.target.value)} /></FormField>
      </fieldset>
      {job && <p>匯入識別碼：{job.id}。{metadataSaved === false && "此瀏覽器無法保存續傳識別，重新整理後不保證可找回；請保留識別碼並在本頁完成。"}</p>}
      {job && <p className="break-words">來源：{job.sourceDocument}／{job.sourceRevision}。狀態：{{ draft: "草稿", reviewed: "已審查", confirming: "確認建立中", completed: "已匯入" }[job.status]}；已建立 {job.rows.filter(row => row.resultObjectId).length}/{job.rows.length}。本機檔案比對{fileMatches ? "相符" : "尚未相符，請重選原PDF"}；此比對不是PDF內容真實性認證。</p>}
      {pages.map(page => <FieldGroup key={page.number}>
        {!frozen && !job && <FormField><Checkbox aria-label={`選取第${page.number}頁`} checked={chosen.includes(page.number)} disabled={busy} onCheckedChange={checked => setChosen(previous => checked ? [...previous, page.number] : previous.filter(value => value !== page.number))} />選取第{page.number}頁傳送文字（預設不選）</FormField>}
        <FormField>第{page.number}頁原始本機文字<Textarea readOnly value={originalPages.find(value => value.number === page.number)?.text || ""} /></FormField>
        <FormField>第{page.number}頁待傳送文字<Textarea readOnly={busy || !!frozen || !!job} value={page.text} onChange={event => { const text = event.target.value; setPages(previous => previous.map(value => value.number === page.number ? { ...value, text } : value)); }} /></FormField>
      </FieldGroup>)}
      {!job && !frozen && <>
        <FormField><Checkbox checked={ack} onCheckedChange={value => setAck(value === true)} disabled={busy} />我已核對所選文字，確認這些內容可送出</FormField>
        <FormField><Checkbox checked={consent} onCheckedChange={value => setConsent(value === true)} disabled={busy} />我同意將所選文字經伺服器傳送至OpenAI解析</FormField>
        <Button disabled={busy || !ack || !consent || !chosen.length} onClick={parse}>解析所選文字</Button>
      </>}
      {frozen && !job && <><Button variant="outline" disabled={busy || uncertain} onClick={() => { if (lock.current || uncertainRef.current) return; everUnknown.current = false; attempt.current = null; expectedSourceHash.current = ""; setFrozen(null); setAck(false); setConsent(false); setMessage("已放棄尚未保存的候選；請重新核對所選文字並同意解析。"); }}>放棄候選重新解析</Button><p>已解析 {frozen.rows.length} 列；來源文字與候選已凍結。保存草稿不會建立工程項目。</p><Button disabled={busy} onClick={createDraft}>{uncertain ? "重試原草稿建立請求" : "保存匯入草稿"}</Button></>}
      {job && <>
        <Button variant="outline" disabled={busy} onClick={() => reload()}>重新讀取匯入進度</Button>
        {job.rows.map((row, index) => <FieldGroup key={row.id}>
          <p>來源項次 {row.sourceItemNo} · {row.resultObjectId ? "已建立" : "未建立"}</p>
          {row.evidence.map((evidence, i) => <blockquote key={i} className="break-words">第{evidence.page}頁：「{evidence.quote}」</blockquote>)}
          <fieldset className="grid min-w-0 gap-4" disabled={busy || needsRead || !fileMatches || job.status === "confirming" || job.status === "completed"}>
            {rowFields.map(([field, label, max]) => <FormField key={field}>{label}{field === "description" ? <Textarea maxLength={max} value={rows[index]?.[field] || ""} onChange={event => editRow(index, field, event.target.value)} /> : <Input inputMode={field === "quantity" ? "decimal" : undefined} maxLength={max} value={rows[index]?.[field] || ""} onChange={event => editRow(index, field, event.target.value)} />}</FormField>)}
          </fieldset>
        </FieldGroup>)}
        <Button disabled={busy || needsRead || !fileMatches || job.status === "confirming" || job.status === "completed"} onClick={review}>保存人工審查</Button>
        <Button disabled={busy || needsRead || reviewDirty || !fileMatches || !["reviewed", "confirming"].includes(job.status)} onClick={confirm}>{job.status === "confirming" ? "依原審查版本繼續確認" : "明確確認建立工程項目"}</Button>
      </>}
    </FieldGroup>
  </PagePanel>;
}
