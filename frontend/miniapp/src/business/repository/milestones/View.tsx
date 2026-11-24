import { Progress } from "./MilestoneProgress";
import { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import { useRepositoryGuard } from "../NavigationGuard";
import { useDiscardChanges } from "@/shared/useDiscardChanges";
import { ReadFeedback } from "@/shared/ReadFeedback";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { rejectedAttempt, useLatestRequest, errorMessage } from "@/shared/http/lifecycle";
import { PagePanel } from "@/shared/PagePanel";
import { FormField } from "@/shared/FormField";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Textarea } from "@/shared/ui/textarea";
import { FieldGroup, FieldSet } from "@/shared/ui/field";
import { Badge } from "@/shared/ui/badge";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from "@/shared/ui/empty";
import { Item, ItemGroup, ItemContent, ItemTitle, ItemActions, ItemFooter } from "@/shared/ui/item";

type Milestone = components["schemas"]["Milestone"];
type Creation = components["schemas"]["MilestoneCreation"];
const states = { open: "進行中", completed: "已完成" };
const nextState = { open: "completed", completed: "open" } as const;
const action = { open: "標記完成", completed: "重新開啟" };

export function View({ repositoryId }: { repositoryId: string }) {
  const { user, request } = useSession();
  const { run: readRun, cancel: cancelRead } = useLatestRequest();
  const detailRequest = useLatestRequest(),
    write = useLatestRequest();
  const [items, setItems] = useState<Milestone[]>([]),
    [cursor, setCursor] = useState("");
  const [listBusy, setListBusy] = useState(false),
    [loaded, setLoaded] = useState(false);
  const [listStatus, setListStatus] = useState("");
  const [mode, setMode] = useState<"create" | "detail" | null>(null);
  const [selected, setSelected] = useState<Milestone | null>(null);
  const [dueDate, setDueDate] = useState("");
  const [title, setTitle] = useState(""),
    [body, setBody] = useState("");
  const [busy, setBusy] = useState(false),
    [detailBusy, setDetailBusy] = useState(false);
  const [uncertain, setUncertain] = useState(false),
    [needsRefresh, setNeedsRefresh] = useState(false);
  const [status, setStatus] = useState("");
  const lock = useRef(false);
  const attempt = useRef<{ key: string; body: Creation } | null>(null);
  const canEdit = selected?.canEdit === true;
  const [activeId, setActiveId] = useState("");
  const baseline = useRef(JSON.stringify(["", "", ""]));
  const dirty = !!mode && JSON.stringify([title, body, dueDate]) !== baseline.current;
  const titleInput = useRef<HTMLInputElement>(null),
    editorPanel = useRef<HTMLElement>(null),
    opener = useRef<HTMLElement | null>(null);
  const navigating = useRef(false);
  function focusEditor() {
    if (titleInput.current && !titleInput.current.matches(":disabled")) titleInput.current.focus();
    else editorPanel.current?.focus();
  }
  const discard = useDiscardChanges(dirty, focusEditor);
  const guard = useRepositoryGuard(dirty, focusEditor, { busy, uncertain, needsRefresh });
  useEffect(() => {
    navigating.current = false;
  });
  useEffect(() => {
    if (!mode) opener.current?.focus();
    else if (!detailBusy && !busy) focusEditor();
  }, [mode, detailBusy, busy]);
  function clearEditor() {
    detailRequest.cancel();
    setDetailBusy(false);
    setSelected(null);
    setMode(null);
    setActiveId("");
    setNeedsRefresh(false);
    setStatus("");
  }
  function close() {
    if (lock.current || uncertain || needsRefresh) return;
    discard.navigate(() => {
      navigating.current = true;
      clearEditor();
    });
  }
  function saved(item: Milestone, message: string) {
    guard.block(false);
    setUncertain(false);
    attempt.current = null;
    clearEditor();
    setItems((previous) =>
      previous.some((value) => value.id === item.id)
        ? previous.map((value) => (value.id === item.id ? item : value))
        : [item, ...previous],
    );
    setListStatus(message);
  }

  const load = useCallback(
    (nextCursor = "") => {
      if (!user) return;
      setListBusy(true);
      setLoaded(false);
      setCursor("");
      if (!nextCursor) setItems([]);
      setListStatus("");
      void readRun(
        (signal) =>
          request(
            (client, s) =>
              client.GET("/api/v1/repositories/{repositoryId}/milestones", {
                signal: s,
                params: { path: { repositoryId }, query: { cursor: nextCursor || undefined } },
              }),
            { signal },
          ),
        (page) => {
          setItems((previous) => [
            ...new Map(
              [...(nextCursor ? previous : []), ...page.items].map((item) => [item.id, item]),
            ).values(),
          ]);
          setCursor(page.nextCursor);
          setListBusy(false);
          setLoaded(true);
          setListStatus("");
        },
        (error) => {
          setItems([]);
          setListBusy(false);
          setLoaded(false);
          setListStatus(errorMessage(error));
        },
      );
    },
    [user, request, repositoryId, readRun],
  );
  useEffect(() => {
    load();
    return cancelRead;
  }, [load, cancelRead]);

  function show(item: Milestone) {
    baseline.current = JSON.stringify([item.title, item.body, item.dueDate]);
    setActiveId(item.id);
    guard.block(false);
    setSelected(item);
    setTitle(item.title);
    setBody(item.body);
    setDueDate(item.dueDate);
    setMode("detail");
    setNeedsRefresh(false);
    setUncertain(false);
    attempt.current = null;
  }
  function open(id: string, trigger?: HTMLElement) {
    if (
      !user ||
      lock.current ||
      uncertain ||
      (needsRefresh && id !== activeId) ||
      !guard.canWrite()
    )
      return;
    discard.navigate(() => {
      if (trigger) opener.current = trigger;
      navigating.current = true;
      cancelRead();
      setListBusy(false);
      baseline.current = JSON.stringify(["", "", ""]);
      setTitle("");
      setBody("");
      setDueDate("");
      setActiveId(id);
      setSelected(null);
      setMode("detail");
      setDetailBusy(true);
      setStatus("");
      void detailRequest.run(
        (signal) =>
          request(
            (client, s) =>
              client.GET("/api/v1/repositories/{repositoryId}/milestones/{id}", {
                signal: s,
                params: { path: { repositoryId, id } },
              }),
            { signal },
          ),
        (item) => {
          show(item);
          setDetailBusy(false);
        },
        (error) => {
          setDetailBusy(false);
          setStatus(errorMessage(error));
        },
      );
    });
  }
  function openCreateForm(trigger: HTMLElement) {
    if (lock.current || uncertain || needsRefresh || !guard.canWrite()) return;
    discard.navigate(() => {
      opener.current = trigger;
      navigating.current = true;
      cancelRead();
      setListBusy(false);
      clearEditor();
      baseline.current = JSON.stringify(["", "", ""]);
      setMode("create");
      setTitle("");
      setBody("");
      setDueDate("");
      attempt.current = null;
    });
  }
  function submit(event: FormEvent) {
    event.preventDefault();
    if (
      !user ||
      lock.current ||
      navigating.current ||
      needsRefresh ||
      discard.isConfirming() ||
      !guard.canWrite()
    )
      return;
    if (mode === "detail") {
      if (selected) save(selected.state);
      return;
    }
    const content = { title: title.trim(), body: body.trim(), dueDate };
    if (!content.title) {
      setStatus("請填寫里程碑標題。");
      return;
    }
    if (!uncertain || !attempt.current)
      attempt.current = { key: crypto.randomUUID(), body: content };
    const pending = attempt.current;
    const wasUnknown = uncertain;
    guard.block(true);
    lock.current = true;
    setBusy(true);
    setUncertain(true);
    setStatus("正在建立里程碑……");
    void write.run(
      (signal) =>
        request(
          (client, s) =>
            client.POST("/api/v1/repositories/{repositoryId}/milestones", {
              signal: s,
              params: { path: { repositoryId }, header: { "Idempotency-Key": pending.key } },
              body: pending.body,
            }),
          { signal },
        ),
      (item) => {
        lock.current = false;
        setBusy(false);
        saved(item, "已建立里程碑。");
      },
      (error) => {
        lock.current = false;
        setBusy(false);
        if (!wasUnknown && rejectedAttempt(error)) {
          setUncertain(false);
          guard.block(false);
        }
        setStatus(
          errorMessage(error) +
            " 結果未確認時請沿用原內容重試，或先查詢里程碑；重新開頁前請先確認紀錄。",
        );
      },
    );
  }
  function save(state: Milestone["state"]) {
    if (
      !user ||
      !selected ||
      !canEdit ||
      lock.current ||
      navigating.current ||
      needsRefresh ||
      discard.isConfirming() ||
      !guard.canWrite()
    )
      return;
    const id = selected.id;
    guard.block(true);
    lock.current = true;
    setBusy(true);
    setStatus("正在更新里程碑……");
    void write.run(
      (signal) =>
        request(
          (client, s) =>
            client.PATCH("/api/v1/repositories/{repositoryId}/milestones/{id}", {
              signal: s,
              params: { path: { repositoryId, id } },
              body: {
                title: title.trim(),
                body: body.trim(),
                dueDate,
                state,
                expectedVersion: selected.version,
              },
            }),
          { signal },
        ),
      (item) => {
        lock.current = false;
        setBusy(false);
        saved(item, "已更新里程碑。");
      },
      (error) => {
        lock.current = false;
        setBusy(false);
        setNeedsRefresh(true);
        setStatus(errorMessage(error) + " 請重新讀取里程碑後再修改。");
      },
    );
  }
  const percent = selected?.completionPercent;
  const editor = (
    <section
      ref={editorPanel}
      tabIndex={-1}
      aria-label="里程碑編輯"
      className="flex w-full min-w-0 flex-col gap-3"
    >
      <ReadFeedback busy={detailBusy} message={status} label="正在讀取里程碑……" />
      {selected &&
        (percent === undefined ? (
          <p>尚無問題</p>
        ) : (
          <>
            <p>
              已關閉 {selected.completedIssueCount}/{selected.issueCount} 個問題（{percent}%）
            </p>
            <Progress value={percent} aria-label="里程碑完成度" />
          </>
        ))}
      {mode === "detail" && (
        <div className="mb-4 flex flex-wrap items-center gap-2">
          {!detailBusy && (!selected || needsRefresh) && (
            <Button variant="outline" disabled={busy || detailBusy} onClick={() => open(activeId)}>
              重新讀取里程碑
            </Button>
          )}
        </div>
      )}
      {(mode === "create" || selected) && (
        <form onSubmit={submit}>
          <FieldGroup>
            <FieldSet
              className="grid min-w-0 gap-4"
              disabled={
                busy || detailBusy || uncertain || needsRefresh || (mode === "detail" && !canEdit)
              }
            >
              <FormField>
                到期日
                <Input type="date" value={dueDate} onChange={(e) => setDueDate(e.target.value)} />
              </FormField>
              <FormField>
                標題
                <Input
                  ref={titleInput}
                  required
                  maxLength={120}
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                />
              </FormField>
              <FormField>
                內容
                <Textarea maxLength={8000} value={body} onChange={(e) => setBody(e.target.value)} />
              </FormField>
            </FieldSet>

            {(mode === "create" || canEdit) && (
              <div className="flex flex-wrap gap-2">
                <Button type="submit" disabled={busy || needsRefresh}>
                  {mode === "create" ? (uncertain ? "重試原建立請求" : "建立里程碑") : "儲存內容"}
                </Button>
                {selected && (
                  <Button
                    type="button"
                    variant="outline"
                    disabled={busy || needsRefresh}
                    onClick={() => save(nextState[selected.state])}
                  >
                    {action[selected.state]}
                  </Button>
                )}
              </div>
            )}
          </FieldGroup>
        </form>
      )}
      <Button
        type="button"
        variant="outline"
        disabled={busy || uncertain || needsRefresh}
        onClick={close}
      >
        {canEdit || mode === "create" ? "取消" : "關閉"}
      </Button>
    </section>
  );
  return (
    <>
      <PagePanel heading="里程碑">
        <div className="mb-4 flex flex-wrap gap-2">
          <Button
            disabled={!user || busy || uncertain || needsRefresh}
            onClick={(event) => openCreateForm(event.currentTarget)}
          >
            新增里程碑
          </Button>
        </div>
        <ReadFeedback
          busy={listBusy}
          message={listStatus}
          label="正在讀取里程碑……"
          onRetry={
            user && listStatus && !loaded && !busy && !mode && !uncertain && !needsRefresh
              ? () => load()
              : undefined
          }
        />
        {mode === "create" && <PagePanel heading="新增里程碑">{editor}</PagePanel>}
        <ItemGroup className="gap-3">
          {items.map((item) => (
            <Item key={item.id} variant="outline" role="listitem" className="min-w-0">
              <ItemContent className="min-w-0">
                <ItemTitle className="min-w-0 max-w-full">
                  <Button
                    variant="link"
                    className="h-auto min-w-0 max-w-full justify-start whitespace-normal text-left [overflow-wrap:anywhere]"
                    disabled={busy || detailBusy || uncertain || needsRefresh}
                    onClick={(event) => open(item.id, event.currentTarget)}
                  >
                    {item.title}
                  </Button>
                </ItemTitle>
                <Badge variant="outline">{states[item.state]}</Badge>
              </ItemContent>
              {item.canEdit && (
                <ItemActions>
                  <Button
                    variant="outline"
                    aria-label={`編輯里程碑：${item.title}`}
                    disabled={busy || detailBusy || uncertain || needsRefresh}
                    onClick={(event) => open(item.id, event.currentTarget)}
                  >
                    編輯
                  </Button>
                </ItemActions>
              )}
              {mode === "detail" && activeId === item.id && <ItemFooter>{editor}</ItemFooter>}
            </Item>
          ))}
        </ItemGroup>
        {loaded && !items.length && (
          <Empty>
            <EmptyHeader>
              <EmptyTitle>{cursor ? "本頁沒有里程碑" : "目前沒有里程碑"}</EmptyTitle>
              <EmptyDescription>
                {cursor ? "請繼續載入下一頁。" : "可建立新的里程碑。"}
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
        {cursor && (
          <Button
            variant="outline"
            disabled={listBusy || busy || !!mode || uncertain || needsRefresh}
            onClick={() => load(cursor)}
          >
            載入更多里程碑
          </Button>
        )}
      </PagePanel>

      {discard.dialog}
    </>
  );
}
