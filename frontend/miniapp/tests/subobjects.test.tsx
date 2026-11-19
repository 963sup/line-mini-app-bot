import React, { act } from "react";
import { createRoot } from "react-dom/client";
import test from "node:test";
import assert from "node:assert/strict";
import { pressButton } from "./sidebar-dom";
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
    if (path === `/api/v1/repositories/${rid}`) return Response.json({ id: rid, ownerId: "C" + rid, name: "測試儲存庫", status: "ACTIVE", version: 1, createdAt: "2026-11-01T00:00:00Z", updatedAt: "2026-11-01T00:00:00Z" });
    if (path === "/api/v1/repositories") return Response.json({ items: [], nextCursor: "" });
    if (path.endsWith("/statistics")) return Response.json({});
    return handle(request, actor);
  };
  const root = createRoot(document.getElementById("root")!);
  const render = async () => { await act(async () => { root.render(<Shell title="測試" workflow={integrated ? "repositories" : "projects"}>{integrated ? <RepositoryView /> : <Objects repositoryId={repository} extract={extractor} />}</Shell>); await flush(); }); };
  await render();
  return {
    async click(text: string) { await act(async () => { const node = [...document.querySelectorAll<HTMLButtonElement>("button")].find(node => node.textContent === text || node.getAttribute("aria-label") === text); assert.ok(node, `missing ${text}`); pressButton(node); await flush(); }); },
    async fill(label: string, value: string) { await act(async () => { const node = [...document.querySelectorAll("label")].find(node => node.textContent === label)?.querySelector<HTMLInputElement>("input,textarea"); assert.ok(node, `missing field ${label}`); const proto = node.tagName === "TEXTAREA" ? window.HTMLTextAreaElement.prototype : window.HTMLInputElement.prototype; Object.getOwnPropertyDescriptor(proto, "value")!.set!.call(node, value); node.dispatchEvent(new window.Event("input", { bubbles: true })); await flush(); }); },
    async selectType(value: string) { await act(async () => { const node = document.querySelector<HTMLSelectElement>("fieldset select")!; node.value = value; node.dispatchEvent(new window.Event("change", { bubbles: true })); await flush(); }); },
    async switchRepository() { repository = other; await render(); },
    async verifyAs(value: string) { actor = value; await act(async () => { document.querySelector<HTMLElement>('[aria-label="身分選單"]')!.dispatchEvent(new window.KeyboardEvent("keydown", { key: "Enter", bubbles: true })); await flush(); }); await act(async () => { document.getElementById("verify")!.click(); await flush(); }); },
    async dispose() { await act(async () => { root.unmount(); await flush(); }); globalThis.fetch = oldFetch; window.liff = oldSDK; },
  };
}

const sid = "d".repeat(32);
const child = (overrides={}) => ({id:sid,parentObjectId:oid,repositoryId:rid,name:"測試工序",type:"work",quantity:"9007199254740993.123456",unit:"M2",completedQuantity:"0",status:"not_started",creatorId:"A",version:1,...overrides});

test("children lazy load only expanded parent, recover read errors, native trigger exposes keyboard semantics", async () => {
  const reads: string[]=[];let fail=true;
  const f=await fixture(async request=>{
    const path=new URL(request.url).pathname;
    if(path.endsWith('/subobjects')) {reads.push(path);if(fail) {fail=false;return Response.json({error:{code:"unavailable",message:"讀取失敗"}},{status:503});}return Response.json({items:[child()],nextCursor:""});}
    return Response.json({items:[resource(),resource({id:other,name:"另一父物件"})],nextCursor:""});
  });
  try{
    assert.equal(reads.length,0);const trigger=document.querySelector<HTMLButtonElement>('[aria-label="子物件：測試物件"]')!;
    trigger.focus();assert.equal(document.activeElement,trigger);assert.equal(trigger.tagName,"BUTTON");assert.equal(trigger.getAttribute("aria-expanded"),"false");
    await f.click("子物件：測試物件");assert.equal(reads.length,1);assert.match(document.body.textContent!,/讀取失敗/);assert.equal(trigger.getAttribute("aria-expanded"),"true");assert.ok(trigger.getAttribute("aria-controls"));
    await f.click("重新讀取子物件清單");assert.match(document.body.textContent!,/測試工序/);assert.equal(reads.length,2);assert.ok(reads.every(path=>path.includes(oid)));
    assert.equal(document.querySelectorAll('button button').length,0);
    await f.click("子物件：測試物件");assert.equal(document.querySelector('[aria-label="子物件"]'),null);
  }finally{await f.dispose();}
});

