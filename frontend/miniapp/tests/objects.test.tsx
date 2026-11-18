import React, { act } from "react";
import { once } from "node:events";
import { createRoot } from "react-dom/client";
import test from "node:test";
import assert from "node:assert/strict";
import "./sidebar-dom";
import { Shell } from "../src/shell/Shell";
import { View as Objects } from "../src/business/object/View";
import { View as RepositoryView } from "../src/business/repository/View";

const rid = "a".repeat(32), other = "b".repeat(32), oid = "c".repeat(32);
const resource = (overrides = {}) => ({ id: oid, repositoryId: rid, name: "測試物件", description: "完整工作說明", type: "work", completedQuantity: "0", quantity: "11", unit: "SET", status: "not_started", sourceDocument: "sample-contract", sourceRevision: "v0", sourceItemNo: "40", sourcePages: "2-3", materialCode: "sample", creatorId: "A", version: 1, createdAt: "2026-11-01T00:00:00Z", updatedAt: "2026-11-01T00:00:00Z", ...overrides });
const flush = async () => { for (let i = 0; i < 10; i++) await new Promise<void>(resolve => setImmediate(resolve)); };
async function fixture(handle: (request: Request, actor: string) => Promise<Response>, integrated = false, extractor?: typeof import("../src/business/object/pdf").extractPdf) {
  document.body.innerHTML = '<div id="root"></div>';
  window.history.replaceState(null, "", integrated ? `/miniapp/repositories?selected=${rid}` : "/miniapp/projects");
  let actor = "A", repository = rid;
  const oldFetch = globalThis.fetch, oldSDK = window.liff;
  window.liff = { init: async () => {}, isLoggedIn: () => true, isInClient: () => false, getIDToken: () => "token", getProfile: async () => ({ userId: actor, displayName: actor }), login() {}, logout() {}, closeWindow() {} };
  globalThis.fetch = async (input, init) => {
    const request = input instanceof Request ? input : new Request(new URL(String(input), location.origin), init);
    const path = new URL(request.url).pathname;
    if (path.endsWith("/config")) return Response.json({ liffId: "object-test" });
    if (path.endsWith("/session")) return Response.json({ user: { userId: actor, displayName: actor }, csrfToken: "c".repeat(43), expiresAt: new Date(Date.now() + 1800000).toISOString() });
    if (path === `/api/v1/repositories/${rid}`) return Response.json({ id: rid, groupId: "C" + rid, name: "測試儲存庫", status: "ACTIVE", version: 1, createdAt: "2026-11-01T00:00:00Z", updatedAt: "2026-11-01T00:00:00Z" });
    if (path === "/api/v1/repositories") return Response.json({ items: [], nextCursor: "" });
    if (path.endsWith("/statistics")) return Response.json({});
    return handle(request, actor);
  };
  const root = createRoot(document.getElementById("root")!);
  const render = async () => { await act(async () => { root.render(<Shell title="測試" workflow={integrated ? "repositories" : "projects"}>{integrated ? <RepositoryView /> : <Objects repositoryId={repository} extract={extractor} />}</Shell>); await flush(); }); };
  await render();
  return {
    async click(text: string) { await act(async () => { const node = [...document.querySelectorAll<HTMLButtonElement>("button")].find(node => node.textContent === text || node.getAttribute("aria-label") === text); assert.ok(node, `missing ${text}`); node.click(); await flush(); }); },
    async fill(label: string, value: string) { await act(async () => { const node = [...document.querySelectorAll("label")].find(node => node.textContent === label)?.querySelector<HTMLInputElement>("input,textarea"); assert.ok(node, `missing field ${label}`); const proto = node.tagName === "TEXTAREA" ? window.HTMLTextAreaElement.prototype : window.HTMLInputElement.prototype; Object.getOwnPropertyDescriptor(proto, "value")!.set!.call(node, value); node.dispatchEvent(new window.Event("input", { bubbles: true })); await flush(); }); },
    async selectType(value: string) { await act(async () => { const node = document.querySelector<HTMLSelectElement>("fieldset select")!; node.value = value; node.dispatchEvent(new window.Event("change", { bubbles: true })); await flush(); }); },
    async switchRepository() { repository = other; await render(); },
    async verifyAs(value: string) { actor = value; await act(async () => { document.querySelector<HTMLElement>('[aria-label="身分選單"]')!.dispatchEvent(new window.KeyboardEvent("keydown", { key: "Enter", bubbles: true })); await flush(); }); await act(async () => { document.getElementById("verify")!.click(); await flush(); }); },
    async dispose() { await act(async () => { root.unmount(); await flush(); }); globalThis.fetch = oldFetch; window.liff = oldSDK; },
  };
}

