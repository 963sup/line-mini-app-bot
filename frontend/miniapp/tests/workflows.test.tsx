import { QuantityHistory } from "../src/business/object/QuantityHistory";
import { OrganizationPage } from "../src/business/organization/OrganizationPage";
import { OrganizationTeams } from "../src/business/team/OrganizationTeams";
import { pressButton } from "./sidebar_dom";
import { Enterprise } from "../src/business/enterprise/Enterprise";
import React from "react";
import { test } from "node:test";
import assert from "node:assert/strict";
import { act, StrictMode, type ReactNode } from "react";
import { createRoot } from "react-dom/client";
import { SessionProvider, useSession } from "../src/identity/Session";
import { SessionControls } from "./session_controls";
import { attendanceLaunchToken } from "../src/business/attendance/launchToken";
import { View as Punch } from "../src/business/attendance/punch/View";
import { View as Leave } from "../src/business/attendance/leave/View";
import { View as History } from "../src/business/attendance/records/View";
import { View as Correction } from "../src/business/attendance/corrections/View";
import { View as Geo } from "../src/diagnostics/geolocation/View";
import { View as Review } from "../src/business/attendance/approvals/View";
import { locate } from "../src/shared/location/position";
const launch = "b".repeat(64),
  requestID = "a".repeat(64);
