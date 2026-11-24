import { useRef, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { errorMessage, rejectedAttempt, useLatestRequest } from "@/shared/http/lifecycle";
import { useRepositoryGuard } from "../repository/NavigationGuard";
import { Button } from "@/shared/ui/button";
import { Checkbox } from "@/shared/ui/checkbox";
import { Input } from "@/shared/ui/input";

type ImportRow = {
  id: string;
  key: string;
  body: components["schemas"]["ObjectCreation"];
};
type ImportResult = { objectId?: string; message?: string };

type Result = components["schemas"]["ObjectCandidateResult"];
export type Candidate = components["schemas"]["ObjectCandidate"];
export type Selection = { id: string; filename: string; candidate: Candidate };

const messages: Record<Exclude<Result["status"], "completed">, string> = {
  refused: "模型拒絕辨識此文件。",
  incomplete: "模型回應不完整，請縮小文件範圍後再試。",
  invalid_output: "模型回應缺少有效的候選或引文。",
  provider_error: "OpenAI 請求失敗。請檢查回應或稍後再試。",
};

export function encodeFile(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new window.FileReader();
    reader.onload = () => resolve(String(reader.result).split(",")[1]);
    reader.onerror = () => reject(new Error("無法讀取檔案。"));
    reader.readAsDataURL(file);
  });
}

