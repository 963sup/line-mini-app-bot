import React, { act } from "react";
import { createRoot } from "react-dom/client";
import { test } from "node:test";
import assert from "node:assert/strict";
import "./sidebar-dom";
import { Shell } from "../src/shell/Shell";
import { View as RepositoryView } from "../src/business/repository/View";

const a = "a".repeat(32), b = "b".repeat(32);
const space = (id: string, name = "同名空間", status = "ACTIVE") => ({ id, name, status, groupId: "C" + id, version: 1, createdAt: "2026-11-01T00:00:00Z", updatedAt: "2026-11-01T00:00:00Z" });
const flush = async () => { for (let i = 0; i < 10; i++) await new Promise<void>((resolve) => setImmediate(resolve)); };
async function fixture(path: string, handle: (request: Request, actor: string) => Promise<Response>, width = 1024) {
  Object.defineProperty(window, "innerWidth", { value: width, configurable: true });
  window.history.replaceState(null, "", path);
  document.body.innerHTML = '<div id="root"></div>';
  let actor = "A", loggedIn = true;
  const oldFetch = globalThis.fetch, oldSDK = window.liff, requests: Request[] = [];
  window.liff = { init: async () => {}, isLoggedIn: () => loggedIn, isInClient: () => false,
    getIDToken: () => "test-token", getProfile: async () => ({ userId: actor, displayName: actor }), login() {}, logout: () => { loggedIn = false; }, closeWindow() {} };
  globalThis.fetch = async (input, init) => {
    const request = input instanceof Request ? input : new Request(new URL(String(input), location.origin), init);
    requests.push(request.clone());
    const url = new URL(request.url);
    if (url.pathname.endsWith("/config")) return Response.json({ liffId: "repository-test" });
    if (url.pathname.endsWith("/session")) {
      if (request.method === "DELETE") return new Response(null, { status: 204 });
      return Response.json({ user: { userId: actor, displayName: actor }, csrfToken: "c".repeat(43), expiresAt: new Date(Date.now() + 1800000).toISOString() });
    }
    return handle(request, actor);
  };
  const workflow = path.startsWith("/miniapp/repositories") ? "repositories" : "review";
  const root = createRoot(document.getElementById("root")!);
  await act(async () => { root.render(<Shell title="空間" workflow={workflow}>{workflow === "repositories" ? <RepositoryView /> : <p>審核內容</p>}</Shell>); await flush(); });
  const clickText = async (text: string) => {
    await act(async () => {
      const button = Array.from(document.querySelectorAll("button")).find((node) => node.textContent === text);
      assert.ok(button, `missing button: ${text}`); button.click(); await flush();
    });
  };
  const openSidebar = async () => {
    if (!document.querySelector('[aria-label="身分選單"]')) await act(async () => { document.querySelector<HTMLElement>('[aria-label="切換側欄"]')!.click(); await flush(); });
  };
  const account = async (id: string) => {
    await openSidebar();
    await act(async () => { document.querySelector<HTMLElement>('[aria-label="身分選單"]')!.dispatchEvent(new window.KeyboardEvent("keydown", { key: "Enter", bubbles: true })); await flush(); });
    await act(async () => { document.getElementById(id)!.click(); await flush(); });
  };
  return { requests, clickText, openSidebar, account,
    setActor(value: string) { actor = value; loggedIn = true; },
    async dispose() { await act(async () => { root.unmount(); await flush(); }); globalThis.fetch = oldFetch; window.liff = oldSDK; },
  };
}
const repositoryRequests = (requests: Request[]) => requests.filter((request) => new URL(request.url).pathname.startsWith("/api/v1/repositories"));

test("spaces lazy-load authorized entries, retain empty-page continuation and use distinct clean links for duplicate names", async () => {
  const f = await fixture("/miniapp/review?id=correction-id&launch=old", async (request) => {
    const url = new URL(request.url);
    assert.equal(url.pathname, "/api/v1/repositories", "unrelated workflow id must not become repository scope");
    return Response.json(url.searchParams.get("cursor") ? { items: [space(a), space(b), space("c".repeat(32), "停用", "INACTIVE")], nextCursor: "" } : { items: [], nextCursor: "more" });
  });
  try {
    assert.equal(repositoryRequests(f.requests).length, 0);
    await f.clickText("空間");
    assert.match(document.body.textContent!, /本頁沒有可開啟的空間，可繼續載入/);
    await f.clickText("載入更多空間");
    const links = Array.from(document.querySelectorAll<HTMLAnchorElement>('a[href*="selected="]'));
    assert.deepEqual(links.map((link) => link.getAttribute("href")).sort(), [`/miniapp/repositories?selected=${a}`, `/miniapp/repositories?selected=${b}`]);
    assert.equal(links[0].textContent, links[1].textContent);
    assert.ok(!document.body.textContent!.includes("停用"));
    assert.equal(repositoryRequests(f.requests).length, 2);
  } finally { await f.dispose(); }
});