test("unknown child create freezes exact identity, blocks parent navigation and same child retry succeeds without parent update", async()=>{
  const posts:{key:string|null;body:unknown}[]=[];let parentWrites=0;
  const f=await fixture(async request=>{
    const path=new URL(request.url).pathname;
    if(path.endsWith('/subobjects')) {
      if(request.method==='POST'){posts.push({key:request.headers.get('Idempotency-Key'),body:await request.json()});if(posts.length===1)throw new Error('unknown');return Response.json(child());}
      return Response.json({items:[],nextCursor:""});
    }
    if(request.method==='PATCH'||request.method==='POST')parentWrites++;
    return Response.json({items:[resource(),resource({id:other,name:"另一父物件"})],nextCursor:""});
  });
  try{
    await f.click("子物件：測試物件");await f.click("新增子物件");await f.fill("子物件名稱","測試工序");await f.fill("子物件數量","9007199254740993.123456");await f.fill("子物件單位","M2");await f.click("建立子物件");
    assert.equal(posts.length,1);for(const label of ['子物件：測試物件','子物件：另一父物件','另一父物件'])assert.equal(document.querySelector<HTMLButtonElement>(`[aria-label="${label}"]`)!.disabled,true);
    for(const text of ['新增物件','PDF匯入','重新載入物件']) assert.equal([...document.querySelectorAll<HTMLButtonElement>('button')].find(button=>button.textContent===text)!.disabled,true);
    await f.click("子物件：測試物件");assert.ok(document.querySelector('[aria-label="子物件"]'));
    await f.click("重試原子物件建立請求");assert.equal(posts.length,2);assert.deepEqual(posts[0],posts[1]);assert.equal((posts[0].body as {quantity:string}).quantity,"9007199254740993.123456");assert.equal(parentWrites,0);assert.match(document.body.textContent!,/父物件進度不變/);
  }finally{await f.dispose();}
});

test("non-owner may read children but has no child write controls",async()=>{
 const f=await fixture(async request=>Response.json(new URL(request.url).pathname.endsWith('/subobjects')?{items:[child()],nextCursor:""}:{items:[resource({creatorId:"B"})],nextCursor:""}));
 try{await f.click("子物件：測試物件");assert.match(document.body.textContent!,/測試工序/);assert.equal([...document.querySelectorAll('button')].some(button=>button.textContent?.includes('新增子物件')||button.textContent?.includes('編輯子物件')),false);}finally{await f.dispose();}
});

test("unknown then forbidden preserves original child create retry identity",async()=>{
 const attempts:{key:string|null;body:unknown}[]=[];
 const f=await fixture(async request=>{
  if(request.method==='POST'){attempts.push({key:request.headers.get('Idempotency-Key'),body:await request.json()});if(attempts.length===1)throw new Error('unknown');if(attempts.length===2)return Response.json({error:{code:'forbidden',message:'forbidden'}},{status:403});return Response.json(child());}
  return Response.json({items:new URL(request.url).pathname.endsWith('/subobjects')?[]:[resource()],nextCursor:""});
 });
 try{await f.click('子物件：測試物件');await f.click('新增子物件');await f.fill('子物件名稱','原始內容');await f.click('建立子物件');await f.click('重試原子物件建立請求');assert.equal(document.querySelector<HTMLButtonElement>('[aria-label="子物件：測試物件"]')!.disabled,true);await f.click('重試原子物件建立請求');assert.equal(attempts.length,3);assert.deepEqual(attempts[0],attempts[1]);assert.deepEqual(attempts[0],attempts[2]);}finally{await f.dispose();}
});

