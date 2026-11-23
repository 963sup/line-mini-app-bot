import { flushSync } from "react-dom";
import { ChevronDown, ChevronRight, Folder, Plus } from "lucide-react";
import { isPositiveQuantity, isValidCompletedQuantity } from "@/business/object/quantity";
import {
  useCallback,
  useImperativeHandle,
  useEffect,
  useRef,
  useState,
  type Ref,
  type FormEvent,
} from "react";
import { useRepositoryGuard } from "../repository/NavigationGuard";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { useDiscardChanges } from "@/shared/useDiscardChanges";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import {
  rejectedAttempt,
  errorMessage,
  isTemporaryReadError,
  useLatestRequest,
} from "@/shared/http/lifecycle";
import { FormField } from "@/shared/FormField";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { Collapsible, CollapsibleTrigger, CollapsibleContent } from "@/shared/ui/collapsible";
import { Item, ItemGroup, ItemContent, ItemTitle, ItemActions, ItemFooter } from "@/shared/ui/item";
import { Input } from "@/shared/ui/input";
import { FieldGroup, FieldSet } from "@/shared/ui/field";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";

type Resource = components["schemas"]["Object"];
type Creation = {
  name: string;
  type: "work" | "non_work";
  quantity: string;
  unit: string;
  completedQuantity?: string;
};
const initial: Creation = {
  name: "",
  type: "work",
  quantity: "1",
  unit: "SET",
  completedQuantity: "0",
};
const states = { not_started: "未開始", in_progress: "進行中", completed: "已完成" };
type BranchEditor = { add: (trigger: HTMLElement) => void };

export function useChildCreation(setExpanded: (id: string) => void) {
  const editor = useRef<BranchEditor>(null);
  function add(id: string, trigger: HTMLElement) {
    // Mount the correct branch before asking its existing editor to open.
    flushSync(() => setExpanded(id));
    editor.current?.add(trigger);
  }
  return { editor, add };
}

export function AddChildButton({
  name,
  disabled,
  onClick,
}: {
  name: string;
  disabled: boolean;
  onClick: (trigger: HTMLElement) => void;
}) {
  const label = `在「${name}」下新增子物件`;
  return (
    <Button
      type="button"
      variant="ghost"
      size="icon"
      title={label}
      aria-label={label}
      disabled={disabled}
      onClick={(event) => onClick(event.currentTarget)}
    >
      <Plus />
    </Button>
  );
}

type Props = {
  editorRef: Ref<BranchEditor>;
  repositoryId: string;
  objectId: string;
  canEdit: boolean;
  parentEditing: boolean;
  onEdit: () => void;
  onBlocked: (value: boolean) => void;
  canWrite: () => boolean;
};

const emptyChildrenMessage = "尚無子物件。";

