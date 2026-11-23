import { ChevronDown, ChevronRight, Plus, Ellipsis } from "lucide-react";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/shared/ui/collapsible";
import { useCallback, useEffect, useRef, useState, type ReactNode, type FormEvent } from "react";
import { RepositoryTarget } from "../repository/RepositoryTarget";
import { Evidence, TemplatePreparation, type Review } from "./TemplatePreparation";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { rejectedAttempt, useLatestRequest, errorMessage } from "@/shared/http/lifecycle";
import { useRepositoryGuard } from "../repository/NavigationGuard";
import { useDiscardChanges } from "@/shared/useDiscardChanges";
import { FormField } from "@/shared/FormField";
import { PagePanel } from "@/shared/PagePanel";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Textarea } from "@/shared/ui/textarea";
import { FieldGroup, FieldSet, FieldLegend } from "@/shared/ui/field";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";

type Snapshot = components["schemas"]["ObjectTemplate"];
type Creation = components["schemas"]["ObjectTemplateCreation"];
type Props = {
  standalone?: boolean;
  initiallyOpen?: boolean;
  repositoryId: string;
  disabled: boolean;
  onBlocked: (blocked: boolean) => void;
  onApplied: (
    result: components["schemas"]["ObjectTemplateInstance"],
    repositoryId: string,
  ) => void;
};
type Attempt = { key: string } & (
  | { kind: "create"; body: Creation }
  | {
      kind: "apply";
      id: string;
      repositoryId: string;
      body: components["schemas"]["ObjectTemplateInstantiation"];
    }
);
const initial = (): Creation => ({ name: "", nodes: [{ id: "root", name: "" }] });

// Parent links define the hierarchy; source order only orders siblings.
function treeRows(nodes: Creation["nodes"], folded: string[] = []) {
  type Node = Creation["nodes"][number];
  const children = new Map<string, Node[]>();
  for (const node of nodes) {
    const parent = node.parentId || "";
    const siblings = children.get(parent) ?? [];
    siblings.push(node);
    children.set(parent, siblings);
  }
  const rows: { node: Node; depth: number; hasChildren: boolean }[] = [];
  function visit(parent: string, depth: number) {
    for (const node of children.get(parent) ?? []) {
      rows.push({ node, depth, hasChildren: children.has(node.id) });
      if (!folded.includes(node.id)) visit(node.id, depth + 1);
    }
  }
  visit("", 0);
  return rows;
}

function TreeRow({ depth, children }: { depth: number; children: ReactNode }) {
  return (
    <li style={{ marginInlineStart: Math.min(depth, 6) * 16 }} className="min-w-0 py-1">
      {children}
    </li>
  );
}

const inlineField =
  "min-h-8 h-auto resize-none rounded-sm border-transparent bg-transparent px-1 py-1 text-sm shadow-none hover:bg-muted focus:bg-background focus:border-ring dark:bg-transparent";

type NodePatch = Partial<
  Pick<Creation["nodes"][number], "name" | "unit" | "quantity" | "parentId">
>;

function updateNode(nodes: Creation["nodes"], id: string, patch: NodePatch) {
  return nodes.map((node) => (node.id === id ? { ...node, ...patch } : node));
}

function removeDraftBranch(nodes: Creation["nodes"], id: string) {
  const removed = new Set([id]);
  while (true) {
    const before = removed.size;
    for (const node of nodes) if (node.parentId && removed.has(node.parentId)) removed.add(node.id);
    if (before === removed.size) return nodes.filter((node) => !removed.has(node.id));
  }
}

