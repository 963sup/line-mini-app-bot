import React, { act } from "react";
import { createRoot } from "react-dom/client";
import test from "node:test";
import assert from "node:assert/strict";
import { pressButton } from "./sidebar-dom";
import { Shell } from "./shell-fixture";
import { View } from "../src/business/object/View";

const rid = "a".repeat(32), tid = "b".repeat(32), rootId = "c".repeat(32), childId = "d".repeat(32);
const base = `/api/v1/repositories/${rid}`;
const snapshot = { id: tid, repositoryId: rid, name: "範本", version: 1, creatorId: "A", createdAt: "2026-11-21T00:00:00Z", nodes: [{ id: "root", name: "工程" }, { id: "child", parentId: "root", name: "工序" }] };
const instance = { rootObjectId: rootId, objectIds: [rootId, childId], templateId: tid, templateVersion: 1, nodes: [{ objectId: rootId, parentId: "", name: "工程" }, { objectId: childId, parentId: rootId, name: "工序" }] };
const resource = (id: string) => ({ id, repositoryId: rid, name: id === rootId ? "工程" : "工序", parentId: id === childId ? rootId : undefined, description: "", type: "", quantity: "", unit: "", status: "not_started", creatorId: "A", canEdit: true, version: 1, createdAt: snapshot.createdAt, updatedAt: snapshot.createdAt });
const flush = async () => { for (let i = 0; i < 10; i++) await new Promise<void>(resolve => setImmediate(resolve)); };
async function fixture(handle: (r: Request) => Promise<Response>) {
  document.body.innerHTML = '<div id="root"></div>';
  window.history.replaceState(null, "", "/miniapp/projects");
  const oldFetch = globalThis.fetch, oldSDK = window.liff;
  window.liff = { init: async () => {}, isLoggedIn: () => true, isInClient: () => false, getIDToken: () => "token", getProfile: async () => ({ userId: "A", displayName: "A" }), login() {}, logout() {}, closeWindow() {} };
  globalThis.fetch = async (input, init) => {
    const r = input instanceof Request ? input : new Request(new URL(String(input), location.origin), init);
    const path = new URL(r.url).pathname;
    if (path.endsWith("/config")) return Response.json({ liffId: "template-test" });
    if (path.endsWith("/session")) return Response.json({ user: { userId: "A", displayName: "A" }, csrfToken: "c".repeat(43), expiresAt: new Date(Date.now() + 1800000).toISOString() });
    if (path === `${base}/objects`) return Response.json({ items: [], nextCursor: "" });
    if (path === `${base}/objects/${rootId}` || path === `${base}/objects/${childId}`) return Response.json(resource(path.split("/").at(-1)!));
    return handle(r);
  };
  const root = createRoot(document.getElementById("root")!);
  await act(async () => { root.render(<Shell title="測試" workflow="projects"><View repositoryId={rid} /></Shell>); await flush(); });
  return {
    async click(label: string) { await act(async () => { const b = [...document.querySelectorAll<HTMLButtonElement>("button")].find(b => b.textContent === label || b.getAttribute("aria-label") === label); assert.ok(b, `missing ${label}`); pressButton(b); await flush(); }); },
    async fill(label: string, value: string) { await act(async () => { const input = [...document.querySelectorAll("label")].find(n => n.textContent === label)?.querySelector<HTMLInputElement | HTMLTextAreaElement>("input, textarea"); assert.ok(input, label); Object.getOwnPropertyDescriptor(Object.getPrototypeOf(input), "value")!.set!.call(input, value); input.dispatchEvent(new window.Event("input", { bubbles: true })); await flush(); }); },
    async select(id: string, label = "選擇範本") { await act(async () => { const select = [...document.querySelectorAll("label")].find(node => node.textContent?.startsWith(label))?.querySelector<HTMLSelectElement>("select")!; select.value = id; select.dispatchEvent(new window.Event("change", { bubbles: true })); await flush(); }); },
    async dispose() { await act(async () => { root.unmount(); await flush(); }); globalThis.fetch = oldFetch; window.liff = oldSDK; },
  };
}

test("template creation and application preserve original retries and open the returned tree", async () => {
  const creates: { key: string | null; body: unknown }[] = [], applies: typeof creates = [];
  let listed = false;
  const f = await fixture(async r => {
    const path = new URL(r.url).pathname;
    if (path === `${base}/object-templates` && r.method === "GET") { listed = true; return Response.json({ items: [], nextCursor: "" }); }
    const log = path.endsWith("/instances") ? applies : creates;
    log.push({ key: r.headers.get("Idempotency-Key"), body: await r.json() });
    if (log.length === 1) throw new Error("response lost");
    return Response.json(path.endsWith("/instances") ? instance : snapshot);
  });
  try {
    assert.equal(listed, false, "templates load only when opened");
    await f.click("物件範本"); await f.click("新增範本");
    await f.fill("範本名稱", "範本"); await f.fill("根物件名稱", "工程"); await f.click("新增節點"); await f.fill("節點 2 名稱", "工序");
    await f.click("建立範本");
    assert.equal([...document.querySelectorAll<HTMLInputElement>('[aria-label="物件範本操作"] input')].every(n => n.matches(":disabled")), true);
    await f.click("重試原範本請求");
    assert.equal(creates.length, 2); assert.deepEqual(creates[0], creates[1]);
    await f.click("使用範本");
    assert.equal(applies.length, 0);
    await f.fill("數量", "6400");
    await f.click("建立 2 個物件"); await f.click("重試原範本請求");
    assert.equal(applies.length, 2); assert.deepEqual(applies[0], applies[1]); assert.deepEqual(applies[0].body, { nodes: [{ ...snapshot.nodes[0], quantity: "6400" }, snapshot.nodes[1]] });
    assert.ok(document.querySelector('[aria-label="新物件樹"]')); assert.match(document.body.textContent!, /父節點：工程/);
    assert.equal(document.querySelector<HTMLInputElement>('[aria-label="物件編輯"] input')!.value, "工程");
    await f.click("工序");
    assert.equal(document.querySelector<HTMLInputElement>('[aria-label="物件編輯"] input')!.value, "工序");
  } finally { await f.dispose(); }
});

