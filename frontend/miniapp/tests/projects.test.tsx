import React, { act } from "react";
import { createRoot } from "react-dom/client";
import test from "node:test";
import assert from "node:assert/strict";
import "./sidebar-dom";
import { Shell } from "../src/shell/Shell";
import { Workspace } from "../src/business/project/Workspace";

const sa = "C" + "a".repeat(32), sb = "C" + "b".repeat(32), rid = "c".repeat(32), pid = "d".repeat(32);
const space = (id: string) => ({ id, groupId: id, name: id === sa ? "Space A" : "Space B", status: "ACTIVE", version: 1, createdAt: "2026-11-01T00:00:00Z", updatedAt: "2026-11-01T00:00:00Z" });
const project = { id: pid, spaceId: sa, sourceRepositoryId: rid, repositoryId: rid, title: "Project A", body: "Project body", creatorId: "A", version: 3, issueCount: 0, completedIssueCount: 0, createdAt: "2026-11-01T00:00:00Z", updatedAt: "2026-11-01T00:00:00Z" };
const repository = { id: rid, groupId: sa, name: "Source A", status: "ACTIVE", version: 1, createdAt: "2026-11-01T00:00:00Z", updatedAt: "2026-11-01T00:00:00Z" };
const flush = async () => { for (let i = 0; i < 10; i++) await new Promise<void>(resolve => setImmediate(resolve)); };

async function fixture(path: string, handle: (request: Request, actor: string) => Promise<Response>) {
  window.history.replaceState(null, "", path);
  document.body.innerHTML = '<div id="root"></div>';
  let actor = "A";
  const oldFetch = globalThis.fetch, oldSDK = window.liff, requests: Request[] = [];
  window.liff = { init: async () => {}, isLoggedIn: () => true, isInClient: () => false,
    getIDToken: () => "token", getProfile: async () => ({ userId: actor, displayName: actor }), login() {}, logout() {}, closeWindow() {} };
  globalThis.fetch = async (input, init) => {
    const request = input instanceof Request ? input : new Request(new URL(String(input), location.origin), init);
    requests.push(request.clone());
    const path = new URL(request.url).pathname;
    if (path.endsWith("/config")) return Response.json({ liffId: "projects-test" });
    if (path.endsWith("/session")) return Response.json({ user: { userId: actor, displayName: actor }, csrfToken: "c".repeat(43), expiresAt: new Date(Date.now() + 1800000).toISOString() });
    return handle(request, actor);
  };
  const root = createRoot(document.getElementById("root")!);
  await act(async () => { root.render(<Shell title="專案" workflow="projects"><Workspace /></Shell>); await flush(); });
  return {
    requests,
    async click(text: string) {
      await act(async () => {
        const button = [...document.querySelectorAll("button")].find(node => node.textContent === text);
        assert.ok(button, `missing ${text}`); button.click(); await flush();
      });
    },
    async verifyAs(value: string) {
      actor = value;
      await act(async () => { document.querySelector<HTMLElement>('[aria-label="身分選單"]')!.dispatchEvent(new window.KeyboardEvent("keydown", { key: "Enter", bubbles: true })); await flush(); });
      await act(async () => { document.getElementById("verify")!.click(); await flush(); });
    },
    async route(value: string) { await act(async () => { window.history.replaceState(null, "", value); window.dispatchEvent(new window.PopStateEvent("popstate")); await flush(); }); },
    async dispose() { await act(async () => { root.unmount(); await flush(); }); globalThis.fetch = oldFetch; window.liff = oldSDK; },
  };
}

async function normal(request: Request): Promise<Response> {
  const path = new URL(request.url).pathname;
  if (path === "/api/v1/spaces") return Response.json({ items: [space(sa), space(sb)], nextCursor: "" });
  if (path === "/api/v1/spaces/" + sa) return Response.json(space(sa));
  if (path === "/api/v1/spaces/" + sb) return Response.json(space(sb));
  if (path.endsWith("/projects")) return Response.json({ items: path.includes(sa) ? [project] : [], nextCursor: "" });
  if (path.endsWith("/projects/" + pid)) return Response.json(project);
  if (path.endsWith("/items") || path.endsWith("/issues")) return Response.json({ items: [], nextCursor: "" });
  if (path === "/api/v1/repositories") return Response.json({ items: [repository], nextCursor: "" });
  if (path === "/api/v1/repositories/" + rid) return Response.json(repository);
  throw new Error("unexpected " + path);
}

test("projects are a sidebar peer and open independently through Space APIs", async () => {
  const f = await fixture("/miniapp/projects", normal);
  try {
    const projectLink = document.querySelector<HTMLAnchorElement>('a[href="/miniapp/projects"]')!;
    const repoLink = Array.from(document.querySelectorAll<HTMLButtonElement>('button[aria-expanded]')).find((button) => button.textContent === "儲存庫")!;
    assert.ok(projectLink); assert.ok(repoLink);
    assert.equal(projectLink.closest('[data-sidebar="menu-item"]')?.parentElement, repoLink.closest('[data-sidebar="menu-item"]')?.parentElement);
    assert.equal(projectLink.getAttribute("aria-current"), "page");
    assert.equal(f.requests.filter(request => new URL(request.url).pathname.startsWith("/api/v1/repositories")).length, 0, "Project landing must not require repository listing");
    await f.click("Space A");
    assert.ok(f.requests.some(request => new URL(request.url).pathname === `/api/v1/spaces/${sa}/projects`));
    assert.equal(f.requests.filter(request => new URL(request.url).pathname.startsWith("/api/v1/repositories")).length, 0);
  } finally { await f.dispose(); }
});