test("child independent totals, exact CQ validation, nonwork N/A and PATCH version recovery",async()=>{
 const patches:Record<string,unknown>[]=[];let conflict=true;
 const f=await fixture(async request=>{
  const path=new URL(request.url).pathname;
  if(request.method==='PATCH'){patches.push(await request.json());if(conflict)return Response.json({error:{code:'version_conflict',message:'版本衝突'}},{status:409});return Response.json(child({type:'non_work',completedQuantity:undefined,version:3}));}
  if(path.endsWith(sid))return Response.json(child({quantity:'100',completedQuantity:'0',version:conflict?1:2}));
  return Response.json({items:path.endsWith('/subobjects')?[child({quantity:'100',completedQuantity:'0'})]:[resource({quantity:'1'})],nextCursor:""});
 });
 try{
  await f.click('子物件：測試物件');await f.click('編輯子物件：測試工序');await f.fill('子物件完成數','100.000001');await f.click('儲存子物件');assert.equal(patches.length,0);assert.match(document.body.textContent!,/完成數須介於/);
  await f.fill('子物件完成數','99.999999');await f.click('儲存子物件');assert.equal(patches[0].completedQuantity,'99.999999');assert.equal(patches[0].quantity,'100');assert.equal(patches[0].expectedVersion,1);
  conflict=false;await f.click('重新讀取子物件');await f.click('放棄修改');
  await act(async()=>{const select=document.querySelector<HTMLSelectElement>('[aria-label="子物件"] select')!;select.value='non_work';select.dispatchEvent(new window.Event('change',{bubbles:true}));await flush();});
  assert.match(document.body.textContent!,/狀態與進度：不適用/);await f.click('儲存子物件');assert.equal(patches[1].expectedVersion,2);assert.ok(!('completedQuantity' in patches[1]));assert.ok(!('status' in patches[1]));
 }finally{await f.dispose();}
});

for(const scope of ['repository','identity'])test(`child context cancels and clears late reads after ${scope} change`,async()=>{
 let release:((value:Response)=>void)|undefined;let aborted=false;
 const f=await fixture(async request=>{
  if(new URL(request.url).pathname.endsWith('/subobjects')){request.signal.addEventListener('abort',()=>{aborted=true;});return new Promise<Response>(resolve=>{release=resolve;});}
  return Response.json({items:[resource()],nextCursor:""});
 });
 try{await f.click('子物件：測試物件');if(scope==='repository')await f.switchRepository();else await f.verifyAs('B');assert.equal(aborted,true);await act(async()=>{release?.(Response.json({items:[child({name:'舊範圍內容'})],nextCursor:""}));await flush();});assert.ok(!document.body.textContent!.includes('舊範圍內容'));assert.equal(document.querySelector('[aria-label="子物件"]'),null);}finally{await f.dispose();}
});

for(const first of ['child','parent'])test(`child write and parent navigation handshake is synchronous (${first} first)`,async()=>{
 let posts=0;
 const f=await fixture(async request=>{
  if(request.method==='POST'){posts++;throw new Error('unknown');}
  return Response.json({items:new URL(request.url).pathname.endsWith('/subobjects')?[]:[resource()],nextCursor:""});
 },true);
 try{
  await f.click('物件');await f.click('子物件：測試物件');await f.click('新增子物件');await f.fill('子物件名稱','同步子工序');const form=document.querySelector('[aria-label="子物件"] form')!;const tab=document.querySelector<HTMLButtonElement>('[aria-label="問題"]')!;
  await act(async()=>{if(first==='child'){form.dispatchEvent(new window.Event('submit',{bubbles:true,cancelable:true}));pressButton(tab);}else{pressButton(tab);form.dispatchEvent(new window.Event('submit',{bubbles:true,cancelable:true}));}await flush();});
  assert.equal(posts,first==='child'?1:0,'navigation confirmation protects the child draft');if(first==='parent')assert.ok(document.querySelector('[role=alertdialog]'));if(first==='child'){assert.equal(tab.disabled,true);assert.match(document.body.textContent!,/重試原子物件建立請求/);}
 }finally{await f.dispose();}
});