const record = {
  RecordID: "r1",
  GroupID: "g1",
  ClockInAt: "2026-11-01T01:00:00Z",
  ClockOutAt: "2026-11-01T02:00:00Z",
  Revision: 1,
  EventSequence: 2,
  Breaks: [],
  RestConfirmed: true,
};
const flush = async () => {
  for (let i = 0; i < 8; i++) await new Promise<void>((resolve) => setImmediate(resolve));
};
function Epoch({ children }: { children: ReactNode }) {
  const { epoch } = useSession();
  return <div key={epoch}>{children}</div>;
}
async function fixture(
  view: ReactNode,
  search = "",
  override?: (request: Request, next: (r: Request) => Promise<Response>) => Promise<Response>,
  client = true,
) {
  window.history.replaceState(null, "", "/miniapp" + search);
  document.body.innerHTML = '<div id="root"></div>';
  let init = 0,
    close = 0,
    logout = 0;
  const sdkEvents: string[] = [];
  const requests: Request[] = [],
    watchers: {
      ok: PositionCallback;
      fail: PositionErrorCallback | null;
      cleared: boolean;
    }[] = [];
  const sdk = {
    init: async () => {
      init++;
    },
    isLoggedIn: () => true,
    isInClient: () => client,
    getProfile: async () => ({
      userId: "u1",
      pictureUrl: "https://profile.line-scdn.net/test-avatar",
    }),
    getIDToken: () => "token",
    login: () => {
      sdkEvents.push("login");
    },
    logout: () => {
      logout++;
      sdkEvents.push("logout");
    },
    closeWindow: () => {
      close++;
    },
  };
  const loader = async () => sdk;
  Object.defineProperty(navigator, "geolocation", {
    configurable: true,
    value: {
      watchPosition(ok: PositionCallback, fail: PositionErrorCallback | null) {
        watchers.push({ ok, fail, cleared: false });
        return watchers.length;
      },
      clearWatch(id: number) {
        watchers[id - 1].cleared = true;
      },
    },
  });
  let handler: (r: Request) => Promise<Response> = async (r) => {
    const path = new URL(r.url).pathname;
    if (path.endsWith("/config")) return Response.json({ liffId: "test" });
    if (path.endsWith("/session/refresh"))
      return Response.json({
        user: { userId: "u1", displayName: "Tester" },
        csrfToken: "c".repeat(43),
        expiresAt: new Date(Date.now() + 1800000).toISOString(),
      });
    if (path.endsWith("/session")) {
      if (r.method === "DELETE") return new Response(null, { status: 204 });
      return Response.json({
        user: { userId: "u1", displayName: "Tester" },
        csrfToken: "c".repeat(43),
        expiresAt: new Date(Date.now() + 1800000).toISOString(),
      });
    }
    if (path.endsWith("/launch")) return Response.json({ requestId: requestID });
    if (path.endsWith("/punches/preflight"))
      return Response.json({ code: "needs_position", accepted: false, message: "需要位置" });
    if (path.endsWith("/punches"))
      return Response.json({ code: "clocked_in", accepted: true, message: "success" });
    if (path.endsWith("/groups/page"))
      return Response.json({
        items: [{ id: "g1", name: "Group" }],
        nextCursor: "",
      });
    if (path.endsWith("/records/r1")) return Response.json(record);
    if (path.endsWith("/records")) return Response.json({ items: [record], nextCursor: "" });
    if (path.endsWith("/summary"))
      return Response.json({
        Calculation: { Minutes: 60, Pending: 0 },
        count: 1,
      });
    if (path.endsWith("/review"))
      return Response.json({
        reason: "reason",
        start: record.ClockInAt,
        end: record.ClockOutAt,
        breaks: [],
        restConfirmed: true,
      });
    return Response.json({ id: "c1", status: "accepted" });
  };
  if (override) {
    const next = handler;
    handler = (r) => override(r, next);
  }
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async (input, init) => {
    const r =
      input instanceof Request
        ? input
        : new Request(typeof input === "string" ? new URL(input, location.origin) : input, init);
    requests.push(r.clone());
    return handler(r);
  };
  const root = createRoot(document.getElementById("root")!);
  await act(async () => {
    root.render(
      <StrictMode>
        <SessionProvider sdkLoader={loader}>
          <SessionControls />
          <Epoch>{view}</Epoch>
        </SessionProvider>
      </StrictMode>,
    );
    await flush();
  });
  const node = <T extends HTMLElement>(id: string) => document.getElementById(id) as T;
  const click = async (id: string) => {
    await act(async () => {
      node(id).click();
      await flush();
    });
  };
  const input = async (id: string, value: string) => {
    await act(async () => {
      const n = node<HTMLInputElement>(id);
      const prototype =
        n.tagName === "TEXTAREA"
          ? window.HTMLTextAreaElement.prototype
          : window.HTMLInputElement.prototype;
      Object.getOwnPropertyDescriptor(prototype, "value")!.set!.call(n, value);
      n.dispatchEvent(new Event("input", { bubbles: true }));
      await flush();
    });
  };
  return {
    node,
    click,
    input,
    requests,
    sdkEvents,
    watchers,
    get init() {
      return init;
    },
    get close() {
      return close;
    },
    get logout() {
      return logout;
    },
    setHandler(fn: typeof handler) {
      const previous = handler;
      handler = fn;
      return previous;
    },
    async dispose() {
      await act(async () => {
        root.unmount();
        await flush();
      });
      globalThis.fetch = originalFetch;
    },
  };
}
const position = {
  coords: { latitude: 25, longitude: 121, accuracy: 5 },
  timestamp: Date.now(),
} as GeolocationPosition;
test("Strict Mode initializes SDK once, auto punches once, ignores duplicate callbacks and clears state on logout", async () => {
  const f = await fixture(<Punch />, "?launch=" + launch, async (r, next) => {
    if (r.url.endsWith("/punches"))
      return Response.json({ code: "new_server_action", accepted: true, message: "success" });
    const response = await next(r);
    if (r.url.endsWith("/session") && r.method === "POST") {
      const value = await response.json();
      return Response.json({ ...value, expiresAt: new Date(Date.now() + 300000).toISOString() });
    }
    return response;
  });
  try {
    assert.equal(f.init, 1);
    assert.equal(f.requests.filter((r) => r.url.endsWith("/session/refresh")).length, 1);
    assert.equal(f.watchers.length, 1);
    assert.equal(f.requests.filter((r) => r.url.endsWith("/launch")).length, 1);
    await act(async () => {
      f.watchers[0].ok(position);
      f.watchers[0].ok(position);
      await flush();
    });
    assert.equal(f.close, 1);
    assert.equal(f.requests.filter((r) => r.url.endsWith("/punches")).length, 1);
    const punchRequest = f.requests.find((r) => r.url.endsWith("/punches"))!;
    assert.equal(punchRequest.credentials, "same-origin");
    assert.equal(punchRequest.headers.get("Authorization"), null);
    assert.equal(punchRequest.headers.get("X-CSRF-Token"), "c".repeat(43));
    assert.equal(f.watchers[0].cleared, true);
    assert.equal(f.node("punch-status").textContent, "success");
    const paths = f.requests.map((r) => new URL(r.url).pathname);
    assert.ok(
      paths.indexOf("/api/v1/attendance/punches/preflight") <
        paths.indexOf("/api/v1/attendance/punches"),
    );
    // A completed submission is recovered before any new location request.
    const previous = f.setHandler(async (r) =>
      r.url.endsWith("/punches/preflight")
        ? Response.json({ code: "clocked_in", accepted: true, message: "saved result" })
        : previous(r),
    );
    await f.click("punch");
    assert.equal(f.watchers.length, 1);
    assert.equal(f.node("punch-status").textContent, "saved result");
    assert.equal(f.close, 1);
    f.setHandler(async (r) =>
      r.url.endsWith("/punches/preflight")
        ? Response.json({ code: "already_in", accepted: false, message: "already clocked in" })
        : previous(r),
    );
    await f.click("punch");
    assert.equal(f.watchers.length, 1);
    assert.equal(f.requests.filter((r) => r.url.endsWith("/punches")).length, 1);
    assert.equal(f.node("punch-status").textContent, "already clocked in");
    assert.equal(f.close, 1);
    await f.click("verify");
    assert.equal(f.watchers.length, 1);
    await f.click("logout");
    assert.equal(f.node<HTMLDListElement>("identity").hidden, true);
    assert.equal(f.logout, 0);
    assert.match(f.node("punch-status").textContent!, /清除/);
  } finally {
    await f.dispose();
  }
});
test("ambiguous punch retry preserves request ID, pagehide aborts and ignores late coordinates", async () => {
  const f = await fixture(<Punch />, "?launch=" + launch);
  try {
    const previous = f.setHandler(async (r) => {
      if (r.url.endsWith("/punches")) throw new TypeError("network");
      return previous(r);
    });
    await act(async () => {
      f.watchers[0].ok(position);
      await flush();
    });
    assert.match(f.node("punch-status").textContent!, /結果尚未確認/);
    await f.click("punch");
    assert.equal(f.requests.filter((r) => r.url.endsWith("/launch")).length, 1);
    await act(async () => {
      f.watchers[1].ok(position);
      await flush();
    });
    const bodies = await Promise.all(
      f.requests.filter((r) => r.url.endsWith("/punches")).map((r) => r.json()),
    );
    assert.equal(bodies.length, 2);
    assert.equal(bodies[0].requestId, bodies[1].requestId);
    await f.click("punch");
    await act(async () => {
      window.dispatchEvent(new Event("pagehide"));
      await flush();
    });
    assert.equal(f.watchers[2].cleared, true);
    await act(async () => {
      f.watchers[2].ok(position);
      await flush();
    });
    assert.equal(f.requests.filter((r) => r.url.endsWith("/punches")).length, 2);
  } finally {
    await f.dispose();
  }
});
test("geolocation abort, permission failure, invalid coordinates and synchronous callback clean watchers", async () => {
  let ok: PositionCallback = () => {},
    fail: PositionErrorCallback | null = null,
    cleared = 0;
  Object.defineProperty(navigator, "geolocation", {
    configurable: true,
    value: {
      watchPosition(success: PositionCallback, error: PositionErrorCallback) {
        ok = success;
        fail = error;
        return 7;
      },
      clearWatch() {
        cleared++;
      },
    },
  });
  const c = new AbortController(),
    pending = locate(c.signal);
  c.abort();
  await assert.rejects(pending, { name: "AbortError" });
  assert.equal(cleared, 1);
  ok(position);
  const denied = locate(new AbortController().signal);
  fail!({ code: 1 } as GeolocationPositionError);
  await assert.rejects(denied, /權限/);
  assert.equal(cleared, 2);
  const invalid = locate(new AbortController().signal);
  ok({ ...position, coords: { ...position.coords, latitude: 100 } });
  await assert.rejects(invalid, /無效/);
  Object.defineProperty(navigator, "geolocation", {
    configurable: true,
    value: {
      watchPosition(success: PositionCallback) {
        success(position);
        return 8;
      },
      clearWatch() {
        cleared++;
      },
    },
  });
  await locate(new AbortController().signal);
  assert.equal(cleared, 4);
  assert.equal(attendanceLaunchToken("?request=" + launch, "https://example.test", "launch"), "");
  assert.equal(
    attendanceLaunchToken(
      "?liff.state=" + encodeURIComponent("?launch=" + launch),
      "https://example.test",
      "launch",
    ),
    launch,
  );
});
test("personal history ignores legacy group filters and switches months automatically", async () => {
  let decemberReads = 0;
  const f = await fixture(<History />, "?group=g1&month=2026-11", async (r, next) => {
    const url = new URL(r.url);
    if (!url.pathname.endsWith("/records")) return next(r);
    if (url.searchParams.get("month") === "2026-12") {
      if (++decemberReads === 1) return new Response(null, { status: 503 });
      return Response.json({ items: [], nextCursor: "" });
    }
    return Response.json({
      items: [record, { ...record, RecordID: "r2", GroupID: "g2" }],
      nextCursor: "",
    });
  });
  try {
    const initial = f.requests.find((r) => new URL(r.url).pathname.endsWith("/records"))!;
    assert.equal(new URL(initial.url).searchParams.has("group"), false);
    assert.equal(new URLSearchParams(location.search).has("group"), false);
    assert.match(f.node("record-list").innerHTML, /group=g2.*record=r2/);
    assert.equal(document.getElementById("next-records"), null);
    await act(async () => {
      const input = f.node<HTMLInputElement>("history-month");
      const setter = Object.getOwnPropertyDescriptor(
        window.HTMLInputElement.prototype,
        "value",
      )!.set!;
      setter.call(input, "2026-12");
      input.dispatchEvent(new Event("input", { bubbles: true }));
      await flush();
    });
    assert.equal(new URLSearchParams(location.search).get("month"), "2026-12");
    assert.match(f.node("history").textContent!, /無法載入出勤紀錄/);
    assert.doesNotMatch(f.node("history").textContent!, /本月沒有出勤紀錄/);
    await act(async () => {
      [...f.node("history").querySelectorAll("button")]
        .find((button) => button.textContent === "重試讀取")!
        .click();
      await flush();
    });
    assert.equal(decemberReads, 2);
    assert.match(f.node("history").textContent!, /本月沒有出勤紀錄/);
    assert.equal(document.getElementById("record-list"), null);
    await f.click("logout");
    assert.equal(document.getElementById("record-list"), null);
  } finally {
    await f.dispose();
  }
});
test("review controls and details clear on logout", async () => {
  const f = await fixture(<Review />);
  try {
    assert.equal(f.node<HTMLButtonElement>("approve-correction").disabled, true);
    await f.click("logout");
    assert.equal(f.node<HTMLFieldSetElement>("review-controls").disabled, true);
    assert.equal(f.node("review-details").textContent, "");
  } finally {
    await f.dispose();
  }
});
test("personal history navigates with cursors without loading a group summary", async () => {
  const f = await fixture(<History />, "?group=g1&month=2026-11", async (r, next) => {
    if (new URL(r.url).pathname.endsWith("/records"))
      return Response.json({
        items: [
          {
            ...record,
            RecordID: new URL(r.url).searchParams.get("cursor") ? "page2" : "page1",
          },
        ],
        nextCursor: new URL(r.url).searchParams.get("cursor") ? "" : "cursor2",
      });
    return next(r);
  });
  try {
    assert.match(f.node("record-list").innerHTML, /page1/);
    await f.click("next-records");
    assert.match(f.node("record-list").innerHTML, /page2/);
    assert.equal(f.node("record-page").textContent, "第 2 頁");
    await f.click("previous-records");
    assert.match(f.node("record-list").innerHTML, /page1/);
    assert.equal(
      f.requests.filter((r) => r.url.includes("/monthly/summary") || r.url.includes("/groups/page"))
        .length,
      0,
    );
  } finally {
    await f.dispose();
  }
});
test("correction double submit and ambiguous retry retain the original key and body", async () => {
  const f = await fixture(<Correction />, "?group=g1&record=r1", async (r, next) =>
    new URL(r.url).pathname.endsWith("/records/r1")
      ? Response.json({ ...record, ClockOutAt: "0001-01-01T00:00:00Z" })
      : r.url.endsWith("/corrections")
        ? Promise.reject(new TypeError("network"))
        : next(r),
  );
  try {
    assert.equal(document.getElementById("correction-start"), null);
    await f.input("correction-end", "2026-11-01T18:00");
    await f.input("correction-reason", "Reason");
    await f.click("correction-confirm");
    const submit = () =>
      f
        .node("correction-form")
        .dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await act(async () => {
      submit();
      submit();
      await flush();
    });
    assert.equal(f.requests.filter((r) => r.url.endsWith("/corrections")).length, 1);
    await act(async () => {
      submit();
      await flush();
    });
    let posts = f.requests.filter((r) => r.url.endsWith("/corrections"));
    assert.equal(posts.length, 2);
    assert.equal(posts[0].headers.get("Idempotency-Key"), posts[1].headers.get("Idempotency-Key"));
    assert.equal((await posts[0].clone().json()).reason, "Reason");
    assert.equal((await posts[0].clone().json()).start, new Date(record.ClockInAt).toISOString());
    await f.input("correction-reason", "Changed");
    await act(async () => {
      submit();
      await flush();
    });
    posts = f.requests.filter((r) => r.url.endsWith("/corrections"));
    assert.equal(posts.length, 3);
    assert.deepEqual(await posts[2].clone().json(), await posts[1].clone().json());
    assert.equal(posts[1].headers.get("Idempotency-Key"), posts[2].headers.get("Idempotency-Key"));
  } finally {
    await f.dispose();
  }
  // Exercise the same submission flow without an original shift.
  const missing = await fixture(<Correction />, "?group=g1", async (r, next) => {
    if (new URL(r.url).pathname.endsWith("/corrections")) {
      if (r.method === "POST") return Response.json({ id: "missing1", status: "accepted" });
      return Response.json({
        items: [
          {
            id: "missing1",
            groupId: "g1",
            recordId: "correction-missing1",
            start: record.ClockInAt,
            end: record.ClockOutAt,
            reason: "整班漏卡",
            status: "rejected",
            decisionReason: "請提供出勤佐證",
          },
        ],
        nextCursor: "",
      });
    }
    return next(r);
  });
  try {
    await missing.input("correction-start", "2026-11-01T09:00");
    await missing.input("correction-end", "2026-11-01T18:00");
    await missing.input("correction-reason", "整班漏卡");
    await missing.click("correction-confirm");
    await act(async () => {
      missing
        .node("correction-form")
        .dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
      await flush();
    });
    const post = missing.requests.find(
      (r) => r.method === "POST" && new URL(r.url).pathname.endsWith("/corrections"),
    );
    assert.ok(post);
    const body = await post.clone().json();
    assert.equal(body.groupId, "g1");
    assert.equal("recordId" in body, false);
    assert.match(missing.node("correction-status").textContent!, /等待審核/);
    assert.match(missing.node("requests-list").textContent!, /已駁回/);
    assert.match(missing.node("requests-list").textContent!, /請提供出勤佐證/);
  } finally {
    await missing.dispose();
  }
});
test("editing review ID rejects a late proposal and prevents approval of another ID", async () => {
  let decision = "pending";
  let resolve!: (response: Response) => void;
  const f = await fixture(<Review />, "", async (r, next) =>
    r.url.endsWith("/reviewer-groups")
      ? Response.json([{ id: "g1", name: "Group" }])
      : new URL(r.url).pathname.endsWith("/corrections")
        ? Response.json({
            items: [
              {
                id: "c2",
                status: decision,
                reason: "漏下班",
                start: record.ClockInAt,
                end: record.ClockOutAt,
              },
            ],
            nextCursor: "",
          })
        : r.url.endsWith("/review")
          ? new Promise<Response>((done) => (resolve = done))
          : next(r),
  );
  try {
    const listRequest = f.requests.find(
      ({ url }) => new URL(url).searchParams.get("scope") === "review",
    );
    assert.ok(listRequest);
    assert.equal(new URL(listRequest.url).searchParams.has("group"), false);
    assert.equal(
      f.requests.some(({ url }) => url.endsWith("/reviewer-groups")),
      false,
    );
    assert.match(f.node("requests-list").textContent!, /待審核/);
    await f.input("correction-id", "c1");
    await f.click("review-correction");
    await f.input("correction-id", "c2");
    await act(async () => {
      resolve(Response.json({ reason: "obsolete" }));
      await flush();
    });
    assert.equal(f.node("review-details").textContent, "");
    assert.equal(f.node<HTMLButtonElement>("approve-correction").disabled, true);
    assert.equal(f.requests.filter((r) => r.url.endsWith("/approve")).length, 0);
    await f.click("review-correction");
    await act(async () => {
      resolve(
        Response.json({
          id: "c2",
          reason: "漏下班",
          start: record.ClockInAt,
          end: record.ClockOutAt,
          breaks: [],
          restConfirmed: true,
          status: "pending",
          canApprove: true,
          canReject: true,
          canCancel: false,
        }),
      );
      await flush();
    });
    await f.input("rejection-reason", "請提供出勤佐證");
    const original = f.setHandler(async (r) =>
      r.url.endsWith("/reject")
        ? ((decision = "rejected"), Response.json({ status: "rejected" }))
        : original(r),
    );
    await f.click("reject-correction");
    const rejected = f.requests.find((r) => r.url.endsWith("/c2/reject"));
    assert.ok(rejected);
    assert.equal((await rejected.clone().json()).reason, "請提供出勤佐證");
    assert.match(f.node("review-status").textContent!, /已駁回/);
    assert.match(f.node("requests-list").textContent!, /已駁回/);
    assert.doesNotMatch(f.node("requests-list").textContent!, /待審核/);
    assert.equal(
      f.requests.filter((r) => new URL(r.url).pathname.endsWith("/corrections")).length,
      2,
    );
  } finally {
    await f.dispose();
  }
});
test("own location and HTTP deadlines allow retry and reject late success", async () => {
  const originalSet = globalThis.setTimeout,
    originalClear = globalThis.clearTimeout;
  let sequence = -1;
  const timers = new Map<number, { callback: () => void; delay: number }>();
  globalThis.setTimeout = ((callback: () => void, delay: number) => {
    if (delay === 10000 || delay === 15000) {
      const id = sequence--;
      timers.set(id, { callback, delay });
      return id;
    }
    return originalSet(callback, delay);
  }) as typeof setTimeout;
  globalThis.clearTimeout = ((id: ReturnType<typeof setTimeout>) => {
    if (typeof id === "number" && id < 0) {
      timers.delete(id);
      return;
    }
    originalClear(id);
  }) as typeof clearTimeout;
  const expire = async (delay: number) => {
    const entry = [...timers.entries()].find(([, t]) => t.delay === delay);
    assert.ok(entry);
    timers.delete(entry[0]);
    await act(async () => {
      entry[1].callback();
      await flush();
    });
  };
  let resolve!: (response: Response) => void;
  const f = await fixture(<Punch />, "?launch=" + launch, async (r, next) =>
    r.url.endsWith("/punches") ? new Promise<Response>((done) => (resolve = done)) : next(r),
  );
  try {
    await expire(15000);
    assert.equal(f.watchers[0].cleared, true);
    assert.equal(f.node<HTMLButtonElement>("punch").disabled, false);
    await act(async () => {
      f.watchers[0].ok(position);
      await flush();
    });
    assert.equal(f.requests.filter((r) => r.url.endsWith("/punches")).length, 0);
    await f.click("punch");
    await act(async () => {
      f.watchers[1].ok(position);
      await flush();
    });
    await expire(10000);
    assert.match(f.node("punch-status").textContent!, /尚未確認/);
    assert.equal(f.requests.find((r) => r.url.endsWith("/punches"))!.signal.aborted, true);
    await act(async () => {
      resolve(Response.json({ code: "clocked_in", accepted: true, message: "late success" }));
      await flush();
    });
    assert.match(f.node("punch-status").textContent!, /尚未確認/);
    assert.equal(f.close, 0);
  } finally {
    await f.dispose();
    globalThis.setTimeout = originalSet;
    globalThis.clearTimeout = originalClear;
  }
});
test("failed server verification cannot enable a workflow; retry and external logout work", async () => {
  let denied = true;
  const f = await fixture(
    <History />,
    "",
    async (r, next) =>
      denied && r.url.endsWith("/session") && r.method === "POST"
        ? Response.json(
            { error: { code: "unauthorized", message: "invalid token" } },
            { status: 401 },
          )
        : next(r),
    false,
  );
  try {
    assert.equal(f.node<HTMLFieldSetElement>("history-controls").disabled, true);
    assert.equal(f.node<HTMLElement>("identity").hidden, true);
    assert.match(f.node("mini-status").textContent!, /invalid token/);
    assert.equal(f.requests.filter((r) => r.url.includes("/groups/")).length, 0);
    denied = false;
    await f.click("verify");
    assert.equal(f.node<HTMLFieldSetElement>("history-controls").disabled, false);
    assert.equal(
      f.requests
        .find((r) => r.url.endsWith("/session") && r.method === "POST")!
        .headers.get("Authorization"),
      "Bearer token",
    );
    await f.click("logout");
    assert.equal(f.logout, 1);
    assert.equal(f.node<HTMLElement>("login").hidden, false);
  } finally {
    await f.dispose();
  }
});
for (const client of [false, true])
  test("failed session login control respects LINE browser context: " + client, async () => {
    const f = await fixture(
      <History />,
      "",
      async (r, next) =>
        r.url.endsWith("/session") && r.method === "POST"
          ? Response.json(
              {
                error: {
                  code: "unauthenticated",
                  message: "LINE 登入已失效，請重新登入。（LINE HTTP 400）",
                },
              },
              { status: 401 },
            )
          : next(r),
      client,
    );
    try {
      assert.match(f.node("mini-status").textContent!, /HTTP 401.*LINE HTTP 400/);
      const writes = f.requests.filter((r) => r.method !== "GET").length;
      await f.click("login");
      assert.deepEqual(f.sdkEvents, client ? [] : ["logout", "login"]);
      assert.equal(
        f.requests.filter((r) => r.method !== "GET").length,
        writes,
        "login redirect does not create another application session",
      );
    } finally {
      await f.dispose();
    }
  });
