import { useCallback, useState, type FormEvent } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Status } from "@/shared/Status";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";

type Item = components["schemas"]["ExploreItem"];

export function View() {
  const { user, request } = useSession();
  const [input, setInput] = useState("");
  const [query, setQuery] = useState("");
  const [items, setItems] = useState<Item[]>([]);
  const [cursor, setCursor] = useState("");
  const [busy, setBusy] = useState(false);
  const [status, setStatus] = useState("");
  const { run } = useLatestRequest();

  const load = useCallback(
    (name: string, next = "") => {
      if (!user) return;
      setBusy(true);
      setStatus("正在搜尋……");
      void run(
        (signal) =>
          request(
            (client, requestSignal) =>
              client.GET("/api/v1/explore", {
                signal: requestSignal,
                params: { query: { type: "repository", q: name, cursor: next || undefined } },
              }),
            { signal },
          ),
        (page) => {
          setItems((previous) => {
            const values = next ? [...previous, ...page.items] : page.items;
            return [...new Map(values.map((item) => [item.id, item])).values()];
          });
          setCursor(page.nextCursor);
          setBusy(false);
          setStatus(
            page.items.length
              ? ""
              : page.nextCursor
                ? "本頁沒有符合項目，可繼續載入。"
                : next
                  ? "已載入全部結果。"
                  : "找不到結果。",
          );
        },
        (error) => {
          setBusy(false);
          setStatus(errorMessage(error));
        },
      );
    },
    [request, run, user],
  );

  function submit(event: FormEvent) {
    event.preventDefault();
    const name = input.trim();
    if (!name) {
      setStatus("請輸入儲存庫名稱。");
      return;
    }
    setQuery(name);
    setItems([]);
    setCursor("");
    load(name);
  }

  return (
    <section
      className="mx-auto flex w-full max-w-3xl flex-col gap-4"
      aria-labelledby="explore-title"
    >
      <div>
        <h1 id="explore-title" className="text-2xl font-semibold">
          Explore
        </h1>
        <p className="text-muted-foreground">搜尋擁有者允許被發現的儲存庫。</p>
      </div>
      <form className="flex gap-2" onSubmit={submit} role="search">
        <Input
          aria-label="儲存庫名稱"
          maxLength={128}
          value={input}
          onChange={(event) => setInput(event.target.value)}
          placeholder="搜尋儲存庫"
        />
        <Button type="submit" disabled={!user || busy}>
          搜尋
        </Button>
      </form>
      {status && (
        <Status aria-live="polite" aria-busy={busy}>
          {status}
        </Status>
      )}
      {items.length > 0 && (
        <ul className="divide-y rounded-md border" aria-label="搜尋結果">
          {items.map((item) => (
            <li key={`${item.type}:${item.id}`}>
              <a
                className="block px-4 py-3 font-medium hover:bg-muted"
                href={`/miniapp/repositories?${new URLSearchParams({ id: item.id })}`}
              >
                {item.displayName || "未命名儲存庫"}
              </a>
            </li>
          ))}
        </ul>
      )}
      {cursor && (
        <Button type="button" variant="outline" disabled={busy} onClick={() => load(query, cursor)}>
          載入更多
        </Button>
      )}
    </section>
  );
}
