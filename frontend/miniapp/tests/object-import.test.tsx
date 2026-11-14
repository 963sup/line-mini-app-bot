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
async function fixture(handle: (request: Request, blocked: boolean) => Promise<Response>, extractor: typeof extractPdf = async () => ({ pages: [{ number: 1, text: "項次10 工程A 1 SET 敏感價格" }, { number: 2, text: "不要傳送的條款" }], fileHash }), seed?: { id: string; fileHash: string }) {
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
    async file() { await act(async () => { const input = document.querySelector<HTMLInputElement>('input[type="file"]')!; Object.defineProperty(input, "files", { value: [new File(["%PDF-1.7"], "local.pdf", { type: "application/pdf" })], configurable: true }); input.dispatchEvent(new window.Event("change", { bubbles: true })); await flush(); }); },
    async click(text: string) { await act(async () => { const node = [...document.querySelectorAll<HTMLButtonElement>("button")].find(node => node.textContent === text); assert.ok(node, `missing ${text}`); node.click(); await flush(); }); },
    async check(text: string) { await act(async () => { const label = [...document.querySelectorAll("label")].find(label => label.textContent?.includes(text)); assert.ok(label, `missing ${text}`); label.querySelector<HTMLElement>('[role="checkbox"]')!.click(); await flush(); }); },
    async fill(label: string, value: string) { await act(async () => { const input = [...document.querySelectorAll("label")].find(node => node.textContent?.startsWith(label))?.querySelector<HTMLInputElement>("input,textarea"); assert.ok(input, `missing ${label}`); const proto = input.tagName === "TEXTAREA" ? window.HTMLTextAreaElement.prototype : window.HTMLInputElement.prototype; Object.getOwnPropertyDescriptor(proto, "value")!.set!.call(input, value); input.dispatchEvent(new window.Event("input", { bubbles: true })); await flush(); }); },
    async dispose() { await act(async () => { root.unmount(); await flush(); }); globalThis.fetch = oldFetch; window.liff = oldSDK; window.localStorage.clear(); },
  };
}

test("selected redacted text requires explicit consent; parse freezes one draft identity and confirmation resumes the reviewed version", async () => {
  const parses: unknown[] = [], drafts: { key: string | null; body: unknown }[] = [], reviews: unknown[] = [], confirmations: unknown[] = [];
  const f = await fixture(async (request, blocked) => {
    const path = new URL(request.url).pathname;
    if (path.endsWith("/parse")) { parses.push(await request.json()); return Response.json({ rows: [candidate], sourceHash: hash }); }
    if (path.endsWith("/review")) { reviews.push(await request.json()); return Response.json(job("reviewed", reviews.length + 1)); }
    if (path.endsWith("/confirm")) { assert.equal(blocked, true, "pending confirmation blocks parent navigation"); confirmations.push(await request.json()); return Response.json({ ...job(confirmations.length === 1 ? "confirming" : "completed", confirmations.length + 3, true), confirmedFromReviewVersion: 3 }); }
    if (request.method === "GET") return Response.json({ ...job("confirming", 4, true), confirmedFromReviewVersion: 3 });
    drafts.push({ key: request.headers.get("Idempotency-Key"), body: await request.json() }); if (drafts.length === 1) throw new Error("network lost"); return Response.json(job());
  });
  try {
    await f.file(); assert.equal(document.querySelectorAll('[role="checkbox"][data-state="checked"]').length, 0); assert.equal(parses.length, 0);
    await f.fill("來源文件／訂單號", "sample"); await f.fill("文件版本", "v0"); await f.fill("第1頁待傳送文字", "項次10 工程A 1 SET");
    await f.check("選取第1頁傳送"); await f.check("我已核對"); assert.equal(parses.length, 0); await f.check("我同意"); await f.click("解析所選文字");
    assert.deepEqual(parses, [{ sourceDocument: "sample", sourceRevision: "v0", pages: [{ number: 1, text: "項次10 工程A 1 SET" }], consent: true }]);
    assert.equal(drafts.length, 0); await f.click("保存匯入草稿"); await f.click("重試原草稿建立請求"); assert.deepEqual(drafts[0], drafts[1]); assert.equal(parses.length, 1);
    await f.fill("名稱", "人工核對工程A"); await f.click("保存人工審查");
    assert.equal((reviews[0] as { rows: { name: string }[] }).rows[0].name, "人工核對工程A"); assert.ok(!("evidence" in (reviews[0] as { rows: object[] }).rows[0]));
    assert.equal(confirmations.length, 0);
    const confirmButton = [...document.querySelectorAll<HTMLButtonElement>("button")].find(node => node.textContent === "明確確認建立工程項目")!;
    await act(async () => {
      const input = [...document.querySelectorAll("label")].find(node => node.textContent === "名稱")!.querySelector("input")!;
      Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!.call(input, "再修改未保存"); input.dispatchEvent(new window.Event("input", { bubbles: true })); confirmButton.click(); await flush();
    });
    assert.equal(confirmations.length, 0, "unsaved edits must block confirmation synchronously");
    await f.click("保存人工審查"); await f.click("明確確認建立工程項目"); await f.click("重新讀取匯入進度"); await f.click("依原審查版本繼續確認");
    assert.deepEqual(confirmations, [{ expectedVersion: 3 }, { expectedVersion: 3 }]);
    const stored = window.localStorage.getItem(`object-import:${rid}:A`)!; assert.deepEqual(JSON.parse(stored), { id: jid, fileHash }); assert.ok(!stored.includes("工程A"));
  } finally { await f.dispose(); }
});