test("sidebar selection retains general catalog but reauthorizes the selected repository detail", async () => {
  const f = await fixture(`/miniapp/repositories?selected=${b}`, async (request) => {
    const path = new URL(request.url).pathname;
    return Response.json(path.endsWith("/" + b) ? space(b, "選取空間") : { items: [space(a), space(b)], nextCursor: "" });
  });
  try {
    assert.equal(repositoryRequests(f.requests).filter((r) => new URL(r.url).pathname.endsWith("/" + b)).length, 1);
    assert.match(document.body.textContent!, /選取空間/);
    assert.ok(document.querySelector(`a[href="/miniapp/repositories?selected=${a}"]`));
    assert.ok(document.querySelector(`a[href="/miniapp/history?group=C${b}"]`));
    assert.equal(document.querySelectorAll('a[href*="selected="]').length, 2, "detail page must not duplicate sidebar list");
  } finally { await f.dispose(); }
});

for (const query of [`id=${a}`, `repository=${a}`, "liff.state=" + encodeURIComponent(`?repositories=1&repository=${a}`)]) {
  test(`scoped entry only requests its detail and preserves scope after denial/retry: ${query}`, async () => {
    const f = await fixture("/miniapp/repositories?" + query, async (request) => {
      assert.equal(new URL(request.url).pathname, "/api/v1/repositories/" + a);
      return Response.json({ error: { message: "無法開啟此空間" } }, { status: 404 });
    });
    try {
      assert.match(document.body.textContent!, /無法開啟此空間/);
      assert.equal(document.querySelectorAll('a[href*="selected="]').length, 0);
      await f.clickText("重新載入空間");
      assert.equal(repositoryRequests(f.requests).length, 2);
    } finally { await f.dispose(); }
  });
}

test("logout and actor changes discard stale catalog responses that ignore abort", async () => {
  let late!: (response: Response) => void;
  const f = await fixture("/miniapp/repositories", async (_request, actor) => actor === "A"
    ? new Promise<Response>((resolve) => { late = resolve; })
    : Response.json({ items: [space(b, "B空間")], nextCursor: "" }));
  try {
    await f.account("logout");
    assert.equal(document.querySelectorAll('a[href*="selected="]').length, 0);
    f.setActor("B");
    await f.account("verify");
    assert.match(document.body.textContent!, /B空間/);
    await act(async () => { late(Response.json({ items: [space(a, "A秘密空間")], nextCursor: "" })); await flush(); });
    assert.ok(!document.body.textContent!.includes("A秘密空間"));
    assert.match(document.body.textContent!, /B空間/);
    await f.account("logout");
    assert.ok(!document.body.textContent!.includes("B空間"));
  } finally { await f.dispose(); }
});

test("list failure clears entries and exposes retry; final empty response has no continuation", async () => {
  let attempt = 0;
  const f = await fixture("/miniapp/repositories", async () => {
    attempt++;
    return attempt === 1 ? Response.json({ items: [space(a, "過期空間")], nextCursor: "" })
      : attempt === 2 ? Response.json({ error: { message: "暫時無法讀取空間" } }, { status: 503 })
      : Response.json({ items: [], nextCursor: "" });
  });
  try {
    await f.clickText("重新載入空間清單");
    assert.ok(!document.body.textContent!.includes("過期空間"));
    assert.match(document.body.textContent!, /暫時無法讀取空間/);
    await f.clickText("載入空間");
    assert.match(document.body.textContent!, /目前沒有可開啟的空間/);
    assert.equal(Array.from(document.querySelectorAll("button")).find((button) => button.textContent === "載入更多空間"), undefined);
  } finally { await f.dispose(); }
});

test("mobile space link closes drawer and targets the real repository rather than preview", async () => {
  const f = await fixture("/miniapp/repositories", async () => Response.json({ items: [space(a, "手機空間")], nextCursor: "" }), 375);
  try {
    await f.openSidebar();
    const link = document.querySelector<HTMLAnchorElement>(`a[href="/miniapp/repositories?selected=${a}"]`)!;
    assert.ok(link);
    await act(async () => { link.addEventListener("click", (event) => event.preventDefault()); link.click(); await flush(); });
    assert.equal(document.querySelector('[aria-label="身分選單"]'), null);
  } finally { await f.dispose(); }
});

test("denied detail for the next actor cannot be overwritten by a previous actor's late detail", async () => {
  let late!: (response: Response) => void;
  const f = await fixture(`/miniapp/repositories?selected=${a}`, async (request, actor) => {
    if (new URL(request.url).pathname === "/api/v1/repositories") return Response.json({ items: [], nextCursor: "" });
    if (actor === "A") return new Promise<Response>((resolve) => { late = resolve; });
    return Response.json({ error: { message: "無權開啟" } }, { status: 404 });
  });
  try {
    await f.account("logout");
    f.setActor("B");
    await f.account("verify");
    await act(async () => { late(Response.json(space(a, "A專用詳情"))); await flush(); });
    assert.ok(!document.body.textContent!.includes("A專用詳情"));
    assert.match(document.body.textContent!, /無權開啟/);
    assert.equal(document.querySelector(`a[href="/miniapp/repositories?selected=${a}"]`), null);
  } finally { await f.dispose(); }
});