test("pagehide rejects late identity responses; BFCache restoration verifies again", async () => {
  const f = await fixture(<History />);
  try {
    let resolve!: (response: Response) => void;
    const previous = f.setHandler((r) =>
      r.url.endsWith("/session") && r.method === "POST"
        ? new Promise<Response>((done) => (resolve = done))
        : previous(r),
    );
    await f.click("verify");
    await act(async () => {
      window.dispatchEvent(new Event("pagehide"));
      await flush();
      resolve(
        Response.json({
          user: { userId: "obsolete", displayName: "Old" },
          csrfToken: "c".repeat(43),
          expiresAt: new Date(Date.now() + 1800000).toISOString(),
        }),
      );
      await flush();
    });
    assert.equal(f.node<HTMLElement>("identity").hidden, true);
    assert.equal(f.node("user-id").textContent, "");
    f.setHandler(previous);
    await act(async () => {
      window.dispatchEvent(new PageTransitionEvent("pageshow", { persisted: true }));
      await flush();
    });
    assert.equal(f.node("user-id").textContent, "u1");
  } finally {
    await f.dispose();
  }
});
test("diagnostic location renders only validated coordinates and retains the last success after failure", async () => {
  const f = await fixture(<Geo />, "", async (r, next) => {
    if (!r.url.endsWith("/api/location/validate")) return next(r);
    assert.equal(r.credentials, "omit");
    return Response.json({ valid: true, receivedAt: new Date().toISOString() });
  });
  try {
    assert.equal(f.watchers.length, 0);
    await f.click("locate");
    await act(async () => {
      f.watchers[0].ok(position);
      await flush();
    });
    assert.equal(f.node("latitude").textContent, "25");
    assert.equal(f.node<HTMLDListElement>("result").hidden, false);
    await f.click("locate");
    await act(async () => {
      f.watchers[1].fail!({ code: 1 } as GeolocationPositionError);
      await flush();
    });
    assert.equal(f.node("latitude").textContent, "25");
    assert.match(f.node("captured").textContent!, /上次取得/);
    assert.equal(f.watchers[1].cleared, true);
  } finally {
    await f.dispose();
  }
});

