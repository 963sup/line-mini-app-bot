import "./sidebar-dom";
import React, { act } from "react";
import { createRoot } from "react-dom/client";
import { test } from "node:test";
import assert from "node:assert/strict";
import { SessionProvider } from "../src/identity/Session";
import { JournalWorkspace } from "../src/business/journal/Workspace";

const flush = async () => {
  for (let i = 0; i < 12; i++) await new Promise<void>(resolve => setImmediate(resolve));
};

for (const scenario of ["cancel", "publish", "publish with reference"]) test("dedicated journal composer: " + scenario, async () => {
  document.body.innerHTML = '<div id="root"></div>';
  const originalFetch = globalThis.fetch;
  const requests: string[] = [];
  const writes: {body: string; key: string | null}[] = [];
  const firstGroup = "C" + "1".repeat(32), secondGroup = "C" + "2".repeat(32);
  let closes = 0;
  globalThis.fetch = async (input, init) => {
    const request = input instanceof Request ? input : new Request(new URL(String(input), location.origin), init);
    const url = new URL(request.url);
    requests.push(url.pathname + url.search);
    if (url.pathname.endsWith("/config")) return Response.json({ liffId: "test" });
    if (url.pathname.endsWith("/session")) return Response.json({ user: { userId: "u1", displayName: "Tester" }, csrfToken: "c".repeat(43), expiresAt: new Date(Date.now() + 1800000).toISOString() });
    if (url.pathname === "/api/v1/journals" && request.method === "POST") {
      writes.push({body: await request.text(), key: request.headers.get("Idempotency-Key")});
      return writes.length === 1 ? Response.json({message: "請重試"}, {status: 503}) : Response.json({id: "published"});
    }
    if (url.pathname.endsWith("/projects")) return Response.json({items: [{id: "p1", title: "工程甲"}], nextCursor: ""});
    if (url.pathname.endsWith("/repositories")) return Response.json({items: [{id: "r1", ownerId: secondGroup, name: "工程甲"}], nextCursor: ""});
    if (url.pathname === "/api/v1/groups") {
      const next = url.searchParams.has("cursor");
      const id = next ? secondGroup : firstGroup;
      return Response.json({ items: [{ id, groupId: id, name: next ? "Second group" : "First group" }], nextCursor: next ? "" : "next-page" });
    }
    return Response.json({ items: [], nextCursor: "" });
  };
  const sdk = {
    init: async () => {}, isLoggedIn: () => true, isInClient: () => false,
    getIDToken: () => "test-token", getProfile: async () => ({ userId: "u1" }),
    login() {}, logout() {}, closeWindow() {},
  };
  const root = createRoot(document.getElementById("root")!);
  try {
    await act(async () => {
      root.render(<SessionProvider sdkLoader={async () => sdk}><JournalWorkspace dashboard={false} onCloseComposer={() => closes++} /></SessionProvider>);
      await flush();
    });
    assert.ok(requests.includes("/api/v1/groups?cursor=next-page"));
    assert.equal(requests.some(path => path.startsWith("/api/v1/journals")), false);
    const group = document.querySelector<HTMLSelectElement>("select")!;
    assert.equal(group.options[2].value, secondGroup);
    await act(async () => {
      group.value = secondGroup;
      group.dispatchEvent(new Event("change", { bubbles: true }));
      await flush();
    });
    assert.match(document.body.textContent!, /Second group/);
    assert.ok(document.querySelector('input[name="firstname"]'), "selecting a group must open the draft");
    if (scenario !== "cancel") {
      const click = async (name: string) => {
        const button = Array.from(document.querySelectorAll("button")).find(b => b.textContent === name && !b.closest("[hidden]"));
        assert.ok(button, name);
        await act(async () => { button.click(); await flush(); });
      };
      const input = (name: string) => document.querySelector<HTMLInputElement>(`input[name="${name}"]`)!;
      input("workers").value = "甲、乙";
      await click("下一步");
      await click("下一步");
      assert.equal(document.querySelector("[data-active] legend")?.textContent, "今日工作");
      input("firstname").value = "施作";
      input("firstpercent").value = "35";
      if (scenario === "publish with reference") {
        const search = document.querySelector<HTMLInputElement>("#firstref")!;
        const count = requests.length;
        await act(async () => {
          search.focus();
          search.dispatchEvent(new window.KeyboardEvent("keydown", {key: "ArrowDown", bubbles: true}));
          await flush();
          Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!.call(search, "工程");
          search.dispatchEvent(new Event("input", {bubbles: true}));
          await flush();
        });
        const results = [...document.querySelectorAll<HTMLElement>('[role="option"]')];
        assert.deepEqual(results.map(item => item.textContent).sort(), ["專案 · 工程甲", "儲存庫 · 工程甲"].sort());
        await act(async () => { results.find(item => item.textContent === "專案 · 工程甲")!.click(); await flush(); });
        assert.equal(requests.length, count);
      }
      await click("新增今日工作");
      await click("移除第 2 項");
      await click("下一步");
      await click("上一步");
      assert.equal(input("firstname").value, "施作");
      await click("上一步");
      assert.equal(input("workers").value, "甲、乙");
      await click("下一步");
      await click("下一步");
      assert.equal(writes.length, 0);
      await click("發布日誌");
      await click("發布日誌");
      assert.equal(closes, 1);
      assert.equal(writes.length, 2);
      assert.ok(writes[0].key);
      assert.deepEqual(writes[1], writes[0]);
      const draft = JSON.parse(writes[0].body);
      assert.deepEqual(draft.todayItems, [{name: "施作", completionPercent: 35, ...(scenario === "publish with reference" ? {reference: {type: "project", id: "p1"}} : {})}]);
      assert.deepEqual(draft.tomorrowItems, []);
      assert.deepEqual(draft.personnel.workers, ["甲", "乙"]);
      assert.equal(draft.groupId, secondGroup);
      return;
    }
    const cancel = Array.from(document.querySelectorAll("button")).find(button => button.textContent === "取消填寫")!;
    await act(async () => { cancel.click(); await flush(); });
    const keep = document.querySelector<HTMLButtonElement>('[data-slot="alert-dialog-cancel"]')!;
    assert.equal(document.activeElement, keep);
    await act(async () => { keep.click(); await flush(); });
    assert.equal(document.activeElement, cancel);
    assert.equal(closes, 0, "declining cancellation must keep the draft open");
    assert.ok(document.querySelector('input[name="firstname"]'));
    await act(async () => { cancel.click(); await flush(); });
    await act(async () => { document.querySelector<HTMLButtonElement>('[data-slot="alert-dialog-action"]')!.click(); await flush(); });
    assert.equal(closes, 1);
    assert.equal(requests.some(path => path.startsWith("/api/v1/journals")), false);
  } finally {
    await act(async () => { root.unmount(); await flush(); });
    globalThis.fetch = originalFetch;
  }
});
