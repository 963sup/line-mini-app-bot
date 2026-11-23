import React, { act } from "react";
import { once } from "node:events";
import { createRoot } from "react-dom/client";
import test from "node:test";
import assert from "node:assert/strict";
import { pressButton } from "./sidebar-dom";
import { Shell } from "./shell-fixture";
import { View as Objects } from "../src/business/object/View";
import { View as RepositoryView } from "../src/business/repository/View";

const rid = "a".repeat(32),
  other = "b".repeat(32),
  oid = "c".repeat(32);
const resource = (overrides = {}) => ({
  id: oid,
  repositoryId: rid,
  name: "測試物件",
  description: "完整工作說明",
  type: "work",
  completedQuantity: "0",
  quantity: "11",
  unit: "SET",
  status: "not_started",
  sourceDocument: "sample-contract",
  sourceRevision: "v0",
  sourceItemNo: "40",
  sourcePages: "2-3",
  materialCode: "sample",
  canEdit: true,
  creatorId: "A",
  version: 1,
  createdAt: "2026-11-01T00:00:00Z",
  updatedAt: "2026-11-01T00:00:00Z",
  ...overrides,
});
const flush = async () => {
  for (let i = 0; i < 10; i++) await new Promise<void>((resolve) => setImmediate(resolve));
};
async function fixture(
  handle: (request: Request, actor: string) => Promise<Response>,
  integrated = false,
  documentTool = false,
) {
  document.body.innerHTML = '<div id="root"></div>';
  window.history.replaceState(
    null,
    "",
    integrated ? `/miniapp/repositories?selected=${rid}` : "/miniapp/projects",
  );
  let actor = "A",
    repository = rid;
  const oldFetch = globalThis.fetch,
    oldSDK = window.liff;
  window.liff = {
    init: async () => {},
    isLoggedIn: () => true,
    isInClient: () => false,
    getIDToken: () => "token",
    getProfile: async () => ({ userId: actor, displayName: actor }),
    login() {},
    logout() {},
    closeWindow() {},
  };
  globalThis.fetch = async (input, init) => {
    const request =
      input instanceof Request ? input : new Request(new URL(String(input), location.origin), init);
    const path = new URL(request.url).pathname;
    if (path.endsWith("/config")) return Response.json({ liffId: "object-test" });
    if (path.endsWith("/session"))
      return Response.json({
        user: { userId: actor, displayName: actor },
        csrfToken: "c".repeat(43),
        expiresAt: new Date(Date.now() + 1800000).toISOString(),
      });
    if (path === `/api/v1/repositories/${rid}`)
      return Response.json({
        id: rid,
        ownerId: "C" + rid,
        name: "測試儲存庫",
        status: "ACTIVE",
        version: 1,
        createdAt: "2026-11-01T00:00:00Z",
        updatedAt: "2026-11-01T00:00:00Z",
      });
    if (path === "/api/v1/repositories") return Response.json({ items: [], nextCursor: "" });
    if (path.endsWith("/statistics")) return Response.json({});
    return handle(request, actor);
  };
  const root = createRoot(document.getElementById("root")!);
  const render = async () => {
    await act(async () => {
      root.render(
        <Shell title="測試" workflow={integrated ? "repositories" : "projects"}>
          {integrated ? (
            <RepositoryView />
          ) : (
            <Objects repositoryId={repository} documentTool={documentTool} />
          )}
        </Shell>,
      );
      await flush();
    });
  };
  await render();
  return {
    async click(text: string) {
      await act(async () => {
        const node = [...document.querySelectorAll<HTMLButtonElement>("button")].find(
          (node) => (node.getAttribute("aria-label") ?? node.textContent) === text,
        );
        assert.ok(node, `missing ${text}`);
        pressButton(node);
        await flush();
      });
    },
    async fill(label: string, value: string) {
      await act(async () => {
        const node = [...document.querySelectorAll("label")]
          .find((node) => node.textContent === label)
          ?.querySelector<HTMLInputElement>("input,textarea");
        assert.ok(node, `missing field ${label}`);
        const proto =
          node.tagName === "TEXTAREA"
            ? window.HTMLTextAreaElement.prototype
            : window.HTMLInputElement.prototype;
        Object.getOwnPropertyDescriptor(proto, "value")!.set!.call(node, value);
        node.dispatchEvent(new window.Event("input", { bubbles: true }));
        await flush();
      });
    },
    async selectType(value: string) {
      await act(async () => {
        const node = document.querySelector<HTMLSelectElement>("fieldset select")!;
        node.value = value;
        node.dispatchEvent(new window.Event("change", { bubbles: true }));
        await flush();
      });
    },
    async switchRepository() {
      repository = other;
      await render();
    },
    async verifyAs(value: string) {
      actor = value;
      await act(async () => {
        document
          .querySelector<HTMLElement>('[aria-label="身分選單"]')!
          .dispatchEvent(new window.KeyboardEvent("keydown", { key: "Enter", bubbles: true }));
        await flush();
      });
      await act(async () => {
        document.getElementById("verify")!.click();
        await flush();
      });
    },
    async dispose() {
      await act(async () => {
        root.unmount();
        await flush();
      });
      globalThis.fetch = oldFetch;
      window.liff = oldSDK;
    },
  };
}

