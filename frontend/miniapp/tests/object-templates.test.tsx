import React, { act } from "react";
import { createRoot } from "react-dom/client";
import test from "node:test";
import assert from "node:assert/strict";
import { pressButton } from "./sidebar-dom";
import { Shell } from "./shell-fixture";
import TemplatePage from "../src/pages/miniapp/object-templates";
import { MiniAppProviders } from "../src/shell/MiniAppProviders";
import { View } from "../src/business/object/View";

const rid = "a".repeat(32),
  tid = "b".repeat(32),
  rootId = "c".repeat(32),
  childId = "d".repeat(32);
const base = `/api/v1/repositories/${rid}`;
const snapshot = {
  id: tid,
  repositoryId: rid,
  name: "範本",
  version: 1,
  creatorId: "A",
  createdAt: "2026-11-21T00:00:00Z",
  nodes: [
    { id: "root", name: "工程" },
    { id: "child", parentId: "root", name: "工序" },
  ],
};
const instance = {
  rootObjectId: rootId,
  objectIds: [rootId, childId],
  templateId: tid,
  templateVersion: 1,
  nodes: [
    { objectId: rootId, parentId: "", name: "工程" },
    { objectId: childId, parentId: rootId, name: "工序" },
  ],
};
const resource = (id: string) => ({
  id,
  repositoryId: rid,
  name: id === rootId ? "工程" : "工序",
  parentId: id === childId ? rootId : undefined,
  description: "",
  type: "",
  quantity: "",
  unit: "",
  status: "not_started",
  creatorId: "A",
  canEdit: true,
  version: 1,
  createdAt: snapshot.createdAt,
  updatedAt: snapshot.createdAt,
});
const flush = async () => {
  for (let i = 0; i < 10; i++) await new Promise<void>((resolve) => setImmediate(resolve));
};
async function fixture(handle: (r: Request) => Promise<Response>, standalone = false) {
  document.body.innerHTML = '<div id="root"></div>';
  window.history.replaceState(null, "", "/miniapp/projects");
  const oldFetch = globalThis.fetch,
    oldSDK = window.liff;
  window.liff = {
    init: async () => {},
    isLoggedIn: () => true,
    isInClient: () => false,
    getIDToken: () => "token",
    getProfile: async () => ({ userId: "A", displayName: "A" }),
    login() {},
    logout() {},
    closeWindow() {},
  };
  globalThis.fetch = async (input, init) => {
    const r =
      input instanceof Request ? input : new Request(new URL(String(input), location.origin), init);
    const path = new URL(r.url).pathname;
    if (path.endsWith("/config")) return Response.json({ liffId: "template-test" });
    if (path.endsWith("/session"))
      return Response.json({
        user: { userId: "A", displayName: "A" },
        csrfToken: "c".repeat(43),
        expiresAt: new Date(Date.now() + 1800000).toISOString(),
      });
    if (path === `${base}/objects`) return Response.json({ items: [], nextCursor: "" });
    if (path === `${base}/objects/${rootId}` || path === `${base}/objects/${childId}`)
      return Response.json(resource(path.split("/").at(-1)!));
    return handle(r);
  };
  const root = createRoot(document.getElementById("root")!);
  await act(async () => {
    root.render(
      standalone ? (
        <MiniAppProviders>
          <TemplatePage />
        </MiniAppProviders>
      ) : (
        <Shell title="測試" workflow="projects">
          <View repositoryId={rid} />
        </Shell>
      ),
    );
    await flush();
  });
  return {
    async click(label: string) {
      await act(async () => {
        const b = [...document.querySelectorAll<HTMLButtonElement>("button")].find(
          (b) => b.textContent?.trim() === label || b.getAttribute("aria-label") === label,
        );
        assert.ok(b, `missing ${label}`);
        pressButton(b);
        await flush();
      });
    },
    async fill(label: string, value: string) {
      await act(async () => {
        const input = [...document.querySelectorAll("label")]
          .find((n) => n.firstChild?.textContent === label || n.textContent === label)
          ?.querySelector<HTMLInputElement | HTMLTextAreaElement>("input, textarea");
        assert.ok(input, label);
        Object.getOwnPropertyDescriptor(Object.getPrototypeOf(input), "value")!.set!.call(
          input,
          value,
        );
        input.dispatchEvent(new window.Event("input", { bubbles: true }));
        await flush();
      });
    },
    async select(id: string, label = "選擇範本") {
      await act(async () => {
        const select = [...document.querySelectorAll("label")]
          .find((node) => node.textContent?.startsWith(label))
          ?.querySelector<HTMLSelectElement>("select")!;
        select.value = id;
        select.dispatchEvent(new window.Event("change", { bubbles: true }));
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

test("repository objects expose object creation and the sidebar link only", async () => {
  const f = await fixture(async () => Response.json({ items: [], nextCursor: "" }));
  try {
    assert.ok(
      document.querySelector('button[aria-label="新增物件"]') !== null,
    );
    assert.ok(
      ![...document.querySelectorAll("button")].some((button) => button.textContent === "物件範本"),
    );
    assert.ok(document.querySelector('a[href="/miniapp/object-templates"]'));
  } finally {
    await f.dispose();
  }
});

test("template creation and application preserve original retries and return the creation result", async () => {
  const creates: { key: string | null; body: unknown }[] = [],
    applies: typeof creates = [];

  const f = await fixture(async (r) => {
    const path = new URL(r.url).pathname;
    if (r.method === "GET")
      return Response.json({
        items: path === "/api/v1/repositories" ? [{ id: rid, name: "目標庫" }] : [],
        nextCursor: "",
      });
    const log = path.endsWith("/instances") ? applies : creates;
    log.push({ key: r.headers.get("Idempotency-Key"), body: await r.json() });
    if (log.length === 1) throw new Error("response lost");
    return Response.json(path.endsWith("/instances") ? instance : snapshot);
  }, true);
  try {
    await f.click("新增範本");
    await f.click("從空白建立");
    await f.fill("物件名稱", "工程");
    await f.click("在工程下新增子項");
    await f.fill("子項名稱", "工序");
    await f.click("在工序下新增子項");
    await act(async () => {
      const inputs = document.querySelectorAll("textarea");
      const last = inputs[inputs.length - 1];
      Object.getOwnPropertyDescriptor(Object.getPrototypeOf(last), "value")!.set!.call(
        last,
        "檢查",
      );
      last.dispatchEvent(new window.Event("input", { bubbles: true }));
    });
    await f.click("建立範本");
    assert.equal(
      [...document.querySelectorAll<HTMLInputElement>('[aria-label="物件範本操作"] input')].every(
        (n) => n.matches(":disabled"),
      ),
      true,
    );
    await f.click("重試原範本請求");
    const { nodes } = creates[0].body as { nodes: typeof snapshot.nodes };
    assert.equal(nodes[2].parentId, nodes[1].id);
    assert.equal(creates.length, 2);
    assert.deepEqual(creates[0], creates[1]);
    await f.click("範本v1");
    assert.doesNotMatch(document.body.textContent ?? "", /使用範本/);
    await f.click("範本v1");
    assert.match(document.body.textContent ?? "", /使用範本/);
    await f.click("使用範本");
    assert.equal(applies.length, 0);
    await f.select(rid, "目標儲存庫");
    await f.click("工程");
    await f.fill("數量", "1.1234567");
    assert.equal(
      [...document.querySelectorAll<HTMLButtonElement>('button[type="submit"]')].find(
        (button) => button.textContent === "建立 2 個物件",
      )?.disabled,
      true,
    );
    assert.equal(applies.length, 0);
    await f.fill("數量", "9007199254740993.123456");
    await f.click("建立 2 個物件");
    await f.click("重試原範本請求");
    assert.equal(applies.length, 2);
    assert.deepEqual(applies[0], applies[1]);
    assert.deepEqual(applies[0].body, {
      templateVersion: 1,
      nodes: [{ ...snapshot.nodes[0], quantity: "9007199254740993.123456" }, snapshot.nodes[1]],
    });
    assert.match(document.body.textContent!, /已建立 2 個物件/);
  } finally {
    await f.dispose();
  }
});

test("template validation rejection unlocks editing and permits a corrected request", async () => {
  let tries = 0;
  const keys: string[] = [];
  const f = await fixture(async (r) => {
    if (r.method === "GET") return Response.json({ items: [], nextCursor: "" });
    if (new URL(r.url).pathname !== "/api/v1/object-templates") return Response.json({});
    keys.push(r.headers.get("Idempotency-Key")!);
    tries++;
    if (tries === 1)
      return Response.json(
        { error: { code: "invalid_input", message: "範本無效" } },
        { status: 400 },
      );
    return Response.json(snapshot);
  }, true);
  try {
    await f.click("新增範本");
    await f.click("從空白建立");
    await f.fill("物件名稱", "工程");
    await f.click("建立範本");
    assert.equal(
      document
        .querySelector<HTMLInputElement>('[aria-label="物件範本操作"] textarea')!
        .matches(":disabled"),
      false,
    );
    await f.fill("物件名稱", "修正工程");
    await f.click("建立範本");
    assert.equal(tries, 2);
    assert.notEqual(keys[0], keys[1]);
  } finally {
    await f.dispose();
  }
});

test("personal library uses the existing editor and submits only after confirmation", async () => {
  const calls: string[] = [];
  const f = await fixture(async (request) => {
    const url = new URL(request.url);
    if (url.pathname === "/api/v1/repositories")
      return Response.json({ items: [{ id: rid, name: "目標庫" }], nextCursor: "" });
    if (url.pathname.endsWith("/instances"))
      assert.equal(url.searchParams.get("scope"), "personal");
    if (request.method === "GET") return Response.json({ items: [], nextCursor: "" });
    calls.push(url.pathname);
    const body = await request.json();
    if (url.pathname.endsWith("/instances")) {
      assert.equal(body.nodes[0].quantity, "6400");
      return Response.json(instance);
    }
    assert.equal(body.nodes[0].quantity, undefined);
    assert.equal(body.nodes[0].unit, "米");
    assert.equal(body.name, body.nodes[0].name);
    return Response.json({ ...snapshot, repositoryId: "" });
  }, true);
  try {
    await f.click("新增範本");
    await f.click("從空白建立");
    assert.doesNotMatch(document.body.textContent!, /根物件名稱|範本名稱/);
    await f.fill("物件名稱", "工程");
    await f.fill("單位", "米");
    await f.click("建立範本");
    await f.click("使用範本");
    assert.equal(calls.length, 1);
    await f.select(rid, "目標儲存庫");
    await f.click("工程");
    await f.fill("數量", "6400");
    await f.click("建立 2 個物件");
    assert.equal(calls.length, 2);
  } finally {
    await f.dispose();
  }
});

test("document extraction creates only an editable personal draft and preserves the work tree", async () => {
  const saves: Array<typeof snapshot> = [];
  const f = await fixture(async (request) => {
    const url = new URL(request.url);
    if (request.method === "GET") return Response.json({ items: [], nextCursor: "" });
    if (url.pathname === "/api/v1/object-template-drafts") {
      const input = await request.json();
      assert.equal(input.source, "document");
      assert.equal(input.filename, "工程.txt");
      return Response.json({
        draft: {
          name: "工程",
          nodes: [
            { id: "root", name: "工程" },
            {
              id: "a",
              parentId: "root",
              name: "（一）SCADA站內工程：\n既設161kv RTU",
              unit: "SET",
            },
            { id: "b", parentId: "root", name: "電力系統" },
            { id: "c", parentId: "b", name: "高壓配電" },
            { id: "d", parentId: "c", name: "地板保護", unit: "式" },
            { id: "e", parentId: "b", name: "接地工程" },
            { id: "f", parentId: "e", name: "地板保護", unit: "式" },
            { id: "g", parentId: "e", name: "搬運定位（=6組）", unit: "式" },
          ],
        },
        candidateIndexByNode: { a: 3, d: 0, f: 1, g: 2 },
        warnings: [],
        candidates: [
          {
            categoryPath: ["電力系統", "高壓配電"],
            name: "地板保護",
            quote: "地板保護 1 式",
            quantity: "1",
            unit: "式",
          },
          {
            categoryPath: ["電力系統", "接地工程"],
            name: "地板保護",
            quote: "地板保護 1 式",
            quantity: "1",
            unit: "式",
          },
          {
            categoryPath: ["電力系統", "接地工程"],
            name: "搬運定位（=6組）",
            quote: "搬運定位（=6組） 12 式",
            quantity: "12",
            unit: "式",
          },
          {
            name: "（一）SCADA站內工程：\n既設161kv RTU",
            quote: "既設161kv RTU",
            quantity: "1",
            unit: "SET",
          },
        ],
      });
    }
    if (url.pathname.endsWith("/instances"))
      assert.equal(url.searchParams.get("scope"), "personal");
    const body = await request.json();
    saves.push(body);
    return Response.json({ ...snapshot, ...body, repositoryId: "" });
  }, true);
  try {
    await f.click("新增範本");
    await f.click("從文件建立");
    await act(async () => {
      const panel = document.querySelector('[aria-label="物件範本操作"]')!;
      const input = panel.querySelector<HTMLInputElement>('input[type="file"]')!;
      Object.defineProperty(input, "files", { value: [new window.File(["sample"], "工程.txt")] });
      input.dispatchEvent(new window.Event("change", { bubbles: true }));
      await flush();
      pressButton(
        [...panel.querySelectorAll<HTMLButtonElement>("button")].find(
          (button) => button.textContent === "辨識候選",
        )!,
      );
      await flush();
    });
    const moving = [...document.querySelectorAll("textarea")].find((field) =>
      field.value.includes("搬運定位"),
    );
    assert.match(moving?.closest("li")?.textContent ?? "", /搬運定位（=6組） 12 式/);
    assert.match(document.body.textContent!, /12 式/);
    assert.equal(saves.length, 0);
    await f.click("建立範本");
    assert.equal(saves.length, 1);
    const nodes = saves[0].nodes;
    const floors = nodes.filter((node) => node.name === "地板保護");
    assert.equal(floors.length, 2);
    assert.notEqual(floors[0].parentId, floors[1].parentId);
    assert.ok(nodes.some((node) => node.name === "（一）SCADA站內工程：\n既設161kv RTU"));
    assert.ok(nodes.every((node) => !("quantity" in node)));
  } finally {
    await f.dispose();
  }
});

test("standalone templates create without repositories and select a target only when applying", async () => {
  let repositoryReads = 0;
  const writes: { path: string; key: string | null; body: string }[] = [];
  let attempts = 0;
  const f = await fixture(async (request) => {
    const path = new URL(request.url).pathname;
    if (path === "/api/v1/repositories") {
      repositoryReads++;
      return Response.json({
        items: [{ id: rid, name: "目標庫", status: "ACTIVE" }],
        nextCursor: "",
      });
    }
    if (path === "/api/v1/object-templates" && request.method === "GET")
      return Response.json({ items: [], nextCursor: "" });
    if (request.method === "POST") {
      writes.push({
        path,
        key: request.headers.get("Idempotency-Key"),
        body: await request.text(),
      });
      if (path === "/api/v1/object-templates") return Response.json(snapshot);
      if (++attempts === 1) throw new TypeError("offline");
      return Response.json(instance);
    }
    throw new Error(`unexpected ${path}`);
  }, true);
  try {
    assert.equal(repositoryReads, 0);
    await f.click("新增範本");
    await f.click("從空白建立");
    await f.fill("物件名稱", "工程");
    await f.click("建立範本");
    assert.equal(writes[0].path, "/api/v1/object-templates");
    assert.equal(repositoryReads, 0);
    await f.click("使用範本");
    assert.ok(repositoryReads > 0);
    const submit = [...document.querySelectorAll<HTMLButtonElement>("button")].find(
      (b) => b.textContent === "建立 2 個物件",
    )!;
    assert.equal(submit.disabled, true);
    await f.select(rid, "目標儲存庫");
    await f.click("建立 2 個物件");
    assert.equal(document.querySelector<HTMLSelectElement>("select")!.disabled, true);
    await f.click("重試原範本請求");
    assert.deepEqual(writes[1], writes[2]);
    assert.equal(writes[1].path, `${base}/object-templates/${tid}/instances`);
    assert.match(document.body.textContent!, /已建立 2 個物件/);
    const link = document.querySelector<HTMLAnchorElement>('a[href*="object="]')!;
    assert.equal(link.pathname, "/miniapp/repositories");
    assert.equal(new URL(link.href).searchParams.get("object"), rootId);
    assert.equal(new URL(link.href).searchParams.get("id"), rid);
    await f.click("返回我的範本");
    assert.match(document.body.textContent!, /我的物件範本/);
  } finally {
    await f.dispose();
  }
});

test("standalone procedure suggestions remain drafts until reviewed and explicitly saved", async () => {
  const saves: unknown[] = [];
  let preparations = 0;
  const draft = {
    name: "浴室防水",
    nodes: [
      { id: "root", name: "浴室防水" },
      { id: "server-step", parentId: "root", name: "基面處理" },
    ],
  };
  const f = await fixture(async (request) => {
    const path = new URL(request.url).pathname;
    if (request.method === "GET") return Response.json({ items: [], nextCursor: "" });
    if (path === "/api/v1/object-template-drafts") {
      preparations++;
      assert.deepEqual(await request.json(), {
        source: "procedure",
        name: "浴室防水",
        conditions: "水泥基面",
      });
      if (preparations === 1)
        return Response.json({ error: "provider_error", message: "暫時無法推導" }, { status: 503 });
      return Response.json({
        draft,
        candidateIndexByNode: { "server-step": 0 },
        warnings: ["工序為推導建議，請核對。"],
        candidates: [
          {
            name: "基面處理",
            quote: "需先確認基面狀態",
            quantity: null,
            unit: null,
            location: null,
          },
        ],
      });
    }
    assert.equal(path, "/api/v1/object-templates");
    saves.push(await request.json());
    return Response.json({ ...snapshot, ...draft, repositoryId: "" });
  }, true);
  try {
    await f.click("新增範本");
    await f.click("從工項推導工序");
    await f.fill("工項名稱", "浴室防水");
    await f.fill("施工條件", "水泥基面");
    await f.click("推導工序");
    assert.equal(document.body.textContent!.includes("編輯範本"), false);
    assert.equal(saves.length, 0);
    await f.click("推導工序");
    assert.match(document.body.textContent!, /推導說明：需先確認基面狀態/);
    assert.equal(saves.length, 0);
    assert.match(document.body.textContent!, /推導說明：需先確認基面狀態/);
    assert.match(document.body.textContent!, /工序為推導建議，請核對/);
    await f.click("收合子工序");
    assert.doesNotMatch(document.body.textContent!, /推導說明：需先確認基面狀態/);
    await f.click("展開子工序");
    await f.click("建立範本");
    assert.deepEqual(saves, [draft]);
  } finally {
    await f.dispose();
  }
});

test("cancelled preparation keeps input and ignores a late result", async () => {
  let finish!: (response: Response) => void;
  const f = await fixture(async (r) => {
    if (r.method === "GET") return Response.json({ items: [], nextCursor: "" });
    return new Promise<Response>((resolve) => {
      finish = resolve;
    });
  }, true);
  try {
    await f.click("新增範本");
    await f.click("從工項推導工序");
    await f.fill("工項名稱", "測試工項");
    await f.click("推導工序");
    await f.click("取消等待");
    await act(async () => {
      finish(Response.json({ draft: snapshot, candidates: [], warnings: [] }));
      await flush();
    });
    assert.doesNotMatch(document.body.textContent!, /編輯範本/);
    assert.match(document.body.textContent!, /已取消等待/);
    assert.ok([...document.querySelectorAll("input")].some((input) => input.value === "測試工項"));
  } finally {
    await f.dispose();
  }
});

test("empty library leads to creation methods and supports returning", async () => {
  const f = await fixture(async () => Response.json({ items: [], nextCursor: "" }), true);
  try {
    assert.doesNotMatch(document.body.textContent!, /來源文件|選擇範本/);
    await f.click("新增範本");
    await f.click("從文件建立");
    assert.ok(document.querySelector("input[type=file]"));
    await f.click("返回建立方式");
    await f.click("從空白建立");
    await f.click("返回範本庫");
    assert.match(document.body.textContent!, /尚無範本/);
  } finally {
    await f.dispose();
  }
});

test("edit a saved personal template and apply its next version", async () => {
  let revision = { ...snapshot, repositoryId: "" };
  const f = await fixture(async (request) => {
    const path = new URL(request.url).pathname;
    if (path === "/api/v1/repositories")
      return Response.json({ items: [{ id: rid, name: "目標庫" }], nextCursor: "" });
    if (request.method === "GET") return Response.json({ items: [revision], nextCursor: "" });
    const body = await request.json();
    if (path === "/api/v1/object-templates") {
      assert.deepEqual(body, {
        templateId: tid,
        baseVersion: 1,
        name: "新版",
        nodes: [{ ...snapshot.nodes[0], name: "新版" }, snapshot.nodes[1]],
      });
      revision = { ...revision, name: body.name, nodes: body.nodes, version: 2 };
      return Response.json(revision);
    }
    assert.equal(body.templateVersion, 2);
    assert.equal(body.nodes[0].name, "A 棟");
    assert.equal(revision.nodes[0].name, "新版");
    return Response.json({ ...instance, templateVersion: 2 });
  }, true);
  try {
    await f.click(snapshot.name + "v1");
    await f.click("編輯範本");
    await f.fill("物件名稱", "新版");
    await f.click("儲存為 v2");
    assert.match(document.body.textContent!, /新版 · v2/);
    await f.click("使用範本");
    await f.select(rid, "目標儲存庫");
    await f.click("新版");
    await f.fill("物件名稱", "A 棟");
    await f.click("建立 2 個物件");
    assert.match(document.body.textContent!, /已建立 2 個物件/);
  } finally {
    await f.dispose();
  }
});

test("a stale personal revision reloads the library before editing the latest version", async () => {
  let currentVersion = 1;
  const submitted: number[] = [];
  const f = await fixture(async (request) => {
    if (request.method === "GET")
      return Response.json({ items: [{ ...snapshot, version: currentVersion }], nextCursor: "" });
    const body = await request.json();
    submitted.push(body.baseVersion);
    if (body.baseVersion !== currentVersion)
      return Response.json(
        { error: { code: "version_conflict", message: "範本已有新版" } },
        { status: 409 },
      );
    return Response.json({
      ...snapshot,
      name: body.name,
      nodes: body.nodes,
      version: ++currentVersion,
    });
  }, true);
  try {
    await f.click(snapshot.name + "v1");
    await f.click("編輯範本");
    currentVersion = 2;
    await f.click("儲存為 v2");
    assert.match(document.body.textContent!, /範本已有新版/);
    assert.doesNotMatch(document.body.textContent!, /重試原範本請求/);
    await f.click("返回範本庫");
    await f.click("放棄修改");
    await f.click(snapshot.name + "v2");
    await f.click("編輯範本");
    await f.click("儲存為 v3");
    assert.deepEqual(submitted, [1, 2]);
  } finally {
    await f.dispose();
  }
});