export function View(props: Props) {
  const { user, epoch } = useSession();
  return (
    <Content key={`${props.repositoryId}:${props.objectId}:${user?.userId}:${epoch}`} {...props} />
  );
}
function Content({
  editorRef,
  repositoryId,
  objectId,
  canEdit,
  parentEditing,
  onEdit,
  onBlocked,
  canWrite,
}: Props) {
  const { request } = useSession(),
    read = useLatestRequest(),
    write = useLatestRequest();
  const [expanded, setExpanded] = useState("");
  const childCreation = useChildCreation(setExpanded);
  useImperativeHandle(editorRef, () => ({
    add(trigger) {
      if (!branchBlocked) edit(null, trigger);
    },
  }));
  const [branchBlocked, setBranchBlocked] = useState(false);
  const [items, setItems] = useState<Resource[]>([]),
    [cursor, setCursor] = useState("");
  const [failedPage, setFailedPage] = useState<string | null>(null);
  const [loading, setLoading] = useState(false),
    [readStatus, setReadStatus] = useState("");
  const [selected, setSelected] = useState<Resource | null>(null),
    [editing, setEditing] = useState(false),
    [content, setContent] = useState<Creation>(initial);
  const [busy, setBusy] = useState(false),
    [uncertain, setUncertain] = useState(false),
    [refresh, setRefresh] = useState(false),
    [status, setStatus] = useState("");
  const lock = useRef(false),
    unknown = useRef(false),
    navigation = useRef(false),
    pending = useRef<{ key: string; body: Creation } | null>(null);
  const baseline = useRef<Creation>(initial);
  const dirty = editing && JSON.stringify(content) !== JSON.stringify(baseline.current);
  const nameInput = useRef<HTMLInputElement>(null),
    editorPanel = useRef<HTMLFormElement>(null),
    opener = useRef<HTMLElement | null>(null);
  function focusEditor() {
    if (nameInput.current && !nameInput.current.matches(":disabled")) nameInput.current.focus();
    else editorPanel.current?.focus();
  }
  const discard = useDiscardChanges(dirty, focusEditor);
  const repositoryGuard = useRepositoryGuard(dirty, focusEditor);
  useEffect(() => {
    repositoryGuard.block(busy || uncertain);
  }, [busy, uncertain]);
  useEffect(() => {
    if (!editing) {
      opener.current?.focus();
      return;
    }
    if (!busy) focusEditor();
  }, [editing, busy]);
  useEffect(() => {
    if (parentEditing) {
      opener.current = null;
      setEditing(false);
      setSelected(null);
      baseline.current = initial;
      setContent(initial);
    }
  }, [parentEditing]);
  const block = (value: boolean) => {
    repositoryGuard.block(value);
    unknown.current = value;
    setUncertain(value);
    onBlocked(value);
  };
  useEffect(() => {
    navigation.current = false;
  });
  useEffect(() => () => onBlocked(false), [onBlocked]);
  useEffect(() => {
    onBlocked(busy || uncertain || dirty || branchBlocked);
  }, [busy, uncertain, dirty, branchBlocked, onBlocked]);
  const load = useCallback(
    (next = "") => {
      if (lock.current || unknown.current) return;
      setLoading(true);
      setFailedPage(null);
      setReadStatus("");
      void read.run(
        (signal) =>
          request(
            (client, s) =>
              client.GET("/api/v1/repositories/{repositoryId}/objects", {
                signal: s,
                params: {
                  path: { repositoryId },
                  query: { parentId: objectId, cursor: next || undefined },
                },
              }),
            { signal },
          ),
        (page) => {
          const children = page.items;
          setItems((previous) => (next ? [...previous, ...children] : children));
          setCursor(page.nextCursor);
          setLoading(false);
          setReadStatus(!next && !page.nextCursor && !children.length ? emptyChildrenMessage : "");
        },
        (error) => {
          if (next && isTemporaryReadError(error)) {
            setFailedPage(next);
          } else {
            setItems([]);
            setCursor("");
            setExpanded("");
            setFailedPage("");
          }
          setLoading(false);
          setReadStatus(errorMessage(error));
        },
      );
    },
    [repositoryId, objectId, request, read.run],
  );
  useEffect(() => {
    load();
    return read.cancel;
  }, [load, read.cancel]);
  function close() {
    if (lock.current || unknown.current) return;
    discard.navigate(() => {
      navigation.current = true;
      setEditing(false);
      setSelected(null);
      baseline.current = initial;
      setContent(initial);
      setStatus("");
    });
  }
  function edit(item: Resource | null, trigger: HTMLElement) {
    if (
      !canEdit ||
      (item && !item.canEdit) ||
      lock.current ||
      unknown.current ||
      !repositoryGuard.canWrite() ||
      !canWrite()
    )
      return;
    discard.navigate(() => {
      opener.current = trigger;
      onEdit();
      if (item) {
        read.cancel();
        setLoading(false);
      }
      navigation.current = true;
      setSelected(item);
      setEditing(true);
      setRefresh(false);
      setStatus("");
      pending.current = null;
      baseline.current = item
        ? {
            name: item.name,
            type: item.type === "non_work" ? "non_work" : "work",
            quantity: item.quantity || "",
            unit: item.unit || "",
            completedQuantity: item.completedQuantity,
          }
        : { ...initial };
      setContent(baseline.current);
      if (item) reloadDetail(item);
    });
  }
  function reloadDetail(target = selected) {
    if (!target || lock.current || !repositoryGuard.canWrite() || !canWrite()) return;
    lock.current = true;
    setBusy(true);
    repositoryGuard.block(true);
    onBlocked(true);
    void write.run(
      (signal) =>
        request(
          (client, s) =>
            client.GET("/api/v1/repositories/{repositoryId}/objects/{id}", {
              signal: s,
              params: { path: { repositoryId, id: target.id } },
            }),
          { signal },
        ),
      (item) => {
        lock.current = false;
        setBusy(false);
        block(false);
        setSelected(item);
        baseline.current = {
          name: item.name,
          type: item.type === "non_work" ? "non_work" : "work",
          quantity: item.quantity || "",
          unit: item.unit || "",
          completedQuantity: item.completedQuantity,
        };
        setContent(baseline.current);
        setRefresh(false);
        setStatus("");
        setItems((previous) => previous.map((value) => (value.id === item.id ? item : value)));
      },
      (error) => {
        lock.current = false;
        setBusy(false);
        onBlocked(unknown.current);
        setRefresh(true);
        setStatus(errorMessage(error));
      },
    );
  }
  const editable = canEdit && (!selected || selected.canEdit);
  function submit(event: FormEvent) {
    event.preventDefault();
    if (
      discard.isConfirming() ||
      !editable ||
      lock.current ||
      navigation.current ||
      refresh ||
      !repositoryGuard.canWrite() ||
      !canWrite()
    )
      return;
    let body =
      pending.current && unknown.current
        ? pending.current.body
        : {
            ...content,
            name: content.name.trim(),
            quantity: content.quantity.trim(),
            unit: content.unit.trim(),
            completedQuantity: content.completedQuantity?.trim(),
          };
    if (!body.name || !body.unit || !isPositiveQuantity(body.quantity)) {
      setStatus("請填寫名稱、單位及大於零的數量（最多18位整數及6位小數）。");
      return;
    }
    if (body.type === "work") {
      if (!isValidCompletedQuantity(body.completedQuantity, body.quantity)) {
        setStatus("完成數須介於0與子物件數量之間（最多18位整數及6位小數）。");
        return;
      }
    } else {
      body = { ...body };
      delete body.completedQuantity;
    }
    if (selected?.requiresProgressReset && body.type === "non_work") {
      setStatus("請先將完成數改為0並儲存，再改為非施作。");
      return;
    }
    const wasUnknown = unknown.current;
    lock.current = true;
    setBusy(true);
    block(true);
    setStatus(selected ? "正在更新子物件……" : "正在建立子物件……");
    if (!selected && !pending.current) pending.current = { key: crypto.randomUUID(), body };
    void write.run(
      (signal) =>
        selected
          ? request(
              (client, s) =>
                client.PATCH("/api/v1/repositories/{repositoryId}/objects/{id}", {
                  signal: s,
                  params: { path: { repositoryId, id: selected.id } },
                  body: {
                    ...body,
                    description: selected.description || "",
                    expectedVersion: selected.version,
                  },
                }),
              { signal },
            )
          : request(
              (client, s) =>
                client.POST("/api/v1/repositories/{repositoryId}/objects", {
                  signal: s,
                  params: {
                    path: { repositoryId },
                    header: { "Idempotency-Key": pending.current!.key },
                  },
                  body: { ...pending.current!.body, parentId: objectId },
                }),
              { signal },
            ),
      (item) => {
        lock.current = false;
        setBusy(false);
        block(false);
        pending.current = null;
        setEditing(false);
        setSelected(null);
        baseline.current = initial;
        setContent(initial);
        setItems((previous) =>
          previous.some((value) => value.id === item.id)
            ? previous.map((value) => (value.id === item.id ? item : value))
            : [...previous, item],
        );
        setReadStatus("已儲存子物件；父物件進度不變。");
      },
      (error) => {
        lock.current = false;
        setBusy(false);
        const rejected = rejectedAttempt(error, ["idempotency_conflict", "version_conflict"]);
        if (selected) {
          setRefresh(true);
          if (rejected) block(false);
          setStatus(errorMessage(error) + " 請重新讀取子物件以確認更新結果。");
        } else {
          if (rejected && !wasUnknown) {
            block(false);
            pending.current = null;
          }
          setStatus(errorMessage(error) + " 結果未確認時，請重試原子物件建立請求。");
        }
      },
    );
  }
  const editor = (
    <form
      ref={editorPanel}
      tabIndex={-1}
      className="w-full min-w-0"
      aria-label="子物件編輯"
      onSubmit={submit}
    >
      <FieldGroup>
        <Status role="status">{status}</Status>
        {selected && refresh && (
          <Button
            type="button"
            variant="outline"
            disabled={busy}
            onClick={() => discard.navigate(reloadDetail)}
          >
            重新讀取子物件
          </Button>
        )}
        <FieldSet
          className="grid min-w-0 gap-3"
          disabled={!editable || busy || uncertain || refresh}
        >
          <FormField>
            子物件名稱
            <Input
              ref={nameInput}
              required
              maxLength={120}
              value={content.name}
              onChange={(event) =>
                setContent((previous) => ({ ...previous, name: event.target.value }))
              }
            />
          </FormField>
          <FormField>
            子物件類型
            <NativeSelect
              value={content.type}
              onChange={(event) =>
                setContent((previous) => ({
                  ...previous,
                  type: event.target.value as Creation["type"],
                  completedQuantity:
                    previous.completedQuantity ?? (event.target.value === "work" ? "0" : undefined),
                }))
              }
            >
              <NativeSelectOption value="work">施作</NativeSelectOption>
              <NativeSelectOption value="non_work">非施作</NativeSelectOption>
            </NativeSelect>
          </FormField>
          <FormField>
            子物件數量
            <Input
              required
              maxLength={25}
              inputMode="decimal"
              value={content.quantity}
              onChange={(event) =>
                setContent((previous) => ({ ...previous, quantity: event.target.value }))
              }
            />
          </FormField>
          <FormField>
            子物件單位
            <Input
              required
              maxLength={32}
              value={content.unit}
              onChange={(event) =>
                setContent((previous) => ({ ...previous, unit: event.target.value }))
              }
            />
          </FormField>
          {content.type === "work" ? (
            <FormField>
              子物件完成數
              <Input
                required
                maxLength={25}
                inputMode="decimal"
                value={content.completedQuantity || ""}
                onChange={(event) =>
                  setContent((previous) => ({ ...previous, completedQuantity: event.target.value }))
                }
              />
            </FormField>
          ) : (
            <p>狀態與進度：不適用</p>
          )}
        </FieldSet>
        {editable && (
          <Button type="submit" disabled={busy || refresh}>
            {selected ? "儲存子物件" : uncertain ? "重試原子物件建立請求" : "建立子物件"}
          </Button>
        )}
        <Button type="button" variant="outline" disabled={busy || uncertain} onClick={close}>
          {editable ? "取消" : "關閉"}
        </Button>
      </FieldGroup>
    </form>
  );
  return (
    <section aria-label="子物件" className="grid min-w-0 gap-3">
      {editing && !selected && editor}
      <ReadFeedback
        plain={failedPage === null}
        busy={loading}
        message={readStatus}
        label="正在讀取子物件……"
        onRetry={
          failedPage !== null && !busy && !uncertain && !editing && !branchBlocked
            ? () => load(failedPage)
            : undefined
        }
      />
      <ItemGroup className="min-w-0 gap-1">
        {items.map((item) => (
          <Collapsible
            key={item.id}
            open={expanded === item.id}
            onOpenChange={(value) => setExpanded(value ? item.id : "")}
          >
            <Item role="listitem" className="min-w-0 border-0 px-0 py-1">
              <ItemContent className="min-w-0">
                <ItemTitle className="w-full min-w-0 max-w-full">
                  <CollapsibleTrigger asChild>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      className="min-w-0 flex-1 justify-start font-normal"
                      aria-label={`子物件：${item.name}`}
                      aria-expanded={expanded === item.id}
                      disabled={busy || uncertain || editing || parentEditing || branchBlocked}
                    >
                      {expanded === item.id ? <ChevronDown /> : <ChevronRight />}
                      <Folder />
                      <span className="truncate">{item.name}</span>
                    </Button>
                  </CollapsibleTrigger>
                  {item.canEdit && canEdit ? (
                    <Button
                      variant="link"
                      aria-label={item.name}
                      className="h-auto min-w-0 max-w-full whitespace-normal text-left [overflow-wrap:anywhere]"
                      disabled={busy || uncertain || branchBlocked}
                      onClick={(event) => edit(item, event.currentTarget)}
                    >
                      詳情
                    </Button>
                  ) : null}
                  {canEdit && item.canEdit && (
                    <AddChildButton
                      name={item.name}
                      disabled={busy || uncertain || editing || parentEditing || branchBlocked}
                      onClick={(trigger) => childCreation.add(item.id, trigger)}
                    />
                  )}
                </ItemTitle>
                <p className="whitespace-normal [overflow-wrap:anywhere]">
                  {item.quantity} {item.unit} ·{" "}
                  {item.type === "work"
                    ? `施作 · ${states[item.status]} · ${item.completedQuantity}/${item.quantity} ${item.unit}`
                    : "非施作 · 狀態與進度不適用"}
                </p>
              </ItemContent>
              {item.canEdit && canEdit && (
                <ItemActions>
                  <Button
                    variant="outline"
                    aria-label={`編輯子物件：${item.name}`}
                    disabled={busy || uncertain || branchBlocked}
                    onClick={(event) => edit(item, event.currentTarget)}
                  >
                    編輯
                  </Button>
                </ItemActions>
              )}
              {editing && selected?.id === item.id && <ItemFooter>{editor}</ItemFooter>}
              {expanded === item.id && !editing && !parentEditing && (
                <CollapsibleContent className="w-full min-w-0 ps-4">
                  <View
                    editorRef={childCreation.editor}
                    repositoryId={repositoryId}
                    objectId={item.id}
                    canEdit={item.canEdit}
                    parentEditing={false}
                    onEdit={onEdit}
                    onBlocked={setBranchBlocked}
                    canWrite={canWrite}
                  />
                </CollapsibleContent>
              )}
            </Item>
          </Collapsible>
        ))}
      </ItemGroup>
      {cursor && failedPage === null && (
        <Button
          disabled={loading || busy || uncertain || editing || branchBlocked}
          onClick={() => load(cursor)}
        >
          載入更多子物件
        </Button>
      )}
      {discard.dialog}
    </section>
  );
}