test("Object creation needs only a name and retries an unknown result with the original identity", async () => {
  const posts: { key: string | null; body: unknown }[] = [];
  const f = await fixture(async (request) => {
    if (request.method === "POST") {
      posts.push({ key: request.headers.get("Idempotency-Key"), body: await request.json() });
      if (posts.length === 1) throw new Error("network lost");
      return Response.json(resource());
    }
    return Response.json({ items: [], nextCursor: "" });
  });
  try {
    await f.click("新增物件");
    await f.fill("物件名稱", "測試物件");
    assert.equal(document.querySelectorAll("[aria-label=新增物件] input").length, 1);
    for (const label of ["來源項次", "來源頁碼", "料號"])
      assert.ok(
        ![...document.querySelectorAll("label")].some((node) => node.textContent === label),
      );
    await f.click("建立物件");
    assert.equal(
      document.querySelector<HTMLInputElement>("[aria-label=新增物件] input")!.matches(":disabled"),
      true,
    );
    await f.click("重試原物件建立請求");
    assert.equal(posts.length, 2);
    assert.deepEqual(posts[0], posts[1]);
    assert.deepEqual(posts[0].body, { name: "測試物件" });
    for (const key of ["sourceItemNo", "sourcePages", "materialCode"])
      assert.ok(!(key in (posts[0].body as object)));
    assert.ok(!("status" in (posts[0].body as object)), "server owns initial tracking status");
    assert.match(document.body.textContent!, /已建立物件/);
  } finally {
    await f.dispose();
  }
});