for (const newRecord of [true, false])
  test(`review compares a proposed shift with ${newRecord ? "no original shift" : "the original shift"}`, async () => {
    const f = await fixture(<Review />, "", async (request, next) =>
      request.url.endsWith("/review")
        ? Response.json({
            newRecord,
            before: newRecord ? null : record,
            status: "pending",
            start: record.ClockInAt,
            end: record.ClockOutAt,
            restConfirmed: false,
            applicantName: "員工甲",
            reason: "班次異動",
          })
        : next(request),
    );
    try {
      await f.input("correction-id", "c1");
      await f.click("review-correction");
      const [original, proposed] = f.node("review-details").querySelectorAll('[data-slot="item"]');
      assert.match(original.textContent!, newRecord ? /無原班次/ : /上班：.+下班：/);
      assert.match(proposed.textContent!, /申報班次.*休息完整確認：否/);
      assert.match(f.node("review-details").textContent!, /申請原因：班次異動/);
    } finally {
      await f.dispose();
    }
  });

test("review without a reported status never offers approval or rejection", async () => {
  const f = await fixture(<Review />);
  try {
    await f.input("correction-id", "c1");
    await f.click("review-correction");
    assert.match(f.node("review-details").textContent!, /狀態無法確認/);
    await f.input("rejection-reason", "查無證據");
    for (const id of ["approve-correction", "reject-correction"])
      assert.ok(f.node<HTMLButtonElement>(id).disabled);
  } finally {
    await f.dispose();
  }
});

