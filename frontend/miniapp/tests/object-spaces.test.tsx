import React, { act } from "react";
import { createRoot } from "react-dom/client";
import test from "node:test";
import assert from "node:assert/strict";
import "./sidebar-dom";
import { Shell } from "../src/shell/Shell";
import { Spaces } from "../src/business/repository/Spaces";
import { View } from "../src/business/object/View";

const rid = "a".repeat(32), sid = "b".repeat(32), oid = "c".repeat(32);
const item = (spaceId?: string) => ({ id: oid, repositoryId: rid, name: "光纖", description: "", type: "work", completedQuantity: "0", quantity: "140", unit: "SET", status: "not_started", sourceDocument: "", sourceRevision: "", sourceItemNo: "", sourcePages: "", materialCode: "", canEdit: true, creatorId: "A", version: 1, createdAt: "2026-11-01T00:00:00Z", updatedAt: "2026-11-01T00:00:00Z", ...(spaceId === undefined ? {} : { spaceId }) });
const space = { id: sid, repositoryId: rid, name: "機房", creatorId: "A", version: 1, createdAt: "2026-11-01T00:00:00Z", updatedAt: "2026-11-01T00:00:00Z" };
const flush = async () => { for (let i = 0; i < 10; i++) await new Promise<void>(resolve => setImmediate(resolve)); };
async function fixture(handle: (request: Request) => Promise<Response>, Panel = View) {
  document.body.innerHTML = '<div id="root"></div>';
  window.history.replaceState(null, "", "/miniapp/projects");
  const oldFetch = globalThis.fetch, oldSDK = window.liff;
  window.liff = { init: async () => {}, isLoggedIn: () => true, isInClient: () => false, getIDToken: () => "token", getProfile: async () => ({ userId: "A", displayName: "A" }), login() {}, logout() {}, closeWindow() {} };
  globalThis.fetch = async (input, init) => {
    const request = input instanceof Request ? input : new Request(new URL(String(input), location.origin), init);
    const path = new URL(request.url).pathname;
    if (path.endsWith("/config")) return Response.json({ liffId: "space-test" });
    if (path.endsWith("/session")) return Response.json({ user: { userId: "A", displayName: "A" }, csrfToken: "c".repeat(43), expiresAt: new Date(Date.now() + 1800000).toISOString() });
    if (path === "/api/v1/repositories") return Response.json({ items: [], nextCursor: "" });
    return handle(request);
  };
  const root = createRoot(document.getElementById("root")!);
  await act(async () => { root.render(<Shell title="測試" workflow="projects"><Panel repositoryId={rid} /></Shell>); await flush(); });
  return {
    async click(text: string) { await act(async () => { const node = [...document.querySelectorAll<HTMLButtonElement>("button")].find(node => node.textContent === text || node.getAttribute("aria-label") === text); assert.ok(node, `missing ${text}`); node.click(); await flush(); }); },
    async fill(label: string, value: string) { await act(async () => { const node = [...document.querySelectorAll("label")].find(node => node.textContent === label)?.querySelector<HTMLInputElement>("input"); assert.ok(node, `missing ${label}`); Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!.call(node, value); node.dispatchEvent(new window.Event("input", { bubbles: true })); await flush(); }); },
    async select(value: string) { await act(async () => { const node = [...document.querySelectorAll("label")].find(node => node.textContent?.startsWith("業務空間"))?.querySelector("select"); assert.ok(node); node.value = value; node.dispatchEvent(new window.Event("change", { bubbles: true })); await flush(); }); },
    async dispose() { await act(async () => { root.unmount(); await flush(); }); globalThis.fetch = oldFetch; window.liff = oldSDK; },
  };
}

test("business spaces are lazy, paginate, and preserve or explicitly clear an object's assignment", async () => {
  const reads: string[] = [], patches: Record<string, unknown>[] = [];
  const f = await fixture(async request => {
    const url = new URL(request.url);
    if (url.pathname.endsWith("/spaces")) { reads.push(url.search); return Response.json({ items: url.search ? [space] : [], nextCursor: url.search ? "" : "next" }); }
    if (request.method === "PATCH") { const body = await request.json(); patches.push(body); return Response.json(item(body.spaceId)); }
    if (url.pathname.endsWith(oid)) return Response.json(item(sid));
    return Response.json({ items: [item(sid)], nextCursor: "" });
  });
  try {
    assert.equal(reads.length, 0, "no space request for each list item");
    await f.click("光纖");
    assert.ok(!document.body.textContent?.includes("新增業務空間"));
    await f.click("載入更多業務空間");
    assert.deepEqual(reads, ["", "?cursor=next"]);
    await f.fill("物件名稱", "光纖更新"); await f.click("儲存物件");
    assert.equal(patches[0].spaceId, sid);
    await f.click("光纖"); await f.select(""); await f.click("儲存物件");
    assert.equal(patches[1].spaceId, "");
  } finally { await f.dispose(); }
});

test("repository space creation retries the original key and lists the new space", async () => {
  const posts: { key: string | null; body: unknown }[] = [];
  const f = await fixture(async request => {
    if (request.method === "POST" && new URL(request.url).pathname.endsWith("/spaces")) {
      posts.push({ key: request.headers.get("Idempotency-Key"), body: await request.json() });
      if (posts.length === 1) throw new Error("lost reply");
      return Response.json(space);
    }
    return Response.json({ items: [], nextCursor: "" });
  }, Spaces);
  try {
    await f.fill("空間名稱", "機房"); await f.click("新增業務空間");
    assert.equal(document.querySelector<HTMLInputElement>("form input")?.disabled, true);
    await f.click("重試原空間建立請求");
    assert.equal(posts.length, 2); assert.deepEqual(posts[0], posts[1]);
    assert.ok([...document.querySelectorAll("li")].some(row => row.textContent === "機房"));
    assert.match(document.body.textContent!, /已建立業務空間/);
  } finally { await f.dispose(); }
});