test("Object detail is read-only for other creators and version conflicts require an explicit fresh read", async () => {
  let conflict = true;
  const patches: unknown[] = [];
  const f = await fixture(async (request, actor) => {
    if (request.method === "PATCH") {
      patches.push(await request.json());
      if (conflict)
        return Response.json(
          { error: { code: "version_conflict", message: "版本已變更" } },
          { status: 409 },
        );
      return Response.json(resource({ version: 3, status: "completed" }));
    }
    return Response.json(
      new URL(request.url).pathname.endsWith(oid)
        ? resource({ version: 2, canEdit: actor === "A" })
        : { items: [resource({ canEdit: actor === "A" })], nextCursor: "" },
    );
  });
  try {
    await f.click("測試物件");
    await f.click("儲存物件");
    assert.equal((patches[0] as { expectedVersion: number }).expectedVersion, 2);
    const save = [...document.querySelectorAll<HTMLButtonElement>("button")].find(
      (node) => node.textContent === "儲存物件",
    )!;
    assert.equal(save.disabled, true);
    conflict = false;
    await f.click("重新讀取物件");
    assert.equal(
      [...document.querySelectorAll<HTMLButtonElement>("button")].find(
        (node) => node.textContent === "儲存物件",
      )!.disabled,
      false,
    );
    await f.fill("完成數", "11");
    await f.selectType("non_work");
    await f.selectType("work");
    await f.click("儲存物件");
    assert.equal((patches[1] as { completedQuantity: string }).completedQuantity, "11");
    assert.ok(!("status" in (patches[1] as object)));
    assert.equal((patches[1] as { sourceRevision: string }).sourceRevision, "v0");
    for (const key of ["sourceItemNo", "sourcePages", "materialCode"])
      assert.ok(
        !(key in (patches[1] as object)),
        "hidden provenance is omitted so the API preserves its stored value",
      );
    await f.verifyAs("B");
    await f.click("測試物件");
    assert.equal(
      [...document.querySelectorAll("button")].some((node) => node.textContent === "儲存物件"),
      false,
    );
    assert.equal(
      document.querySelector<HTMLInputElement>("fieldset input")!.matches(":disabled"),
      true,
    );
  } finally {
    await f.dispose();
  }
});

for (const scope of ["repository", "identity"]) {
  test(`Object state rejects late list responses after ${scope} changes`, async () => {
    let finish!: (response: Response) => void;
    const f = await fixture(async (request, actor) =>
      new URL(request.url).pathname.includes(rid) && actor === "A"
        ? new Promise((resolve) => {
            finish = resolve;
          })
        : Response.json({ items: [resource({ id: other, name: "新範圍項目" })], nextCursor: "" }),
    );
    try {
      if (scope === "repository") await f.switchRepository();
      else await f.verifyAs("B");
      await act(async () => {
        finish(Response.json({ items: [resource({ name: "舊範圍秘密" })], nextCursor: "" }));
        await flush();
      });
      assert.ok(!document.body.textContent!.includes("舊範圍秘密"));
      assert.match(document.body.textContent!, /新範圍項目/);
    } finally {
      await f.dispose();
    }
  });
}

test("repository tab focus keeps the current panel until keyboard activation", async () => {
  const reads: string[] = [];
  const f = await fixture(async (request) => {
    reads.push(new URL(request.url).pathname);
    return Response.json({ items: [], nextCursor: "" });
  }, true);
  try {
    const issues = document.querySelector<HTMLButtonElement>('[role=tab][aria-label="問題"]')!;
    await act(async () => {
      issues.focus();
      await flush();
    });
    assert.equal(issues.getAttribute("aria-selected"), "false");
    assert.ok(!reads.some((path) => path.endsWith("/issues")));
    await act(async () => {
      issues.dispatchEvent(new window.KeyboardEvent("keydown", { key: "Enter", bubbles: true }));
      await flush();
    });
    assert.equal(issues.getAttribute("aria-selected"), "true");
    assert.ok(reads.some((path) => path.endsWith("/issues")));
  } finally {
    await f.dispose();
  }
});

test("repository exposes Problems and Objects and retains an uncertain creation when tab navigation is attempted", async () => {
  const f = await fixture(
    async (request) =>
      request.method === "POST"
        ? Promise.reject(new Error("unknown"))
        : Response.json({ items: [], nextCursor: "" }),
    true,
  );
  try {
    assert.ok(document.querySelector('[aria-label="問題"]'));
    assert.equal(document.querySelector('[aria-label="事項"]'), null);
    assert.equal(document.querySelector('[aria-label="空間功能"] [aria-label="專案"]'), null);
    assert.ok(document.querySelector('[aria-label="空間功能"] [aria-label="稽核"]'));
    assert.equal(document.querySelector('[aria-label="空間功能"] [aria-label="稽核紀錄"]'), null);
    assert.equal(document.querySelector('a[href^="/miniapp/projects?"]'), null);
    await f.click("物件");
    await f.click("新增物件");
    await f.fill("物件名稱", "保留請求");
    await f.click("建立物件");
    assert.equal(document.querySelector<HTMLButtonElement>('[aria-label="問題"]')!.disabled, true);
    await f.click("問題");
    assert.match(document.body.textContent!, /重試原物件建立請求/);
    assert.equal(
      [...document.querySelectorAll<HTMLButtonElement>("button")].find(
        (node) => node.textContent === "重新載入空間",
      ),
      undefined,
    );
  } finally {
    await f.dispose();
  }
});

