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

for (const dashboard of [false,true]) test((dashboard ? "dashboard" : "feed") + " reads only the selected repository", async () => {
  document.body.innerHTML = '<div id="root"></div>';
  const original = globalThis.fetch;
  const paths: URL[] = [];
  globalThis.fetch = async (input,init) => {
    const req = input instanceof Request ? input : new Request(new URL(String(input), location.origin), init);
    const url = new URL(req.url);
    paths.push(url);
    if (url.pathname.endsWith("/config")) return Response.json({ liffId: "test" });
    if (url.pathname.endsWith("/session")) return Response.json({ user: { userId: "u1", displayName: "Tester" }, csrfToken: "c".repeat(43), expiresAt: new Date(Date.now() + 1800000).toISOString() });
    if (url.pathname.endsWith("/repositories")) return Response.json({items:["a","b"].map(id=>({id,name:id,status:"ACTIVE"})),nextCursor:""});
    return Response.json({items:[],nextCursor:""});
  };
  const sdk = {
    init: async () => {}, isLoggedIn: () => true, isInClient: () => false,
    getIDToken: () => "test-token", getProfile: async () => ({ userId: "u1" }),
    login() {}, logout() {}, closeWindow() {},
  };
  const root = createRoot(document.getElementById("root")!);
  try {
    await act(async () => {
      root.render(<SessionProvider sdkLoader={async () => sdk}><JournalWorkspace dashboard={dashboard}/></SessionProvider>);
      await flush();
    });
    const select = document.querySelector("select")!;
    assert.equal(select.value,"a");
    await act(async () => {
      select.value="b";
      select.dispatchEvent(new Event("change",{bubbles:true}));
      await flush();
    });
    const queries = paths.filter(url=>url.pathname.endsWith("/journals")).map(url=>url.searchParams);
    assert.deepEqual(queries.map(query=>query.get("repositoryId")),["a","b"]);
    assert.ok(queries.every(query=>query.get("includeProgress") === (dashboard ? "true" : null)));
    assert.ok(paths.every(url=>!url.pathname.includes("/groups") && !url.searchParams.has("groupId")));
  } finally {
    await act(async()=>root.unmount());
    globalThis.fetch=original;
  }
});