test("Object creation preserves exact quantity and source fields and retries an unknown result with the original identity", async () => {
  const posts: { key: string | null; body: unknown }[] = [];
  const f = await fixture(async request => {
    if (request.method === "POST") { posts.push({ key: request.headers.get("Idempotency-Key"), body: await request.json() }); if (posts.length === 1) throw new Error("network lost"); return Response.json(resource()); }
    return Response.json({ items: [], nextCursor: "" });
  });
  try {
    await f.click("新增物件"); await f.fill("物件名稱", "測試物件"); await f.fill("數量", "9007199254740993.123400");
    for (const label of ["來源項次", "來源頁碼", "料號"]) assert.ok(![...document.querySelectorAll("label")].some(node => node.textContent === label));
    await f.click("建立物件");
    assert.equal(document.querySelector<HTMLInputElement>('input[inputmode="decimal"]')!.matches(':disabled'), true);
    await f.click("重試原物件建立請求");
    assert.equal(posts.length, 2); assert.deepEqual(posts[0], posts[1]);
    assert.equal((posts[0].body as { quantity: string }).quantity, "9007199254740993.123400");
    assert.equal((posts[0].body as { sourceRevision: string }).sourceRevision, "");
    for (const key of ["sourceItemNo", "sourcePages", "materialCode"]) assert.ok(!(key in (posts[0].body as object)));
    assert.ok(!("status" in (posts[0].body as object)), "server owns initial tracking status");
    assert.match(document.body.textContent!, /已建立物件/);
  } finally { await f.dispose(); }
});

test("Object detail is read-only for other creators and version conflicts require an explicit fresh read", async () => {
  let conflict = true; const patches: unknown[] = [];
  const f = await fixture(async request => {
    if (request.method === "PATCH") { patches.push(await request.json()); if (conflict) return Response.json({ error: { code: "version_conflict", message: "版本已變更" } }, { status: 409 }); return Response.json(resource({ version: 3, status: "completed" })); }
    return Response.json(new URL(request.url).pathname.endsWith(oid) ? resource({ version: 2 }) : { items: [resource()], nextCursor: "" });
  });
  try {
    await f.click("測試物件"); await f.click("儲存物件");
    assert.equal((patches[0] as { expectedVersion: number }).expectedVersion, 2);
    const save = [...document.querySelectorAll<HTMLButtonElement>("button")].find(node => node.textContent === "儲存物件")!; assert.equal(save.disabled, true);
    conflict = false; await f.click("重新讀取物件");
    assert.equal([...document.querySelectorAll<HTMLButtonElement>("button")].find(node => node.textContent === "儲存物件")!.disabled, false);
    await f.fill("完成數", "11");
    await f.click("儲存物件");
    assert.equal((patches[1] as { completedQuantity: string }).completedQuantity, "11");
    assert.ok(!("status" in (patches[1] as object)));
    assert.equal((patches[1] as { sourceRevision: string }).sourceRevision, "v0");
    for (const key of ["sourceItemNo", "sourcePages", "materialCode"]) assert.ok(!(key in (patches[1] as object)), "hidden provenance is omitted so the API preserves its stored value");
    await f.verifyAs("B"); await f.click("測試物件");
    assert.equal([...document.querySelectorAll("button")].some(node => node.textContent === "儲存物件"), false);
    assert.equal(document.querySelector<HTMLInputElement>("fieldset input")!.matches(":disabled"), true);
  } finally { await f.dispose(); }
});

for (const scope of ["repository", "identity"]) {
  test(`Object state rejects late list responses after ${scope} changes`, async () => {
    let finish!: (response: Response) => void;
    const f = await fixture(async (request, actor) => new URL(request.url).pathname.includes(rid) && actor === "A" ? new Promise(resolve => { finish = resolve; }) : Response.json({ items: [resource({ id: other, name: "新範圍項目" })], nextCursor: "" }));
    try {
      if (scope === "repository") await f.switchRepository(); else await f.verifyAs("B");
      await act(async () => { finish(Response.json({ items: [resource({ name: "舊範圍秘密" })], nextCursor: "" })); await flush(); });
      assert.ok(!document.body.textContent!.includes("舊範圍秘密")); assert.match(document.body.textContent!, /新範圍項目/);
    } finally { await f.dispose(); }
  });
}