for (const first of ["create", "tab"] as const) {
  test(`Object write and parent navigation handshake is synchronous (${first} first)`, async () => {
    let posts = 0;
    const f = await fixture(async (request) => {
      if (request.method === "POST") {
        posts++;
        throw new Error("unknown");
      }
      return Response.json({ items: [], nextCursor: "" });
    }, true);
    try {
      await f.click("物件");
      await f.click("新增物件");
      await f.fill("物件名稱", "同步建立");
      const form = document.querySelector("form")!;
      const tab = document.querySelector<HTMLButtonElement>('[aria-label="問題"]')!;
      await act(async () => {
        if (first === "create") {
          form.dispatchEvent(new window.Event("submit", { bubbles: true, cancelable: true }));
          pressButton(tab);
        } else {
          pressButton(tab);
          form.dispatchEvent(new window.Event("submit", { bubbles: true, cancelable: true }));
        }
        await flush();
      });
      assert.equal(
        posts,
        first === "create" ? 1 : 0,
        "pending navigation confirmation prevents a stale form submission",
      );
      if (first !== "create") assert.ok(document.querySelector("[role=alertdialog]"));
      if (first === "create") assert.match(document.body.textContent!, /重試原物件建立請求/);
    } finally {
      await f.dispose();
    }
  });
}

for (const direction of ["create-to-detail", "detail-to-create"] as const) {
  test(`Object internal navigation rejects a stale same-batch form submission (${direction})`, async () => {
    let writes = 0;
    const f = await fixture(async (request) => {
      if (request.method === "POST" || request.method === "PATCH") {
        writes++;
        return Response.json(resource());
      }
      return Response.json(
        new URL(request.url).pathname.endsWith(oid)
          ? resource()
          : { items: [resource()], nextCursor: "" },
      );
    });
    try {
      if (direction === "create-to-detail") {
        await f.click("新增物件");
        await f.fill("物件名稱", "舊建立意圖");
      } else {
        await f.click("測試物件");
        await f.fill("物件名稱", "舊修改意圖");
      }
      const oldForm = document.querySelector("form")!;
      await act(async () => {
        const target = [...document.querySelectorAll<HTMLButtonElement>("button")].find(
          (button) =>
            (button.getAttribute("aria-label") || button.textContent) ===
            (direction === "create-to-detail" ? "測試物件" : "新增物件"),
        )!;
        target.click();
        oldForm.dispatchEvent(new window.Event("submit", { bubbles: true, cancelable: true }));
        await flush();
      });
      assert.equal(writes, 0, "navigation must invalidate the previous form before React commits");
    } finally {
      await f.dispose();
    }
  });
}

test("object list displays five labeled lines and omits source item metadata", async () => {
  const name = "ExtremelyLongUnbrokenObjectNameWithoutSpacesThatMustRemainComplete1234567890";
  const f = await fixture(async () =>
    Response.json({
      items: [
        resource({
          name,
          type: "Engineering",
          sourceItemNo: "450",
          quantity: "9007199254740993.123456",
        }),
      ],
      nextCursor: "",
    }),
  );
  try {
    const item = document.querySelector('[data-slot="item-group"] [data-slot="item"]')!;
    assert.equal(item.tagName, "DIV");
    assert.ok(item.querySelector(`[aria-label="在「${name}」下新增子物件"]`));
    assert.equal(item.querySelectorAll("button button").length, 0);
    assert.match(item.textContent!, /9007199254740993\.123456 SET/);
    const lines = [
      ...item.querySelectorAll('[data-slot="item-title"] button[data-variant="link"],p'),
    ].map((node) => node.textContent);
    assert.deepEqual(lines, [
      `名稱：${name}`,
      "數量：9007199254740993.123456 SET",
      "類型：Engineering（待確認）",
      "狀態：未開始",
      "進度：待補／待確認",
    ]);
    assert.ok(!item.textContent!.includes("來源項次"));
  } finally {
    await f.dispose();
  }
});

