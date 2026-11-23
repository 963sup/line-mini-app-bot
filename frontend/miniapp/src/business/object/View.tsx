import { Candidates, type Selection } from "./Candidates";
import { isPositiveQuantity, isValidCompletedQuantity } from "@/business/object/quantity";
import React, { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import { ChevronDownIcon, ChevronRightIcon, Folder, Pencil, Plus } from "lucide-react";
import { useRepositoryGuard } from "../repository/NavigationGuard";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { useDiscardChanges } from "@/shared/useDiscardChanges";
import { AddChildButton, useChildCreation, View as ObjectChildren } from "./Children";
import { Collapsible, CollapsibleTrigger, CollapsibleContent } from "@/shared/ui/collapsible";
import { useSpaces } from "../repository/useSpaces";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import {
  rejectedAttempt,
  errorMessage,
  isTemporaryReadError,
  useLatestRequest,
} from "@/shared/http/lifecycle";
import { PagePanel } from "@/shared/PagePanel";
import { FormField } from "@/shared/FormField";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Textarea } from "@/shared/ui/textarea";
import { FieldError, FieldGroup, FieldSet } from "@/shared/ui/field";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";
import { Item, ItemGroup, ItemContent, ItemTitle, ItemFooter } from "@/shared/ui/item";

type Resource = components["schemas"]["Object"];
type Creation = components["schemas"]["ObjectCreation"];
type Props = { repositoryId: string; initialObjectId?: string; documentTool?: boolean };
const states = { not_started: "未開始", in_progress: "進行中", completed: "已完成" };
const empty: Creation = { name: "" };
const typeLabel = (type: string) =>
  !type
    ? "尚未分類"
    : type === "work"
      ? "施作"
      : type === "non_work"
        ? "非施作"
        : `${type}（待確認）`;
const statusLabel = (item: Resource) =>
  !item.type ? "尚未設定" : item.type === "non_work" ? "不適用" : states[item.status];
const progressLabel = (item: Resource) =>
  item.type === "non_work"
    ? "不適用"
    : item.type === "work" && item.completedQuantity !== undefined
      ? `${item.completedQuantity}/${item.quantity} ${item.unit}`
      : "待補／待確認";
const fields = [
  ["name", "物件名稱", 120],
  ["quantity", "數量", 25],
  ["unit", "單位", 32],
] as const;

export function View(props: Props) {
  const { epoch, user } = useSession();
  return <Content key={`${props.repositoryId}:${epoch}:${user?.userId || ""}`} {...props} />;
}

