import {
  Accordion,
  AccordionItem,
  AccordionTrigger,
  AccordionContent,
} from "@/shared/ui/accordion";
import { Item, ItemContent, ItemGroup, ItemTitle } from "@/shared/ui/item";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription, EmptyMedia } from "@/shared/ui/empty";
import { FieldGroup } from "@/shared/ui/field";
import { FileTextIcon } from "lucide-react";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { useEffect, useRef, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { PagePanel } from "@/shared/PagePanel";
import { Button } from "@/shared/ui/button";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/shared/ui/collapsible";
import { Status } from "@/shared/Status";
import { Input } from "@/shared/ui/input";
import { Textarea } from "@/shared/ui/textarea";
import { FormField } from "@/shared/FormField";
import { rejectedAttempt, errorMessage, useLatestRequest } from "@/shared/http/lifecycle";
import { useRepositoryGuard } from "../NavigationGuard";

type Page = components["schemas"]["WikiPage"];
type Summary = components["schemas"]["WikiSummary"];
const collection = "/api/v1/repositories/{repositoryId}/wiki-pages";
const detail = "/api/v1/repositories/{repositoryId}/wiki-pages/{id}";
const revisions = "/api/v1/repositories/{repositoryId}/wiki-pages/{id}/revisions";

function Markdown({ body }: { body: string }) {
  return (
    <div className="whitespace-pre-wrap break-words">
      {body.split("\n").map((line, i) => {
        const parts = line
          .replace(/^#{1,3} /, "")
          .split(/(\[[^\]\n]+\]\(https?:\/\/[^\s()]+\))/g)
          .map((part, j) => {
            const link = /^\[([^\]]+)\]\((https?:\/\/[^\s()]+)\)$/.exec(part);
            return link ? (
              <a className="underline" key={j} href={link[2]}>
                {link[1]}
              </a>
            ) : (
              part
            );
          });
        return /^#{1,3} /.test(line) ? (
          <h3 className="mt-4 mb-2 font-semibold" key={i}>
            {parts}
          </h3>
        ) : (
          <p key={i}>{line ? parts : <br />}</p>
        );
      })}
    </div>
  );
}