test("work completion editing uses exact decimals and non-work omits completion and status", async () => {
  const writes: Record<string, unknown>[] = [];
  let current = resource();
  const f = await fixture(async (request) => {
    if (request.method === "PATCH") {
      const body = await request.json();
      writes.push(body);
      current = resource({ ...body, version: current.version + 1 });
      return Response.json(current);
    }
    return Response.json(
      new URL(request.url).pathname.endsWith(oid) ? current : { items: [current], nextCursor: "" },
    );
  });
  try {
    await f.click("數量：11 SET");
    assert.equal(document.activeElement?.getAttribute("data-edit-field"), "quantity");
    await f.fill("數量", "9007199254740993.123455");
    await f.fill("完成數", "9007199254740993.123456");
    await f.click("儲存物件");
    assert.equal(writes.length, 0);
    assert.match(document.body.textContent!, /0至數量之間/);
    await f.fill("完成數", "9007199254740993.123454");
    await f.click("儲存物件");
    assert.equal(writes[0].completedQuantity, "9007199254740993.123454");
    assert.ok(!("status" in writes[0]));
    await f.click("測試物件");
    await f.fill("完成數", "0");
    await f.click("儲存物件");
    await f.click("測試物件");
    await f.selectType("non_work");
    assert.equal(
      [...document.querySelectorAll("label")].some((node) => node.textContent === "完成數"),
      false,
    );
    await f.click("儲存物件");
    assert.equal(writes[2].type, "non_work");
    assert.ok(!("completedQuantity" in writes[2]));
  } finally {
    await f.dispose();
  }
});

test("legacy progress requires an explicit value and cannot silently reclassify", async () => {
  const patches: Record<string, unknown>[] = [];
  const old = resource({
    type: "Engineering",
    completedQuantity: undefined,
    status: "in_progress",
  });
  const f = await fixture(async (request) => {
    if (request.method === "PATCH") {
      patches.push(await request.json());
      return Response.json(old);
    }
    return Response.json(
      new URL(request.url).pathname.endsWith(oid) ? old : { items: [old], nextCursor: "" },
    );
  });
  try {
    await f.click("測試物件");
    assert.equal(
      document.querySelector<HTMLSelectElement>("fieldset select")!.value,
      "Engineering",
    );
    await f.click("儲存物件");
    assert.equal(patches[0].type, "Engineering");
    assert.ok(!("completedQuantity" in patches[0]));
    await f.click("測試物件");
    await f.selectType("work");
    await f.click("儲存物件");
    assert.equal(patches.length, 1);
    await f.fill("完成數", "3.500000");
    await f.click("儲存物件");
    assert.equal(patches[1].completedQuantity, "3.500000");
  } finally {
    await f.dispose();
  }
});

test("work with progress cannot change to non-work before saving zero", async () => {
  let writes = 0;
  const item = resource({
    completedQuantity: "0.000001",
    status: "in_progress",
    requiresProgressReset: true,
  });
  const f = await fixture(async (request) => {
    if (request.method === "PATCH") writes++;
    return Response.json(
      new URL(request.url).pathname.endsWith(oid) ? item : { items: [item], nextCursor: "" },
    );
  });
  try {
    await f.click("測試物件");
    await f.selectType("non_work");
    await f.click("儲存物件");
    assert.equal(writes, 0);
    assert.match(document.body.textContent!, /先將完成數改為0並儲存/);
  } finally {
    await f.dispose();
  }
});