export function Candidates({
  repositoryId,
  disabled,
  created,
  choose,
}: {
  repositoryId: string;
  disabled: boolean;
  created?: ReadonlyMap<string, string>;
  choose?: (selection: Selection, trigger: HTMLButtonElement) => void;
}) {
  const { request } = useSession();
  const { run } = useLatestRequest();
  const writes = useLatestRequest();
  const [selected, setSelected] = useState<Map<string, string>>(() => new Map());
  const [outcomes, setOutcomes] = useState<Map<string, ImportResult>>(() => new Map());
  const [pending, setPending] = useState<ImportRow[]>([]);
  const [importing, setImporting] = useState(false);
  const heading = useRef<HTMLHeadingElement>(null);
  const guard = useRepositoryGuard(selected.size > 0, () => heading.current?.focus());
  const recovery = importing || pending.length > 0;

  function importRows() {
    if (!file || !result || disabled || locked.current || !guard.canWrite()) return;
    const rows = pending.length
      ? pending
      : result.candidates.flatMap((candidate, index) => {
          const id = `${batch}-${index}`;
          const name = selected.get(id)?.trim();
          if (!name || created?.has(id) || outcomes.get(id)?.objectId) return [];
          return [
            {
              id,
              key: crypto.randomUUID(),
              body: {
                name,
                description: candidate.quote,
                sourceDocument: file.name,
                sourcePages: candidate.location ?? "",
              },
            },
          ];
        });
    if (!rows.length) return;
    locked.current = true;
    guard.block(true);
    setImporting(true);
    void writes.run(
      async (signal) => {
        const unknown: ImportRow[] = [];
        for (const row of rows) {
          if (signal.aborted) break;
          try {
            const item = await request(
              (client, s) =>
                client.POST("/api/v1/repositories/{repositoryId}/objects", {
                  signal: s,
                  params: {
                    path: { repositoryId },
                    header: { "Idempotency-Key": row.key },
                  },
                  body: row.body,
                }),
              { signal },
            );
            if (signal.aborted) break;
            setOutcomes((previous) => new Map(previous).set(row.id, { objectId: item.id }));
            setSelected((previous) => {
              const next = new Map(previous);
              next.delete(row.id);
              return next;
            });
          } catch (error) {
            if (signal.aborted) break;
            const uncertain = pending.length > 0 || !rejectedAttempt(error);
            if (uncertain) unknown.push(row);
            setOutcomes((previous) =>
              new Map(previous).set(row.id, {
                message: uncertain ? `結果未確認：${errorMessage(error)}` : errorMessage(error),
              }),
            );
          }
        }
        return unknown;
      },
      (unknown) => {
        locked.current = false;
        setPending(unknown);
        setImporting(false);
        guard.block(unknown.length > 0);
      },
      (error) => {
        locked.current = false;
        setImporting(false);
        setPending(rows);
        setMessage(errorMessage(error));
      },
    );
  }
  const locked = useRef(false);
  const [file, setFile] = useState<File>();
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const [result, setResult] = useState<Result>();
  const [batch, setBatch] = useState("");

  function extract() {
    if (!file || disabled || recovery || selected.size || locked.current) return;
    if (file.size === 0 || file.size > 8 * 1024 * 1024) {
      setMessage("請選擇 8 MiB 以內的非空文件。");
      return;
    }
    locked.current = true;
    setBusy(true);
    setMessage("");
    setResult(undefined);
    setOutcomes(new Map());
    void run(
      async (signal) => {
        const data = await encodeFile(file);
        if (signal.aborted) throw new DOMException("操作已取消", "AbortError");
        return request(
          (client, s) =>
            client.POST("/api/v1/repositories/{repositoryId}/objects/candidates", {
              signal: s,
              params: { path: { repositoryId } },
              body: { filename: file.name, data },
            }),
          { signal, timeoutMs: 95000 },
        );
      },
      (value) => {
        locked.current = false;
        setBusy(false);
        setBatch(crypto.randomUUID());
        setResult(value);
        setMessage(
          value.status === "completed"
            ? `找到 ${value.candidates.length} 筆候選，請與原件核對。`
            : messages[value.status],
        );
      },
      (error) => {
        locked.current = false;
        setBusy(false);
        setMessage(errorMessage(error));
      },
    );
  }

  const available = (result?.candidates ?? []).flatMap((candidate, index) => {
    const id = `${batch}-${index}`;
    return created?.has(id) || outcomes.get(id)?.objectId ? [] : [{ id, name: candidate.name }];
  });
  const showBatch = !!choose && available.length > 0;

  return (
    <section aria-label="文件辨識" className="grid gap-3">
      <p>
        辨識時會將原檔送至 OpenAI，格式依 Responses API 支援範圍判定，每檔最多 8
        MiB。試算表每張工作表最多處理前 1,000 列。
        <a
          href="https://developers.openai.com/api/docs/guides/file-inputs"
          target="_blank"
          rel="noreferrer"
        >
          官方文件格式與限制
        </a>
      </p>
      <label>
        來源文件
        <Input
          type="file"
          disabled={busy || disabled || recovery || selected.size > 0}
          onChange={(e) => {
            setFile(e.target.files?.[0]);
            setResult(undefined);
            setMessage("");
          }}
        />
      </label>
      <Button
        type="button"
        disabled={!file || disabled || busy || recovery || selected.size > 0}
        onClick={extract}
      >
        辨識候選
      </Button>
      <p role="status">{busy ? "正在辨識文件……" : message}</p>
      {result?.candidates.length ? (
        <h3 ref={heading} tabIndex={-1}>
          覆核候選
        </h3>
      ) : null}
      {showBatch && (
        <div className="grid gap-2">
          <label className="flex items-center gap-2">
            <Checkbox
              checked={selected.size === available.length || (selected.size > 0 && "indeterminate")}
              disabled={disabled || recovery}
              onCheckedChange={(checked) =>
                setSelected((previous) =>
                  checked ? new Map(available.map(({ id, name }) => [id, previous.get(id) ?? name])) : new Map(),
                )
              }
            />
            全選
          </label>
          <p role="status">已選 {selected.size}／{available.length} 筆可匯入候選</p>
          <p>批次匯入會儲存名稱、引文、來源檔名與頁面位置。需要填寫數量、單位或分類時，請使用單筆覆核表單。</p>
        </div>
      )}
      {result?.candidates.map((candidate, index) => {
        const id = `${batch}-${index}`;
        const outcome = outcomes.get(id);
        const objectId = created?.get(id) ?? outcome?.objectId;
        const name = selected.get(id);
        return (
          <article key={id} className="my-3">
            <strong>{candidate.name}</strong>
            {!!candidate.categoryPath?.length && <p>{candidate.categoryPath.join(" › ")}</p>}
            <p>
              {candidate.quantity ?? "數量待補"} {candidate.unit ?? "單位待補"} ·{" "}
              {candidate.location ?? "位置待核對"}
            </p>
            <blockquote>{candidate.quote.trim() || "引文待核對"}</blockquote>
            {objectId && (
              <a
                href={
                  "/miniapp/repositories?" +
                  new URLSearchParams({ id: repositoryId, object: objectId })
                }
              >
                查看已建立物件
              </a>
            )}
            {outcome?.message && !objectId && <p role="status">{outcome.message}</p>}
            {choose && !objectId && (
              <div>
                <label>
                  <input
                    type="checkbox"
                    checked={name !== undefined}
                    disabled={disabled || recovery}
                    onChange={(event) => {
                      const checked = event.currentTarget.checked;
                      setSelected((previous) => {
                        const next = new Map(previous);
                        if (checked) next.set(id, candidate.name);
                        else next.delete(id);
                        return next;
                      });
                    }}
                  />
                  匯入 {candidate.name}
                </label>
                {name !== undefined && (
                  <label>
                    匯入名稱
                    <Input
                      value={name}
                      maxLength={120}
                      disabled={disabled || recovery}
                      onChange={(event) => {
                        const value = event.currentTarget.value;
                        setSelected((previous) => new Map(previous).set(id, value));
                      }}
                    />
                  </label>
                )}
              </div>
            )}
            {choose && (
              <Button
                type="button"
                variant="outline"
                disabled={disabled || recovery || selected.size > 0 || !!objectId}
                onClick={(e) => choose({ id, candidate, filename: file!.name }, e.currentTarget)}
              >
                {objectId ? "已建立" : "覆核並填入新增表單"}
              </Button>
            )}
          </article>
        );
      })}
      {(showBatch || recovery) && (
        <Button
          type="button"
          disabled={
            disabled ||
            importing ||
            (!pending.length && (!selected.size || [...selected.values()].some((name) => !name.trim())))
          }
          onClick={importRows}
        >
          {importing
            ? "正在匯入……"
            : pending.length
              ? `重試 ${pending.length} 筆未確認結果`
              : `確認匯入 ${selected.size} 筆`}
        </Button>
      )}
    </section>
  );
}