export function View({ repositoryId, initialId }: { repositoryId: string; initialId: string }) {
  const { request } = useSession();
  const { run, cancel } = useLatestRequest();
  const [id, setId] = useState(initialId),
    [page, setPage] = useState<Page | null>(null);
  const [items, setItems] = useState<Summary[]>([]),
    [history, setHistory] = useState<Page[]>([]);
  const [cursor, setCursor] = useState(""),
    [historyCursor, setHistoryCursor] = useState("");
  const [draft, setDraft] = useState<{ title: string; body: string } | null>(null);
  const [busy, setBusy] = useState(true),
    [message, setMessage] = useState("");
  const [uncertain, setUncertain] = useState(false),
    [compare, setCompare] = useState(false);
  const lock = useRef(false),
    key = useRef(""),
    titleInput = useRef<HTMLInputElement>(null);
  const guard = useRepositoryGuard(!!draft, () => titleInput.current?.focus());
  const path = { repositoryId, id };
  const link = (pageId: string) =>
    "/miniapp/repositories?" + new URLSearchParams({ id: repositoryId, wiki: pageId });
  function fail(e: unknown, keepAttempt = false) {
    const code = (e as { code?: string }).code;
    if (code === "not_found" || code === "forbidden" || code === "unauthorized") {
      setPage(null);
      setHistory([]);
      setItems([]);
      setCursor("");
      setHistoryCursor("");
      setCompare(false);
      if (!keepAttempt) {
        setDraft(null);
        setUncertain(false);
        guard.block(false);
      }
    }
    setBusy(false);
    setMessage(errorMessage(e));
  }
  function load(next = "") {
    setBusy(true);
    setMessage("");
    if (id)
      void run(
        (signal) => request((c, s) => c.GET(detail, { signal: s, params: { path } }), { signal }),
        (p) => {
          setPage(p);
          setBusy(false);
        },
        fail,
      );
    else
      void run(
        (signal) =>
          request(
            (c, s) =>
              c.GET(collection, {
                signal: s,
                params: { path: { repositoryId }, query: { cursor: next || undefined } },
              }),
            { signal },
          ),
        (p) => {
          setItems((old) => (next ? [...old, ...p.items] : p.items));
          setCursor(p.nextCursor);
          setBusy(false);
        },
        fail,
      );
  }
  useEffect(() => {
    setHistory([]);
    setHistoryCursor("");
    load();
    return cancel;
  }, [repositoryId, id]);
  function versions(next = "") {
    setBusy(true);
    setMessage("");
    void run(
      (signal) =>
        request(
          (c, s) =>
            c.GET(revisions, { signal: s, params: { path, query: { cursor: next || undefined } } }),
          { signal },
        ),
      (p) => {
        setHistory((old) => (next ? [...old, ...p.items] : p.items));
        setHistoryCursor(p.nextCursor);
        setBusy(false);
      },
      fail,
    );
  }
  function save() {
    if (!draft || lock.current || !guard.canWrite() || compare) return;
    lock.current = true;
    guard.block(true);
    setBusy(true);
    setMessage("");
    void run(
      (signal) =>
        request(
          (c, s) =>
            page
              ? c.PATCH(detail, {
                  signal: s,
                  params: { path },
                  body: { ...draft, expectedVersion: page.version },
                })
              : c.POST(collection, {
                  signal: s,
                  params: { path: { repositoryId }, header: { "Idempotency-Key": key.current } },
                  body: draft,
                }),
          { signal },
        ),
      (p) => {
        lock.current = false;
        guard.block(false);
        setUncertain(false);
        setBusy(false);
        setDraft(null);
        setPage(p);
        setHistory([]);
        setHistoryCursor("");
        setId(p.id);
      },
      (e) => {
        lock.current = false;
        const { code, conflictReason } = e as { code?: string; conflictReason?: string };
        const unknown = uncertain || !rejectedAttempt(e, ["conflict"]);
        guard.block(unknown);
        setUncertain(unknown);
        setCompare(
          !!page &&
            (unknown ||
              (code === "conflict" && (!conflictReason || conflictReason === "stale_version"))),
        );
        fail(e, unknown && !page);
      },
    );
  }
  return (
    <PagePanel heading="知識庫" description="整理儲存庫的共用知識，並保留每次編輯的版本紀錄。">
      {message && <Status role="status">{message}</Status>}
      {!draft && !id && (
        <>
          <Button
            disabled={busy}
            onClick={() => {
              key.current = crypto.randomUUID();
              setDraft({ title: "", body: "" });
            }}
          >
            建立頁面
          </Button>
          <ReadFeedback busy={busy} label="正在讀取 Wiki 目錄……" />
          {!busy && !message && !items.length && !cursor && (
            <Empty>
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <FileTextIcon />
                </EmptyMedia>
                <EmptyTitle>目前沒有 Wiki 頁面。</EmptyTitle>
                <EmptyDescription>選擇「建立頁面」，開始整理共用知識。</EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}
          <ItemGroup className="gap-3">
            {items.map((p) => (
              <div key={p.id} role="listitem">
                <Item asChild variant="outline">
                  <a href={link(p.id)}>
                    <ItemContent className="min-w-0">
                      <ItemTitle className="max-w-full [overflow-wrap:anywhere]">
                        {p.title}
                      </ItemTitle>
                    </ItemContent>
                  </a>
                </Item>
              </div>
            ))}
          </ItemGroup>
          {cursor && (
            <Button disabled={busy} onClick={() => load(cursor)}>
              更多頁面
            </Button>
          )}
        </>
      )}
      {!draft && page && (
        <>
          <h2 className="text-xl font-semibold [overflow-wrap:anywhere]">{page.title}</h2>
          <Markdown body={page.body} />
          <div className="flex flex-wrap gap-2">
            <Button
              disabled={busy}
              onClick={() => setDraft({ title: page.title, body: page.body })}
            >
              編輯
            </Button>
            <Button variant="outline" disabled={busy} onClick={() => versions()}>
              版本歷史
            </Button>
            <Button asChild variant="outline">
              <a href={link(page.id)}>此頁永久連結</a>
            </Button>
          </div>
          {history.length > 0 && (
            <Accordion type="multiple" key={page.id}>
              {history.map((p) => (
                <AccordionItem key={p.version} value={String(p.version)}>
                  <AccordionTrigger>
                    第 {p.version} 版 · {p.title} · {p.authorId} · {p.updatedAt}
                  </AccordionTrigger>
                  <AccordionContent>
                    <Markdown body={p.body} />
                  </AccordionContent>
                </AccordionItem>
              ))}
            </Accordion>
          )}
          {historyCursor && (
            <Button disabled={busy} onClick={() => versions(historyCursor)}>
              更早版本
            </Button>
          )}
        </>
      )}
      {!draft && id && (
        <Button
          variant="outline"
          disabled={busy}
          onClick={() => {
            setPage(null);
            setId("");
          }}
        >
          返回目錄
        </Button>
      )}
      {!draft && message && (
        <Button variant="outline" disabled={busy} onClick={() => load()}>
          重新讀取
        </Button>
      )}
      {draft && (
        <form
          onSubmit={(e) => {
            e.preventDefault();
            save();
          }}
        >
          <FieldGroup>
            <FormField>
              標題
              <Input
                ref={titleInput}
                required
                maxLength={120}
                disabled={busy || uncertain}
                value={draft.title}
                onChange={(e) => setDraft({ ...draft, title: e.target.value })}
              />
            </FormField>
            <FormField>
              Markdown 內容
              <Textarea
                rows={12}
                className="min-h-64"
                maxLength={32000}
                disabled={busy || uncertain}
                value={draft.body}
                onChange={(e) => setDraft({ ...draft, body: e.target.value })}
              />
            </FormField>
            {compare && (
              <Button
                type="button"
                disabled={busy}
                onClick={() => {
                  setBusy(true);
                  void run(
                    (signal) =>
                      request((c, s) => c.GET(detail, { signal: s, params: { path } }), { signal }),
                    (p) => {
                      setPage(p);
                      setCompare(false);
                      setUncertain(false);
                      guard.block(false);
                      setBusy(false);
                      setMessage("已載入目前版本。請比對草稿後再儲存。");
                    },
                    fail,
                  );
                }}
              >
                讀取目前版本以比對
              </Button>
            )}
            {page && (
              <Collapsible key={page.version}>
                <CollapsibleTrigger asChild>
                  <Button
                    type="button"
                    variant="outline"
                    className="h-auto max-w-full whitespace-normal text-left"
                  >
                    <span className="min-w-0 break-words">
                      對照第 {page.version} 版 · {page.title}
                    </span>
                  </Button>
                </CollapsibleTrigger>
                <CollapsibleContent>
                  <Markdown body={page.body} />
                </CollapsibleContent>
              </Collapsible>
            )}
            <div className="flex flex-wrap gap-2">
              <Button type="submit" disabled={busy || compare}>
                {uncertain && !page ? "重試原建立請求" : "儲存"}
              </Button>
              <Button
                type="button"
                variant="outline"
                disabled={busy || uncertain}
                onClick={() => {
                  setDraft(null);
                  setCompare(false);
                  setMessage("");
                }}
              >
                取消編輯
              </Button>
            </div>
          </FieldGroup>
        </form>
      )}
    </PagePanel>
  );
}