test("list distinguishes measured, non-work, and legacy missing progress", async () => {
  const f = await fixture(async () =>
    Response.json({
      items: [
        resource({
          id: "1".repeat(32),
          name: "有進度",
          completedQuantity: "2.500000",
          status: "in_progress",
        }),
        resource({
          id: "2".repeat(32),
          name: "不追蹤",
          type: "non_work",
          completedQuantity: undefined,
        }),
        resource({
          id: "3".repeat(32),
          name: "舊工作",
          completedQuantity: undefined,
          status: "in_progress",
        }),
      ],
      nextCursor: "",
    }),
  );
  try {
    const text = (name: string) =>
      document.querySelector(`[aria-label="${name}"]`)!.closest('[data-slot="item"]')!.textContent!;
    assert.match(text("有進度"), /類型：施作狀態：進行中進度：2.500000\/11 SET/);
    assert.match(text("不追蹤"), /類型：非施作狀態：不適用進度：不適用/);
    assert.match(text("舊工作"), /進度：待補／待確認/);
  } finally {
    await f.dispose();
  }
});

test("object editor stays in its row, cancel preserves a draft until confirmed, save closes and updates only that row", async () => {
  let writes = 0;
  const f = await fixture(async (request) => {
    if (request.method === "PATCH") {
      writes++;
      return Response.json(resource({ name: "更新名稱", version: 2 }));
    }
    return Response.json(
      new URL(request.url).pathname.endsWith(oid)
        ? resource()
        : { items: [resource(), resource({ id: other, name: "另一物件" })], nextCursor: "" },
    );
  });
  try {
    await f.click("測試物件");
    const row = document.querySelector('[aria-label="測試物件"]')!.closest('[data-slot="item"]')!;
    assert.ok(row.querySelector("form"));
    assert.equal(document.querySelectorAll("form").length, 1);
    await f.fill("物件名稱", "更新名稱");
    await f.click("另一物件");
    assert.ok(document.querySelector("[role=alertdialog]"));
    assert.equal(document.activeElement?.textContent, "繼續編輯");
    const input = row.querySelector("input")!;
    async function continueEditing() {
      // Radix restores focus from a timer after the dialog unmounts.
      await Promise.all([
        once(input, "focus", { signal: AbortSignal.timeout(2_000) }),
        f.click("繼續編輯"),
      ]);
      assert.ok(document.activeElement === input, "focus returns to the editor");
    }
    await continueEditing();
    await f.click("取消");
    await continueEditing();
    assert.equal(writes, 0);
    assert.equal(row.querySelector<HTMLInputElement>("input")!.value, "更新名稱");
    await f.click("儲存物件");
    assert.equal(writes, 1);
    assert.equal(document.querySelector("form"), null);
    assert.ok(document.querySelector('[aria-label="更新名稱"]'));
    assert.ok(document.querySelector('[aria-label="另一物件"]'));
    await f.click("更新名稱");
    await f.fill("物件名稱", "放棄內容");
    await f.click("取消");
    await f.click("放棄修改");
    assert.equal(document.querySelector("form"), null);
    assert.equal(writes, 1);
  } finally {
    await f.dispose();
  }
});

