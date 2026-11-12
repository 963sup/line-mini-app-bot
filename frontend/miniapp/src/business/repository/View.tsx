import { PagePanel } from "@/shared/PagePanel";
import { Button } from "@/shared/ui/button";
import { Status } from "@/shared/Status";
import React, { useEffect, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "../../identity/Session";
import { errorMessage, useLatestRequest } from "../../shared/http/lifecycle";
import { displayTime } from "../../shared/time/time";

type Repository = components["schemas"]["Repository"];

export function View() {
  const { user, request } = useSession();
  const read = useLatestRequest();
  const [items, setItems] = useState<Repository[]>([]);
  const [selected, setSelected] = useState<Repository | null>(null);
  const [scopeId, setScopeId] = useState("");
  const [cursor, setCursor] = useState("");
  const [busy, setBusy] = useState(false);
  const [status, setStatus] = useState("請先完成身分驗證。");

  function load(next = "") {
    setBusy(true);
    setSelected(null);
    setStatus("正在讀取空間……");
    void read.run(
      (signal) =>
        request(
          (client, s) =>
            client.GET("/api/v1/repositories", {
              signal: s,
              params: { query: { cursor: next } },
            }),
          { signal },
        ),
      (page) => {
        setItems(page.items);
        setCursor(page.nextCursor);
        setBusy(false);
        setStatus(
          page.items.length
            ? "選擇空間查看資料。"
            : page.nextCursor
              ? "本頁沒有可存取的空間，可繼續下一頁。"
              : "沒有可顯示的空間；請先在 Bot 所在的群組互動後重新載入。",
        );
      },
      (error) => {
        setItems([]);
        setCursor("");
        setBusy(false);
        setStatus(errorMessage(error));
      },
    );
  }

  function detail(id: string) {
    setBusy(true);
    setSelected(null);
    setStatus("正在核驗空間存取資格……");
    void read.run(
      (signal) =>
        request(
          (client, s) =>
            client.GET("/api/v1/repositories/{id}", {
              signal: s,
              params: { path: { id } },
            }),
          { signal },
        ),
      (value) => {
        setSelected(value);
        setBusy(false);
        setStatus("已載入空間。");
      },
      (error) => {
        setItems([]);
        setCursor("");
        setBusy(false);
        setStatus(errorMessage(error));
      },
    );
  }

  useEffect(() => {
    if (user) {
      const params = new URLSearchParams(location.search);
      let id = params.get("id") || params.get("repository") || "";
      try {
        const nested = params.get("liff.state");
        if (nested) {
          const query = new URL(nested, location.origin).searchParams;
          id ||= query.get("id") || query.get("repository") || "";
        }
      } catch {}
      setScopeId(id);
      if (id) detail(id);
      else load();
    }
    return read.cancel;
  }, [user, request]);

  return (
    <PagePanel heading="空間">
      <a
        className="break-words text-primary underline underline-offset-4 focus-visible:outline-2 focus-visible:outline-ring"
        href="/miniapp/workspace-preview"
      >
        預覽新版空間 UI（示範資料）
      </a>
      <p className="text-sm leading-relaxed text-muted-foreground">
        {scopeId
          ? "此連結僅開啟原群組空間，需具備該群組的存取資格。"
          : "每個群組一個空間，由 Bot 管理。僅顯示已互動且目前仍有存取資格的群組。"}
      </p>
      <Button
        disabled={!user || busy}
        onClick={() => (scopeId ? detail(scopeId) : load())}
      >
        重新載入
      </Button>
      <Status role="status">{status}</Status>
      <ul className="my-4 flex flex-col gap-4 [&>li]:border-b [&>li]:pb-4 [&_h3]:font-semibold [&_p]:my-2">
        {items.map((item) => (
          <li key={item.id}>
            <Button disabled={busy} onClick={() => detail(item.id)}>
              {item.name || "群組空間"}
            </Button>
          </li>
        ))}
      </ul>
      {cursor && (
        <Button disabled={!user || busy} onClick={() => load(cursor)}>
          下一頁
        </Button>
      )}
      {selected && (
        <article>
          <h3 className="font-semibold">{selected.name || "群組空間"}</h3>
          <dl className="grid gap-2 text-sm [&>dt]:font-medium [&>dd]:break-all [&>dd]:text-muted-foreground">
            <dt>儲存庫 ID</dt>
            <dd>{selected.id}</dd>
            <dt>群組 ID</dt>
            <dd>{selected.groupId}</dd>
            <dt>狀態</dt>
            <dd>{selected.status === "ACTIVE" ? "啟用" : "停用"}</dd>
            <dt>版本</dt>
            <dd>{selected.version}</dd>
            <dt>建立時間</dt>
            <dd>{displayTime(selected.createdAt)}</dd>
            <dt>更新時間</dt>
            <dd>{displayTime(selected.updatedAt)}</dd>
          </dl>
          <a
            className="break-words text-primary underline underline-offset-4 focus-visible:outline-2 focus-visible:outline-ring"
            href={
              "/miniapp/history?" +
              new URLSearchParams({ group: selected.groupId })
            }
          >
            查看本人出勤
          </a>
          <p className="text-sm leading-relaxed text-muted-foreground">
            摘要請回原群組操作。
          </p>
        </article>
      )}
    </PagePanel>
  );
}