test("repository exposes Problems and Objects and retains an uncertain creation when tab navigation is attempted", async () => {
  const f = await fixture(async request => request.method === "POST" ? Promise.reject(new Error("unknown")) : Response.json({ items: [], nextCursor: "" }), true);
  try {
    assert.ok(document.querySelector('[aria-label="問題"]'));
    assert.equal(document.querySelector('[aria-label="事項"]'), null);
    assert.equal(document.querySelector('[aria-label="空間功能"] [aria-label="專案"]'), null);
    assert.ok(document.querySelector('[aria-label="空間功能"] [aria-label="稽核"]'));
    assert.equal(document.querySelector('[aria-label="空間功能"] [aria-label="稽核紀錄"]'), null);
    assert.equal(document.querySelector('a[href^="/miniapp/projects?"]'), null);
    await f.click("物件"); await f.click("新增物件"); await f.fill("物件名稱", "保留請求"); await f.click("建立物件");
    assert.equal(document.querySelector<HTMLButtonElement>('[aria-label="問題"]')!.disabled, true);
    await f.click("問題");
    assert.match(document.body.textContent!, /重試原物件建立請求/);
    assert.equal([...document.querySelectorAll<HTMLButtonElement>("button")].find(node => node.textContent === "重新載入空間")!.disabled, true);
  } finally { await f.dispose(); }
});

for (const first of ["create", "reload", "tab"] as const) {
  test(`Object write and parent navigation handshake is synchronous (${first} first)`, async () => {
    let posts = 0;
    const f = await fixture(async request => {
      if (request.method === "POST") { posts++; throw new Error("unknown"); }
      return Response.json({ items: [], nextCursor: "" });
    }, true);
    try {
      await f.click("物件"); await f.click("新增物件"); await f.fill("物件名稱", "同步建立");
      const form = document.querySelector("form")!;
      const reload = [...document.querySelectorAll<HTMLButtonElement>("button")].find(node => node.textContent === "重新載入空間")!;
      const tab = document.querySelector<HTMLButtonElement>('[aria-label="問題"]')!;
      await act(async () => {
        if (first === "create") { form.dispatchEvent(new window.Event("submit", { bubbles: true, cancelable: true })); reload.click(); tab.click(); }
        else { (first === "reload" ? reload : tab).click(); form.dispatchEvent(new window.Event("submit", { bubbles: true, cancelable: true })); }
        await flush();
      });
      assert.equal(posts, first === "create" ? 1 : 0, "pending navigation confirmation prevents a stale form submission");
      if (first !== "create") assert.ok(document.querySelector("[role=dialog]"));
      if (first === "create") assert.match(document.body.textContent!, /重試原物件建立請求/);
    } finally { await f.dispose(); }
  });
}

for (const direction of ["create-to-detail", "detail-to-create"] as const) {
  test(`Object internal navigation rejects a stale same-batch form submission (${direction})`, async () => {
    let writes = 0;
    const f = await fixture(async request => {
      if (request.method === "POST" || request.method === "PATCH") { writes++; return Response.json(resource()); }
      return Response.json(new URL(request.url).pathname.endsWith(oid) ? resource() : { items: [resource()], nextCursor: "" });
    });
    try {
      if (direction === "create-to-detail") { await f.click("新增物件"); await f.fill("物件名稱", "舊建立意圖"); }
      else { await f.click("測試物件"); await f.fill("物件名稱", "舊修改意圖"); }
      const oldForm = document.querySelector("form")!;
      await act(async () => {
        const target = [...document.querySelectorAll<HTMLButtonElement>("button")].find(button => (button.getAttribute("aria-label") || button.textContent) === (direction === "create-to-detail" ? "測試物件" : "新增物件"))!;
        target.click();
        oldForm.dispatchEvent(new window.Event("submit", { bubbles: true, cancelable: true }));
        await flush();
      });
      assert.equal(writes, 0, "navigation must invalidate the previous form before React commits");
    } finally { await f.dispose(); }
  });
}