test("PDF extraction cancellation aborts local work without sending text to any API", async () => {
  let calls = 0, aborted = false;
  const f = await fixture(async () => { calls++; throw new Error("must not call API"); }, async (_file, signal) => new Promise((_resolve, reject) => { signal.addEventListener("abort", () => { aborted = true; reject(new Error("cancelled")); }); }));
  try { await f.file(); await f.click("取消目前操作"); assert.equal(aborted, true); assert.equal(calls, 0); } finally { await f.dispose(); }
});

test("discarding unsaved candidates requires renewed consent; unavailable AI never creates a draft", async () => {
  let parses = 0, creates = 0;
  const f = await fixture(async request => {
    if (new URL(request.url).pathname.endsWith("/parse")) { parses++; return parses === 1 ? Response.json({ rows: [candidate], sourceHash: hash }) : Response.json({ error: { code: "not_configured", message: "AI解析尚未啟用" } }, { status: 503 }); }
    creates++; throw new Error("must not create draft");
  });
  try {
    await f.file(); await f.fill("來源文件／訂單號", "sample"); await f.check("選取第1頁傳送"); await f.check("我已核對"); await f.check("我同意"); await f.click("解析所選文字");
    await f.click("放棄候選重新解析"); assert.equal(document.querySelectorAll('[role="checkbox"][data-state="checked"]').length, 1);
    await f.check("我已核對"); await f.check("我同意"); await f.click("解析所選文字");
    assert.match(document.body.textContent!, /AI解析尚未啟用/); assert.equal(creates, 0); assert.equal(parses, 2);
    assert.equal([...document.querySelectorAll("button")].some(node => node.textContent === "保存匯入草稿"), false);
  } finally { await f.dispose(); }
});

test("resuming a job with a different local PDF cannot overwrite the stored file hash or unlock confirmation", async () => {
  const f = await fixture(async () => Response.json(job("reviewed", 2)), async () => ({ pages: [{ number: 1, text: "unrelated" }], fileHash: "0".repeat(64) }), { id: jid, fileHash });
  try {
    await f.click("繼續已保存的匯入"); await f.file(); await f.click("重新讀取匯入進度");
    assert.match(document.body.textContent!, /尚未相符/);
    assert.equal([...document.querySelectorAll<HTMLButtonElement>("button")].find(button => button.textContent === "明確確認建立工程項目")!.disabled, true);
    assert.equal(JSON.parse(window.localStorage.getItem(`object-import:${rid}:A`)!).fileHash, fileHash);
  } finally { await f.dispose(); }
});


