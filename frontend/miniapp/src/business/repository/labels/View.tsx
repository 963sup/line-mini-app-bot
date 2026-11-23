import { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import { useRepositoryGuard } from "../NavigationGuard";
import { ReadFeedback } from "@/shared/ReadFeedback";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { rejectedAttempt, useLatestRequest, errorMessage } from "@/shared/http/lifecycle";
import { PagePanel } from "@/shared/PagePanel";
import { FormField } from "@/shared/FormField";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { FieldGroup } from "@/shared/ui/field";
import { Empty, EmptyHeader, EmptyTitle } from "@/shared/ui/empty";
import { Item, ItemGroup, ItemContent, ItemTitle } from "@/shared/ui/item";

type Label = components["schemas"]["Label"];
export function View({ repositoryId }: { repositoryId: string }) {
  const { user, request } = useSession();
  const { run: readRun, cancel: cancelRead } = useLatestRequest(),
    write = useLatestRequest();
  const [items, setItems] = useState<Label[]>([]),
    [cursor, setCursor] = useState("");
  const [name, setName] = useState(""),
    [status, setStatus] = useState(""),
    [listStatus, setListStatus] = useState("");
  const [loaded, setLoaded] = useState(false),
    [readBusy, setReadBusy] = useState(false),
    [busy, setBusy] = useState(false),
    [uncertain, setUncertain] = useState(false);
  const lock = useRef(false),
    attempt = useRef<{ key: string; name: string } | null>(null);
  const [editing, setEditing] = useState<Label | null>(null);
  const input = useRef<HTMLInputElement>(null);
  const guard = useRepositoryGuard(!!name, () => input.current?.focus());
  useEffect(() => {
    guard.block(busy || uncertain);
  }, [busy, uncertain]);
  const reconcile = useRef<{ id: string; name: string } | null>(null);
  const load = useCallback(
    (nextCursor = "") => {
      if (!user) return;
      setReadBusy(true);
      setLoaded(false);
      setCursor("");
      if (!nextCursor) setItems([]);
      setListStatus("");
      void readRun(
        (signal) =>
          request(
            (client, s) =>
              client.GET("/api/v1/repositories/{repositoryId}/labels", {
                signal: s,
                params: { path: { repositoryId }, query: { cursor: nextCursor || undefined } },
              }),
            { signal },
          ),
        (page) => {
          const current =
            reconcile.current && page.items.find((item) => item.id === reconcile.current?.id);
          if (current && !lock.current) {
            const confirmed = current.name === reconcile.current?.name;
            reconcile.current = null;
            setUncertain(false);
            setEditing(null);
            setName("");
            attempt.current = null;
            setStatus(confirmed ? "已確認標籤名稱。" : "已重新讀取標籤，目前名稱：" + current.name);
          }
          setItems((previous) => [
            ...new Map(
              [...(nextCursor ? previous : []), ...page.items].map((item) => [item.id, item]),
            ).values(),
          ]);
          setCursor(page.nextCursor);
          setReadBusy(false);
          setLoaded(true);
          setListStatus("");
        },
        (error) => {
          setItems([]);
          setReadBusy(false);
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
  function submit(event: FormEvent) {
    event.preventDefault();
    if (!user || lock.current || !guard.canWrite()) return;
    if (!uncertain || !attempt.current) {
      const value = name.trim();
      if (!value) {
        setStatus("請填寫標籤名稱。");
        return;
      }
      attempt.current = { key: crypto.randomUUID(), name: value };
    }
    const wasUnknown = uncertain;
    guard.block(true);
    const pending = attempt.current;
    lock.current = true;
    setBusy(true);
    setUncertain(true);
    setStatus(editing ? "正在更新名稱……" : "正在建立標籤……");
    void write.run(
      (signal) =>
        request(
          (client, s) =>
            editing
              ? client.PATCH("/api/v1/repositories/{repositoryId}/labels/{id}", {
                  signal: s,
                  params: { path: { repositoryId, id: editing.id } },
                  body: { name: pending.name, expectedName: editing.name },
                })
              : client.POST("/api/v1/repositories/{repositoryId}/labels", {
                  signal: s,
                  params: { path: { repositoryId }, header: { "Idempotency-Key": pending.key } },
                  body: { name: pending.name },
                }),
          { signal },
        ),
      () => {
        lock.current = false;
        setBusy(false);
        setUncertain(false);
        guard.block(false);
        attempt.current = null;
        setName("");
        setEditing(null);
        setStatus(editing ? "已更新標籤名稱。" : "已建立標籤。");
        load();
      },
      (error) => {
        lock.current = false;
        setBusy(false);
        const code = (error as { code?: string })?.code;
        if (
          !wasUnknown &&
          rejectedAttempt(error, ["name_conflict", "idempotency_conflict", "version_conflict"])
        ) {
          setUncertain(false);
          guard.block(false);
        }
        if (!wasUnknown && code === "version_conflict") {
          setEditing(null);
          setName("");
        }
        if (code === "name_conflict" || code === "version_conflict") {
          reconcile.current = wasUnknown && editing ? { id: editing.id, name: pending.name } : null;
          load();
        }
        setStatus(
          errorMessage(error) + " 結果未確認時請沿用原名稱重試，離頁或重新整理前請先確認目錄。",
        );
      },
    );
  }
  return (
    <PagePanel heading="標籤">
      <form onSubmit={submit}>
        <FieldGroup>
          <FormField>
            標籤名稱
            <Input
              ref={input}
              required
              maxLength={40}
              value={name}
              disabled={!user || busy || uncertain}
              onChange={(e) => setName(e.target.value)}
            />
          </FormField>
          <Button type="submit" disabled={!user || busy}>
            {uncertain
              ? editing
                ? "重試改名"
                : "重試原建立請求"
              : editing
                ? "儲存新名稱"
                : "建立標籤"}
          </Button>
          {editing && (
            <Button
              type="button"
              variant="outline"
              disabled={busy || uncertain}
              onClick={() => {
                setEditing(null);
                setName("");
              }}
            >
              取消改名
            </Button>
          )}
        </FieldGroup>
      </form>
      <Status role="status">{status}</Status>
      <ReadFeedback
        busy={readBusy}
        message={listStatus}
        label="正在讀取標籤……"
        onRetry={user && listStatus && !loaded && !busy ? () => load() : undefined}
      />
      <ItemGroup className="gap-3">
        {items.map((item) => (
          <Item key={item.id} variant="outline" role="listitem">
            <ItemContent>
              <ItemTitle>{item.name}</ItemTitle>
            </ItemContent>
            <Button
              variant="outline"
              disabled={!user || busy || uncertain || !!name}
              onClick={() => {
                setEditing(item);
                setName(item.name);
                input.current?.focus();
              }}
            >
              改名
            </Button>
          </Item>
        ))}
      </ItemGroup>
      {loaded && !items.length && (
        <Empty>
          <EmptyHeader>
            <EmptyTitle>目前沒有標籤</EmptyTitle>
          </EmptyHeader>
        </Empty>
      )}
      {cursor && (
        <Button variant="outline" disabled={readBusy || busy} onClick={() => load(cursor)}>
          載入更多標籤
        </Button>
      )}
    </PagePanel>
  );
}
