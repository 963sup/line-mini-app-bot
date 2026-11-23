import { useOrganizationSelection } from "../src/business/organization/OrganizationSelection";
import React, { act } from "react";
import Router from "next/router";
import { MiniAppLayout } from "../src/shell/MiniAppProviders";
import { createRoot } from "react-dom/client";
import { test } from "node:test";
import assert from "node:assert/strict";
import { Enterprise } from "../src/business/enterprise/Enterprise";
import { Shell } from "./shell-fixture";
import { Shell as ProductionShell } from "../src/shell/Shell";
import { useSession } from "../src/identity/Session";

import "./sidebar-dom";

function PrivateContent() {
  const { user, pictureUrl } = useSession();
  return (
    <p id="private-content" data-avatar={pictureUrl}>
      {user ? "私人內容" : "尚未驗證"}
    </p>
  );
}
const flush = async () => {
  for (let i = 0; i < 12; i++) await new Promise<void>((resolve) => setImmediate(resolve));
};

for (const width of [1024, 375])
  test(`production account menu at ${width}px retries initialization, verifies and retries logout`, async () => {
    Object.defineProperty(window, "innerWidth", { value: width, configurable: true });
    window.history.replaceState(null, "", "/miniapp/review");
    document.body.innerHTML = '<div id="root"></div>';
    const originalFetch = globalThis.fetch,
      originalSDK = window.liff;
    let organizationRevoked = false,
      organizationReads = 0;
    let initialization = 0,
      login = 0,
      logout = 0,
      loggedIn = false,
      revocations = 0;
    let failInitialization: (() => void) | undefined;
    window.liff = {
      init: async () => {
        if (++initialization === 1)
          await new Promise<void>((_, reject) => {
            failInitialization = () => reject(new Error("初始化暫時失敗"));
          });
      },
      isLoggedIn: () => loggedIn,
      isInClient: () => false,
      getProfile: async () => ({
        userId: "u1",
        pictureUrl: "https://profile.line-scdn.net/test-avatar",
      }),
      getIDToken: () => "test-token",
      login: () => {
        login++;
      },
      logout: () => {
        logout++;
        loggedIn = false;
      },
      closeWindow() {},
    };
    globalThis.fetch = async (input, init) => {
      const request =
        input instanceof Request
          ? input
          : new Request(new URL(String(input), location.origin), init);
      if (new URL(request.url).pathname === "/api/v1/me/organization-context") {
        if (request.method === "PUT") {
          assert.deepEqual(await request.json(), { organizationId: "org2" });
          organizationRevoked = true;
          return Response.json(
            { error: { code: "not_found", message: "找不到組織。" } },
            { status: 404 },
          );
        }
        organizationReads++;
        const options = (organizationRevoked ? ["org1"] : ["org1", "org2"]).map((id) => ({
          organizationId: id,
          name: "台北營運",
          detail: id,
        }));
        return Response.json({
          enterprises: [{ id: "hq", name: "總公司" }],
          canCreateEnterprise: !organizationRevoked,
          groups: [{ enterpriseId: "hq", label: "總公司", organizations: options }],
          selected: { organizationId: "org1", name: "台北營運", groupLabel: "總公司" },
          canSwitch: options.length > 1,
        });
      }
      if (new URL(request.url).pathname.endsWith("/config"))
        return Response.json({ liffId: "test-shell" });
      if (request.method === "DELETE") {
        if (++revocations === 1)
          return Response.json({ error: { message: "retry" } }, { status: 503 });
        return new Response(null, { status: 204 });
      }
      return Response.json({
        user: { userId: "u1", displayName: "Tester" },
        csrfToken: "c".repeat(43),
        expiresAt: new Date(Date.now() + 1800000).toISOString(),
      });
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
        document
          .querySelector<HTMLElement>('[aria-label="身分選單"]')!
          .dispatchEvent(new window.KeyboardEvent("keydown", { key: "Enter", bubbles: true }));
        await flush();
      });
    };
    const click = async (id: string) => {
      await act(async () => {
        document.getElementById(id)!.click();
        await flush();
      });
    };
    try {
      function EnterpriseCreation() {
        const { context, error } = useOrganizationSelection();
        return (
          <Enterprise
            id="new"
            canCreate={context?.canCreateEnterprise}
            eligibilityError={error}
            onSaved={() => {}}
          />
        );
      }
      await act(async () => {
        root.render(
          <Shell title="授權審核" workflow="review">
            <PrivateContent />
            <EnterpriseCreation />
          </Shell>,
        );
        await flush();
      });
      const notice = () => document.querySelector<HTMLElement>('[aria-label="LINE 身分狀態"]');
      assert.equal(notice()!.getAttribute("role"), "status");
      assert.equal(notice()!.getAttribute("aria-busy"), "true");
      assert.match(notice()!.textContent!, /正在初始化 LINE/);
      await act(async () => {
        failInitialization!();
        await flush();
      });
      assert.equal(notice()!.getAttribute("aria-busy"), "false");
      assert.match(notice()!.textContent!, /初始化暫時失敗/);
      assert.match(notice()!.textContent!, /重試初始化/);
      await showSidebar();
      assert.equal(document.querySelector('[aria-label="切換工作區"]'), null);
      assert.match(
        document.querySelector('[aria-label="組織切換狀態"]')!.textContent!,
        /初始化暫時失敗/,
      );
      await openMenu();
      assert.equal(document.getElementById("verify")!.textContent, "重試初始化");
      await click("verify");
      assert.match(notice()!.textContent!, /請登入 LINE/);
      await openMenu();
      assert.ok(document.getElementById("login"));
      await click("login");
      assert.equal(login, 1);
      loggedIn = true;
      await openMenu();
      await click("verify");
      assert.equal(document.getElementById("private-content")!.textContent, "私人內容");
      assert.equal(notice(), null, "verified users must not retain an authentication notice");
      assert.equal(
        document.getElementById("private-content")!.getAttribute("data-avatar"),
        "https://profile.line-scdn.net/test-avatar",
      );
      await showSidebar();
      await act(async () => {
        document
          .querySelector<HTMLElement>('[aria-label="切換組織"]')!
          .dispatchEvent(new window.KeyboardEvent("keydown", { key: "Enter", bubbles: true }));
        await flush();
      });
      assert.ok(document.querySelector('a[href="/miniapp?enterprise=hq"]'));
      assert.ok(document.querySelector('a[href="/miniapp?enterprise=new"]'));
      assert.equal(organizationReads, 1);
      assert.match(document.body.textContent!, /建立企業/);
      const choices = [...document.querySelectorAll<HTMLElement>('[role="menuitemradio"]')];
      assert.equal(choices.length, 2);
      assert.equal(choices[0].getAttribute("aria-checked"), "true");
      assert.match(choices[1].textContent!, /org2/);
      await act(async () => {
        choices[1].click();
        await flush();
      });
      assert.ok(organizationRevoked);
      assert.equal(organizationReads, 2);
      assert.match(document.body.textContent!, /帳號尚無可用的企業開通資格|找不到組織/);
      assert.ok(
        ![...document.querySelectorAll("button")].some((b) => b.textContent === "建立企業"),
      );
      assert.equal(document.querySelector('a[href="/miniapp?enterprise=new"]'), null);
      assert.equal(document.querySelector('[aria-label="切換組織"]'), null);
      await openMenu();
      assert.equal(document.getElementById("login"), null);
      assert.ok(document.getElementById("logout"));
      await click("logout");
      assert.equal(document.getElementById("private-content")!.textContent, "尚未驗證");
      assert.equal(document.getElementById("private-content")!.getAttribute("data-avatar"), null);
      assert.equal(logout, 0, "SDK logout must wait for server revocation");
      await openMenu();
      assert.ok(
        document.getElementById("logout"),
        "failed server revocation must remain retryable",
      );
      await click("logout");
      assert.equal(revocations, 2);
      assert.equal(logout, 1);
      await openMenu();
      assert.equal(document.getElementById("logout"), null);
      assert.ok(document.getElementById("login"));
    } finally {
      await act(async () => {
        root.unmount();
        await flush();
      });
      globalThis.fetch = originalFetch;
      window.liff = originalSDK;
    }
  });

