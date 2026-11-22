import "./sidebar-dom";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { test } from "node:test";
import assert from "node:assert/strict";
import { SessionProvider } from "../src/identity/Session";
import { Picker as Labels } from "../src/business/repository/labels/Picker";
import { Picker as Milestones } from "../src/business/repository/milestones/Picker";
import { SourcePicker } from "../src/business/project/SourcePicker";

const flush = async () => { for (let i=0;i<10;i++) await new Promise<void>(resolve=>setImmediate(resolve)); };
const props = { repositoryId:"r1", disabled:false, onChange:()=>{} };
for (const [kind, more, picker] of [
  ["labels","載入更多分類標籤",<Labels {...props} value={[]} />],
  ["milestones","載入更多可選里程碑",<Milestones {...props} value="" />],
  ["repositories","載入更多來源儲存庫",<SourcePicker {...props} initialSource="" ownerId="" value="" />],
] as const) test(kind + ": retry same page after interruption, discard options after denial", async()=>{
  document.body.innerHTML='<div id="root"></div>';
  const root=createRoot(document.getElementById("root")!);
  const original=globalThis.fetch;
  const cursors:(string|null)[]=[];
  globalThis.fetch=async input=>{
    const url=new URL(input instanceof Request ? input.url : String(input),location.origin);
    if(url.pathname.endsWith("/config")) return Response.json({liffId:"test"});
    if(url.pathname.endsWith("/session")) return Response.json({user:{userId:"a"},csrfToken:"c".repeat(43),expiresAt:new Date(Date.now()+1800000).toISOString()});
    assert.ok(url.pathname.endsWith("/"+kind));
    cursors.push(url.searchParams.get("cursor"));
    if(cursors.length===1) return Response.json({items:[{id:"1",name:"測試選項",title:"測試選項"}],nextCursor:"page2"});
    const status=cursors.length===2 ? 503 : 403;
    return Response.json({error:{code:status===503?"unavailable":"forbidden",message:"讀取失敗"}},{status});
  };
  async function click(text:string){
    const button=Array.from(document.querySelectorAll("button")).find(b=>b.textContent===text);
    assert.ok(button,text);
    await act(async()=>{button.click(); await flush();});
  }
  try{
    await act(async()=>{root.render(<SessionProvider sdkLoader={async()=>({init:async()=>{},isLoggedIn:()=>true,isInClient:()=>false,getIDToken:()=>"token",getProfile:async()=>({userId:"a"}),login(){},logout(){},closeWindow(){}})}>{picker}</SessionProvider>); await flush();});
    await click(more);
    assert.match(document.body.textContent!,/測試選項/);
    await click("重試讀取");
    assert.doesNotMatch(document.body.textContent!,/測試選項/);
    await click("重試讀取");
    assert.deepEqual(cursors,[null,"page2","page2",null]);
  }finally{await act(async()=>root.unmount());globalThis.fetch=original;}
});