function Content({ repositoryId, initialObjectId, documentTool }: Props) {
  const { user, request } = useSession();
  const read = useLatestRequest(),
    detail = useLatestRequest(),
    write = useLatestRequest();
  const [items, setItems] = useState<Resource[]>([]),
    [cursor, setCursor] = useState("");
  const [listBusy, setListBusy] = useState(false),
    [loaded, setLoaded] = useState(false),
    [listStatus, setListStatus] = useState("");
  const [mode, setMode] = useState<"create" | "detail" | null>(null),
    [selected, setSelected] = useState<Resource | null>(null);
  const { spaces, spaceCursor, spacesBusy, spaceStatus, loadSpaces } = useSpaces(
    repositoryId,
    mode === "detail",
  );
  const [invalidField, setInvalidField] = useState<
    "name" | "unit" | "quantity" | "completedQuantity" | null
  >(null);
  const [content, setContent] = useState<Creation>(empty);
  const [review, setReview] = useState<Selection>();
  const [createdCandidates, setCreatedCandidates] = useState<Map<string, string>>(() => new Map());
  const baseline = useRef<Creation>(empty);
  const dirty = !!mode && JSON.stringify(content) !== JSON.stringify(baseline.current);
  const editField = useRef("name"),
    editorPanel = useRef<HTMLElement>(null),
    opener = useRef<HTMLElement | null>(null);
  function focusEditor() {
    const input = editorPanel.current?.querySelector<HTMLElement>(
      `[data-edit-field="${editField.current}"]`,
    );
    if (input && !input.matches(":disabled")) input.focus();
    else editorPanel.current?.focus();
  }
  const discard = useDiscardChanges(dirty, focusEditor);
  const repositoryGuard = useRepositoryGuard(dirty, focusEditor);
  const [activeId, setActiveId] = useState("");
  const [busy, setBusy] = useState(false),
    [detailBusy, setDetailBusy] = useState(false),
    [uncertain, setUncertain] = useState(false),
    [needsRefresh, setNeedsRefresh] = useState(false),
    [status, setStatus] = useState("");
  useEffect(() => {
    repositoryGuard.block(busy || uncertain || needsRefresh);
  }, [busy, uncertain, needsRefresh]);
  const [expanded, setExpanded] = useState("");
  const childCreation = useChildCreation(setExpanded);
  const [subBlocked, setSubBlocked] = useState(false);
  const subLock = useRef(false);
  const subBlock = useCallback((value: boolean) => {
    subLock.current = value;
    setSubBlocked(value);
  }, []);
  const lock = useRef(false),
    formNavigation = useRef(false),
    objectCreateAttempt = useRef<{ idempotencyKey: string; body: Creation } | null>(null);
  useEffect(() => {
    formNavigation.current = false;
  });
  const canSubWrite = useCallback(
    () =>
      !formNavigation.current &&
      !lock.current &&
      !busy &&
      !uncertain &&
      !listBusy &&
      !detailBusy &&
      !needsRefresh &&
      !dirty &&
      !discard.isConfirming() &&
      repositoryGuard.canWrite(),
    [repositoryGuard.canWrite, busy, uncertain, listBusy, detailBusy, needsRefresh, dirty],
  );
  const canEdit = selected?.canEdit === true;

  const load = useCallback(
    (next = "") => {
      if (!user || subLock.current) return;
      formNavigation.current = true;
      setListBusy(true);
      setLoaded(false);
      setCursor("");
      setListStatus("");
      void read.run(
        (signal) =>
          request(
            (client, s) =>
              client.GET("/api/v1/repositories/{repositoryId}/objects", {
                signal: s,
                params: { path: { repositoryId }, query: { cursor: next || undefined } },
              }),
            { signal },
          ),
        (page) => {
          setItems((previous) => [
            ...new Map(
              [...(next ? previous : []), ...page.items].map((item) => [item.id, item]),
            ).values(),
          ]);
          setCursor(page.nextCursor);
          setListBusy(false);
          setLoaded(true);
          setListStatus("");
        },
        (error) => {
          if (next && isTemporaryReadError(error)) {
            setCursor(next);
          } else {
            setItems([]);
            setExpanded("");
          }
          setListBusy(false);
          setListStatus(errorMessage(error));
        },
      );
    },
    [user, request, repositoryId, read.run],
  );
  useEffect(() => {
    if (!documentTool) load();
    return read.cancel;
  }, [documentTool, load, read.cancel]);
  function show(item: Resource) {
    setInvalidField(null);
    setSelected(item);
    setActiveId(item.id);
    baseline.current = {
      name: item.name,
      description: item.description,
      type: item.type,
      quantity: item.quantity,
      unit: item.unit,
      spaceId: item.spaceId || "",
      sourceDocument: item.sourceDocument || "",
      sourceRevision: item.sourceRevision || "",
      completedQuantity: item.completedQuantity,
    };
    setContent(baseline.current);
    setMode("detail");
    setNeedsRefresh(false);
    setUncertain(false);
    objectCreateAttempt.current = null;
    repositoryGuard.block(false);
  }
  useEffect(() => {
    if (!mode) {
      opener.current?.focus();
      return;
    }
    if (!detailBusy && !busy) focusEditor();
  }, [mode, detailBusy, busy]);
  function clearEditor() {
    setInvalidField(null);
    setReview(undefined);
    detail.cancel();
    setDetailBusy(false);
    setMode(null);
    setSelected(null);
    setActiveId("");
    baseline.current = empty;
    setContent(empty);
    setStatus("");
  }
  function close() {
    if (lock.current || uncertain || needsRefresh || subLock.current) return;
    discard.navigate(() => {
      formNavigation.current = true;
      clearEditor();
    });
  }
  function saved(item: Resource, message: string) {
    if (mode === "create" && review)
      setCreatedCandidates((previous) => new Map(previous).set(review.id, item.id));
    setReview(undefined);
    setItems((previous) =>
      previous.some((value) => value.id === item.id)
        ? previous.map((value) => (value.id === item.id ? item : value))
        : [item, ...previous],
    );
    setMode(null);
    setSelected(null);
    setActiveId("");
    baseline.current = empty;
    setContent(empty);
    setUncertain(false);
    setNeedsRefresh(false);
    objectCreateAttempt.current = null;
    repositoryGuard.block(false);
    setLoaded(true);
    setListStatus(message);
  }
  function open(id: string, trigger?: HTMLElement, field = "name") {
    if (
      !user ||
      subLock.current ||
      lock.current ||
      uncertain ||
      (needsRefresh && id !== activeId) ||
      formNavigation.current
    )
      return;
    discard.navigate(() => {
      editField.current = field;
      if (trigger) opener.current = trigger;
      formNavigation.current = true;
      read.cancel();
      setListBusy(false);
      setMode("detail");
      setActiveId(id);
      setSelected(null);
      baseline.current = empty;
      setContent(empty);
      setDetailBusy(true);
      setStatus("");
      void detail.run(
        (signal) =>
          request(
            (client, s) =>
              client.GET("/api/v1/repositories/{repositoryId}/objects/{id}", {
                signal: s,
                params: { path: { repositoryId, id } },
              }),
            { signal },
          ),
        (item) => {
          show(item);
          setDetailBusy(false);
          setStatus("");
        },
        (error) => {
          setDetailBusy(false);
          setStatus(errorMessage(error));
        },
      );
    });
  }
  useEffect(() => {
    if (initialObjectId) open(initialObjectId);
  }, [initialObjectId]);
  function openCreateForm(trigger: HTMLElement, selection?: Selection) {
    if (subLock.current || lock.current || uncertain || needsRefresh || formNavigation.current)
      return;
    discard.navigate(() => {
      opener.current = trigger;
      formNavigation.current = true;
      read.cancel();
      setListBusy(false);
      clearEditor();
      setReview(selection);
      const candidate = selection?.candidate;
      editField.current = "name";
      setMode("create");
      setContent(
        candidate
          ? {
              ...empty,
              name: candidate.name,
              description: candidate.quote,
              sourceDocument: selection!.filename,
              sourcePages: candidate.location ?? "",
            }
          : { ...empty },
      );
      setNeedsRefresh(false);
      objectCreateAttempt.current = null;
    });
  }
  function submit(event: FormEvent) {
    event.preventDefault();
    if (
      discard.isConfirming() ||
      !user ||
      subLock.current ||
      lock.current ||
      formNavigation.current ||
      needsRefresh ||
      !repositoryGuard.canWrite() ||
      (mode === "detail" && !canEdit)
    )
      return;
    const body =
      mode === "create"
        ? {
            name: content.name.trim(),
            ...(review
              ? {
                  description: content.description,
                  sourceDocument: content.sourceDocument,
                  sourcePages: content.sourcePages,
                }
              : {}),
          }
        : ({
            name: content.name,
            description: "",
            type: "",
            quantity: "",
            unit: "",
            ...Object.fromEntries(
              Object.entries(content).map(([key, value]) => [key, value?.trim() || ""]),
            ),
          } as Creation);
    setInvalidField(null);
    if (!body.name) {
      setStatus("");
      setInvalidField("name");
      return;
    }
    if (mode === "detail" && body.quantity && !isPositiveQuantity(body.quantity)) {
      setStatus("");
      setInvalidField("quantity");
      return;
    }
    if (mode === "detail" && body.type === "work" && (body.quantity || body.completedQuantity)) {
      if (!body.unit) {
        setStatus("");
        setInvalidField("unit");
        return;
      }
      if (!isValidCompletedQuantity(body.completedQuantity, body.quantity || "")) {
        setStatus("");
        setInvalidField("completedQuantity");
        return;
      }
    } else {
      delete body.completedQuantity;
    }
    if (selected?.requiresProgressReset && body.type === "non_work") {
      setStatus("請先將完成數改為0並儲存，再將類型改為非施作。");
      return;
    }
    if (mode === "detail" && selected) {
      lock.current = true;
      setBusy(true);
      repositoryGuard.block(true);
      setStatus("正在更新物件……");
      void write.run(
        (signal) =>
          request(
            (client, s) =>
              client.PATCH("/api/v1/repositories/{repositoryId}/objects/{id}", {
                signal: s,
                params: { path: { repositoryId, id: selected.id } },
                body: {
                  ...body,
                  description: body.description || "",
                  type: body.type || "",
                  quantity: body.quantity || "",
                  unit: body.unit || "",
                  expectedVersion: selected.version,
                },
              }),
            { signal },
          ),
        (item) => {
          lock.current = false;
          setBusy(false);
          saved(item, "已更新物件。");
        },
        (error) => {
          lock.current = false;
          setBusy(false);
          repositoryGuard.block(true);
          setNeedsRefresh(true);
          setStatus(errorMessage(error) + " 請重新讀取物件後再修改。");
        },
      );
      return;
    }
    if (!uncertain || !objectCreateAttempt.current)
      objectCreateAttempt.current = { idempotencyKey: crypto.randomUUID(), body };
    const attempt = objectCreateAttempt.current;
    const wasUnknown = uncertain;
    lock.current = true;
    setBusy(true);
    setUncertain(true);
    repositoryGuard.block(true);
    setStatus("正在建立物件……");
    void write.run(
      (signal) =>
        request(
          (client, s) =>
            client.POST("/api/v1/repositories/{repositoryId}/objects", {
              signal: s,
              params: {
                path: { repositoryId },
                header: { "Idempotency-Key": attempt.idempotencyKey },
              },
              body: attempt.body,
            }),
          { signal },
        ),
      (item) => {
        lock.current = false;
        setBusy(false);
        saved(item, "已建立物件。");
      },
      (error) => {
        lock.current = false;
        setBusy(false);
        if (!wasUnknown && rejectedAttempt(error)) {
          setUncertain(false);
          repositoryGuard.block(false);
        }
        setStatus(
          errorMessage(error) +
            " 結果未確認時請重試原建立請求，避免重複建立；離開頁面前請先確認紀錄。",
        );
      },
    );
  }
  const cancelButton = (
    <Button
      type="button"
      variant="ghost"
      disabled={busy || uncertain || needsRefresh || subBlocked}
      onClick={close}
    >
      {canEdit || mode === "create" ? "取消" : "關閉"}
    </Button>
  );
  const editor = (
    <section
      ref={editorPanel}
      tabIndex={-1}
      aria-label={mode === "create" ? "新增物件" : "物件編輯"}
      className="flex w-full min-w-0 flex-col items-start gap-3 rounded-lg border bg-background p-4 sm:max-w-3xl"
    >
      {mode === "create" && review && (
        <aside>
          <p>
            {review.filename} · {review.candidate.location ?? "位置待核對"}
          </p>
          <blockquote>{review.candidate.quote}</blockquote>
        </aside>
      )}
      <ReadFeedback busy={detailBusy} message={status} label="正在讀取物件……" />
      {mode === "detail" && (
        <Button
          variant="outline"
          disabled={busy || detailBusy || subBlocked}
          onClick={() => open(activeId)}
        >
          重新讀取物件
        </Button>
      )}
      {(mode === "create" || selected) && (
        <form className="w-full" onSubmit={submit} onChange={() => setInvalidField(null)}>
          <FieldGroup className="gap-3">
            <FieldSet
              className="grid min-w-0 gap-3 sm:grid-cols-2"
              disabled={
                busy ||
                detailBusy ||
                subBlocked ||
                uncertain ||
                needsRefresh ||
                (mode === "detail" && !canEdit)
              }
            >
              {mode === "detail" && (
                <FormField>
                  類型
                  <NativeSelect
                    data-edit-field="type"
                    value={content.type}
                    onChange={(event) => {
                      const type = event.target.value;
                      setContent((previous) => {
                        let completedQuantity = previous.completedQuantity;
                        if (type === "work" && completedQuantity === undefined) {
                          completedQuantity =
                            selected?.workConversionDefault === "quantity"
                              ? previous.quantity
                              : selected?.workConversionDefault === "zero"
                                ? "0"
                                : (selected?.completedQuantity ?? "");
                        }
                        return { ...previous, type, completedQuantity };
                      });
                    }}
                  >
                    {selected &&
                      selected.type &&
                      selected.type !== "work" &&
                      selected.type !== "non_work" && (
                        <NativeSelectOption value={selected.type}>
                          {typeLabel(selected.type)}
                        </NativeSelectOption>
                      )}
                    <NativeSelectOption value="">尚未分類</NativeSelectOption>
                    <NativeSelectOption value="work">施作</NativeSelectOption>
                    <NativeSelectOption value="non_work">非施作</NativeSelectOption>
                  </NativeSelect>
                </FormField>
              )}
              {fields
                .filter(([key]) => mode === "detail" || key === "name")
                .map(([key, label, max]) => {
                  const invalid = invalidField === key;
                  return (
                    <React.Fragment key={key}>
                      <FormField data-invalid={invalid || undefined}>
                        {label}
                        <Input
                          data-edit-field={key}
                          required={key === "name"}
                          maxLength={max}
                          inputMode={key === "quantity" ? "decimal" : undefined}
                          aria-invalid={invalid || undefined}
                          aria-describedby={invalid ? `object-${key}-error` : undefined}
                          value={content[key] || ""}
                          onChange={(event) =>
                            setContent((previous) => ({ ...previous, [key]: event.target.value }))
                          }
                        />
                      </FormField>
                      {invalid && (
                        <FieldError id={`object-${key}-error`}>
                          {key === "name"
                            ? "請填寫物件名稱。"
                            : key === "unit"
                              ? "追蹤進度時請補填單位。"
                              : "請填寫大於零的數量（最多18位整數及6位小數）。"}
                        </FieldError>
                      )}
                    </React.Fragment>
                  );
                })}
              {mode === "detail" && (
                <>
                  <FormField>
                    空間
                    <NativeSelect
                      value={content.spaceId || ""}
                      onChange={(event) =>
                        setContent((previous) => ({ ...previous, spaceId: event.target.value }))
                      }
                    >
                      <NativeSelectOption value="">未分配</NativeSelectOption>
                      {content.spaceId && !spaces.some((value) => value.id === content.spaceId) && (
                        <NativeSelectOption value={content.spaceId}>
                          目前空間（未載入）
                        </NativeSelectOption>
                      )}
                      {spaces.map((space) => (
                        <NativeSelectOption key={space.id} value={space.id}>
                          {space.name}
                        </NativeSelectOption>
                      ))}
                    </NativeSelect>
                  </FormField>
                  <ReadFeedback
                    busy={spacesBusy}
                    message={spaceStatus}
                    label="正在讀取空間……"
                    onRetry={spaceStatus ? () => loadSpaces(spaceCursor) : undefined}
                  />
                  {spaceCursor && !spaceStatus && (
                    <Button
                      type="button"
                      variant="outline"
                      disabled={spacesBusy}
                      onClick={() => loadSpaces(spaceCursor)}
                    >
                      載入更多空間
                    </Button>
                  )}
                </>
              )}
              {mode === "detail" && (
                <>
                  <FormField>
                    完整說明
                    <Textarea
                      maxLength={8000}
                      value={content.description}
                      onChange={(event) =>
                        setContent((previous) => ({ ...previous, description: event.target.value }))
                      }
                    />
                  </FormField>
                  {content.type === "work" && (
                    <FormField data-invalid={invalidField === "completedQuantity" || undefined}>
                      完成數
                      <Input
                        aria-invalid={invalidField === "completedQuantity" || undefined}
                        aria-describedby={
                          invalidField === "completedQuantity"
                            ? "object-completed-error"
                            : undefined
                        }
                        data-edit-field="completedQuantity"
                        required={!!content.quantity}
                        inputMode="decimal"
                        maxLength={25}
                        value={content.completedQuantity || ""}
                        onChange={(event) =>
                          setContent((previous) => ({
                            ...previous,
                            completedQuantity: event.target.value,
                          }))
                        }
                      />
                    </FormField>
                  )}
                  {invalidField === "completedQuantity" && (
                    <FieldError id="object-completed-error">
                      請填寫0至數量之間的完成數（最多18位整數及6位小數）。
                    </FieldError>
                  )}
                  <p>
                    狀態：
                    {content.type === "non_work"
                      ? "不適用"
                      : selected
                        ? statusLabel(selected)
                        : "未開始"}
                  </p>
                  <p className="whitespace-normal [overflow-wrap:anywhere]">
                    進度：
                    {content.type === "non_work"
                      ? "不適用"
                      : selected
                        ? progressLabel(selected)
                        : "儲存後依完成數更新"}
                  </p>
                </>
              )}
            </FieldSet>

            <div className="flex flex-wrap items-center gap-2">
              {(mode === "create" || canEdit) && (
                <Button
                  className="self-start"
                  type="submit"
                  disabled={busy || detailBusy || needsRefresh || subBlocked}
                >
                  {mode === "create" ? (uncertain ? "重試原物件建立請求" : "建立物件") : "儲存物件"}
                </Button>
              )}
              {cancelButton}
            </div>
          </FieldGroup>
        </form>
      )}
      {mode === "detail" && !selected && cancelButton}
    </section>
  );
  if (documentTool)
    return (
      <>
        <Candidates
          repositoryId={repositoryId}
          disabled={!user || !!mode || busy || uncertain || needsRefresh || subBlocked}
          created={createdCandidates}
          choose={(selection, trigger) => openCreateForm(trigger, selection)}
        />
        {mode === "create" && editor}
        <ReadFeedback busy={false} plain={loaded} message={listStatus} label="" />
        {discard.dialog}
      </>
    );
  return (
    <>
      <div className="flex w-full min-w-0 flex-col gap-4">
        <ReadFeedback
          busy={listBusy}
          plain={loaded}
          message={listStatus}
          label="正在讀取物件……"
          onRetry={
            user && listStatus && !loaded && !busy && !subBlocked && !uncertain && !mode
              ? () => load(cursor)
              : undefined
          }
        />
        {mode === "detail" && !items.some((item) => item.id === activeId) && (
          <PagePanel heading="物件詳情">{editor}</PagePanel>
        )}
        <ItemGroup className="w-full min-w-0 gap-0">
          {items.map((item) => (
            <Collapsible
              key={item.id}
              open={expanded === item.id}
              onOpenChange={(value) => {
                if (
                  subLock.current ||
                  lock.current ||
                  uncertain ||
                  formNavigation.current ||
                  discard.isConfirming()
                )
                  return;
                formNavigation.current = true;
                setExpanded(value ? item.id : "");
              }}
              className="w-full min-w-0"
            >
              <Item role="listitem" className="w-full min-w-0 border-0 px-0 py-0.5">
                <ItemContent className="w-full min-w-0">
                  <ItemTitle className="w-full min-w-0 max-w-full">
                    <CollapsibleTrigger asChild>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="min-w-0 flex-1 justify-start font-normal"
                        disabled={busy || detailBusy || uncertain || needsRefresh || subBlocked}
                        aria-label={`子物件：${item.name}`}
                        title={expanded === item.id ? "收合子物件" : "展開子物件"}
                      >
                        {expanded === item.id ? (
                          <ChevronDownIcon aria-hidden="true" />
                        ) : (
                          <ChevronRightIcon aria-hidden="true" />
                        )}
                        <Folder aria-hidden="true" />
                        <span className="truncate">{item.name}</span>
                      </Button>
                    </CollapsibleTrigger>
                    <Button
                      variant="link"
                      aria-label={item.name}
                      title={item.canEdit ? `編輯物件：${item.name}` : "物件詳情"}
                      className="h-auto min-w-0 max-w-full justify-start whitespace-normal text-left [overflow-wrap:anywhere]"
                      disabled={busy || detailBusy || uncertain || needsRefresh || subBlocked}
                      onClick={(event) => open(item.id, event.currentTarget)}
                    >
                      {item.canEdit ? <Pencil aria-hidden="true" className="size-3.5" /> : "詳情"}
                      <span className="sr-only">名稱：{item.name}</span>
                    </Button>
                    {item.canEdit && (
                      <AddChildButton
                        name={item.name}
                        disabled={
                          busy ||
                          detailBusy ||
                          uncertain ||
                          needsRefresh ||
                          !!mode ||
                          (subBlocked && expanded !== item.id)
                        }
                        onClick={(trigger) => childCreation.add(item.id, trigger)}
                      />
                    )}
                  </ItemTitle>
                  <div className="flex flex-wrap items-center gap-x-4 gap-y-1 ps-2 text-xs text-muted-foreground sm:ps-12">
                    {(
                      [
                        [
                          "數量",
                          item.quantity && `${item.quantity} ${item.unit ?? ""}`.trim(),
                          "quantity",
                        ],
                        ["類型", item.type && typeLabel(item.type), "type"],
                        ["狀態", item.type && statusLabel(item), ""],
                        [
                          "進度",
                          item.type && progressLabel(item),
                          item.type === "work" ? "completedQuantity" : "",
                        ],
                      ] as const
                    ).map(
                      ([label, value, field]) =>
                        value && (
                          <p key={label} className="min-w-0 [overflow-wrap:anywhere]">
                            {item.canEdit && field ? (
                              <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                className="h-auto whitespace-normal px-1 py-1 text-xs"
                                disabled={
                                  busy || detailBusy || uncertain || needsRefresh || subBlocked
                                }
                                onClick={(event) => open(item.id, event.currentTarget, field)}
                              >
                                {label}：{value}
                              </Button>
                            ) : (
                              <>
                                {label}：{value}
                              </>
                            )}
                          </p>
                        ),
                    )}
                  </div>
                </ItemContent>
                {mode === "detail" && activeId === item.id && <ItemFooter>{editor}</ItemFooter>}
              </Item>
              <CollapsibleContent className="min-w-0 ps-4 pt-1">
                {expanded === item.id && (
                  <ObjectChildren
                    editorRef={childCreation.editor}
                    repositoryId={repositoryId}
                    objectId={item.id}
                    canEdit={item.canEdit}
                    parentEditing={!!mode}
                    onEdit={() => {
                      opener.current = null;
                      clearEditor();
                    }}
                    onBlocked={subBlock}
                    canWrite={canSubWrite}
                  />
                )}
              </CollapsibleContent>
            </Collapsible>
          ))}
        </ItemGroup>
        {loaded && cursor && (
          <Button
            variant="outline"
            disabled={busy || listBusy || subBlocked || !!mode}
            onClick={() => load(cursor)}
          >
            載入更多物件
          </Button>
        )}
        {mode === "create" && editor}
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="self-start"
          title="新增物件"
          disabled={!user || busy || uncertain || needsRefresh || subBlocked || mode === "create"}
          onClick={(event) => openCreateForm(event.currentTarget)}
        >
          <Plus aria-hidden="true" />
          新增物件
        </Button>
      </div>
      {discard.dialog}
    </>
  );
}