test("changing organizations keeps the authenticated context across Next.js pages", async (t) => {
  function ReviewPage() {
    return (
      <ProductionShell title="審核" workflow="review">
        <PrivateContent />
      </ProductionShell>
    );
  }
  function OrganizationPage() {
    return (
      <ProductionShell title="組織" workflow="home">
        <PrivateContent />
      </ProductionShell>
    );
  }
  Object.defineProperty(window, "innerWidth", { value: 1024, configurable: true });
  window.history.replaceState(null, "", "/miniapp/review");
  document.body.innerHTML = '<div id="root"></div>';
  const root = createRoot(document.getElementById("root")!);
  const originalFetch = globalThis.fetch,
    originalSDK = window.liff;
  t.after(async () => {
    await act(async () => root.unmount());
    globalThis.fetch = originalFetch;
    window.liff = originalSDK;
  });

  let startSession!: () => void;
  const sessionPending = new Promise<void>((resolve) => {
    startSession = resolve;
  });
  let sessions = 0,
    reads = 0,
    writes = 0;
  const organizations = [
    { organizationId: "org1", name: "台北營運" },
    { organizationId: "org2", name: "高雄營運" },
  ];
  const context = (index: number) =>
    Response.json({
      groups: [{ label: "獨立組織", organizations }],
      selected: organizations[index],
      canSwitch: true,
    });
  window.liff = {
    init: async () => {},
    isLoggedIn: () => true,
    isInClient: () => true,
    getProfile: async () => ({ userId: "u1" }),
    getIDToken: () => "test-token",
    login() {},
    logout() {},
    closeWindow() {},
  };
  globalThis.fetch = async (input, init) => {
    const request =
      input instanceof Request ? input : new Request(new URL(String(input), location.origin), init);
    switch (new URL(request.url).pathname) {
      case "/api/v1/miniapp/config":
        return Response.json({ liffId: "test-shell" });
      case "/api/v1/session":
        sessions++;
        await sessionPending;
        return Response.json({
          user: { userId: "u1" },
          csrfToken: "test",
          expiresAt: new Date(Date.now() + 1800000).toISOString(),
        });
      case "/api/v1/me/organization-context":
        if (request.method === "PUT") {
          writes++;
          assert.deepEqual(await request.json(), { organizationId: "org2" });
          return context(1);
        }
        reads++;
        return context(0);
      default:
        throw new Error("unexpected request: " + request.url);
    }
  };

  function render(Component: React.ComponentType, pathname: string) {
    root.render(
      <MiniAppLayout pathname={pathname}>
        <Component />
      </MiniAppLayout>,
    );
  }
  const push = t.mock.method(Router, "push", async (href: string) => {
    assert.equal(href, "/miniapp?organization=org2");
    window.history.pushState(null, "", href);
    render(OrganizationPage, "/miniapp");
    return true;
  });
  await act(async () => {
    render(ReviewPage, "/miniapp/review");
    await flush();
  });
  const loading = document.querySelector('[aria-label="組織切換狀態"]');
  assert.equal(loading?.getAttribute("aria-busy"), "true");
  assert.match(loading?.textContent ?? "", /正在驗證身分/);
  await act(async () => {
    startSession();
    await flush();
  });
  await act(async () => {
    document
      .querySelector('[aria-label="切換組織"]')!
      .dispatchEvent(new window.KeyboardEvent("keydown", { key: "Enter", bubbles: true }));
    await flush();
  });
  assert.equal(writes, 0);
  assert.equal(push.mock.callCount(), 0);
  const choice = [...document.querySelectorAll<HTMLElement>('[role="menuitemradio"]')].find(
    (node) => node.textContent?.trim() === "高雄營運",
  );
  assert.ok(choice);
  await act(async () => {
    choice.click();
    await flush();
  });
  assert.equal(push.mock.callCount(), 1);
  assert.equal(writes, 1);
  assert.equal(sessions, 1);
  assert.equal(reads, 1);
  assert.match(document.querySelector('[aria-label="切換組織"]')!.textContent!, /高雄營運/);
  assert.equal(document.querySelector('[aria-label="組織切換狀態"]'), null);
});