test("temporary type selection does not reset stored child completion", async () => {
  const patches: Record<string, unknown>[] = [];
  const f = await fixture(async request => {
    if (request.method === "PATCH") { patches.push(await request.json()); return Response.json(child({quantity:"100",completedQuantity:"25"})); }
    if (new URL(request.url).pathname.endsWith(sid)) return Response.json(child({quantity:"100",completedQuantity:"25"}));
    return Response.json({items: new URL(request.url).pathname.endsWith("/subobjects") ? [child({quantity:"100",completedQuantity:"25"})] : [resource()], nextCursor:""});
  });
  try {
    await f.click("子物件：測試物件"); await f.click("編輯子物件：測試工序");
    for (const type of ["non_work", "work"]) await act(async () => {
      const select = document.querySelector<HTMLSelectElement>('[aria-label="子物件"] select')!;
      select.value = type; select.dispatchEvent(new window.Event("change", {bubbles:true})); await flush();
    });
    await f.click("儲存子物件"); assert.equal(patches[0].completedQuantity, "25");
  } finally { await f.dispose(); }
});

test("child name opens a fresh inline editor; failure retains draft and cancel does not write", async () => {
  let reads = 0, writes = 0;
  const f = await fixture(async request => {
    const path = new URL(request.url).pathname;
    if (request.method === "PATCH") { writes++; return Response.json({error:{code:"version_conflict",message:"版本衝突"}}, {status:409}); }
    if (path.endsWith(sid)) { reads++; return Response.json(child({quantity:"100",version:7})); }
    return Response.json({items:path.endsWith("/subobjects")?[child()]:[resource()],nextCursor:""});
  });
  try {
    await f.click("子物件：測試物件"); await f.click("測試工序");
    const row = document.querySelector('[aria-label="測試工序"]')!.closest('[data-slot="item"]')!;
    assert.ok(row.querySelector("form")); assert.equal(reads, 1);
    await f.fill("子物件名稱", "尚未儲存");
    await f.click("新增子物件"); await f.click("繼續編輯"); await f.click("取消"); await f.click("繼續編輯");
    assert.equal(writes, 0); assert.equal(row.querySelector<HTMLInputElement>("input")!.value,"尚未儲存");
    await f.click("儲存子物件"); assert.equal(writes, 1);
    assert.equal(row.querySelector<HTMLInputElement>("input")!.value,"尚未儲存");
    assert.match(row.textContent!, /版本衝突/);
    await f.click("取消"); await f.click("放棄修改");
    assert.equal(document.querySelector("form"),null); assert.equal(writes,1);
    assert.equal(document.querySelector<HTMLButtonElement>('[aria-label="子物件：測試物件"]')!.disabled,false);
  } finally { await f.dispose(); }
});

test("arrow keeps expanded children through parent editing and each edit returns focus without duplicate forms", async () => {
  let reads = 0;
  const f = await fixture(async request => {
    const path = new URL(request.url).pathname;
    if (path.endsWith(sid)) return Response.json(child());
    if (path.endsWith(oid)) return Response.json(resource());
    if (path.endsWith("/subobjects")) { reads++; return Response.json({items:[child()],nextCursor:""}); }
    return Response.json({items:[resource()],nextCursor:""});
  });
  try {
    const trigger = document.querySelector<HTMLButtonElement>('[aria-label="子物件：測試物件"]')!;
    assert.equal(trigger.textContent, ""); assert.ok(trigger.querySelector("svg"));
    await f.click("子物件：測試物件");
    await f.click("編輯物件：測試物件");
    assert.equal(trigger.getAttribute("aria-expanded"),"true"); assert.equal(reads,1);
    assert.equal(document.querySelectorAll("form").length,1);
    assert.equal(document.activeElement, document.querySelector('[aria-label="物件編輯"] input'));
    await f.click("編輯子物件：測試工序");
    assert.equal(document.querySelector('[aria-label="物件編輯"]'),null);
    assert.equal(document.querySelectorAll("form").length,1);
    assert.equal(document.activeElement, document.querySelector('[aria-label="子物件編輯"] input'));
    await f.click("取消");
    assert.equal(document.activeElement,document.querySelector('[aria-label="編輯子物件：測試工序"]'));
    await f.click("編輯物件：測試物件"); await f.click("取消");
    assert.equal(document.activeElement, document.querySelector('[aria-label="編輯物件：測試物件"]'));
    assert.equal(trigger.getAttribute("aria-expanded"),"true"); assert.equal(reads,1);
  } finally { await f.dispose(); }
});
