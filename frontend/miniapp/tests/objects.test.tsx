import React, { act } from "react";
import { createRoot } from "react-dom/client";
import test from "node:test";
import assert from "node:assert/strict";
import "./sidebar-dom";
import { Shell } from "../src/shell/Shell";
import { View as Objects } from "../src/business/object/View";
import { View as RepositoryView } from "../src/business/repository/View";

const rid = "a".repeat(32), other = "b".repeat(32), oid = "c".repeat(32);
const resource = (overrides = {}) => ({ id: oid, repositoryId: rid, name: "測試工程項目", description: "完整工作說明", type: "工程", quantity: "11", unit: "SET", status: "not_started", sourceDocument: "sample-contract", sourceRevision: "v0", sourceItemNo: "40", sourcePages: "2-3", materialCode: "sample", creatorId: "A", version: 1, createdAt: "2026-11-01T00:00:00Z", updatedAt: "2026-11-01T00:00:00Z", ...overrides });
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
    async click(text: string) { await act(async () => { const node = [...document.querySelectorAll<HTMLButtonElement>("button")].find(node => node.textContent === text); assert.ok(node, `missing ${text}`); node.click(); await flush(); }); },
    async fill(label: string, value: string) { await act(async () => { const node = [...document.querySelectorAll("label")].find(node => node.textContent === label)?.querySelector<HTMLInputElement>("input,textarea"); assert.ok(node, `missing field ${label}`); const proto = node.tagName === "TEXTAREA" ? window.HTMLTextAreaElement.prototype : window.HTMLInputElement.prototype; Object.getOwnPropertyDescriptor(proto, "value")!.set!.call(node, value); node.dispatchEvent(new window.Event("input", { bubbles: true })); await flush(); }); },
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
    await f.click("新增工程項目"); await f.fill("工程項目名稱", "測試工程項目"); await f.fill("數量", "9007199254740993.123400"); await f.fill("來源文件識別", "sample-contract"); await f.fill("來源版本", "v0"); await f.fill("來源項次", "40");
    await f.click("建立工程項目");
    assert.equal(document.querySelector<HTMLInputElement>('input[inputmode="decimal"]')!.matches(':disabled'), true);
    await f.click("重試原工程項目建立請求");
    assert.equal(posts.length, 2); assert.deepEqual(posts[0], posts[1]);
    assert.equal((posts[0].body as { quantity: string }).quantity, "9007199254740993.123400");
    assert.equal((posts[0].body as { sourceRevision: string }).sourceRevision, "v0");
    assert.ok(!("status" in (posts[0].body as object)), "server owns initial tracking status");
    assert.match(document.body.textContent!, /已建立工程項目/);
  } finally { await f.dispose(); }
});

test("Object detail is read-only for other creators and version conflicts require an explicit fresh read", async () => {
  let conflict = true; const patches: unknown[] = [];
  const f = await fixture(async request => {
    if (request.method === "PATCH") { patches.push(await request.json()); if (conflict) return Response.json({ error: { code: "version_conflict", message: "版本已變更" } }, { status: 409 }); return Response.json(resource({ version: 3, status: "completed" })); }
    return Response.json(new URL(request.url).pathname.endsWith(oid) ? resource({ version: 2 }) : { items: [resource()], nextCursor: "" });
  });
  try {
    await f.click("測試工程項目"); await f.click("儲存工程項目");
    assert.equal((patches[0] as { expectedVersion: number }).expectedVersion, 2);
    const save = [...document.querySelectorAll<HTMLButtonElement>("button")].find(node => node.textContent === "儲存工程項目")!; assert.equal(save.disabled, true);
    conflict = false; await f.click("重新讀取工程項目");
    assert.equal([...document.querySelectorAll<HTMLButtonElement>("button")].find(node => node.textContent === "儲存工程項目")!.disabled, false);
    await act(async () => { const select = document.querySelector<HTMLSelectElement>("fieldset select")!; select.value = "completed"; select.dispatchEvent(new window.Event("change", { bubbles: true })); await flush(); });
    await f.click("儲存工程項目");
    assert.equal((patches[1] as { status: string }).status, "completed");
    assert.equal((patches[1] as { sourceRevision: string }).sourceRevision, "v0");
    await f.verifyAs("B"); await f.click("測試工程項目");
    assert.equal([...document.querySelectorAll("button")].some(node => node.textContent === "儲存工程項目"), false);
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
    await f.click("Objects／工程項目"); await f.click("新增工程項目"); await f.fill("工程項目名稱", "保留請求"); await f.click("建立工程項目");
    assert.equal(document.querySelector<HTMLButtonElement>('[aria-label="問題"]')!.disabled, true);
    await f.click("問題");
    assert.match(document.body.textContent!, /重試原工程項目建立請求/);
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
      await f.click("Objects／工程項目"); await f.click("新增工程項目"); await f.fill("工程項目名稱", "同步建立");
      const form = document.querySelector("form")!;
      const reload = [...document.querySelectorAll<HTMLButtonElement>("button")].find(node => node.textContent === "重新載入空間")!;
      const tab = document.querySelector<HTMLButtonElement>('[aria-label="問題"]')!;
      await act(async () => {
        if (first === "create") { form.dispatchEvent(new window.Event("submit", { bubbles: true, cancelable: true })); reload.click(); tab.click(); }
        else { (first === "reload" ? reload : tab).click(); form.dispatchEvent(new window.Event("submit", { bubbles: true, cancelable: true })); }
        await flush();
      });
      assert.equal(posts, first === "create" ? 1 : 0);
      if (first === "create") assert.match(document.body.textContent!, /重試原工程項目建立請求/);
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
      if (direction === "create-to-detail") { await f.click("新增工程項目"); await f.fill("工程項目名稱", "舊建立意圖"); }
      else { await f.click("測試工程項目"); await f.fill("工程項目名稱", "舊修改意圖"); }
      const oldForm = document.querySelector("form")!;
      await act(async () => {
        const target = [...document.querySelectorAll<HTMLButtonElement>("button")].find(button => button.textContent === (direction === "create-to-detail" ? "測試工程項目" : "新增工程項目"))!;
        target.click();
        oldForm.dispatchEvent(new window.Event("submit", { bubbles: true, cancelable: true }));
        await flush();
      });
      assert.equal(writes, 0, "navigation must invalidate the previous form before React commits");
    } finally { await f.dispose(); }
  });
}