test("manual/import navigation synchronously invalidates old form", async () => {
  let writes = 0;
  const f = await fixture(async request => { if (request.method === "POST") writes++; return Response.json({items: [], nextCursor: ""}); });
  try {
    await f.click("新增物件"); await f.fill("物件名稱", "舊表單"); const oldForm = document.querySelector("form")!;
    await act(async () => { [...document.querySelectorAll<HTMLButtonElement>("button")].find(button => button.textContent === "PDF匯入")!.click(); oldForm.dispatchEvent(new window.Event("submit", {bubbles: true, cancelable: true})); await flush(); });
    assert.equal(writes, 0);
  } finally { await f.dispose(); }
});

for (const scope of ["repository", "identity"]) {
  test(`PDF local extraction aborts and clears text after ${scope} changes`, async () => {
    let aborted = false;
    const f = await fixture(async () => Response.json({ items: [], nextCursor: "" }), false, async (_file, signal) => new Promise((_resolve, reject) => { signal.addEventListener("abort", () => { aborted = true; reject(new Error("cancelled")); }); }));
    try {
      await f.click("PDF匯入");
      await act(async () => { const input = document.querySelector<HTMLInputElement>('input[type="file"]')!; Object.defineProperty(input, "files", { value: [new File(["%PDF-1.7"], "test.pdf", { type: "application/pdf" })], configurable: true }); input.dispatchEvent(new window.Event("change", { bubbles: true })); await flush(); });
      if (scope === "repository") await f.switchRepository(); else await f.verifyAs("B");
      assert.equal(aborted, true); assert.equal(document.querySelector("textarea"), null); assert.match(document.body.textContent!, /物件/);
    } finally { await f.dispose(); }
  });
}

test("object list displays five labeled lines and omits source item metadata", async () => {
 const name="ExtremelyLongUnbrokenObjectNameWithoutSpacesThatMustRemainComplete1234567890";
 const f=await fixture(async()=>Response.json({items:[resource({name,type:"Engineering",sourceItemNo:"450",quantity:"9007199254740993.123456"})],nextCursor:""}));
 try {
  const item=document.querySelector('[data-slot="item-group"] [data-slot="item"]')!;
  assert.equal(item.tagName,"DIV");assert.equal(item.querySelectorAll("button").length,3);assert.equal(item.querySelectorAll("button button").length,0);
  assert.match(item.textContent!,/9007199254740993\.123456 SET/);
  const lines=[...item.querySelectorAll('[data-slot="item-title"] button[data-variant="link"],p')].map(node=>node.textContent);
  assert.deepEqual(lines,[`名稱：${name}`,"數量：9007199254740993.123456 SET","類型：Engineering（待確認）","狀態：未開始","進度：待補／待確認"]);
  assert.ok(!item.textContent!.includes("來源項次"));
 }finally{await f.dispose();}
});


test("work completion uses exact decimals and non-work omits completion and status", async () => {
  const posts: Record<string, unknown>[] = [];
  const f = await fixture(async request => {
    if (request.method === "POST") { posts.push(await request.json()); return Response.json(resource()); }
    return Response.json({ items: [], nextCursor: "" });
  });
  try {
    await f.click("新增物件"); await f.fill("物件名稱", "精確進度");
    await f.fill("數量", "9007199254740993.123455"); await f.fill("完成數", "9007199254740993.123456");
    await f.click("建立物件"); assert.equal(posts.length, 0); assert.match(document.body.textContent!, /0至數量之間/);
    await f.fill("完成數", "9007199254740993.123454"); await f.click("建立物件");
    assert.equal(posts[0].completedQuantity, "9007199254740993.123454"); assert.ok(!("status" in posts[0]));
    await f.click("新增物件"); await f.fill("物件名稱", "非工作項目"); await f.selectType("non_work");
    assert.equal([...document.querySelectorAll("label")].some(node => node.textContent === "完成數"), false);
    await f.click("建立物件"); assert.equal(posts[1].type, "non_work"); assert.ok(!("completedQuantity" in posts[1]));
  } finally { await f.dispose(); }
});

