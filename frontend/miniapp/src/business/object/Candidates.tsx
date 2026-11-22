import { useRef, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";

type Result = components["schemas"]["ObjectCandidateResult"];
export type Candidate = components["schemas"]["ObjectCandidate"];
export type Selection = { id: string; filename: string; candidate: Candidate };

const messages: Record<Exclude<Result["status"], "completed">, string> = {
  refused: "模型拒絕辨識此文件。",
  incomplete: "模型回應不完整，請縮小文件範圍後再試。",
  invalid_output: "模型回應缺少有效的候選或引文。",
  provider_error: "OpenAI 請求失敗。請檢查回應或稍後再試。",
};

function encodeFile(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new window.FileReader();
    reader.onload = () => resolve(String(reader.result).split(",")[1]);
    reader.onerror = () => reject(new Error("無法讀取檔案。"));
    reader.readAsDataURL(file);
  });
}

export function Candidates({ repositoryId, disabled, created, choose, importDraft }: {
  repositoryId: string; disabled: boolean; created?: ReadonlySet<string>;
  choose?: (selection: Selection, trigger: HTMLButtonElement) => void;
  importDraft?: (filename: string, candidates: Candidate[]) => void;
}) {
  const { request } = useSession();
  const { run } = useLatestRequest();
  const locked = useRef(false);
  const [file, setFile] = useState<File>();
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const [result, setResult] = useState<Result>();
  const [batch, setBatch] = useState("");

  function extract() {
    if (!file || disabled || locked.current) return;
    if (file.size === 0 || file.size > 8 * 1024 * 1024) {
      setMessage("請選擇 8 MiB 以內的非空文件。");
      return;
    }
    locked.current = true;
    setBusy(true); setMessage(""); setResult(undefined);
    void run(async signal => {
      const data = await encodeFile(file);
      if (signal.aborted) throw new DOMException("操作已取消", "AbortError");
      return request((client, s) => client.POST("/api/v1/repositories/{repositoryId}/objects/candidates", {
        signal: s, params: { path: { repositoryId } }, body: { filename: file.name, data },
      }), { signal, timeoutMs: 95000 });
    }, value => {
      locked.current = false; setBusy(false); setBatch(crypto.randomUUID()); setResult(value);
      setMessage(value.status === "completed" ? `找到 ${value.candidates.length} 筆候選，請與原件核對。` : messages[value.status]);
    }, error => { locked.current = false; setBusy(false); setMessage(errorMessage(error)); });
  }

  return <details className="my-3">
    <summary>{importDraft ? "從文件建立" : "從文件辨識物件"}</summary>
    <p>辨識時會將原檔送至 OpenAI，格式依 Responses API 支援範圍判定，每檔最多 8 MiB。試算表每張工作表最多處理前 1,000 列。<a href="https://developers.openai.com/api/docs/guides/file-inputs" target="_blank" rel="noreferrer">官方文件格式與限制</a></p>
    <label>來源文件<Input type="file" disabled={busy} onChange={e => {
      setFile(e.target.files?.[0]); setResult(undefined); setMessage("");
    }} /></label>
    <Button type="button" disabled={!file || disabled || busy} onClick={extract}>辨識候選</Button>
    <p role="status">{busy ? "正在辨識文件……" : message}</p>
    {result?.candidates.map((candidate, index) => {
      const id = `${batch}-${index}`;
      return <article key={id} className="my-3">
        <strong>{candidate.name}</strong>
        {!!candidate.categoryPath?.length && <p>{candidate.categoryPath.join(" › ")}</p>}
        <p>{candidate.quantity ?? "數量待補"} {candidate.unit ?? "單位待補"} · {candidate.location ?? "位置待核對"}</p>
        <blockquote>{candidate.quote}</blockquote>
        {choose && <Button type="button" variant="outline" disabled={disabled || created?.has(id)} onClick={e => choose({ id, candidate, filename: file!.name }, e.currentTarget)}>
          {created?.has(id) ? "已建立" : "覆核並填入新增表單"}
        </Button>}
      </article>;
    })}
    {importDraft && result?.status === "completed" && result.candidates.length > 0 &&
      <Button type="button" disabled={disabled || busy} onClick={() => importDraft(file!.name, result.candidates)}>填入範本草稿</Button>}
    {result?.rawResponse && <details><summary>OpenAI 原始回應</summary><pre className="whitespace-pre-wrap break-all">{result.rawResponse}</pre></details>}
  </details>;
}
