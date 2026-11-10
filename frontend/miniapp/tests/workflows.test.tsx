import React from "react";
import { test } from "node:test";
import assert from "node:assert/strict";
import { act, StrictMode, type ReactNode } from "react";
import { createRoot } from "react-dom/client";
import { SessionProvider, useSession } from "../src/identity/Session";
import { IdentityPanel } from "../src/identity/Panel";
import { View as Punch } from "../src/features/attendance/punch/View";
import { View as History } from "../src/features/attendance/history/View";
import { View as Correction } from "../src/features/attendance/correction/View";
import { View as Geo } from "../src/diagnostics/geolocation/View";
import { View as Review } from "../src/features/attendance/review/View";
import { launchToken } from "../src/features/attendance/punch/usePunch";
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
  for (let i = 0; i < 8; i++)
    await new Promise<void>((resolve) => setImmediate(resolve));
};
function Epoch({ children }: { children: ReactNode }) {
  const { epoch } = useSession();
  return <div key={epoch}>{children}</div>;
}
async function fixture(
  view: ReactNode,
  search = "",
  override?: (
    request: Request,
    next: (r: Request) => Promise<Response>,
  ) => Promise<Response>,
  client = true,
) {
  window.history.replaceState(null, "", "/miniapp" + search);
  document.body.innerHTML = '<div id="root"></div>';
  let init = 0,
    close = 0,
    logout = 0;
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
    getIDToken: () => "token",
    login: () => {},
    logout: () => {
      logout++;
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
    if (path.endsWith("/me"))
      return Response.json({ userId: "u1", displayName: "Tester" });
    if (path.endsWith("/launch"))
      return Response.json({ requestId: requestID });
    if (path.endsWith("/punches"))
      return Response.json({ code: "clocked_in", message: "success" });
    if (path.endsWith("/groups/page"))
      return Response.json({
        items: [{ id: "g1", name: "Group" }],
        nextCursor: "",
      });
    if (path.endsWith("/records/r1")) return Response.json(record);
    if (path.endsWith("/records"))
      return Response.json({ items: [record], nextCursor: "" });
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
        : new Request(
            typeof input === "string" ? new URL(input, location.origin) : input,
            init,
          );
    requests.push(r.clone());
    return handler(r);
  };
  const root = createRoot(document.getElementById("root")!);
  await act(async () => {
    root.render(
      <StrictMode>
        <SessionProvider sdkLoader={loader}>
          <IdentityPanel />
          <Epoch>{view}</Epoch>
        </SessionProvider>
      </StrictMode>,
    );
    await flush();
  });
  const node = <T extends HTMLElement>(id: string) =>
    document.getElementById(id) as T;
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
  const f = await fixture(<Punch />, "?launch=" + launch);
  try {
    assert.equal(f.init, 1);
    assert.equal(f.watchers.length, 1);
    assert.equal(f.requests.filter((r) => r.url.endsWith("/launch")).length, 1);
    await act(async () => {
      f.watchers[0].ok(position);
      f.watchers[0].ok(position);
      await flush();
    });
    assert.equal(f.close, 1);
    assert.equal(
      f.requests.filter((r) => r.url.endsWith("/punches")).length,
      1,
    );
    assert.equal(f.watchers[0].cleared, true);
    assert.equal(f.node("punch-status").textContent, "success");
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
    assert.equal(
      f.requests.filter((r) => r.url.endsWith("/punches")).length,
      2,
    );
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
  assert.equal(launchToken("?request=" + launch, "https://example.test"), "");
  assert.equal(
    launchToken(
      "?liff.state=" + encodeURIComponent("?launch=" + launch),
      "https://example.test",
    ),
    launch,
  );
});
test("month reads are cancelled on reload and private records clear on logout", async () => {
  const f = await fixture(<History />, "?group=g1&month=2026-11");
  try {
    assert.match(f.node("month-summary").textContent!, /60/);
    let resolve!: (value: Response) => void;
    const old = f.setHandler(async (r) =>
      r.url.includes("/records?")
        ? new Promise<Response>((r) => (resolve = r))
        : old(r),
    );
    await f.click("load-month");
    await act(async () => {
      const input = f.node<HTMLInputElement>("history-month");
      const setter = Object.getOwnPropertyDescriptor(
        window.HTMLInputElement.prototype,
        "value",
      )!.set!;
      setter.call(input, "2026-12");
      input.dispatchEvent(new Event("input", { bubbles: true }));
      input.dispatchEvent(new Event("change", { bubbles: true }));
      await flush();
    });
    // Trigger cancellation through another explicit read as well, independent of DOM event implementation.
    await f.click("load-groups");
    await act(async () => {
      resolve(
        Response.json({
          items: [{ ...record, RecordID: "obsolete" }],
          nextCursor: "",
        }),
      );
      await flush();
    });
    assert.doesNotMatch(f.node("record-list").textContent!, /obsolete/);
    await f.click("logout");
    assert.equal(f.node("record-list").children.length, 0);
    assert.equal(f.node("month-summary").textContent, "");
  } finally {
    await f.dispose();
  }
});
test("review controls and details clear on logout", async () => {
  const f = await fixture(<Review />);
  try {
    assert.equal(
      f.node<HTMLButtonElement>("approve-correction").disabled,
      true,
    );
    await f.click("logout");
    assert.equal(f.node<HTMLFieldSetElement>("review-controls").disabled, true);
    assert.equal(f.node("review-details").textContent, "");
  } finally {
    await f.dispose();
  }
});
test("cursor pagination keeps the full monthly summary and navigates back", async () => {
  const f = await fixture(
    <History />,
    "?group=g1&month=2026-11",
    async (r, next) => {
      if (new URL(r.url).pathname.endsWith("/records"))
        return Response.json({
          items: [
            {
              ...record,
              RecordID: new URL(r.url).searchParams.get("cursor")
                ? "page2"
                : "page1",
            },
          ],
          nextCursor: new URL(r.url).searchParams.get("cursor")
            ? ""
            : "cursor2",
        });
      return next(r);
    },
  );
  try {
    assert.match(f.node("record-list").innerHTML, /page1/);
    await f.click("next-records");
    assert.match(f.node("record-list").innerHTML, /page2/);
    assert.match(f.node("month-summary").textContent!, /60/);
    assert.equal(f.node("record-page").textContent, "第 2 頁");
    await f.click("previous-records");
    assert.match(f.node("record-list").innerHTML, /page1/);
    assert.equal(
      f.requests.filter((r) => r.url.includes("/monthly/summary")).length,
      1,
    );
  } finally {
    await f.dispose();
  }
});
test("correction double submit and ambiguous retry reuse a key; edited draft rotates it", async () => {
  const f = await fixture(
    <Correction />,
    "?group=g1&record=r1",
    async (r, next) =>
      r.url.endsWith("/corrections")
        ? Promise.reject(new TypeError("network"))
        : next(r),
  );
  try {
    await f.input("correction-reason", "Reason");
    await f.click("correction-confirm");
    const submit = () =>
      f
        .node("correction-form")
        .dispatchEvent(
          new Event("submit", { bubbles: true, cancelable: true }),
        );
    await act(async () => {
      submit();
      submit();
      await flush();
    });
    assert.equal(
      f.requests.filter((r) => r.url.endsWith("/corrections")).length,
      1,
    );
    await act(async () => {
      submit();
      await flush();
    });
    let posts = f.requests.filter((r) => r.url.endsWith("/corrections"));
    assert.equal(posts.length, 2);
    assert.equal(
      posts[0].headers.get("Idempotency-Key"),
      posts[1].headers.get("Idempotency-Key"),
    );
    assert.equal((await posts[0].clone().json()).reason, "Reason");
    await f.input("correction-reason", "Changed");
    await act(async () => {
      submit();
      await flush();
    });
    posts = f.requests.filter((r) => r.url.endsWith("/corrections"));
    assert.equal(posts.length, 3);
    assert.notEqual(
      posts[1].headers.get("Idempotency-Key"),
      posts[2].headers.get("Idempotency-Key"),
    );
  } finally {
    await f.dispose();
  }
});
test("editing review ID rejects a late proposal and prevents approval of another ID", async () => {
  let resolve!: (response: Response) => void;
  const f = await fixture(<Review />, "", async (r, next) =>
    r.url.endsWith("/review")
      ? new Promise<Response>((done) => (resolve = done))
      : next(r),
  );
  try {
    await f.input("correction-id", "c1");
    await f.click("review-correction");
    await f.input("correction-id", "c2");
    await act(async () => {
      resolve(Response.json({ reason: "obsolete" }));
      await flush();
    });
    assert.equal(f.node("review-details").textContent, "");
    assert.equal(
      f.node<HTMLButtonElement>("approve-correction").disabled,
      true,
    );
    assert.equal(
      f.requests.filter((r) => r.url.endsWith("/approve")).length,
      0,
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
    r.url.endsWith("/punches")
      ? new Promise<Response>((done) => (resolve = done))
      : next(r),
  );
  try {
    await expire(15000);
    assert.equal(f.watchers[0].cleared, true);
    assert.equal(f.node<HTMLButtonElement>("punch").disabled, false);
    await act(async () => {
      f.watchers[0].ok(position);
      await flush();
    });
    assert.equal(
      f.requests.filter((r) => r.url.endsWith("/punches")).length,
      0,
    );
    await f.click("punch");
    await act(async () => {
      f.watchers[1].ok(position);
      await flush();
    });
    await expire(10000);
    assert.match(f.node("punch-status").textContent!, /尚未確認/);
    assert.equal(
      f.requests.find((r) => r.url.endsWith("/punches"))!.signal.aborted,
      true,
    );
    await act(async () => {
      resolve(Response.json({ code: "clocked_in", message: "late success" }));
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
      denied && r.url.endsWith("/me")
        ? Response.json(
            { error: { code: "unauthorized", message: "invalid token" } },
            { status: 401 },
          )
        : next(r),
    false,
  );
  try {
    assert.equal(
      f.node<HTMLFieldSetElement>("history-controls").disabled,
      true,
    );
    assert.equal(f.node<HTMLElement>("identity").hidden, true);
    assert.match(f.node("mini-status").textContent!, /invalid token/);
    assert.equal(
      f.requests.filter((r) => r.url.includes("/groups/")).length,
      0,
    );
    denied = false;
    await f.click("verify");
    assert.equal(
      f.node<HTMLFieldSetElement>("history-controls").disabled,
      false,
    );
    assert.equal(
      f.requests
        .find((r) => r.url.endsWith("/me"))!
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
test("pagehide rejects late identity responses; BFCache restoration verifies again", async () => {
  const f = await fixture(<History />);
  try {
    let resolve!: (response: Response) => void;
    const previous = f.setHandler((r) =>
      r.url.endsWith("/me")
        ? new Promise<Response>((done) => (resolve = done))
        : previous(r),
    );
    await f.click("verify");
    await act(async () => {
      window.dispatchEvent(new Event("pagehide"));
      await flush();
      resolve(Response.json({ userId: "obsolete", displayName: "Old" }));
      await flush();
    });
    assert.equal(f.node<HTMLElement>("identity").hidden, true);
    assert.equal(f.node("user-id").textContent, "");
    f.setHandler(previous);
    await act(async () => {
      window.dispatchEvent(
        new PageTransitionEvent("pageshow", { persisted: true }),
      );
      await flush();
    });
    assert.equal(f.node("user-id").textContent, "u1");
  } finally {
    await f.dispose();
  }
});
test("diagnostic location renders only validated coordinates and retains the last success after failure", async () => {
  const f = await fixture(<Geo />, "", async (r, next) =>
    r.url.endsWith("/api/location/validate")
      ? Response.json({ valid: true })
      : next(r),
  );
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