test("new-shift corrections discover membership groups before any attendance records exist", async () => {
  const f = await fixture(<Correction />, "", async (r, next) =>
    new URL(r.url).pathname === "/api/v1/groups"
      ? Response.json({ items: [{ id: "g1", groupId: "g1", name: "First group" }], nextCursor: "" })
      : next(r),
  );
  try {
    assert.equal(document.getElementById("correction-group"), null);
    assert.equal(f.node<HTMLFieldSetElement>("correction-fields").disabled, false);
    assert.ok(f.requests.some((r) => new URL(r.url).pathname === "/api/v1/groups"));
    assert.equal(
      f.requests.some((r) => new URL(r.url).pathname.endsWith("/groups/page")),
      false,
    );
  } finally {
    await f.dispose();
  }
});

test("leave history filters already loaded records by group without issuing another request", async () => {
  const f = await fixture(<Leave />, "", async (r, next) =>
    r.url.endsWith("/leaves")
      ? Response.json(
          ["g1", "g2"].map((groupId) => ({
            id: groupId,
            groupId,
            kind: "personal",
            start: record.ClockInAt,
            end: record.ClockOutAt,
            reason: groupId + " reason",
          })),
        )
      : next(r),
  );
  try {
    const count = f.requests.filter((r) => r.url.endsWith("/leaves")).length;
    const select = document.querySelector<HTMLSelectElement>("select")!;
    assert.equal(document.querySelectorAll('[role="listitem"]').length, 2);
    await act(async () => {
      select.value = "g2";
      select.dispatchEvent(new Event("change", { bubbles: true }));
      await flush();
    });
    const items = document.querySelectorAll('[role="listitem"]');
    assert.equal(items.length, 1);
    assert.match(items[0].textContent!, /g2 reason/);
    assert.equal(f.requests.filter((r) => r.url.endsWith("/leaves")).length, count);
  } finally {
    await f.dispose();
  }
});

