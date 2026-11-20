import "./sidebar-dom";
import React, { act } from "react";
import { createRoot } from "react-dom/client";
import { test } from "node:test";
import assert from "node:assert/strict";
import { SessionProvider } from "../src/identity/Session";
import { JournalWorkspace } from "../src/business/journal/Workspace";

const flush = async () => {
  for (let i = 0; i < 12; i++) await new Promise<void>(resolve => setImmediate(resolve));
};

for (const scenario of ["cancel", "publish", "publish with reference", "space deferred", "space unknown", "space limit"]) test("dedicated journal composer: " + scenario, async () => {
  document.body.innerHTML = '<div id="root"></div>';
  const originalFetch = globalThis.fetch;
  const requests: string[] = [];
  const writes: {body: string; key: string | null}[] = [];
  const spaceWrites: {body:string;key:string|null}[] = [];
  let finishSpace!: (response:Response)=>void;
  const firstGroup = "C" + "1".repeat(32), secondGroup = "C" + "2".repeat(32);
  let closes = 0;
  globalThis.fetch = async (input, init) => {
    const request = input instanceof Request ? input : new Request(new URL(String(input), location.origin), init);
    const url = new URL(request.url);
    requests.push(url.pathname + url.search);
    if (url.pathname.endsWith("/config")) return Response.json({ liffId: "test" });
    if (url.pathname.endsWith("/session")) return Response.json({ user: { userId: "u1", displayName: "Tester" }, csrfToken: "c".repeat(43), expiresAt: new Date(Date.now() + 1800000).toISOString() });
    if (url.pathname.endsWith("/spaces") && request.method === "POST") {
      spaceWrites.push({body:await request.text(),key:request.headers.get("Idempotency-Key")});
      if(scenario === "space deferred")return new Promise(resolve=>{finishSpace=resolve;});
      if(scenario === "space limit")return Response.json({id:"s1",name:"現場空間"});
      if(spaceWrites.length === 1)throw new Error("response lost");
      if(spaceWrites.length === 2)return Response.json({error:{code:"forbidden",message:"permission changed"}},{status:403});
      return Response.json({id:"s1",name:"現場空間"});
    }
    if(url.pathname.endsWith("/spaces") && scenario === "space limit")return Response.json({items:Array.from({length:20},(_,i)=>({id:"pre"+i,name:"既有空間"+i})),nextCursor:""});
    if (url.pathname === "/api/v1/journals" && request.method === "POST") {
      writes.push({body: await request.text(), key: request.headers.get("Idempotency-Key")});
      return writes.length === 1 ? Response.json({message: "請重試"}, {status: 503}) : Response.json({id: "published"});
    }
    if (url.pathname.endsWith("/journal-reference-targets")) {
      if (url.searchParams.has("cursor")) return Response.json({items:[{type:"project",id:"p1",name:"工程甲"}],nextCursor:""});
      return Response.json({items:url.pathname.includes(secondGroup)?[{type:"repository",id:"r1",name:"工程甲"}]:[],nextCursor:"projects"});
    }
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
    if (scenario !== "cancel") {
      const click = async (name: string) => {
        const button = Array.from(document.querySelectorAll("button")).find(b => b.textContent === name && !b.closest("[hidden]"));
        assert.ok(button, name);
        await act(async () => { button.click(); await flush(); });
      };
      const input = (name: string) => document.querySelector<HTMLInputElement>(`input[name="${name}"]`)!;
      input("workers").value = "甲、乙";
      await click("下一步");
      await click("下一步");
      assert.equal(document.querySelector("[data-active] legend")?.textContent, "今日工作");
      input("firstname").value = "施作";
      input("firstpercent").value = "35";
      if (scenario.startsWith("space ")) {
        const search=document.querySelector<HTMLInputElement>("#firstref")!;
        await act(async()=>{search.focus();search.dispatchEvent(new window.KeyboardEvent("keydown",{key:"ArrowDown",bubbles:true}));await flush();});
        await act(async()=>{[...document.querySelectorAll<HTMLElement>('[role="option"]')].find(item=>item.textContent==="儲存庫 · 工程甲")!.click();await flush();});
        if(scenario === "space limit")for(let i=0;i<19;i++){
          const spaces=document.querySelector<HTMLInputElement>("#firstspaces")!;
          await act(async()=>{spaces.focus();spaces.dispatchEvent(new window.KeyboardEvent("keydown",{key:"ArrowDown",bubbles:true}));await flush();});
          await act(async()=>{[...document.querySelectorAll<HTMLElement>('[role="option"]')].find(item=>item.textContent==="既有空間"+i)!.click();await flush();});
        }
        const name=[...document.querySelectorAll("label")].find(label=>label.textContent?.includes("新增空間"))!.querySelector<HTMLInputElement>("input")!;
        await act(async()=>{Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype,"value")!.set!.call(name,"現場空間");name.dispatchEvent(new Event("input",{bubbles:true}));await flush();});
        await act(async()=>{
          [...document.querySelectorAll<HTMLButtonElement>("button")].find(button=>button.textContent==="建立並選用")!.click();
          [...document.querySelectorAll<HTMLButtonElement>("button")].find(button=>button.textContent==="取消填寫")!.click();
          await flush();
        });
        const assertLocked=()=>{
          assert.equal(name.disabled,true);
          assert.equal(search.disabled,true);
          assert.equal([...document.querySelectorAll<HTMLButtonElement>("button")].find(button=>button.textContent==="移除第 1 項")!.disabled,true);
          const cancel=[...document.querySelectorAll<HTMLButtonElement>("button")].find(button=>button.textContent==="取消填寫")!;
          assert.equal(cancel.disabled,true);cancel.click();assert.equal(closes,0);
        };
        if(scenario !== "space limit"){
          assertLocked();
          await act(async()=>{document.querySelector("form")!.dispatchEvent(new Event("submit",{bubbles:true,cancelable:true}));await flush();});
          assert.equal(writes.length,0,"pending space cannot publish or advance the draft");
        }
        if(scenario === "space deferred")await act(async()=>{finishSpace(Response.json({id:"s1",name:"現場空間"}));await flush();});
        else if(scenario === "space unknown") {
          await click("重試原空間建立請求");assertLocked();
          await click("重試原空間建立請求");
          assert.equal(spaceWrites.length,3);assert.ok(spaceWrites[0].key);
          assert.deepEqual(spaceWrites,[spaceWrites[0],spaceWrites[0],spaceWrites[0]]);
        }
        assert.equal(search.disabled,false);
        assert.equal([...document.querySelectorAll<HTMLButtonElement>("button")].find(button=>button.textContent==="取消填寫")!.disabled,false);
        const expectedSpaces=scenario === "space limit" ? [...Array.from({length:19},(_,i)=>"pre"+i),"s1"] : ["s1"];
        assert.deepEqual([...document.querySelectorAll<HTMLInputElement>('input[name="firstspace"]')].map(value=>value.value),expectedSpaces);
        if(scenario === "space limit"){
          await act(async()=>{Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype,"value")!.set!.call(name,"第21個空間");name.dispatchEvent(new Event("input",{bubbles:true}));await flush();});
          const create=[...document.querySelectorAll<HTMLButtonElement>("button")].find(button=>button.textContent==="建立並選用")!;
          assert.equal(create.disabled,true);create.click();assert.equal(spaceWrites.length,1);
        }
      }
      if (scenario === "publish with reference") {
        const search = document.querySelector<HTMLInputElement>("#firstref")!;
        const count = requests.length;
        await act(async () => {
          search.focus();
          search.dispatchEvent(new window.KeyboardEvent("keydown", {key: "ArrowDown", bubbles: true}));
          await flush();
          Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!.call(search, "工程");
          search.dispatchEvent(new Event("input", {bubbles: true}));
          await flush();
        });
        const results = [...document.querySelectorAll<HTMLElement>('[role="option"]')];
        assert.deepEqual(results.map(item => item.textContent).sort(), ["專案 · 工程甲", "儲存庫 · 工程甲"].sort());
        await act(async () => { results.find(item => item.textContent === "專案 · 工程甲")!.click(); await flush(); });
        assert.equal(requests.length, count);
      }
      await click("新增今日工作");
      await click("移除第 2 項");
      await click("下一步");
      await click("上一步");
      assert.equal(input("firstname").value, "施作");
      await click("上一步");
      assert.equal(input("workers").value, "甲、乙");
      await click("下一步");
      await click("下一步");
      assert.equal(writes.length, 0);
      await click("發布日誌");
      await click("發布日誌");
      assert.equal(closes, 1);
      assert.equal(writes.length, 2);
      assert.ok(writes[0].key);
      assert.deepEqual(writes[1], writes[0]);
      const draft = JSON.parse(writes[0].body);
      assert.deepEqual(draft.todayItems, [{name: "施作", completionPercent: 35, ...(scenario === "publish with reference" ? {reference: {type: "project", id: "p1"}} : scenario.startsWith("space ") ? {reference:{type:"repository",id:"r1"},spaceIds:scenario === "space limit" ? [...Array.from({length:19},(_,i)=>"pre"+i),"s1"] : ["s1"]} : {})}]);
      assert.deepEqual(draft.tomorrowItems, []);
      assert.deepEqual(draft.personnel.workers, ["甲", "乙"]);
      assert.equal(draft.groupId, secondGroup);
      return;
    }
    const cancel = Array.from(document.querySelectorAll("button")).find(button => button.textContent === "取消填寫")!;
    await act(async () => { cancel.click(); await flush(); });
    const keep = document.querySelector<HTMLButtonElement>('[data-slot="alert-dialog-cancel"]')!;
    assert.equal(document.activeElement, keep);
    await act(async () => { keep.click(); await flush(); });
    assert.equal(document.activeElement, cancel);
    assert.equal(closes, 0, "declining cancellation must keep the draft open");
    assert.ok(document.querySelector('input[name="firstname"]'));
    await act(async () => { cancel.click(); await flush(); });
    await act(async () => { document.querySelector<HTMLButtonElement>('[data-slot="alert-dialog-action"]')!.click(); await flush(); });
    assert.equal(closes, 1);
    assert.equal(requests.some(path => path.startsWith("/api/v1/journals")), false);
  } finally {
    await act(async () => { root.unmount(); await flush(); });
    globalThis.fetch = originalFetch;
  }
});
