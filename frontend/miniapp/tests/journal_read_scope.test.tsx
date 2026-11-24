import "./sidebar_dom";
import React, { act } from "react";
import { createRoot } from "react-dom/client";
import { test } from "node:test";
import assert from "node:assert/strict";
import { SessionProvider } from "../src/identity/Session";
import { JournalWorkspace } from "../src/business/journal/Workspace";

const flush = async () => {
  for (let i = 0; i < 12; i++) await new Promise<void>((resolve) => setImmediate(resolve));
};

const sdk = {
  init: async () => {},
  isLoggedIn: () => true,
  isInClient: () => false,
  getIDToken: () => "test-token",
  getProfile: async () => ({ userId: "u1" }),
  login() {},
  logout() {},
  closeWindow() {},
};

for (const dashboard of [false, true])
  test((dashboard ? "dashboard" : "feed") + " reads only the selected repository", async () => {
    document.body.innerHTML = '<div id="root"></div>';
    const original = globalThis.fetch;
    const paths: URL[] = [];
    globalThis.fetch = async (input, init) => {
      const req =
        input instanceof Request
          ? input
          : new Request(new URL(String(input), location.origin), init);
      const url = new URL(req.url);
      paths.push(url);
      if (url.pathname.endsWith("/config")) return Response.json({ liffId: "test" });
      if (url.pathname.endsWith("/session"))
        return Response.json({
          user: { userId: "u1", displayName: "Tester" },
          csrfToken: "c".repeat(43),
          expiresAt: new Date(Date.now() + 1800000).toISOString(),
        });
      if (url.pathname.endsWith("/repositories"))
        return Response.json({
          items: ["a", "b"].map((id) => ({ id, name: id, status: "ACTIVE" })),
          nextCursor: "",
        });
      return Response.json({ items: [], nextCursor: "" });
    };
    const root = createRoot(document.getElementById("root")!);
    try {
      await act(async () => {
        root.render(
          <SessionProvider sdkLoader={async () => sdk}>
            <JournalWorkspace dashboard={dashboard} />
          </SessionProvider>,
        );
        await flush();
      });
      const select = document.querySelector("select")!;
      assert.equal(select.value, "a");
      await act(async () => {
        select.value = "b";
        select.dispatchEvent(new Event("change", { bubbles: true }));
        await flush();
      });
      const queries = paths
        .filter((url) => url.pathname.endsWith("/journals"))
        .map((url) => url.searchParams);
      assert.deepEqual(
        queries.map((query) => query.get("repositoryId")),
        ["a", "b"],
      );
      assert.ok(
        queries.every((query) => query.get("includeProgress") === (dashboard ? "true" : null)),
      );
      assert.ok(
        paths.every((url) => !url.pathname.includes("/groups") && !url.searchParams.has("groupId")),
      );
    } finally {
      await act(async () => root.unmount());
      globalThis.fetch = original;
    }
  });

test("feed retries the same page after an outage and clears entries on access denial", async () => {
  document.body.innerHTML = '<div id="root"></div>';
  const original = globalThis.fetch;
  const cursors: (string | null)[] = [];
  globalThis.fetch = async (input, init) => {
    const req =
      input instanceof Request ? input : new Request(new URL(String(input), location.origin), init);
    const url = new URL(req.url);
    if (url.pathname.endsWith("/config")) return Response.json({ liffId: "test" });
    if (url.pathname.endsWith("/session"))
      return Response.json({
        user: { userId: "u1" },
        csrfToken: "c".repeat(43),
        expiresAt: new Date(Date.now() + 1800000).toISOString(),
      });
    if (url.pathname.endsWith("/repositories"))
      return Response.json({ items: [{ id: "a", name: "A" }], nextCursor: "" });
    cursors.push(url.searchParams.get("cursor"));
    if (cursors.length === 2 || cursors.length === 4)
      return Response.json(
        { error: { message: "讀取失敗" } },
        { status: cursors.length === 2 ? 503 : 403 },
      );
    const date = cursors.length === 1 ? "2026-01-01" : "2026-01-02";
    return Response.json({
      items: [
        {
          id: date,
          workDate: date,
          personnel: { management: [], workers: [], control: [] },
          todayItems: [],
          tomorrowItems: [],
        },
      ],
      nextCursor: cursors.length === 1 ? "second" : "third",
    });
  };
  const root = createRoot(document.getElementById("root")!);
  const click = async (label: string) => {
    const button = [...document.querySelectorAll("button")].find(
      (item) => item.textContent === label,
    );
    assert.ok(button, label);
    await act(async () => {
      button.click();
      await flush();
    });
  };
  try {
    await act(async () => {
      root.render(
        <SessionProvider sdkLoader={async () => sdk}>
          <JournalWorkspace dashboard={false} />
        </SessionProvider>,
      );
      await flush();
    });
    await click("載入更多");
    assert.match(document.body.textContent!, /2026-01-01/);
    await click("重試讀取");
    assert.match(document.body.textContent!, /2026-01-02/);
    await click("載入更多");
    assert.deepEqual(cursors, [null, "second", "second", "third"]);
    assert.doesNotMatch(document.body.textContent!, /2026-01-0[12]/);
  } finally {
    await act(async () => root.unmount());
    globalThis.fetch = original;
  }
});
