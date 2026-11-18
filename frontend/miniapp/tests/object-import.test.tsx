import React, { act } from "react";
import { createRoot } from "react-dom/client";
import test from "node:test";
import assert from "node:assert/strict";
import "./sidebar-dom";
import { Shell } from "../src/shell/Shell";
import { Import } from "../src/business/object/Import";
import type { extractPdf } from "../src/business/object/pdf";
const rid = "a".repeat(32), jid = "b".repeat(32), rowid = "c".repeat(32), hash = "d".repeat(64), fileHash = "e".repeat(64);
const candidate = { name: "工程A", description: "測試說明", type: "工程", quantity: "1", unit: "SET", materialCode: "sample", sourceItemNo: "10", evidence: [{ page: 1, quote: "項次10 工程A 1 SET" }] };
const job = (status = "draft", version = 1, progress = false) => ({ id: jid, repositoryId: rid, creatorId: "A", sourceDocument: "sample", sourceRevision: "v0", sourceHash: hash, status, version, confirmedFromReviewVersion: status === "confirming" || status === "completed" ? 2 : 0, rows: [{ ...candidate, id: rowid, resultObjectId: progress ? "f".repeat(32) : "" }], createdAt: "2026-11-01T00:00:00Z", updatedAt: "2026-11-01T00:00:00Z" });
const flush = async () => { for (let i = 0; i < 12; i++) await new Promise<void>(resolve => setImmediate(resolve)); };
async function fixture(handle: (request: Request, blocked: boolean) => Promise<Response>, extractor: typeof extractPdf = async () => ({ pages: [{ number: 1, text: "10 測試物件 1 SET" }], fileHash }), seed?: { id: string; fileHash: string }) {
  document.body.innerHTML = '<div id="root"></div>'; window.history.replaceState(null, "", "/miniapp/projects");
  Object.defineProperty(globalThis, "localStorage", { value: window.localStorage, configurable: true }); window.localStorage.clear(); if (seed) window.localStorage.setItem(`object-import:${rid}:A`, JSON.stringify(seed));
  const oldFetch = globalThis.fetch, oldSDK = window.liff;
  window.liff = { init: async () => {}, isLoggedIn: () => true, isInClient: () => false, getIDToken: () => "token", getProfile: async () => ({ userId: "A", displayName: "A" }), login() {}, logout() {}, closeWindow() {} };
  globalThis.fetch = async (input, init) => {
    const request = input instanceof Request ? input : new Request(new URL(String(input), location.origin), init), path = new URL(request.url).pathname;
    if (path.endsWith("/config")) return Response.json({ liffId: "test" });
    if (path.endsWith("/session")) return Response.json({ user: { userId: "A", displayName: "A" }, csrfToken: "c".repeat(43), expiresAt: new Date(Date.now() + 1800000).toISOString() });
    return handle(request, blocked);
  };
  let blocked = false;
  const root = createRoot(document.getElementById("root")!); const block = (value: boolean) => { blocked = value; };
  await act(async () => { root.render(<Shell title="匯入" workflow="projects"><Import repositoryId={rid} onBlocked={block} extract={extractor} /></Shell>); await flush(); });
  return {
    blocked: () => blocked,
    async selectType(value: string) { await act(async () => { const select = document.querySelector<HTMLSelectElement>("select")!; select.value = value; select.dispatchEvent(new window.Event("change", { bubbles: true })); await flush(); }); },
    async file() { await act(async () => { const input = document.querySelector<HTMLInputElement>('input[type="file"]')!; Object.defineProperty(input, "files", { value: [new File(["%PDF-1.7"], "local.pdf", { type: "application/pdf" })], configurable: true }); input.dispatchEvent(new window.Event("change", { bubbles: true })); await flush(); }); },
    async click(text: string) { await act(async () => { const node = [...document.querySelectorAll<HTMLButtonElement>("button")].find(node => node.textContent === text); assert.ok(node, `missing ${text}`); node.click(); await flush(); }); },
    async check(text: string) { await act(async () => { const label = [...document.querySelectorAll("label")].find(label => label.textContent?.includes(text)); assert.ok(label, `missing ${text}`); label.querySelector<HTMLElement>('[role="checkbox"]')!.click(); await flush(); }); },
    async fill(label: string, value: string) { await act(async () => { const input = [...document.querySelectorAll("label")].find(node => node.textContent?.startsWith(label))?.querySelector<HTMLInputElement>("input,textarea"); assert.ok(input, `missing ${label}`); const proto = input.tagName === "TEXTAREA" ? window.HTMLTextAreaElement.prototype : window.HTMLInputElement.prototype; Object.getOwnPropertyDescriptor(proto, "value")!.set!.call(input, value); input.dispatchEvent(new window.Event("input", { bubbles: true })); await flush(); }); },
    async dispose() { await act(async () => { root.unmount(); await flush(); }); globalThis.fetch = oldFetch; window.liff = oldSDK; window.localStorage.clear(); },
  };
}