async function readyDraft(f: Awaited<ReturnType<typeof fixture>>) {
  await f.file(); await f.fill("來源文件／訂單號", "sample"); await f.check("選取第1頁傳送"); await f.check("我已核對"); await f.check("我同意"); await f.click("解析所選文字");
}

test("a first definitive draft rejection unlocks candidate correction without creating a job", async () => {
  const f = await fixture(async request => new URL(request.url).pathname.endsWith("/parse") ? Response.json({ rows: [candidate], sourceHash: hash }) : Response.json({ error: { code: "invalid_input", message: "請修正資料" } }, { status: 400 }));
  try {
    await readyDraft(f); await f.click("保存匯入草稿");
    const discard = [...document.querySelectorAll<HTMLButtonElement>("button")].find(button => button.textContent === "放棄候選重新解析")!;
    assert.equal(discard.disabled, false); assert.equal(f.blocked(), false);
    await f.click("放棄候選重新解析"); assert.ok(document.querySelector('[role="checkbox"][aria-label="選取第1頁"]'));
  } finally { await f.dispose(); }
});

for (const code of ["invalid_input", "forbidden"]) {
  test(`an ambiguous draft attempt retains its body/key after later ${code} rejection`, async () => {
    const drafts: { key: string | null; body: unknown }[] = [];
    const f = await fixture(async request => {
      if (new URL(request.url).pathname.endsWith("/parse")) return Response.json({ rows: [candidate], sourceHash: hash });
      drafts.push({ key: request.headers.get("Idempotency-Key"), body: await request.json() });
      if (drafts.length === 1) throw new Error("network lost");
      if (drafts.length === 2) return Response.json({ error: { code, message: "rejected" } }, { status: code === "forbidden" ? 403 : 400 });
      return Response.json(job());
    });
    try {
      await readyDraft(f); await f.click("保存匯入草稿"); await f.click("重試原草稿建立請求");
      assert.equal([...document.querySelectorAll<HTMLButtonElement>("button")].find(button => button.textContent === "放棄候選重新解析")!.disabled, true); assert.equal(f.blocked(), true);
      await f.click("重試原草稿建立請求"); assert.deepEqual(drafts[0], drafts[1]); assert.deepEqual(drafts[0], drafts[2]);
    } finally { await f.dispose(); }
  });
}

test("cancelling a pending draft retains the unresolved identity and navigation block", async () => {
  const drafts: { key: string | null; body: unknown }[] = [];
  const f = await fixture(async request => {
    if (new URL(request.url).pathname.endsWith("/parse")) return Response.json({ rows: [candidate], sourceHash: hash });
    drafts.push({ key: request.headers.get("Idempotency-Key"), body: await request.json() });
    return drafts.length === 1 ? new Promise(() => {}) : Response.json(job());
  });
  try {
    await readyDraft(f); await f.click("保存匯入草稿"); await f.click("取消目前操作"); assert.equal(f.blocked(), true);
    await f.click("重試原草稿建立請求"); assert.deepEqual(drafts[0], drafts[1]);
  } finally { await f.dispose(); }
});

test("unavailable browser storage shows the job ID and does not promise refresh recovery", async () => {
  const f = await fixture(async request => new URL(request.url).pathname.endsWith("/parse") ? Response.json({ rows: [candidate], sourceHash: hash }) : Response.json(job()));
  const proto = Object.getPrototypeOf(window.localStorage), original = proto.setItem;
  try {
    await readyDraft(f); proto.setItem = () => { throw new Error("storage denied"); }; await f.click("保存匯入草稿");
    assert.match(document.body.textContent!, new RegExp(jid)); assert.match(document.body.textContent!, /無法保存續傳識別/); assert.match(document.body.textContent!, /重新整理後不保證可找回/);
    assert.equal([...document.querySelectorAll<HTMLButtonElement>("button")].find(button => button.textContent === "保存人工審查")!.disabled, false);
  } finally { proto.setItem = original; await f.dispose(); }
});