test("document candidates require review and use the object creation form", async () => {
  const writes: unknown[] = [];
  const f = await fixture(
    async (request) => {
      if (new URL(request.url).pathname.endsWith("/candidates")) {
        assert.deepEqual(await request.json(), { filename: "list.txt", data: "YWJj" });
        return Response.json({
          status: "completed",
          rawResponse: "test response",
          candidates: [
            { name: "電纜", quantity: "2.50", location: "第 1 行", quote: "電纜 2.50" },
            { name: "配管", quote: "" },
          ],
        });
      }
      if (request.method === "POST") {
        writes.push(await request.json());
        return Response.json(resource());
      }
      return Response.json({ items: [], nextCursor: "" });
    },
    false,
    true,
  );
  try {
    await act(async () => {
      const input = document.querySelector<HTMLInputElement>('input[type="file"]')!;
      assert.equal(
        input.hasAttribute("accept"),
        false,
        "native picker must not filter document formats",
      );
      assert.equal(input.disabled, false);
      Object.defineProperty(input, "files", { value: [new window.File(["abc"], "list.txt")] });
      input.dispatchEvent(new window.Event("change", { bubbles: true }));
    });
    await f.click("辨識候選");
    assert.match(document.body.textContent ?? "", /引文待核對/);
    await f.click("覆核並填入新增表單");
    assert.equal(writes.length, 0);
    await f.click("建立物件");
    assert.equal(writes.length, 1);
    assert.deepEqual(writes[0], {
      name: "電纜",
      description: "電纜 2.50",
      sourceDocument: "list.txt",
      sourcePages: "第 1 行",
    });
    const links = [...document.querySelectorAll("a")].filter(
      (link) => link.textContent === "查看已建立物件",
    );
    assert.equal(links.length, 1);
    assert.equal(new URL(links[0].href).searchParams.get("object"), oid);
    assert.ok(
      [...document.querySelectorAll("button")].find((b) => b.textContent === "已建立")?.disabled,
    );
  } finally {
    await f.dispose();
  }
});

test("work conversion uses the API preset and the edited exact quantity", async () => {
  const old = resource({
    type: "Engineering",
    completedQuantity: undefined,
    workConversionDefault: "quantity",
  });
  let patch: Record<string, unknown> | undefined;
  const f = await fixture(async (request) => {
    if (request.method === "PATCH") patch = await request.json();
    return Response.json(
      new URL(request.url).pathname.endsWith(oid) ? old : { items: [old], nextCursor: "" },
    );
  });
  try {
    await f.click("測試物件");
    await f.fill("數量", "9007199254740993.123456");
    await f.selectType("work");
    await f.click("儲存物件");
    assert.equal(patch?.completedQuantity, "9007199254740993.123456");
  } finally {
    await f.dispose();
  }
});

test("batch import only retries unknown selected rows with their original keys", async () => {
  const posts: { key: string | null; body: { name: string } }[] = [];
  const f = await fixture(
    async (request) => {
      if (new URL(request.url).pathname.endsWith("/candidates")) {
        return Response.json({
          status: "completed",
          candidates: ["電纜", "配管", "開關"].map((name) => ({ name, quote: name })),
        });
      }
      if (request.method === "POST") {
        const body = await request.json();
        posts.push({ key: request.headers.get("Idempotency-Key"), body });
        if (posts.length === 2) throw new Error("network lost");
        return Response.json(resource({ id: posts.length === 1 ? oid : other }));
      }
      return Response.json({ items: [], nextCursor: "" });
    },
    false,
    true,
  );
  try {
    await act(async () => {
      const input = document.querySelector<HTMLInputElement>('input[type="file"]')!;
      Object.defineProperty(input, "files", { value: [new window.File(["abc"], "list.txt")] });
      input.dispatchEvent(new window.Event("change", { bubbles: true }));
    });
    await f.click("辨識候選");
    await act(async () => {
      const boxes = document.querySelectorAll<HTMLInputElement>('input[type="checkbox"]');
      boxes[0].click();
      boxes[1].click();
    });
    assert.equal(posts.length, 0);
    await f.click("確認匯入 2 筆");
    assert.deepEqual(
      posts.map(({ body }) => body.name),
      ["電纜", "配管"],
    );
    assert.ok(posts[0].key && posts[1].key);
    assert.notEqual(posts[0].key, posts[1].key);
    await f.click("重試 1 筆未確認結果");
    assert.equal(posts.length, 3);
    assert.deepEqual(posts[2], posts[1]);
    assert.equal(
      [...document.querySelectorAll("a")].filter((a) => a.textContent === "查看已建立物件").length,
      2,
    );
  } finally {
    await f.dispose();
  }
});