test("choosing PDF automatically imports; ambiguous retry preserves exact body and key", async () => {
  const calls: {key: string | null; body: unknown}[] = [];
  const f = await fixture(async request => {
    assert.match(new URL(request.url).pathname, /object-import-runs$/);
    calls.push({key: request.headers.get("Idempotency-Key"), body: await request.json()});
    if (calls.length === 1) throw new TypeError("network interrupted");
    return Response.json({ id: jid, status: "completed", rows: [{ id: rowid, name: "測試物件", quantity: "1", unit: "SET", resultObjectId: "f".repeat(32) }], reason: "" });
  });
  try {
    await f.file(); assert.equal(calls.length, 1); assert.equal(f.blocked(), true);
    assert.equal(document.querySelectorAll("textarea,[role=checkbox]").length, 0);
    assert.deepEqual(calls[0].body, {type: "work", sourceDocument: `pdf:${fileHash}`, sourceRevision: "", pages: [{number: 1, text: "10 測試物件 1 SET"}]});
    assert.equal(document.querySelector<HTMLSelectElement>("select")!.disabled, true);
    await f.selectType("non_work");
    await f.click("重試原匯入"); assert.deepEqual(calls[0], calls[1]); assert.equal(f.blocked(), false);
    assert.match(document.body.textContent!, /已建立 1 筆物件/);
  } finally { await f.dispose(); }
});
test("partial results report actual successes and retry frozen run", async () => {
  let count = 0; const calls: {key: string | null; body: unknown}[] = [];
  const f = await fixture(async request => { calls.push({key: request.headers.get("Idempotency-Key"), body: await request.json()}); return Response.json({id: jid, status: ++count === 1 ? "processing" : "completed", reason: "creation_unavailable", rows: [{id: rowid, name: "測試物件", quantity: "1", unit: "SET", resultObjectId: count === 1 ? "" : "f".repeat(32)}]}); });
  try { await f.selectType("non_work"); await f.file(); assert.equal(f.blocked(), true); assert.equal(document.querySelector<HTMLSelectElement>("select")!.disabled, true); assert.match(document.body.textContent!, /已建立 0\/1/); await f.selectType("work"); await f.click("重試原匯入"); assert.deepEqual(calls[0], calls[1]); assert.equal(f.blocked(), false); } finally { await f.dispose(); }
});
test("PDF extraction cancellation aborts without API transmission", async () => {
  let calls = 0, aborted = false;
  const f = await fixture(async () => { calls++; throw new Error("unexpected API"); }, async (_file, signal) => new Promise((_resolve, reject) => { signal.addEventListener("abort", () => { aborted = true; reject(new Error("cancelled")); }); }));
  try { await f.file(); assert.equal(document.querySelector<HTMLSelectElement>("select")!.disabled, true); await f.click("取消擷取"); assert.equal(aborted, true); assert.equal(calls, 0); assert.equal(f.blocked(), false); } finally { await f.dispose(); }
});

for (const unknown of [false, true]) {
 test(`known empty pending run ${unknown ? "after unknown keeps original intent" : "can change PDF"}`, async () => {
  const calls: {key:string|null;body:unknown}[]=[];
  const f=await fixture(async request=>{
   calls.push({key:request.headers.get("Idempotency-Key"),body:await request.json()});
   if(unknown&&calls.length===1) throw new TypeError("reply lost");
   return Response.json({id:jid,status:"pending",rows:[],reason:"model_unavailable"});
  });
  try {
   await f.file(); if(unknown) await f.click("重試原匯入");
   const change=[...document.querySelectorAll<HTMLButtonElement>("button")].find(b=>b.textContent==="更換PDF");
   if(unknown) {assert.equal(change,undefined);assert.equal(f.blocked(),true);assert.deepEqual(calls[0],calls[1]);}
   else {assert.ok(change);await f.click("更換PDF");assert.equal(f.blocked(),false);await f.selectType("non_work");await f.file();assert.notEqual(calls[0].key,calls[1].key);assert.equal((calls[1].body as {type:string}).type,"non_work");}
  } finally {await f.dispose();}
 });
}


test("batch type is visible before PDF and explicit non-work selection freezes into new intent", async () => {
  const calls: { key: string | null; body: { type: string } }[] = [];
  const f = await fixture(async request => {
    calls.push({ key: request.headers.get("Idempotency-Key"), body: await request.json() });
    return Response.json({ id: jid, status: "completed", rows: [], reason: "" });
  });
  try {
    const select = document.querySelector<HTMLSelectElement>("select")!;
    assert.equal(select.value, "work"); assert.equal(select.disabled, false);
    assert.deepEqual([...select.options].map(option => option.textContent), ["施作", "非施作"]);
    assert.match(document.body.textContent!, /整批類型/);
    assert.ok(select.compareDocumentPosition(document.querySelector('input[type="file"]')!) & window.Node.DOCUMENT_POSITION_FOLLOWING);
    await f.selectType("non_work"); await f.file(); assert.equal(calls[0].body.type, "non_work");
    await f.selectType("work"); await f.file(); assert.equal(calls[1].body.type, "work"); assert.notEqual(calls[0].key, calls[1].key);
  } finally { await f.dispose(); }
});

test("extraction snapshots batch type and ignores same-batch selector changes", async () => {
  let finish!: (value: Awaited<ReturnType<typeof extractPdf>>) => void;
  const calls: {type: string}[] = [];
  const f = await fixture(async request => { calls.push(await request.json()); return Response.json({id: jid, status: "completed", rows: [], reason: ""}); }, async () => new Promise(resolve => { finish = resolve; }));
  try {
    await f.selectType("non_work"); await f.file(); await f.selectType("work");
    await act(async () => { finish({ pages: [{number: 1, text: "10 測試物件 1 SET"}], fileHash }); await flush(); });
    assert.equal(calls[0].type, "non_work");
  } finally { await f.dispose(); }
});