test("expired leave launch offers history instead of retrying an unusable link", async () => {
  let posts = 0,
    reads = 0;
  const f = await fixture(<Leave />, "?leave=" + launch, async (r, next) => {
    if (!r.url.endsWith("/leaves")) return next(r);
    if (r.method !== "POST") {
      reads++;
      return Response.json([]);
    }
    if (++posts === 1) throw new Error("connection lost");
    return Response.json(
      { error: { code: "expired_launch", message: "請假連結已過期" } },
      { status: 403 },
    );
  });
  try {
    const form = document.querySelector("form")!;
    const fields = form.querySelectorAll("input");
    fields[0].id = "leave-from";
    fields[1].id = "leave-to";
    form.querySelector("textarea")!.id = "leave-why";
    await f.input("leave-from", "2026-11-01T09:00");
    await f.input("leave-to", "2026-11-01T12:00");
    await f.input("leave-why", "看診");
    await act(async () => {
      form.dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
      await flush();
    });
    const before = reads;
    await act(async () => {
      form.dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
      await flush();
    });
    assert.ok(reads > before);
    assert.match(document.body.textContent!, /前次結果待確認/);
    assert.match(document.body.textContent!, /請假連結已過期/);
    assert.equal(document.querySelector("form"), null);
    assert.doesNotMatch(document.body.textContent!, /重試原申請/);
  } finally {
    await f.dispose();
  }
});

test("personal corrections load automatically and distinguish older results from pending requests", async () => {
  const f = await fixture(<Correction />, "?group=g1", async (r, next) =>
    new URL(r.url).pathname.endsWith("/corrections")
      ? Response.json({
          items: [
            {
              id: "c1",
              status: "pending",
              reason: "較早待審核",
              createdAt: "2020-01-01T00:00:00Z",
            },
            {
              id: "c2",
              status: "approved",
              reason: "較早已結案",
              createdAt: "2020-01-01T00:00:00Z",
            },
          ],
          nextCursor: "",
        })
      : next(r),
  );
  try {
    assert.equal(document.getElementById("requests-load"), null);
    assert.equal(document.getElementById("correction-group"), null);
    assert.match(f.node("requests-list").textContent!, /較早待審核/);
    assert.doesNotMatch(f.node("requests-list").textContent!, /較早已結案/);
    await act(async () => {
      [...document.querySelectorAll("button")]
        .find((button) => button.textContent === "查看較早的申請")!
        .click();
      await flush();
    });
    assert.match(f.node("requests-list").textContent!, /較早已結案/);
  } finally {
    await f.dispose();
  }
});

test("a denied corrections page clears previously readable requests", async () => {
  const f = await fixture(<Correction />, "?group=g1", async (r, next) => {
    const url = new URL(r.url);
    if (!url.pathname.endsWith("/corrections")) return next(r);
    return url.searchParams.get("cursor")
      ? Response.json({ error: { code: "forbidden", message: "權限已變更" } }, { status: 403 })
      : Response.json({
          items: [{ id: "c1", status: "pending", reason: "私人申請內容" }],
          nextCursor: "page2",
        });
  });
  try {
    assert.match(f.node("requests-list").textContent!, /私人申請內容/);
    await act(async () => {
      [...document.querySelectorAll("button")]
        .find((button) => button.textContent === "載入更多申請")!
        .click();
      await flush();
    });
    assert.doesNotMatch(f.node("requests-list").textContent!, /私人申請內容/);
    assert.match(document.body.textContent!, /權限已變更/);
    assert.match(f.node("requests-load").textContent!, /重試讀取申請/);
  } finally {
    await f.dispose();
  }
});

test("enterprise creation retains its original name after a lost response and later denial", async () => {
  const names: string[] = [];
  let saved = "";
  const f = await fixture(
    <Enterprise
      id="new"
      canCreate
      onSaved={(id) => {
        saved = id;
      }}
    />,
    "",
    async (r, next) => {
      if (!r.url.endsWith("/enterprises") || r.method !== "POST") return next(r);
      names.push((await r.clone().json()).name);
      if (names.length === 1) throw new Error("connection lost");
      if (names.length === 2)
        return Response.json(
          { error: { code: "forbidden", message: "denied", retrySameRequest: false } },
          { status: 403 },
        );
      return Response.json({ id: "existing", name: names[0] });
    },
  );
  try {
    const input = document.querySelector<HTMLInputElement>("input[name=value]")!;
    const submit = async () =>
      act(async () => {
        input.form!.requestSubmit();
        await flush();
      });
    input.value = "原企業";
    await submit();
    assert.equal(input.readOnly, true);
    input.value = "編輯後名稱";
    await submit();
    await submit();
    assert.deepEqual(names, ["原企業", "原企業", "原企業"]);
    assert.equal(saved, "existing");
  } finally {
    await f.dispose();
  }
});