// Only input, request lifecycle and navigation live here. Go validates and creates the tree.
export function Templates({
  repositoryId,
  disabled,
  onBlocked,
  onApplied,
  initiallyOpen = false,
  standalone = false,
}: Props) {
  const { request, user } = useSession();
  const read = useLatestRequest(),
    write = useLatestRequest();
  const [open, setOpen] = useState(false),
    [creating, setCreating] = useState(false);
  const [draft, setDraft] = useState<Creation>(initial),
    [items, setItems] = useState<Snapshot[]>([]);
  const [method, setMethod] = useState<"choose" | Review["source"]>();
  const [review, setReview] = useState<Review>();
  const [folded, setFolded] = useState<string[]>([]);
  const [scope, setScope] = useState<"repository" | "personal">(
    standalone ? "personal" : "repository",
  );
  const [targetRepository, setTargetRepository] = useState(repositoryId);
  const [plan, setPlan] = useState<Creation["nodes"]>();
  const [excluded, setExcluded] = useState<string[]>([]);
  const [selected, setSelected] = useState(""),
    [cursor, setCursor] = useState("");
  const [busy, setBusy] = useState(false),
    [reading, setReading] = useState(false),
    [message, setMessage] = useState("");
  const [success, setSuccess] = useState(false);
  const [uncertain, setUncertain] = useState(false);
  const attempt = useRef<Attempt | null>(null),
    lock = useRef(false),
    panel = useRef<HTMLElement>(null),
    trigger = useRef<HTMLButtonElement>(null);
  const focus = () => panel.current?.focus();
  const dirty = !!plan || (creating && JSON.stringify(draft) !== JSON.stringify(initial()));
  const guard = useRepositoryGuard(dirty, focus),
    discard = useDiscardChanges(dirty, focus);
  useEffect(() => {
    onBlocked(open);
    return () => onBlocked(false);
  }, [open, onBlocked]);
  useEffect(() => {
    guard.block(busy || uncertain);
  }, [busy, uncertain, guard.block]);
  useEffect(() => {
    if (open) focus();
  }, [open]);
  const load = useCallback(
    (next = "", library = scope) => {
      setReading(true);
      setMessage("");
      if (!next) {
        setItems([]);
        setSelected("");
      }
      void read.run(
        (signal) =>
          request(
            (client, s) =>
              standalone
                ? client.GET("/api/v1/object-templates", {
                    signal: s,
                    params: { query: { cursor: next || undefined } },
                  })
                : client.GET("/api/v1/repositories/{repositoryId}/object-templates", {
                    signal: s,
                    params: {
                      path: { repositoryId },
                      query: {
                        cursor: next || undefined,
                        scope: library === "personal" ? "personal" : undefined,
                      },
                    },
                  }),
            { signal },
          ),
        (page) => {
          setItems((previous) => (next ? [...previous, ...page.items] : page.items));
          setCursor(page.nextCursor);
          setReading(false);
        },
        (error) => {
          setReading(false);
          setSuccess(false);
          setMessage(errorMessage(error));
        },
      );
    },
    [read.run, request, repositoryId, scope, standalone],
  );
  const openedFromEntry = useRef(false);
  // Strict Mode cancels the initial read during effect cleanup. Allow the
  // replayed mount to start it again, just as a fresh mount would.
  useEffect(
    () => () => {
      openedFromEntry.current = false;
    },
    [],
  );
  useEffect(() => {
    if (!user || !initiallyOpen || disabled || openedFromEntry.current) return;
    openedFromEntry.current = true;
    setOpen(true);
    load();
  }, [user, initiallyOpen, disabled, load]);
  function close() {
    if (lock.current || uncertain) return;
    discard.navigate(() => {
      read.cancel();
      onBlocked(false);
      setOpen(false);
      setMethod(undefined);
      setPlan(undefined);
      setExcluded([]);
      setCreating(false);
      setDraft(initial());
      setReview(undefined);
      setFolded([]);
      setMessage("");
      trigger.current?.focus();
    });
  }
  function send(pending: Attempt) {
    if (lock.current || discard.isConfirming() || !guard.canWrite()) return;
    attempt.current = pending;
    const wasUnknown = uncertain;
    lock.current = true;
    setBusy(true);
    setUncertain(true);
    guard.block(true);
    setMessage("正在處理範本……");
    const failed = (error: unknown) => {
      lock.current = false;
      setBusy(false);
      if (!wasUnknown && rejectedAttempt(error)) {
        setUncertain(false);
        guard.block(false);
        attempt.current = null;
      }
      setSuccess(false);
      setMessage(errorMessage(error) + " 結果未確認時，請重試原請求。");
    };
    const settled = () => {
      lock.current = false;
      setBusy(false);
      setUncertain(false);
      guard.block(false);
      attempt.current = null;
    };
    if (pending.kind === "create") {
      void write.run(
        (signal) =>
          request(
            (client, s) =>
              standalone
                ? client.POST("/api/v1/object-templates", {
                    signal: s,
                    params: { header: { "Idempotency-Key": pending.key } },
                    body: pending.body,
                  })
                : client.POST("/api/v1/repositories/{repositoryId}/object-templates", {
                    signal: s,
                    params: {
                      path: { repositoryId },
                      query: { scope: scope === "personal" ? "personal" : undefined },
                      header: { "Idempotency-Key": pending.key },
                    },
                    body: pending.body,
                  }),
            { signal },
          ),
        (snapshot) => {
          settled();
          setSuccess(true);
          setItems((previous) => [snapshot, ...previous.filter((item) => item.id !== snapshot.id)]);
          setSelected(snapshot.id);
          setCreating(false);
          setDraft(initial());
          setReview(undefined);
          if (!pending.body.templateId) setFolded([]);
          setMessage(
            pending.body.templateId
              ? `已儲存為 v${snapshot.version}。`
              : "已建立範本，可直接套用。",
          );
        },
        failed,
      );
    } else {
      void write.run(
        (signal) =>
          request(
            (client, s) =>
              client.POST("/api/v1/repositories/{repositoryId}/object-templates/{id}/instances", {
                signal: s,
                params: {
                  path: { repositoryId: pending.repositoryId, id: pending.id },
                  query: { scope: scope === "personal" ? "personal" : undefined },
                  header: { "Idempotency-Key": pending.key },
                },
                body: pending.body,
              }),
            { signal },
          ),
        (result) => {
          settled();
          setSuccess(true);
          read.cancel();
          setMessage(`已建立 ${result.objectIds.length} 個物件。`);
          setOpen(standalone);
          setPlan(undefined);
          setExcluded([]);
          setCreating(false);
          setDraft(initial());
          setReview(undefined);
          setFolded([]);
          onBlocked(false);
          onApplied(result, pending.repositoryId);
        },
        failed,
      );
    }
  }
  function create(event: FormEvent) {
    event.preventDefault();
    const object = draft.nodes.find((node) => !node.parentId);
    if (!object?.name.trim()) return;
    send({
      kind: "create",
      key: crypto.randomUUID(),
      body: structuredClone({ ...draft, name: object.name }),
    });
  }
  const nameTooLong = (name: string) => Array.from(name.trim()).length > 120;
  const draftTooLong = draft.nodes.some((node) => nameTooLong(node.name));
  const NodeName = scope === "personal" ? Textarea : Input;
  const snapshot = items.find((item) => item.id === selected);
  const byID = new Map(plan?.map((node) => [node.id, node]));
  function isRemoved(id: string): boolean {
    for (
      let node = byID.get(id);
      node;
      node = node.parentId ? byID.get(node.parentId) : undefined
    ) {
      if (excluded.includes(node.id)) return true;
    }
    return false;
  }
  function descendantsOf(id: string) {
    return (
      plan?.filter((node) => {
        for (let parent = node.parentId; parent; parent = byID.get(parent)?.parentId) {
          if (parent === id) return true;
        }
        return false;
      }) ?? []
    );
  }
  const retained = plan?.filter((node) => !isRemoved(node.id)) ?? [];
  function editDraftNode(id: string, patch: NodePatch) {
    setDraft((previous) => ({ ...previous, nodes: updateNode(previous.nodes, id, patch) }));
  }
  function editNode(id: string, patch: NodePatch) {
    setPlan((previous) => previous && updateNode(previous, id, patch));
  }
  const focusChild = useRef<string | null>(null);
  function addChild(parentId: string) {
    const child = { id: crypto.randomUUID(), parentId, name: "" };
    focusChild.current = child.id;
    setFolded((previous) => previous.filter((id) => id !== parentId));
    setDraft((previous) => ({ ...previous, nodes: [...previous.nodes, child] }));
  }
  const visibleNodes = creating ? draft.nodes : (snapshot?.nodes ?? []);
  return (
    <>
      {!standalone && (
        <Button
          ref={trigger}
          variant="outline"
          disabled={disabled || open}
          onClick={() => {
            if (!guard.canWrite()) return;
            onBlocked(true);
            setOpen(true);
            load();
          }}
        >
          物件範本
        </Button>
      )}
      {open && (
        <PagePanel
          className="border-0 shadow-none"
          heading={
            creating
              ? draft.templateId
                ? "編輯範本"
                : "新增範本"
              : plan
                ? "確認本次建立的物件"
                : method === "choose"
                  ? "如何建立範本？"
                  : method === "document"
                    ? "從文件建立範本"
                    : method === "procedure"
                      ? "從工項推導工序"
                      : standalone
                        ? "我的物件範本"
                        : "物件範本"
          }
          description={plan ? snapshot?.name : undefined}
        >
          <section
            ref={panel}
            tabIndex={-1}
            aria-label="物件範本操作"
            className="flex min-w-0 flex-col gap-4 outline-none"
          >
            <ReadFeedback
              plain={success}
              busy={reading || busy}
              message={message}
              label="正在讀取範本……"
            />
            {method === "choose" && (
              <div className="grid gap-3">
                <Button
                  onClick={() => {
                    setMethod(undefined);
                    setCreating(true);
                  }}
                >
                  從空白建立
                </Button>
                <Button variant="outline" onClick={() => setMethod("document")}>
                  從文件建立
                </Button>
                <Button variant="outline" onClick={() => setMethod("procedure")}>
                  從工項推導工序
                </Button>
              </div>
            )}
            {(method === "procedure" || method === "document") && (
              <TemplatePreparation
                key={method}
                source={method}
                disabled={busy || uncertain}
                onDraft={(review) => {
                  setReview(review);
                  setDraft(review.result.draft);
                  setMethod(undefined);
                  setCreating(true);
                }}
              />
            )}
            {method && (
              <Button
                variant="outline"
                onClick={() => setMethod(method === "choose" ? undefined : "choose")}
              >
                {method === "choose" ? "返回範本庫" : "返回建立方式"}
              </Button>
            )}
            {!method && (!creating || draft.templateId) && !plan && (
              <>
                {!standalone && (
                  <FormField>
                    範本庫
                    <NativeSelect
                      value={scope}
                      disabled={busy || uncertain}
                      onChange={(event) => {
                        const library =
                          event.target.value === "personal" ? "personal" : "repository";
                        setScope(library);
                        load("", library);
                      }}
                    >
                      <NativeSelectOption value="repository">本庫</NativeSelectOption>
                      <NativeSelectOption value="personal">我的範本</NativeSelectOption>
                    </NativeSelect>
                  </FormField>
                )}
                <Button
                  className="self-start"
                  disabled={creating || busy || uncertain || reading}
                  onClick={() => (scope === "personal" ? setMethod("choose") : setCreating(true))}
                >
                  新增範本
                </Button>
                {items.length > 0 && (
                  <nav aria-label="選擇範本" className="grid gap-1">
                    {items.map((item) => (
                      <Button
                        key={item.id}
                        variant={selected === item.id ? "secondary" : "ghost"}
                        aria-expanded={selected === item.id}
                        className="justify-between"
                        disabled={creating || busy || uncertain}
                        onClick={() => {
                          setSelected((current) => (current === item.id ? "" : item.id));
                          setFolded([]);
                        }}
                      >
                        <span>{item.name}</span>
                        <span>v{item.version}</span>
                      </Button>
                    ))}
                  </nav>
                )}
                {!reading && !busy && !message && !items.length && (
                  <p className="text-sm text-muted-foreground">尚無範本。</p>
                )}
                {cursor && (
                  <Button
                    variant="outline"
                    disabled={creating || busy || uncertain || reading}
                    onClick={() => load(cursor)}
                  >
                    載入更多範本
                  </Button>
                )}
              </>
            )}
            {plan && (
              <form
                onSubmit={(event) => {
                  event.preventDefault();
                  if (!targetRepository) return;
                  send({
                    kind: "apply",
                    repositoryId: targetRepository,
                    key: crypto.randomUUID(),
                    id: selected,
                    body: {
                      nodes: retained,
                      ...(snapshot && snapshot.version > 1
                        ? { templateVersion: snapshot.version }
                        : {}),
                    },
                  });
                }}
              >
                <FieldGroup>
                  <FieldSet disabled={busy || uncertain}>
                    {standalone && (
                      <RepositoryTarget
                        value={targetRepository}
                        onChange={setTargetRepository}
                        disabled={busy || uncertain}
                      />
                    )}
                    {treeRows(plan).map(({ node, depth }) => {
                      const removed = isRemoved(node.id);
                      const descendants = descendantsOf(node.id);
                      return (
                        <Collapsible
                          key={node.id}
                          style={{ marginInlineStart: Math.min(depth, 6) * 16 }}
                          className="py-1"
                        >
                          <CollapsibleTrigger asChild>
                            <Button
                              type="button"
                              variant="ghost"
                              className="h-auto min-h-11 w-full justify-between gap-3 whitespace-normal text-start"
                            >
                              <span>{node.name || "未命名物件"}</span>
                              <span className="flex shrink-0 items-center gap-2 text-muted-foreground">
                                {removed
                                  ? "本次不建立"
                                  : [node.quantity, node.unit].filter(Boolean).join(" ")}
                                <ChevronDown aria-hidden="true" />
                              </span>
                            </Button>
                          </CollapsibleTrigger>
                          <CollapsibleContent>
                            <FieldSet className="px-2 pb-4">
                              <FieldLegend>{node.parentId ? "工序" : "物件"}</FieldLegend>
                              <FormField>
                                物件名稱
                                <NodeName
                                  required
                                  maxLength={120}
                                  disabled={removed}
                                  value={node.name}
                                  onChange={(event) =>
                                    editNode(node.id, { name: event.target.value })
                                  }
                                />
                              </FormField>
                              {node.parentId && <p>父物件：{byID.get(node.parentId)?.name}</p>}
                              <div className="grid grid-cols-2 gap-2">
                                <FormField>
                                  數量
                                  <Input
                                    inputMode="decimal"
                                    disabled={removed}
                                    value={node.quantity ?? ""}
                                    onChange={(event) =>
                                      editNode(node.id, { quantity: event.target.value })
                                    }
                                  />
                                </FormField>
                                <FormField>
                                  單位
                                  <Input
                                    maxLength={32}
                                    disabled={removed}
                                    value={node.unit ?? ""}
                                    onChange={(event) =>
                                      editNode(node.id, { unit: event.target.value })
                                    }
                                  />
                                </FormField>
                              </div>
                              {node.parentId && !removed && descendants.length > 0 && (
                                <p>
                                  移除時一併移除：
                                  {descendants.map((child) => child.name).join("、")}
                                </p>
                              )}
                              {node.parentId && (!removed || excluded.includes(node.id)) && (
                                <Button
                                  type="button"
                                  variant="outline"
                                  onClick={() =>
                                    setExcluded((previous) =>
                                      removed
                                        ? previous.filter((id) => id !== node.id)
                                        : [...previous, node.id],
                                    )
                                  }
                                >
                                  {removed ? "復原工序" : "移除工序（含子工序）"}
                                </Button>
                              )}
                              {removed && <p>本次不建立</p>}
                            </FieldSet>
                          </CollapsibleContent>
                        </Collapsible>
                      );
                    })}
                    <div className="sticky bottom-0 z-10 flex flex-wrap items-center gap-2 border-t bg-background py-3">
                      <Button
                        type="submit"
                        disabled={!targetRepository || retained.some((node) => !node.name.trim())}
                      >
                        建立 {retained.length} 個物件
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        onClick={() =>
                          discard.navigate(() => {
                            setPlan(undefined);
                            setExcluded([]);
                            if (standalone) load();
                          })
                        }
                      >
                        返回範本
                      </Button>
                    </div>
                  </FieldSet>
                </FieldGroup>
              </form>
            )}
            {creating && draft.nodes.length > 100 && (
              <p role="alert">範本最多 100 個節點（含根物件），請移除不需要的工序後再儲存。</p>
            )}
            {!plan && !method && (creating || snapshot) && (
              <form onSubmit={(event) => (creating ? create(event) : event.preventDefault())}>
                <FieldGroup>
                  <FieldSet disabled={busy || uncertain}>
                    {review &&
                      !review.result.candidateIndexByNode &&
                      review.result.candidates.map((candidate, i) => (
                        <Evidence key={i} candidate={candidate} source={review.source} />
                      ))}
                    {review?.result.warnings.map((warning, i) => (
                      <p role="status" key={i}>
                        {warning}
                      </p>
                    ))}
                    {snapshot && (!creating || draft.templateId) && (
                      <h3 className="font-semibold">
                        {snapshot.name} · v{snapshot.version}
                      </h3>
                    )}
                    <ul aria-label="範本內容">
                      {treeRows(visibleNodes, folded).map(({ node, depth, hasChildren }) => {
                        const index = review?.result.candidateIndexByNode?.[node.id];
                        const candidate =
                          index === undefined ? undefined : review?.result.candidates[index];
                        const nameLabel = node.parentId ? "子項名稱" : "物件名稱";
                        const collapsed = folded.includes(node.id);
                        return (
                          <TreeRow key={node.id} depth={depth}>
                            <Collapsible>
                              <div className="flex items-start gap-1">
                                {hasChildren ? (
                                  <Button
                                    type="button"
                                    size="icon-sm"
                                    variant="ghost"
                                    aria-label={collapsed ? "展開子工序" : "收合子工序"}
                                    aria-expanded={!collapsed}
                                    onClick={() =>
                                      setFolded((current) =>
                                        collapsed
                                          ? current.filter((id) => id !== node.id)
                                          : [...current, node.id],
                                      )
                                    }
                                  >
                                    {collapsed ? <ChevronRight /> : <ChevronDown />}
                                  </Button>
                                ) : (
                                  <span aria-hidden="true" className="size-8 shrink-0" />
                                )}
                                {creating ? (
                                  <label className="min-w-0 flex-1">
                                    <span className="sr-only">{nameLabel}</span>
                                    <Textarea
                                      aria-label={nameLabel}
                                      rows={1}
                                      className={inlineField}
                                      placeholder={nameLabel}
                                      required
                                      maxLength={120}
                                      value={node.name}
                                      ref={(el) => {
                                        if (el && focusChild.current === node.id) {
                                          focusChild.current = null;
                                          el.focus();
                                        }
                                      }}
                                      onChange={(event) =>
                                        editDraftNode(node.id, { name: event.target.value })
                                      }
                                    />
                                  </label>
                                ) : (
                                  <span className="min-w-0 flex-1 whitespace-pre-wrap border border-transparent px-1 py-1 text-sm [overflow-wrap:anywhere]">
                                    {node.name}
                                  </span>
                                )}
                                {scope === "personal" &&
                                  (creating ? (
                                    <label className="w-14">
                                      <span className="sr-only">單位</span>
                                      <Input
                                        className={inlineField}
                                        placeholder="單位"
                                        maxLength={32}
                                        value={node.unit ?? ""}
                                        onChange={(event) =>
                                          editDraftNode(node.id, { unit: event.target.value })
                                        }
                                      />
                                    </label>
                                  ) : (
                                    <span className="w-14 border border-transparent px-1 py-1 text-sm text-muted-foreground">
                                      {node.unit}
                                    </span>
                                  ))}
                                {creating ? (
                                  <Button
                                    type="button"
                                    size="icon-sm"
                                    variant="ghost"
                                    title="新增子項"
                                    aria-label={`在${node.name || "此物件"}下新增子項`}
                                    disabled={draft.nodes.length >= 100}
                                    onClick={() => addChild(node.id)}
                                  >
                                    <Plus />
                                  </Button>
                                ) : (
                                  <span className="size-8 shrink-0" />
                                )}

                                {creating && node.parentId ? (
                                  <CollapsibleTrigger asChild>
                                    <Button
                                      type="button"
                                      variant="ghost"
                                      size="icon-sm"
                                      aria-label="工項設定"
                                      title="工項設定"
                                    >
                                      <Ellipsis />
                                    </Button>
                                  </CollapsibleTrigger>
                                ) : (
                                  <span className="size-8 shrink-0" />
                                )}
                              </div>
                              {candidate && review && (
                                <Evidence candidate={candidate} source={review.source} />
                              )}
                              {nameTooLong(node.name) && (
                                <p role="alert">此工項名稱超過 120 字，請先修正。</p>
                              )}
                              {creating && node.parentId && (
                                <CollapsibleContent className="grid gap-3 py-2">
                                  {node.parentId && (
                                    <FormField>
                                      移至其他父項
                                      <NativeSelect
                                        value={node.parentId}
                                        onChange={(event) =>
                                          editDraftNode(node.id, { parentId: event.target.value })
                                        }
                                      >
                                        {removeDraftBranch(draft.nodes, node.id).map((n) => (
                                          <NativeSelectOption key={n.id} value={n.id}>
                                            {n.name || "未命名節點"}
                                          </NativeSelectOption>
                                        ))}
                                      </NativeSelect>
                                    </FormField>
                                  )}
                                  {scope === "personal" && node.parentId && (
                                    <Button
                                      type="button"
                                      variant="outline"
                                      onClick={() =>
                                        setDraft((previous) => ({
                                          ...previous,
                                          nodes: removeDraftBranch(previous.nodes, node.id),
                                        }))
                                      }
                                    >
                                      移除工序（含子工序）
                                    </Button>
                                  )}
                                </CollapsibleContent>
                              )}
                            </Collapsible>
                          </TreeRow>
                        );
                      })}
                    </ul>

                    {creating ? (
                      <div className="sticky bottom-0 z-10 flex flex-wrap items-center gap-2 border-t bg-background py-3">
                        <Button
                          type="submit"
                          disabled={
                            draft.nodes.length > 100 ||
                            draftTooLong ||
                            draft.nodes.some((node) => !node.name.trim())
                          }
                        >
                          {draft.baseVersion ? `儲存為 v${draft.baseVersion + 1}` : "建立範本"}
                        </Button>
                        <Button
                          type="button"
                          variant="outline"
                          onClick={() =>
                            discard.navigate(() => {
                              setCreating(false);
                              setDraft(initial());
                              setReview(undefined);
                              if (!draft.templateId) setFolded([]);
                            })
                          }
                        >
                          返回範本庫
                        </Button>
                      </div>
                    ) : (
                      <div className="flex flex-wrap gap-2">
                        {snapshot && (
                          <Button
                            type="button"
                            disabled={busy || uncertain}
                            onClick={() => {
                              if (snapshot) setPlan(structuredClone(snapshot.nodes));
                            }}
                          >
                            使用範本
                          </Button>
                        )}
                        {standalone && snapshot && (
                          <Button
                            type="button"
                            variant="outline"
                            disabled={busy || uncertain || reading}
                            onClick={() => {
                              setDraft({
                                templateId: snapshot.id,
                                baseVersion: snapshot.version,
                                name: snapshot.name,
                                nodes: structuredClone(snapshot.nodes),
                              });
                              setCreating(true);
                            }}
                          >
                            編輯範本
                          </Button>
                        )}
                      </div>
                    )}
                  </FieldSet>
                </FieldGroup>
              </form>
            )}
            {uncertain && !busy && (
              <Button
                onClick={() => {
                  if (attempt.current) send(attempt.current);
                }}
              >
                重試原範本請求
              </Button>
            )}
            {!standalone && (
              <Button variant="outline" disabled={busy || uncertain} onClick={close}>
                關閉範本
              </Button>
            )}
          </section>
        </PagePanel>
      )}
      {discard.dialog}
    </>
  );
}