test("template validation rejection unlocks editing and permits a corrected request", async () => {
  let tries = 0;
  const keys: string[] = [];
  const f = await fixture(async r => {
    if (r.method === "GET") return Response.json({ items: [], nextCursor: "" });
    keys.push(r.headers.get("Idempotency-Key")!); tries++;
    if (tries === 1) return Response.json({ error: { code: "invalid_input", message: "範本無效" } }, { status: 400 });
    return Response.json(snapshot);
  });
  try {
    await f.click("物件範本"); await f.click("新增範本"); await f.fill("範本名稱", "範本"); await f.fill("根物件名稱", "工程"); await f.click("建立範本");
    assert.equal(document.querySelector<HTMLInputElement>('[aria-label="物件範本操作"] input')!.matches(":disabled"), false);
    await f.fill("根物件名稱", "修正工程"); await f.click("建立範本");
    assert.equal(tries, 2); assert.notEqual(keys[0], keys[1]);
  } finally { await f.dispose(); }
});

test("personal library uses the existing editor and submits only after confirmation", async () => {
  const calls: string[] = [];
  const f = await fixture(async request => {
    const url = new URL(request.url);
    if (request.method === "GET" && !url.searchParams.has("scope")) return Response.json({ items: [], nextCursor: "" });
    assert.equal(url.searchParams.get("scope"), "personal");
    if (request.method === "GET") return Response.json({ items: [], nextCursor: "" });
    calls.push(url.pathname);
    const body = await request.json();
    if (url.pathname.endsWith("/instances")) {
      assert.equal(body.nodes[0].quantity, "6400");
      return Response.json(instance);
    }
    assert.equal(body.nodes[0].quantity, undefined);
    assert.equal(body.nodes[0].unit, "米");
    return Response.json({ ...snapshot, repositoryId: "" });
  });
  try {
    await f.click("物件範本");
    await f.select("personal", "範本庫");
    await f.click("新增範本");
    await f.fill("範本名稱", "配電");
    await f.fill("根物件名稱", "工程");
    await f.fill("節點 1 單位", "米");
    await f.click("建立範本");
    await f.click("使用範本");
    assert.equal(calls.length, 1);
    await f.fill("數量", "6400");
    await f.click("建立 2 個物件");
    assert.equal(calls.length, 2);
  } finally { await f.dispose(); }
});

test("document extraction creates only an editable personal draft and preserves the work tree", async () => {
  const saves: Array<typeof snapshot> = [];
  const f = await fixture(async request => {
    const url = new URL(request.url);
    if (request.method === "GET") return Response.json({ items: [], nextCursor: "" });
    if (url.pathname.endsWith("/candidates")) {
      return Response.json({ status: "completed", rawResponse: "", candidates: [
        { name: "（一）SCADA站內工程：\n既設161kv RTU", quote: "既設161kv RTU", quantity: "1", unit: "SET" },
        { categoryPath: ["電力系統", "高壓配電"], name: "地板保護", quote: "地板保護 1 式", quantity: "1", unit: "式" },
        { categoryPath: ["電力系統", "接地工程"], name: "地板保護", quote: "地板保護 1 式", quantity: "1", unit: "式" },
        { categoryPath: ["電力系統", "接地工程"], name: "搬運定位（=6組）", quote: "搬運定位（=6組） 12 式", quantity: "12", unit: "式" },
      ] });
    }
    assert.equal(url.searchParams.get("scope"), "personal");
    const body = await request.json();
    saves.push(body);
    return Response.json({ ...snapshot, ...body, repositoryId: "" });
  });
  try {
    await f.click("物件範本");
    await f.select("personal", "範本庫");
    await act(async () => {
      const panel = document.querySelector('[aria-label="物件範本操作"]')!;
      const input = panel.querySelector<HTMLInputElement>('input[type="file"]')!;
      Object.defineProperty(input, "files", { value: [new window.File(["sample"], "工程.txt")] });
      input.dispatchEvent(new window.Event("change", { bubbles: true }));
      await flush();
      pressButton([...panel.querySelectorAll<HTMLButtonElement>("button")].find(button => button.textContent === "辨識候選")!);
      await flush();
    });
    assert.match(document.body.textContent!, /12 式/);
    await f.click("填入範本草稿");
    assert.equal(saves.length, 0);
    await f.click("建立範本");
    assert.equal(saves.length, 1);
    const nodes = saves[0].nodes;
    const floors = nodes.filter(node => node.name === "地板保護");
    assert.equal(floors.length, 2);
    assert.notEqual(floors[0].parentId, floors[1].parentId);
    assert.ok(nodes.some(node => node.name === "（一）SCADA站內工程：\n既設161kv RTU"));
    assert.ok(nodes.every(node => !("quantity" in node)));
  } finally { await f.dispose(); }
});