test("enterprise management denial refreshes permissions and removes stale data", async () => {
  let denied = false;
  const f = await fixture(<Enterprise id="hq" onSaved={() => {}} />, "", async (r, next) => {
    if (r.url.endsWith("/enterprises/hq/owners")) {
      denied = true;
      return Response.json(
        { error: { code: "forbidden", message: "權限已變更" } },
        { status: 403 },
      );
    }
    if (r.url.endsWith("/enterprises/hq"))
      return Response.json({
        id: "hq",
        name: "企業",
        canManage: !denied,
        teams: [],
        organizations: denied ? [] : [{ id: "private", name: "過期組織" }],
      });
    return next(r);
  });
  try {
    assert.ok(document.body.textContent?.includes("過期組織"));
    const management = [...document.querySelectorAll<HTMLElement>('[role="tab"]')].find(
      (tab) => tab.textContent === "企業管理",
    )!;
    await act(async () => {
      pressButton(management as HTMLButtonElement);
      await flush();
    });
    const owner = [...document.querySelectorAll("form")].find((form) =>
      form.textContent?.includes("新增企業 Owner"),
    )!;
    owner.querySelector("input")!.value = "member";
    await act(async () => {
      owner.requestSubmit();
      await flush();
    });
    assert.ok(denied);
    assert.ok(
      [...document.querySelectorAll("h2")].some((heading) => heading.textContent === "企業"),
    );
    assert.doesNotMatch(document.body.textContent!, /新增企業 Owner|加入組織|過期組織/);
  } finally {
    await f.dispose();
  }
});

test("organization team binding uses the selected scope and reloads its list", async () => {
  let bound = false,
    reads = 0;
  const f = await fixture(<OrganizationTeams organizationId="org1" />, "", async (r, next) => {
    if (r.url.endsWith("/teams/group1/management")) {
      assert.equal(r.method, "PUT");
      assert.deepEqual(await r.json(), { managementScope: "organization", managementId: "org1" });
      bound = true;
      return new Response(null, { status: 204 });
    }
    if (r.url.endsWith("/organizations/org1/teams")) {
      reads++;
      return Response.json({
        canBindTeam: true,
        teams: bound ? [{ teamId: "group1", name: "排班組", managementScope: "organization" }] : [],
      });
    }
    return next(r);
  });
  try {
    const form = document.querySelector("form")!;
    form.querySelector("input")!.value = "group1";
    await act(async () => {
      form.requestSubmit();
      await flush();
    });
    assert.ok(bound);
    assert.equal(reads, 2);
    assert.match(document.body.textContent!, /排班組/);
  } finally {
    await f.dispose();
  }
});

test("enterprise teams bind, assign and remove through existing APIs and retain uncertain inputs", async () => {
  let reads = 0,
    writes = 0;
  const calls: { method: string; path: string; body?: unknown }[] = [];
  const f = await fixture(<Enterprise id="hq" onSaved={() => {}} />, "", async (r, next) => {
    if (r.url.endsWith("/enterprises/hq")) {
      reads++;
      return Response.json({
        id: "hq",
        name: "總公司",
        canManage: false,
        canManageTeams: true,
        organizations: [{ id: "ineligible", name: "僅供查看" }],
        assignableOrganizations: [
          { id: "org1", name: "台灣" },
          { id: "org2", name: "日本" },
        ],
        teams: [{ id: "group1", name: "資安組" }],
      });
    }
    if (r.url.endsWith("/teams/group1/management")) {
      calls.push({ method: r.method, path: new URL(r.url).pathname, body: await r.json() });
      return new Response(null, { status: 204 });
    }
    if (r.url.includes("/enterprise-teams/")) {
      calls.push({ method: r.method, path: new URL(r.url).pathname });
      writes++;
      return writes === 1
        ? Response.json(
            { error: { code: "unavailable", message: "連線中斷", retrySameRequest: true } },
            { status: 503 },
          )
        : writes === 2
          ? Response.json(
              { error: { code: "invalid_input", message: "本次被拒絕", retrySameRequest: false } },
              { status: 400 },
            )
          : new Response(null, { status: 204 });
    }
    return next(r);
  });
  try {
    const tab = [...document.querySelectorAll<HTMLButtonElement>('[role="tab"]')].find(
      (t) => t.textContent === "企業團隊",
    )!;
    await act(async () => {
      pressButton(tab);
      await flush();
    });
    const binding = document.querySelector("form")!;
    binding.querySelector("input")!.value = "group1";
    await act(async () => {
      binding.requestSubmit();
      await flush();
    });
    assert.deepEqual(calls[0].body, { managementScope: "enterprise", managementId: "hq" });
    const assignment = [...document.querySelectorAll("form")].find((f) =>
      f.textContent?.includes("送出指派設定"),
    )!;
    assert.doesNotMatch(assignment.textContent!, /僅供查看/);
    await act(async () => {
      assignment.requestSubmit();
      await flush();
    });
    assert.match(assignment.textContent!, /結果待確認/);
    assert.ok(assignment.querySelector("fieldset")!.disabled);
    // Even if the controls are changed programmatically, retry sends the original request.
    assignment.querySelector<HTMLSelectElement>('[name="organizationId"]')!.value = "org2";
    await act(async () => {
      assignment.requestSubmit();
      await flush();
    });
    assert.ok(
      assignment.querySelector("fieldset")!.disabled,
      "a later rejection cannot resolve an earlier unknown outcome",
    );
    await act(async () => {
      assignment.requestSubmit();
      await flush();
    });
    assert.deepEqual(calls[1], calls[3]);
    assert.deepEqual(calls[1], calls[2]);
    assert.equal(calls[2].path, "/api/v1/organizations/org1/enterprise-teams/group1");
    assignment.querySelector<HTMLSelectElement>('[name="action"]')!.value = "remove";
    await act(async () => {
      assignment.requestSubmit();
      await flush();
    });
    assert.equal(calls[4].method, "DELETE");
    assert.equal(calls[4].path, "/api/v1/organizations/org2/enterprise-teams/group1");
    assert.equal(reads, 4);
  } finally {
    await f.dispose();
  }
});

