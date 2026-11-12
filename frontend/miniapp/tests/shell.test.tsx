import React, { act } from "react";
import { createRoot } from "react-dom/client";
import { test } from "node:test";
import assert from "node:assert/strict";
import { Shell } from "../src/shell/Shell";
import { useSession } from "../src/identity/Session";

// jsdom lacks the browser surfaces used by Radix portals and sidebar sizing.
for (const key of ["DocumentFragment", "Element", "Node", "CustomEvent", "MutationObserver", "HTMLInputElement", "NodeFilter"] as const) {
  Object.defineProperty(globalThis, key, { value: window[key], configurable: true });
}
Object.assign(globalThis, {
  getComputedStyle: window.getComputedStyle.bind(window),
  requestAnimationFrame: (callback: FrameRequestCallback) => setTimeout(() => callback(0), 0),
  cancelAnimationFrame: clearTimeout,
  ResizeObserver: class { observe() {} unobserve() {} disconnect() {} },
});
window.matchMedia = (query) => ({ matches: false, media: query, onchange: null,
  addListener() {}, removeListener() {}, addEventListener() {}, removeEventListener() {}, dispatchEvent() { return false; } });

function PrivateContent() {
  const { user } = useSession();
  return <p id="private-content">{user ? "私人內容" : "尚未驗證"}</p>;
}
const flush = async () => {
  for (let i = 0; i < 12; i++) await new Promise<void>((resolve) => setImmediate(resolve));
};

for (const width of [1024, 375]) test(`production account menu at ${width}px retries initialization, verifies and retries logout`, async () => {
  Object.defineProperty(window, "innerWidth", { value: width, configurable: true });
  window.history.replaceState(null, "", "/miniapp/review");
  document.body.innerHTML = '<div id="root"></div>';
  const originalFetch = globalThis.fetch, originalSDK = window.liff;
  let initialization = 0, login = 0, logout = 0, loggedIn = false, revocations = 0;
  window.liff = {
    init: async () => { if (++initialization === 1) throw new Error("初始化暫時失敗"); },
    isLoggedIn: () => loggedIn,
    isInClient: () => false,
    getIDToken: () => "test-token",
    login: () => { login++; },
    logout: () => { logout++; loggedIn = false; },
    closeWindow() {},
  };
  globalThis.fetch = async (input, init) => {
    const request = input instanceof Request ? input : new Request(new URL(String(input), location.origin), init);
    if (new URL(request.url).pathname.endsWith("/config")) return Response.json({ liffId: "test-shell" });
    if (request.method === "DELETE") {
      if (++revocations === 1) return Response.json({ error: { message: "retry" } }, { status: 503 });
      return new Response(null, { status: 204 });
    }
    return Response.json({ user: { userId: "u1", displayName: "Tester" }, csrfToken: "c".repeat(43), expiresAt: new Date(Date.now() + 1800000).toISOString() });
  };
  const root = createRoot(document.getElementById("root")!);
  const showSidebar = async () => {
    if (!document.querySelector('[aria-label="身分選單"]')) {
      await act(async () => {
        document.querySelector<HTMLElement>('[aria-label="切換側欄"]')!.click();
        await flush();
      });
    }
  };
  const openMenu = async () => {
    await showSidebar();
    await act(async () => {
      document.querySelector<HTMLElement>('[aria-label="身分選單"]')!.dispatchEvent(new window.KeyboardEvent("keydown", { key: "Enter", bubbles: true }));
      await flush();
    });
  };
  const click = async (id: string) => {
    await act(async () => { document.getElementById(id)!.click(); await flush(); });
  };
  try {
    await act(async () => { root.render(<Shell title="授權審核" workflow="review"><PrivateContent /></Shell>); await flush(); });
    await showSidebar();
    assert.match(document.getElementById("mini-status")!.textContent!, /初始化暫時失敗/);
    await openMenu();
    assert.equal(document.getElementById("verify")!.textContent, "重試初始化");
    await click("verify");
    await openMenu();
    assert.ok(document.getElementById("login"));
    await click("login");
    assert.equal(login, 1);
    loggedIn = true;
    await openMenu();
    await click("verify");
    assert.equal(document.getElementById("private-content")!.textContent, "私人內容");
    await openMenu();
    assert.equal(document.getElementById("login"), null);
    assert.ok(document.getElementById("logout"));
    await click("logout");
    assert.equal(document.getElementById("private-content")!.textContent, "尚未驗證");
    assert.equal(logout, 0, "SDK logout must wait for server revocation");
    await openMenu();
    assert.ok(document.getElementById("logout"), "failed server revocation must remain retryable");
    await click("logout");
    assert.equal(revocations, 2);
    assert.equal(logout, 1);
    await openMenu();
    assert.equal(document.getElementById("logout"), null);
    assert.ok(document.getElementById("login"));
  } finally {
    await act(async () => { root.unmount(); await flush(); });
    globalThis.fetch = originalFetch;
    window.liff = originalSDK;
  }
});
