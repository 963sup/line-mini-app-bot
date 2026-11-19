import React, { act } from "react";
import { createRoot } from "react-dom/client";
import { test } from "node:test";
import assert from "node:assert/strict";
import { SessionProvider } from "../src/identity/Session";
import { JournalWorkspace } from "../src/business/journal/Workspace";

const flush = async () => {
  for (let i = 0; i < 12; i++) await new Promise<void>(resolve => setImmediate(resolve));
};

test("dedicated journal composer loads group pages without reading journals and preserves cancellation", async () => {
  document.body.innerHTML = '<div id="root"></div>';
  const originalFetch = globalThis.fetch, originalConfirm = window.confirm;
  const requests: string[] = [];
  const firstGroup = "C" + "1".repeat(32), secondGroup = "C" + "2".repeat(32);
  let closes = 0, confirm = false;
  window.confirm = () => confirm;
  globalThis.fetch = async (input, init) => {
    const request = input instanceof Request ? input : new Request(new URL(String(input), location.origin), init);
    const url = new URL(request.url);
    requests.push(url.pathname + url.search);
    if (url.pathname.endsWith("/config")) return Response.json({ liffId: "test" });
    if (url.pathname.endsWith("/session")) return Response.json({ user: { userId: "u1", displayName: "Tester" }, csrfToken: "c".repeat(43), expiresAt: new Date(Date.now() + 1800000).toISOString() });
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
    const cancel = Array.from(document.querySelectorAll("button")).find(button => button.textContent === "取消填寫")!;
    await act(async () => { cancel.click(); await flush(); });
    assert.equal(closes, 0, "declining cancellation must keep the draft open");
    assert.ok(document.querySelector('input[name="firstname"]'));
    confirm = true;
    await act(async () => { cancel.click(); await flush(); });
    assert.equal(closes, 1);
    assert.equal(requests.some(path => path.startsWith("/api/v1/journals")), false);
  } finally {
    await act(async () => { root.unmount(); await flush(); });
    globalThis.fetch = originalFetch;
    window.confirm = originalConfirm;
  }
});
