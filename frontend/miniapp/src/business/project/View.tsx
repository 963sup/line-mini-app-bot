import {
  useCallback,
  useEffect,
  useImperativeHandle,
  type Ref,
  useRef,
  useState,
  type FormEvent,
} from "react";
import { ChevronRight } from "lucide-react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import {
  isTemporaryReadError,
  rejectedAttempt,
  useLatestRequest,
  errorMessage,
} from "@/shared/http/lifecycle";
import { useDiscardChanges } from "@/shared/useDiscardChanges";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { PagePanel } from "@/shared/PagePanel";
import { FormField } from "@/shared/FormField";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Textarea } from "@/shared/ui/textarea";
import { Field, FieldGroup, FieldLabel, FieldSet } from "@/shared/ui/field";
import { SettingsEditor } from "./SettingsEditor";
import { Items } from "./Items";
import { SourcePicker } from "./SourcePicker";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from "@/shared/ui/empty";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";
import { Item, ItemGroup } from "@/shared/ui/item";

type Project = components["schemas"]["Project"];
type Creation = components["schemas"]["OwnerProjectCreation"];

export type EditorNavigation = {
  navigate: (action: () => void) => void;
  allowHistoryChange: () => boolean;
};

export function View({
  ref,
  ownerId,
  projectId,
  initialSource,
  copyFrom,
  destinations,
  onCopy,
  onSelect,
  onBlocked,
  canWrite,
}: {
  ref?: Ref<EditorNavigation>;
  ownerId: string;
  projectId: string;
  initialSource: string;
  copyFrom?: Creation["copyFrom"];
  destinations: { id: string; name: string }[];
  onCopy: (destination: string) => void;
  onSelect: (id: string) => void;
  onBlocked: (value: boolean) => void;
  canWrite: () => boolean;
}) {
  const { user, request } = useSession();
  const { run: readRun, cancel: cancelRead } = useLatestRequest();
  const detailRequest = useLatestRequest(),
    write = useLatestRequest();
  const [items, setItems] = useState<Project[]>([]),
    [cursor, setCursor] = useState("");
  const [listBusy, setListBusy] = useState(false),
    [loaded, setLoaded] = useState(false);
  const [retryCursor, setRetryCursor] = useState<string | null>(null);
  const [listStatus, setListStatus] = useState("");
  const [mode, setMode] = useState<"create" | "detail" | null>(null);
  const [detail, setSelected] = useState<Project | null>(null);
  const selected = detail?.id === projectId ? detail : null;
  const [title, setTitle] = useState(""),
    [body, setBody] = useState("");
  const [settingsDraft, setSettingsDraft] = useState<Project["settings"]>();
  const [prefilledSource, setPrefilledSource] = useState("");
  const [chosenSource, setChosenSource] = useState<string>();
  const sourceRepositoryId = chosenSource ?? prefilledSource;
  const [busy, setBusy] = useState(false),
    [detailBusy, setDetailBusy] = useState(false);
  const [uncertain, setUncertain] = useState(false),
    [needsRefresh, setNeedsRefresh] = useState(false);
  const [status, setStatus] = useState("");
  const lock = useRef(false),
    unknown = useRef(false),
    referenceLock = useRef(false),
    navigation = useRef(false);
  const [referencesBlocked, setReferencesBlocked] = useState(false);
  const referenceBlock = useCallback(
    (value: boolean) => {
      referenceLock.current = value;
      setReferencesBlocked(value);
      onBlocked(value || lock.current || unknown.current);
    },
    [onBlocked],
  );
  const block = useCallback(
    (value: boolean) => {
      unknown.current = value;
      onBlocked(value || referenceLock.current);
    },
    [onBlocked],
  );
  const editorHeading = useRef<HTMLSpanElement>(null);
  const dirty =
    mode === "create"
      ? !!(title || body || sourceRepositoryId !== prefilledSource)
      : !!selected && (!!settingsDraft || title !== selected.title || body !== selected.body);
  const discard = useDiscardChanges(dirty, () => editorHeading.current?.focus());
  useImperativeHandle(ref, () => ({
    navigate: discard.navigate,
    allowHistoryChange: () =>
      !discard.isConfirming() && (!dirty || window.confirm("放棄未儲存的修改？")),
  }));
  useEffect(() => {
    navigation.current = false;
  });
  useEffect(() => () => onBlocked(false), [onBlocked]);
  useEffect(() => {
    if (!dirty && !busy && !uncertain && !referencesBlocked) return;
    const warn = (event: BeforeUnloadEvent) => {
      event.preventDefault();
      event.returnValue = "";
    };
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty, busy, uncertain, referencesBlocked]);
  const attempt = useRef<{ key: string; body: Creation } | null>(null);
  const focusOnOpen = useRef(false);
  useEffect(() => {
    if (mode && focusOnOpen.current && editorHeading.current) {
      editorHeading.current.focus();
      focusOnOpen.current = false;
    }
  }, [mode, projectId]);
  const copyDestinations = destinations.filter((owner) => owner.id !== ownerId);
  const canEdit = selected?.canEdit === true;

  const load = useCallback(
    (nextCursor = "") => {
      if (!user || referenceLock.current) return;
      setRetryCursor(null);
      setListBusy(true);
      setLoaded(false);
      setCursor("");
      if (!nextCursor) setItems([]);
      setListStatus("正在讀取專案……");
      void readRun(
        (signal) =>
          request(
            (client, s) =>
              client.GET("/api/v1/owners/{ownerId}/projects", {
                signal: s,
                params: { path: { ownerId }, query: { cursor: nextCursor || undefined } },
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
          setListStatus(page.nextCursor ? "可繼續載入更多專案。" : "已讀取完畢。");
        },
        (error) => {
          if (isTemporaryReadError(error)) setRetryCursor(nextCursor);
          else setItems([]);
          setListBusy(false);
          setListStatus(errorMessage(error));
        },
      );
    },
    [user, request, ownerId, readRun],
  );
  useEffect(() => {
    load();
    return cancelRead;
  }, [load, cancelRead]);

  const show = useCallback(
    (item: Project) => {
      setSettingsDraft(undefined);
      setSelected(item);
      setTitle(item.title);
      setBody(item.body);
      setMode("detail");
      setNeedsRefresh(false);
      setUncertain(false);
      attempt.current = null;
      block(false);
    },
    [block],
  );
  const open = useCallback(
    (id: string) => {
      if (!user || lock.current || referenceLock.current) return;
      navigation.current = true;
      setSelected(null);
      setMode("detail");
      setDetailBusy(true);
      setStatus("正在核驗並讀取專案……");
      void detailRequest.run(
        (signal) =>
          request(
            (client, s) =>
              client.GET("/api/v1/owners/{ownerId}/projects/{id}", {
                signal: s,
                params: { path: { ownerId, id } },
              }),
            { signal },
          ),
        (item) => {
          show(item);
          setDetailBusy(false);
          setStatus("已讀取專案。");
        },
        (error) => {
          setDetailBusy(false);
          setStatus(errorMessage(error));
        },
      );
    },
    [user, request, ownerId, detailRequest.run, show, canWrite],
  );
  useEffect(() => {
    detailRequest.cancel();
    write.cancel();
    lock.current = false;
    block(false);
    setBusy(false);
    setDetailBusy(false);
    setSelected(null);
    setMode(null);
    setUncertain(false);
    setNeedsRefresh(false);
    setStatus("");
    attempt.current = null;
    setPrefilledSource("");
    setChosenSource(undefined);
    if (projectId === "new") {
      setMode("create");
      setTitle("");
      setBody("");
    } else if (projectId) {
      if (/^[0-9a-f]{32}$/.test(projectId)) open(projectId);
      else setStatus("專案連結無效。");
    }
    return () => {
      detailRequest.cancel();
      write.cancel();
    };
  }, [projectId, open, detailRequest.cancel, write.cancel, block]);
  function openCreateForm() {
    if (lock.current || unknown.current || referenceLock.current || !canWrite()) return;
    if (mode === "create") {
      editorHeading.current?.focus();
      return;
    }
    discard.navigate(() => {
      focusOnOpen.current = true;
      navigation.current = true;
      onSelect("new");
    });
  }
  function submit(event: FormEvent) {
    event.preventDefault();
    if (
      discard.isConfirming() ||
      !user ||
      lock.current ||
      referenceLock.current ||
      navigation.current ||
      needsRefresh ||
      !canWrite()
    )
      return;
    if (mode === "detail") {
      if (selected) save();
      return;
    }
    const content = {
      title: title.trim(),
      body: body.trim(),
      defaultRepositoryId: copyFrom ? undefined : sourceRepositoryId || undefined,
      copyFrom,
    };
    if (!content.title) {
      setStatus("請填寫專案標題。");
      return;
    }
    if (!uncertain || !attempt.current)
      attempt.current = { key: crypto.randomUUID(), body: content };
    const pending = attempt.current;
    const wasUnknown = unknown.current;
    lock.current = true;
    block(true);
    setBusy(true);
    setUncertain(true);
    setStatus("正在建立專案……");
    void write.run(
      (signal) =>
        request(
          (client, s) =>
            client.POST("/api/v1/owners/{ownerId}/projects", {
              signal: s,
              params: { path: { ownerId }, header: { "Idempotency-Key": pending.key } },
              body: pending.body,
            }),
          { signal },
        ),
      (item) => {
        lock.current = false;
        setBusy(false);
        show(item);
        onSelect(item.id);
        setStatus("已建立專案。");
        load();
      },
      (error) => {
        lock.current = false;
        setBusy(false);
        if (!wasUnknown && rejectedAttempt(error)) {
          setUncertain(false);
          block(false);
          setStatus(errorMessage(error));
        } else {
          setStatus(errorMessage(error) + " 結果未確認，請沿用原內容重試。");
        }
      },
    );
  }
  function save() {
    if (
      !user ||
      !selected ||
      !canEdit ||
      lock.current ||
      unknown.current ||
      referenceLock.current ||
      navigation.current ||
      needsRefresh ||
      !canWrite()
    )
      return;
    const id = selected.id;
    lock.current = true;
    block(true);
    setBusy(true);
    setStatus("正在更新專案……");
    void write.run(
      (signal) =>
        request(
          (client, s) =>
            client.PATCH("/api/v1/owners/{ownerId}/projects/{id}", {
              signal: s,
              params: { path: { ownerId, id } },
              body: {
                title: title.trim(),
                body: body.trim(),
                expectedVersion: selected.version,
                settings: settingsDraft,
              },
            }),
          { signal },
        ),
      (item) => {
        lock.current = false;
        setBusy(false);
        show(item);
        setStatus("已更新專案。");
        load();
      },
      (error) => {
        lock.current = false;
        setBusy(false);
        const rejected = rejectedAttempt(error, ["version_conflict", "conflict"]);
        if (rejected && (error as { code?: string })?.code === "invalid_input") {
          block(false);
          setStatus(errorMessage(error));
          return;
        }
        setNeedsRefresh(true);
        if (rejected) block(false);
        else {
          block(true);
          setUncertain(true);
        }
        setStatus(errorMessage(error) + " 請重新讀取專案以確認更新結果，再修改。");
      },
    );
  }
  return (
    <div className="flex min-w-0 flex-col gap-4">
      <PagePanel heading="專案" description="建立個人或組織專案，並整理相關的物件、問題與討論。">
        <div className="flex flex-wrap gap-2">
          <Button
            disabled={!user || busy || uncertain || referencesBlocked}
            onClick={openCreateForm}
          >
            {mode === "create" ? "繼續填寫專案" : "新增專案"}
          </Button>
        </div>
        {listBusy || retryCursor !== null ? (
          <ReadFeedback
            busy={listBusy}
            label="正在讀取專案……"
            message={listStatus}
            onRetry={
              retryCursor !== null && !busy && !uncertain && !referencesBlocked
                ? () => load(retryCursor)
                : undefined
            }
          />
        ) : loaded ? (
          <span role="status" className="sr-only">
            {listStatus}
          </span>
        ) : (
          <Status role="status">{listStatus}</Status>
        )}
        {!mode && <Status role="status">{status}</Status>}
        {items.length > 0 && (
          <ItemGroup className="gap-3" aria-label="專案列表">
            {items.map((item) => (
              <div key={item.id} role="listitem">
                <Item
                  asChild
                  variant="outline"
                  size="sm"
                  className="min-h-11 w-full cursor-pointer flex-nowrap justify-between text-left hover:bg-accent disabled:cursor-default disabled:opacity-50"
                >
                  <button
                    type="button"
                    aria-current={projectId === item.id ? "page" : undefined}
                    disabled={busy || detailBusy || uncertain || referencesBlocked}
                    onClick={() => {
                      if (lock.current || unknown.current || referenceLock.current || !canWrite())
                        return;
                      if (projectId === item.id) {
                        editorHeading.current?.focus();
                        return;
                      }
                      discard.navigate(() => {
                        focusOnOpen.current = true;
                        navigation.current = true;
                        onSelect(item.id);
                      });
                    }}
                  >
                    <span className="min-w-0 font-medium [overflow-wrap:anywhere]">
                      {item.title}
                    </span>
                    <ChevronRight
                      className="size-4 shrink-0 text-muted-foreground"
                      aria-hidden="true"
                    />
                  </button>
                </Item>
              </div>
            ))}
          </ItemGroup>
        )}
        {loaded && !items.length && (
          <Empty>
            <EmptyHeader>
              <EmptyTitle>{cursor ? "本頁沒有專案" : "目前沒有專案"}</EmptyTitle>
              <EmptyDescription>
                {cursor ? "請繼續載入下一頁。" : "可建立新的專案。"}
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
        {cursor && (
          <Button
            variant="outline"
            disabled={listBusy || busy || uncertain || referencesBlocked}
            onClick={() => load(cursor)}
          >
            載入更多專案
          </Button>
        )}
      </PagePanel>
      {mode && (
        <PagePanel
          heading={
            <span ref={editorHeading} tabIndex={-1} className="scroll-mt-20">
              {mode === "create" ? (copyFrom ? "建立專案副本" : "新增專案") : "專案詳情"}
            </span>
          }
          description={
            selected && <span className="block [overflow-wrap:anywhere]">{selected.title}</span>
          }
        >
          <Status role="status">{status}</Status>
          {mode === "detail" && /^[a-f0-9]{32}$/.test(projectId) && (
            <div className="flex flex-wrap items-center gap-2">
              {selected && copyDestinations.length > 0 && (
                <FormField
                  className="min-w-0 max-w-full"
                  labelClassName="[&_[data-slot=native-select-wrapper]]:max-w-full"
                >
                  複製到另一個擁有者
                  <NativeSelect
                    value=""
                    disabled={busy || detailBusy || uncertain || referencesBlocked}
                    onChange={(event) => {
                      const destination = event.target.value;
                      if (destination) discard.navigate(() => onCopy(destination));
                    }}
                  >
                    <NativeSelectOption value="">選擇副本目的地</NativeSelectOption>
                    {copyDestinations.map((owner) => (
                      <NativeSelectOption key={owner.id} value={owner.id}>
                        {owner.name}
                      </NativeSelectOption>
                    ))}
                  </NativeSelect>
                </FormField>
              )}
              <Button
                variant="outline"
                disabled={busy || detailBusy || referencesBlocked}
                onClick={() =>
                  needsRefresh ? open(projectId) : discard.navigate(() => open(projectId))
                }
              >
                重新讀取專案
              </Button>
            </div>
          )}
          {(mode === "create" || selected) && (
            <form onSubmit={submit}>
              <FieldGroup>
                <FieldSet
                  className="grid min-w-0 gap-4"
                  disabled={
                    busy ||
                    detailBusy ||
                    referencesBlocked ||
                    uncertain ||
                    needsRefresh ||
                    (mode === "detail" && !canEdit)
                  }
                >
                  <FormField>
                    標題
                    <Input
                      required
                      maxLength={120}
                      value={title}
                      onChange={(e) => setTitle(e.target.value)}
                    />
                  </FormField>
                  <Field>
                    <FieldLabel htmlFor="project-body">內容</FieldLabel>
                    <Textarea
                      id="project-body"
                      className="min-h-40"
                      maxLength={8000}
                      value={body}
                      onChange={(e) => setBody(e.target.value)}
                    />
                  </Field>
                  {selected && (
                    <SettingsEditor
                      editable={canEdit}
                      value={settingsDraft ?? selected.settings ?? { fields: [], views: [] }}
                      onChange={setSettingsDraft}
                    />
                  )}
                </FieldSet>
                {mode === "create" && !copyFrom && (
                  <SourcePicker
                    key={ownerId}
                    ownerId={ownerId}
                    initialSource={initialSource}
                    value={sourceRepositoryId}
                    onPrefill={setPrefilledSource}
                    onChange={setChosenSource}
                    disabled={busy || uncertain}
                  />
                )}

                {(mode === "create" || canEdit) && (
                  <div className="flex flex-wrap gap-2">
                    <Button type="submit" disabled={busy || needsRefresh || referencesBlocked}>
                      {mode === "create" ? (uncertain ? "重試原建立請求" : "建立專案") : "儲存內容"}
                    </Button>
                  </div>
                )}
              </FieldGroup>
            </form>
          )}
        </PagePanel>
      )}
      {selected && (
        <Items
          key={selected.id + ":" + selected.version}
          ownerId={ownerId}
          sourceRepositoryId={selected.defaultRepositoryId ?? ""}
          project={selected}
          disabled={busy || detailBusy || needsRefresh || dirty}
          canWrite={() =>
            !navigation.current &&
            !lock.current &&
            !unknown.current &&
            !detailBusy &&
            !needsRefresh &&
            !dirty &&
            !discard.isConfirming() &&
            canWrite()
          }
          onBlocked={referenceBlock}
          onChange={(item) => {
            show(item);
            load();
          }}
          onRefresh={() => open(selected.id)}
        />
      )}
      {discard.dialog}
    </div>
  );
}