test("Project deep links and popstate reauthorize Space and reset previous Project detail", async () => {
  const f = await fixture(`/miniapp/projects?space=${sa}&project=${pid}`, normal);
  try {
    assert.ok(document.body.textContent!.includes("Project body"));
    await f.click("Space B");
    assert.equal(new URLSearchParams(location.search).get("space"), sb);
    assert.ok(!document.body.textContent!.includes("Project body"));
    await f.route(`/miniapp/projects?space=${sa}&project=${pid}`);
    assert.ok(document.body.textContent!.includes("Project body"));
    assert.equal(f.requests.filter(request => new URL(request.url).pathname === `/api/v1/spaces/${sa}`).length, 2);
  } finally { await f.dispose(); }
});

test("identity changes reject late Space detail even when fetch ignores cancellation", async () => {
  let finish!: (response: Response) => void;
  const f = await fixture(`/miniapp/projects?space=${sa}&project=${pid}`, async (request, actor) => {
    if (new URL(request.url).pathname === `/api/v1/spaces/${sa}`) {
      if (actor === "A") return new Promise(resolve => { finish = resolve; });
      return Response.json({ error: { code: "not_found", message: "空間不可讀取" } }, { status: 404 });
    }
    return normal(request);
  });
  try {
    await f.verifyAs("B");
    await act(async () => { finish(Response.json(space(sa))); await flush(); });
    assert.ok(document.body.textContent!.includes("空間不可讀取"));
    assert.ok(!document.body.textContent!.includes("Project body"));
    assert.ok(!f.requests.some(request => new URL(request.url).pathname === `/api/v1/spaces/${sa}/projects/${pid}`));
  } finally { await f.dispose(); }
});

test("canonical create retries retain source, payload and key after an unknown result", async () => {
  const posts: { key: string | null; body: unknown }[] = [];
  const f = await fixture(`/miniapp/projects?space=${sa}&source=${rid}`, async request => {
    if (request.method === "POST" && new URL(request.url).pathname.endsWith("/projects")) {
      posts.push({ key: request.headers.get("Idempotency-Key"), body: await request.json() });
      return posts.length === 1 ? Response.json({ error: { code: "unavailable", message: "結果不明" } }, { status: 503 }) : Response.json(project);
    }
    return normal(request);
  });
  try {
    await f.click("新增專案");
    const input = document.querySelector<HTMLInputElement>('input[maxlength="120"]')!;
    await act(async () => {
      Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!.call(input, "Same project");
      input.dispatchEvent(new Event("input", { bubbles: true })); await flush();
    });
    await f.click("建立專案"); await f.click("重試原建立請求");
    assert.equal(posts.length, 2); assert.deepEqual(posts[0], posts[1]);
    assert.deepEqual(posts[0].body, { title: "Same project", body: "", sourceRepositoryId: rid });
    assert.equal(new URLSearchParams(location.search).get("project"), pid);
  } finally { await f.dispose(); }
});

test("a source deep link from another Space cannot become a creation source", async () => {
  const f = await fixture(`/miniapp/projects?space=${sa}&source=${rid}&project=new`, async request => {
    const path = new URL(request.url).pathname;
    if (path === "/api/v1/repositories") return Response.json({ items: [{ ...repository, groupId: sb }], nextCursor: "" });
    if (path === "/api/v1/repositories/" + rid) return Response.json({ ...repository, groupId: sb });
    return normal(request);
  });
  try {
    assert.ok(document.body.textContent!.includes("來源不屬於此空間"));
    assert.equal(document.querySelector<HTMLSelectElement>("select")!.value, "");
    assert.ok(!f.requests.some(request => request.method === "POST" && new URL(request.url).pathname.endsWith("/projects")));
  } finally { await f.dispose(); }
});

test("unreadable source preserves Project metadata without inventing progress or fetching source items", async () => {
  const f = await fixture(`/miniapp/projects?space=${sa}&project=${pid}`, async request => {
    if (new URL(request.url).pathname.endsWith("/projects/" + pid)) return Response.json({ ...project, sourceAvailable: false });
    return normal(request);
  });
  try {
    assert.ok(document.body.textContent!.includes("來源儲存庫目前不可讀取"));
    assert.ok(document.body.textContent!.includes("Project body"));
    assert.ok(!document.body.textContent!.includes("尚無事項"));
    assert.equal(document.querySelector<HTMLInputElement>('input[maxlength="120"]')!.disabled, false);
    assert.ok(!f.requests.some(request => /\/(items|issues)$/.test(new URL(request.url).pathname)));
  } finally { await f.dispose(); }
});