test("legacy progress requires an explicit value and cannot silently reclassify", async () => {
  const patches: Record<string, unknown>[] = [];
  const old = resource({ type: "Engineering", completedQuantity: undefined, status: "in_progress" });
  const f = await fixture(async request => {
    if (request.method === "PATCH") { patches.push(await request.json()); return Response.json(old); }
    return Response.json(new URL(request.url).pathname.endsWith(oid) ? old : { items: [old], nextCursor: "" });
  });
  try {
    await f.click("測試物件"); assert.equal(document.querySelector<HTMLSelectElement>("fieldset select")!.value, "Engineering");
    await f.click("儲存物件"); assert.equal(patches[0].type, "Engineering"); assert.ok(!("completedQuantity" in patches[0]));
    await f.click("測試物件"); await f.selectType("work"); await f.click("儲存物件"); assert.equal(patches.length, 1);
    await f.fill("完成數", "3.500000"); await f.click("儲存物件"); assert.equal(patches[1].completedQuantity, "3.500000");
  } finally { await f.dispose(); }
});

test("work with progress cannot change to non-work before saving zero", async () => {
  let writes = 0;
  const item = resource({ completedQuantity: "0.000001", status: "in_progress" });
  const f = await fixture(async request => {
    if (request.method === "PATCH") writes++;
    return Response.json(new URL(request.url).pathname.endsWith(oid) ? item : { items: [item], nextCursor: "" });
  });
  try {
    await f.click("測試物件"); await f.selectType("non_work"); await f.click("儲存物件");
    assert.equal(writes, 0); assert.match(document.body.textContent!, /先將完成數改為0並儲存/);
  } finally { await f.dispose(); }
});


test("list distinguishes measured, non-work, and legacy missing progress", async () => {
  const f = await fixture(async () => Response.json({ items: [
    resource({ id: "1".repeat(32), name: "有進度", completedQuantity: "2.500000", status: "in_progress" }),
    resource({ id: "2".repeat(32), name: "不追蹤", type: "non_work", completedQuantity: undefined }),
    resource({ id: "3".repeat(32), name: "舊工作", completedQuantity: undefined, status: "in_progress" }),
  ], nextCursor: "" }));
  try {
    const text = (name: string) => document.querySelector(`[aria-label="${name}"]`)!.closest('[data-slot="item"]')!.textContent!;
    assert.match(text("有進度"), /類型：施作狀態：進行中進度：2.500000\/11 SET/);
    assert.match(text("不追蹤"), /類型：非施作狀態：不適用進度：不適用/);
    assert.match(text("舊工作"), /進度：待補／待確認/);
  } finally { await f.dispose(); }
});

test("object editor stays in its row, cancel preserves a draft until confirmed, save closes and updates only that row", async () => {
  let writes = 0;
  const f = await fixture(async request => {
    if (request.method === "PATCH") { writes++; return Response.json(resource({ name: "更新名稱", version: 2 })); }
    return Response.json(new URL(request.url).pathname.endsWith(oid) ? resource() : {items: [resource(), resource({id:other,name:"另一物件"})], nextCursor:""});
  });
  try {
    await f.click("編輯物件：測試物件");
    const row = document.querySelector('[aria-label="測試物件"]')!.closest('[data-slot="item"]')!;
    assert.ok(row.querySelector("form")); assert.equal(document.querySelectorAll("form").length, 1);
    await f.fill("物件名稱", "更新名稱");
    await f.click("另一物件");
    assert.ok(document.querySelector("[role=dialog]"));
    assert.equal(document.activeElement?.textContent, "繼續編輯");
    const input = row.querySelector("input")!;
    async function continueEditing() {
      // Radix restores focus from a timer after the dialog unmounts.
      await Promise.all([
        once(input, "focus", { signal: AbortSignal.timeout(2_000) }),
        f.click("繼續編輯"),
      ]);
      assert.ok(document.activeElement === input, "focus returns to the editor");
    }
    await continueEditing();
    await f.click("取消");
    await continueEditing();
    assert.equal(writes, 0);
    assert.equal(row.querySelector<HTMLInputElement>("input")!.value, "更新名稱");
    await f.click("儲存物件");
    assert.equal(writes, 1); assert.equal(document.querySelector("form"), null);
    assert.ok(document.querySelector('[aria-label="更新名稱"]')); assert.ok(document.querySelector('[aria-label="另一物件"]'));
    await f.click("編輯物件：更新名稱"); await f.fill("物件名稱", "放棄內容");
    await f.click("取消"); await f.click("放棄修改");
    assert.equal(document.querySelector("form"), null); assert.equal(writes, 1);
  } finally { await f.dispose(); }
});