test("manual/import transitions synchronously invalidate the previous active form and parse action", async () => {
  let writes = 0;
  const f = await fixture(async request => {
    if (request.method === "POST" || request.method === "PATCH") { writes++; throw new Error("must not write"); }
    return Response.json({ items: [], nextCursor: "" });
  }, false, async () => ({ pages: [{ number: 1, text: "sample row" }], fileHash: "e".repeat(64) }));
  try {
    await f.click("新增工程項目"); await f.fill("工程項目名稱", "舊表單"); const oldForm = document.querySelector("form")!;
    await act(async () => { [...document.querySelectorAll<HTMLButtonElement>("button")].find(button => button.textContent === "PDF文字匯入")!.click(); oldForm.dispatchEvent(new window.Event("submit", { bubbles: true, cancelable: true })); await flush(); });
    assert.equal(writes, 0);
    await act(async () => { const input = document.querySelector<HTMLInputElement>('input[type="file"]')!; Object.defineProperty(input, "files", { value: [new File(["%PDF-1.7"], "test.pdf", { type: "application/pdf" })], configurable: true }); input.dispatchEvent(new window.Event("change", { bubbles: true })); await flush(); });
    await f.fill("來源文件／訂單號", "sample");
    await act(async () => { for (const checkbox of document.querySelectorAll<HTMLElement>('[role="checkbox"]')) checkbox.click(); await flush(); });
    const parse = [...document.querySelectorAll<HTMLButtonElement>("button")].find(button => button.textContent === "解析所選文字")!;
    await act(async () => { [...document.querySelectorAll<HTMLButtonElement>("button")].find(button => button.textContent === "返回工程項目清單")!.click(); parse.click(); await flush(); });
    assert.equal(writes, 0);
  } finally { await f.dispose(); }
});

for (const scope of ["repository", "identity"]) {
  test(`PDF local extraction aborts and clears text after ${scope} changes`, async () => {
    let aborted = false;
    const f = await fixture(async () => Response.json({ items: [], nextCursor: "" }), false, async (_file, signal) => new Promise((_resolve, reject) => { signal.addEventListener("abort", () => { aborted = true; reject(new Error("cancelled")); }); }));
    try {
      await f.click("PDF文字匯入");
      await act(async () => { const input = document.querySelector<HTMLInputElement>('input[type="file"]')!; Object.defineProperty(input, "files", { value: [new File(["%PDF-1.7"], "test.pdf", { type: "application/pdf" })], configurable: true }); input.dispatchEvent(new window.Event("change", { bubbles: true })); await flush(); });
      if (scope === "repository") await f.switchRepository(); else await f.verifyAs("B");
      assert.equal(aborted, true); assert.equal(document.querySelector("textarea"), null); assert.match(document.body.textContent!, /Objects／工程項目/);
    } finally { await f.dispose(); }
  });
}