test("team binding denial stays visible and enterprise permissions refresh", async () => {
  let denied = false;
  let finishRefresh!: () => void;
  const refresh = new Promise<void>((resolve) => {
    finishRefresh = resolve;
  });
  const f = await fixture(<Enterprise id="hq" onSaved={() => {}} />, "", async (r, next) => {
    if (r.url.endsWith("/teams/group1/management")) {
      denied = true;
      return Response.json(
        { error: { code: "forbidden", message: "團隊管理權限已變更", retrySameRequest: false } },
        { status: 403 },
      );
    }
    if (r.url.endsWith("/enterprises/hq")) {
      if (denied) await refresh;
      return Response.json({
        id: "hq",
        name: "總公司",
        canManage: !denied,
        canManageTeams: !denied,
        organizations: [],
        teams: [],
      });
    }
    return next(r);
  });
  try {
    await act(async () => {
      pressButton(
        [...document.querySelectorAll<HTMLButtonElement>('[role="tab"]')].find(
          (t) => t.textContent === "企業團隊",
        )!,
      );
      await flush();
    });
    const form = document.querySelector("form")!;
    form.querySelector("input")!.value = "group1";
    await act(async () => {
      form.requestSubmit();
      await flush();
    });
    assert.ok(denied);
    assert.doesNotMatch(document.body.textContent!, /綁定團隊|送出指派設定/);
    await act(async () => {
      finishRefresh();
      await flush();
    });
    assert.match(document.body.textContent!, /團隊管理權限已變更/);
    assert.doesNotMatch(document.body.textContent!, /綁定團隊|送出指派設定|企業管理/);
  } finally {
    await f.dispose();
  }
});

for (const scope of ["企業", "組織"]) {
  test(scope + " audit tab and its team's audit tab keep separate scopes", async () => {
    const view =
      scope === "企業" ? (
        <Enterprise id="hq" onSaved={() => {}} />
      ) : (
        <OrganizationPage organizationId="org1" onChanged={() => {}} onLeave={() => {}}>
          <OrganizationTeams organizationId="org1" />
        </OrganizationPage>
      );
    const team = {
      id: "g1",
      teamId: "g1",
      name: "排班組",
      managementScope: "organization",
      auditHistoryStatus: "unavailable",
    };
    const f = await fixture(view, "", async (request, next) => {
      const path = new URL(request.url).pathname;
      if (path.endsWith("/enterprises/hq"))
        return Response.json({
          id: "hq",
          name: "測試企業",
          organizations: [],
          teams: [team],
          auditHistoryStatus: "unavailable",
        });
      if (path.endsWith("/org1/members"))
        return Response.json({
          organizationName: "測試組織",
          members: [],
          auditHistoryStatus: "unavailable",
        });
      if (path.endsWith("/org1/teams")) return Response.json({ teams: [team] });
      return next(request);
    });
    const clickTab = async (list: string, label: string) => {
      const tabs = document.querySelector(`[role=tablist][aria-label="${list}"]`);
      const tab = [...tabs!.querySelectorAll<HTMLButtonElement>("[role=tab]")].find(
        (item) => item.textContent === label,
      );
      assert.ok(tab);
      await act(async () => {
        pressButton(tab);
        await flush();
      });
    };
    try {
      await clickTab(scope + "功能", "稽核");
      assert.match(
        document.body.textContent!,
        new RegExp("測試" + scope + "的修改歷史尚未開放查詢"),
      );
      await clickTab(scope + "功能", scope + "團隊");
      document.querySelector("details")!.open = true;
      await clickTab("排班組團隊功能", "稽核");
      assert.match(document.body.textContent!, /排班組的修改歷史尚未開放查詢/);
    } finally {
      await f.dispose();
    }
  });
}

test("enterprise management retries only the original action and target", async () => {
  const owners: string[] = [];
  let saved = "";
  const f = await fixture(
    <Enterprise
      id="hq"
      onSaved={(id) => {
        saved = id;
      }}
    />,
    "",
    async (r, next) => {
      if (r.url.endsWith("/enterprises/hq"))
        return Response.json({
          id: "hq",
          name: "企業",
          canManage: true,
          teams: [],
          organizations: [],
        });
      if (!r.url.endsWith("/owners")) return next(r);
      owners.push((await r.json()).userId);
      if (owners.length === 1)
        return Response.json(
          { error: { code: "unavailable", message: "待確認", retrySameRequest: true } },
          { status: 503 },
        );
      if (owners.length === 2)
        return Response.json(
          { error: { code: "invalid_input", message: "拒絕", retrySameRequest: false } },
          { status: 400 },
        );
      return Response.json({ id: "hq", name: "企業" });
    },
  );
  try {
    const tab = [...document.querySelectorAll<HTMLElement>('[role="tab"]')].find(
      (t) => t.textContent === "企業管理",
    )!;
    await act(async () => {
      pressButton(tab as HTMLButtonElement);
      await flush();
    });
    const owner = [...document.forms].find((form) => form.textContent?.includes("新增企業 Owner"))!;
    const org = [...document.forms].find((form) => form.textContent?.includes("加入組織"))!;
    const input = owner.querySelector("input")!;
    input.value = "member-A";
    const submit = async () =>
      act(async () => {
        owner.requestSubmit();
        await flush();
      });
    await submit();
    assert.equal(org.querySelector("button")!.disabled, true);
    input.value = "member-B";
    await submit();
    assert.equal(input.readOnly, true);
    await submit();
    assert.deepEqual(owners, ["member-A", "member-A", "member-A"]);
    assert.equal(saved, "hq");
  } finally {
    await f.dispose();
  }
});

test("quantity history keeps pages on interruption and clears them when access is denied", async () => {
  let attempts = 0;
  const f = await fixture(<QuantityHistory repositoryId="repo" />, "", async (r, next) => {
    const url = new URL(r.url);
    if (!url.pathname.endsWith("object-quantity-changes")) return next(r);
    if (!url.searchParams.get("cursor"))
      return Response.json({
        items: [
          {
            objectId: "obj",
            version: 1,
            name: "私人數量紀錄",
            afterQuantity: "2",
            at: "2026-11-23T00:00:00Z",
          },
        ],
        nextCursor: "page2",
      });
    assert.equal(url.searchParams.get("cursor"), "page2");
    const status = ++attempts === 1 ? 503 : 403;
    return Response.json(
      { error: { code: status === 403 ? "forbidden" : "unavailable", message: "讀取失敗" } },
      { status },
    );
  });
  try {
    async function click(text: string) {
      const button = [...document.querySelectorAll("button")].find(
        (item) => item.textContent === text,
      )!;
      await act(async () => {
        button.click();
        await flush();
      });
    }
    await click("載入更早紀錄");
    assert.match(document.body.textContent!, /私人數量紀錄/);
    await click("重試讀取");
    assert.equal(attempts, 2);
    assert.doesNotMatch(document.body.textContent!, /私人數量紀錄/);
  } finally {
    await f.dispose();
  }
});
