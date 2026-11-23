import React, { act } from "react";
import { createRoot } from "react-dom/client";
import test from "node:test";
import assert from "node:assert/strict";
import { pressButton } from "./sidebar-dom";
import { Shell } from "./shell-fixture";
import { View } from "../src/business/repository/View";

const rid = "a".repeat(32),
  id = "b".repeat(32);
const item = (overrides = {}) => ({
  id,
  repositoryId: rid,
  number: 1,
  title: "測試項目",
  body: "原內容",
  dueDate: "2026-12-01",
  state: "open",
  canEdit: true,
  creatorId: "A",
  authorId: "A",
  version: 2,
  labelIds: [],
  milestoneId: null,
  issueCount: 0,
  completedIssueCount: 0,
  ...overrides,
});
const flush = async () => {
  for (let i = 0; i < 10; i++) await new Promise<void>((resolve) => setImmediate(resolve));
};
async function fixture(
  handle: (request: Request) => Promise<Response | undefined> = async () => undefined,
) {
  document.body.innerHTML = '<div id="root"></div>';
  window.history.replaceState(null, "", `/miniapp/repositories?selected=${rid}`);
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
    const request =
      input instanceof Request ? input : new Request(new URL(String(input), location.origin), init);
    const path = new URL(request.url).pathname;
    if (path.endsWith("/config")) return Response.json({ liffId: "test" });
    if (path.endsWith("/session"))
      return Response.json({
        user: { userId: "A", displayName: "A" },
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
      });
    const response = await handle(request);
    if (response) return response;
    if (path.endsWith(id)) return Response.json(item());
    return Response.json({
      items: /\/(issues|milestones|discussions)$/.test(path) ? [item()] : [],
      nextCursor: "",
    });
  };
  const root = createRoot(document.getElementById("root")!);
  await act(async () => {
    root.render(
      <Shell title="測試" workflow="repositories">
        <View />
      </Shell>,
    );
    await flush();
  });
  return {
    async click(text: string) {
      await act(async () => {
        const node = [...document.querySelectorAll<HTMLButtonElement>("button")].find(
          (node) => node.textContent === text || node.getAttribute("aria-label") === text,
        );
        assert.ok(node, `missing ${text}`);
        pressButton(node);
        await flush();
      });
    },
    async fill(label: string, value: string) {
      await act(async () => {
        const node = [...document.querySelectorAll("label")]
          .find((node) => node.textContent?.startsWith(label))
          ?.querySelector<HTMLInputElement | HTMLTextAreaElement>("input,textarea");
        assert.ok(node, `missing ${label}`);
        const proto =
          node.tagName === "TEXTAREA"
            ? window.HTMLTextAreaElement.prototype
            : window.HTMLInputElement.prototype;
        Object.getOwnPropertyDescriptor(proto, "value")!.set!.call(node, value);
        node.dispatchEvent(new window.Event("input", { bubbles: true }));
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

test("issue status selection stays selected and fetches only on apply", async () => {
  const queries: URLSearchParams[] = [];
  const f = await fixture(async (request) => {
    const url = new URL(request.url);
    if (url.pathname.endsWith("/issues")) queries.push(url.searchParams);
    return undefined;
  });
  try {
    await f.click("問題");
    const before = queries.length;
    await f.click("已關閉");
    await f.click("已關閉");
    assert.equal(queries.length, before);
    await f.click("套用篩選");
    assert.equal(queries.at(-1)?.get("state"), "completed");
    await f.click("全部");
    await f.click("套用篩選");
    assert.equal(queries.at(-1)?.has("state"), false);
  } finally {
    await f.dispose();
  }
});

for (const [name, collection] of [
  ["問題", "issues"],
  ["里程碑", "milestones"],
] as const) {
  test(`${name}: inline draft confirms navigation, conflict freezes writes until fresh read, save updates and closes`, async () => {
    let conflict = true;
    const patches: Record<string, unknown>[] = [];
    const f = await fixture(async (request) => {
      if (request.method === "PATCH") {
        patches.push(await request.json());
        return conflict
          ? Response.json(
              { error: { code: "version_conflict", message: "版本衝突" } },
              { status: 409 },
            )
          : Response.json(item({ title: "更新標題", version: 4 }));
      }
      if (new URL(request.url).pathname.endsWith(`/${collection}/${id}`))
        return Response.json(item({ version: conflict ? 2 : 3 }));
    });
    try {
      await f.click(name);
      await f.click(`編輯${name}：測試項目`);
      const editor = document.querySelector(`[aria-label="${name}編輯"]`)!;
      assert.ok(editor.closest('[data-slot="item"]'));
      assert.equal(document.activeElement, editor.querySelector("input:not([type=date])"));
      await f.fill("標題", "更新標題");
      await f.click("標籤");
      assert.ok(document.querySelector("[role=alertdialog]"));
      await f.click("繼續編輯");
      assert.equal(
        editor.querySelector<HTMLInputElement>("input:not([type=date])")!.value,
        "更新標題",
      );
      await f.click("儲存內容");
      assert.equal(patches[0].expectedVersion, 2);
      assert.equal(
        document.querySelector<HTMLButtonElement>('[aria-label="標籤"]')!.disabled,
        true,
      );
      conflict = false;
      await f.click(`重新讀取${name}`);
      await f.click("放棄修改");
      assert.equal(
        document.querySelector<HTMLButtonElement>('[aria-label="標籤"]')!.disabled,
        false,
      );
      await f.fill("標題", "更新標題");
      await f.click("儲存內容");
      assert.equal(patches[1].expectedVersion, 3);
      assert.equal(document.querySelector(`[aria-label="${name}編輯"]`), null);
      assert.match(document.body.textContent!, /更新標題/);
    } finally {
      await f.dispose();
    }
  });

  test(`${name}: unknown then forbidden keeps exact create retry and blocks tabs and reload`, async () => {
    const attempts: { key: string | null; body: unknown }[] = [];
    const f = await fixture(async (request) => {
      if (request.method !== "POST" || !request.url.endsWith(collection)) return;
      attempts.push({ key: request.headers.get("Idempotency-Key"), body: await request.json() });
      if (attempts.length === 1) throw new Error("unknown");
      if (attempts.length === 2)
        return Response.json(
          { error: { code: "forbidden", message: "forbidden" } },
          { status: 403 },
        );
      return Response.json(item());
    });
    try {
      await f.click(name);
      await f.click(`新增${name}`);
      await f.fill("標題", "原始請求");
      await f.click(`建立${name}`);
      await f.click("重試原建立請求");
      assert.equal(
        document.querySelector<HTMLButtonElement>('[aria-label="標籤"]')!.disabled,
        true,
      );
      const reload = [...document.querySelectorAll<HTMLButtonElement>("button")].find(
        (node) => node.textContent === "重新載入空間",
      )!;
      assert.equal(reload.disabled, true);
      await f.click("重試原建立請求");
      assert.equal(attempts.length, 3);
      assert.deepEqual(attempts[0], attempts[1]);
      assert.deepEqual(attempts[0], attempts[2]);
    } finally {
      await f.dispose();
    }
  });
}

for (const name of ["討論", "標籤"] as const)
  test(`${name}: tab and repository reload confirm drafts without submitting`, async () => {
    let writes = 0;
    const f = await fixture(async (request) => {
      if (request.method === "POST") writes++;
    });
    try {
      await f.click(name);
      if (name === "討論") {
        await f.click("發起討論");
        await f.fill("標題", "保留討論");
      } else await f.fill("標籤名稱", "保留標籤");
      await f.click("重新載入空間");
      assert.ok(document.querySelector("[role=alertdialog]"));
      await f.click("繼續編輯");
      await f.click("物件");
      assert.ok(document.querySelector("[role=alertdialog]"));
      await f.click("放棄修改");
      assert.equal(writes, 0);
      assert.ok(
        [...document.querySelectorAll("button")].some((node) => node.textContent === "新增物件"),
      );
    } finally {
      await f.dispose();
    }
  });

test("comment draft participates in repository confirmation and unknown send keeps navigation blocked", async () => {
  const f = await fixture(async (request) => {
    if (request.method === "POST" && request.url.endsWith("/comments")) throw new Error("unknown");
  });
  try {
    await f.click("討論");
    await f.click("#1 測試項目");
    await f.fill("新增留言", "留言草稿");
    await f.click("問題");
    await f.click("繼續編輯");
    // Radix restores focus from a timer when its focus scope unmounts.
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 0));
    });
    assert.equal(document.activeElement?.tagName, "TEXTAREA");
    await f.click("送出留言");
    assert.equal(document.querySelector<HTMLButtonElement>('[aria-label="問題"]')!.disabled, true);
    assert.ok(
      [...document.querySelectorAll("button")].some((node) => node.textContent === "重試原留言"),
    );
  } finally {
    await f.dispose();
  }
});

for (const name of ["問題", "里程碑"] as const)
  for (const first of ["write", "navigation"] as const) {
    test(`${name}: ${first} first synchronously protects writes and navigation`, async () => {
      let writes = 0;
      const f = await fixture(async (request) => {
        if (request.method === "POST") {
          writes++;
          throw new Error("unknown");
        }
      });
      try {
        await f.click(name);
        await f.click(`新增${name}`);
        await f.fill("標題", "同步操作");
        const form = document.querySelector(`[aria-label="${name}編輯"] form`)!;
        const tab = document.querySelector<HTMLButtonElement>('[aria-label="標籤"]')!;
        await act(async () => {
          if (first === "write") {
            form.dispatchEvent(new window.Event("submit", { bubbles: true, cancelable: true }));
            pressButton(tab);
          } else {
            pressButton(tab);
            form.dispatchEvent(new window.Event("submit", { bubbles: true, cancelable: true }));
          }
          await flush();
        });
        assert.equal(writes, first === "write" ? 1 : 0);
        if (first === "navigation") assert.ok(document.querySelector("[role=alertdialog]"));
        else assert.equal(tab.disabled, true);
      } finally {
        await f.dispose();
      }
    });
  }

for (const [name, collection, create, field, save, retry] of [
  ["討論", "discussions", "發起討論", "標題", "建立討論", "重試原建立請求"],
  ["標籤", "labels", "", "標籤名稱", "建立標籤", "重試原建立請求"],
] as const) {
  test(`${name}: uncertain retry survives a subsequent denial without unlocking tabs`, async () => {
    const attempts: { key: string | null; body: unknown }[] = [];
    const f = await fixture(async (request) => {
      if (request.method !== "POST" || !request.url.endsWith(collection)) return;
      attempts.push({ key: request.headers.get("Idempotency-Key"), body: await request.json() });
      if (attempts.length === 1) throw new Error("unknown");
      if (attempts.length === 2)
        return Response.json(
          { error: { code: "forbidden", message: "forbidden" } },
          { status: 403 },
        );
      return Response.json(name === "標籤" ? { id, name: "原始內容" } : item());
    });
    try {
      await f.click(name);
      if (create) await f.click(create);
      await f.fill(field, "原始內容");
      await f.click(save);
      await f.click(retry);
      assert.equal(
        document.querySelector<HTMLButtonElement>('[aria-label="物件"]')!.disabled,
        true,
      );
      await f.click(retry);
      assert.equal(attempts.length, 3);
      assert.deepEqual(attempts[0], attempts[1]);
      assert.deepEqual(attempts[0], attempts[2]);
    } finally {
      await f.dispose();
    }
  });
}

test("object create retains its original retry after unknown then forbidden under the shared guard", async () => {
  const attempts: { key: string | null; body: unknown }[] = [];
  const f = await fixture(async (request) => {
    if (request.method !== "POST" || !request.url.endsWith("/objects")) return;
    attempts.push({ key: request.headers.get("Idempotency-Key"), body: await request.json() });
    if (attempts.length === 1) throw new Error("unknown");
    if (attempts.length === 2)
      return Response.json({ error: { code: "forbidden", message: "forbidden" } }, { status: 403 });
    return Response.json({
      ...item(),
      name: "原始物件",
      type: "work",
      quantity: "1",
      unit: "SET",
      completedQuantity: "0",
      status: "not_started",
    });
  });
  try {
    await f.click("物件");
    await f.click("新增物件");
    await f.fill("物件名稱", "原始物件");
    await f.click("建立物件");
    await f.click("重試原物件建立請求");
    assert.equal(document.querySelector<HTMLButtonElement>('[aria-label="問題"]')!.disabled, true);
    await f.click("重試原物件建立請求");
    assert.equal(attempts.length, 3);
    assert.deepEqual(attempts[0], attempts[1]);
    assert.deepEqual(attempts[0], attempts[2]);
  } finally {
    await f.dispose();
  }
});

const comment = (overrides = {}) => ({
  id: "c".repeat(32),
  canEdit: true,
  authorId: "A",
  body: "原留言",
  version: 1,
  createdAt: "2026-11-01T00:00:00Z",
  ...overrides,
});

test("comment pagination preserves an edit and failed-page retry; switching comments confirms discarding", async () => {
  let fail = true,
    reads = 0,
    writes = 0;
  const f = await fixture(async (request) => {
    if (request.method === "PATCH") {
      writes++;
      assert.equal((await request.json()).expectedVersion, 1);
      return Response.json(comment({ id: "d".repeat(32), body: "伺服器確認的修改", version: 2 }));
    }
    if (!new URL(request.url).pathname.endsWith("/comments")) return;
    reads++;
    if (new URL(request.url).searchParams.has("cursor")) {
      if (fail) {
        fail = false;
        return Response.json(
          { error: { code: "unavailable", message: "下一頁讀取失敗" } },
          { status: 503 },
        );
      }
      return Response.json({
        items: [comment({ id: "d".repeat(32), body: "另一則留言" })],
        nextCursor: "",
      });
    }
    return Response.json({ items: [comment()], nextCursor: "page2" });
  });
  try {
    await f.click("討論");
    await f.click("#1 測試項目");
    await f.click("編輯留言");
    await f.fill("編輯自己的留言", "未儲存修改");
    for (let i = 0; i < 2; i++) {
      await f.click("載入更多留言");
      assert.ok(
        [...document.querySelectorAll("textarea")].some((node) => node.value === "未儲存修改"),
      );
      assert.equal(document.querySelector("[role=alertdialog]"), null);
    }
    await act(async () => {
      [...document.querySelectorAll<HTMLButtonElement>("button")]
        .filter((node) => node.textContent === "編輯留言")[1]
        .click();
      await flush();
    });
    assert.ok(document.querySelector("[role=alertdialog]"));
    await f.click("繼續編輯");
    assert.ok(
      [...document.querySelectorAll("textarea")].some((node) => node.value === "未儲存修改"),
    );
    await act(async () => {
      [...document.querySelectorAll<HTMLButtonElement>("button")]
        .filter((node) => node.textContent === "編輯留言")[1]
        .click();
      await flush();
    });
    await f.click("放棄修改");
    assert.ok(
      [...document.querySelectorAll("textarea")].some((node) => node.value === "另一則留言"),
    );
    await f.fill("編輯自己的留言", "修改第二頁留言");
    const readsBeforeSave = reads;
    await f.click("儲存留言");
    assert.equal(writes, 1);
    assert.equal(reads, readsBeforeSave);
    assert.match(document.body.textContent!, /伺服器確認的修改/);
    assert.match(document.body.textContent!, /原留言/);
  } finally {
    await f.dispose();
  }
});

test("a comments page already in flight cannot overwrite a successful edit", async () => {
  let updated = false,
    firstPageReads = 0;
  let finishPage!: (response: Response) => void;
  const stalePage = new Promise<Response>((resolve) => {
    finishPage = resolve;
  });
  const edited = comment({ body: "已更新的留言", version: 2 });
  const f = await fixture(async (request) => {
    if (request.method === "PATCH") {
      updated = true;
      return Response.json(edited);
    }
    const url = new URL(request.url);
    if (!url.pathname.endsWith("/comments")) return;
    if (url.searchParams.has("cursor")) return stalePage;
    firstPageReads++;
    return Response.json({ items: [updated ? edited : comment()], nextCursor: "page2" });
  });
  try {
    await f.click("討論");
    await f.click("#1 測試項目");
    await f.click("編輯留言");
    await f.fill("編輯自己的留言", "修改");
    await f.click("載入更多留言");
    const before = firstPageReads;
    await f.click("儲存留言");
    assert.equal(firstPageReads, before + 1);
    await act(async () => {
      finishPage(Response.json({ items: [comment({ body: "過期的留言" })], nextCursor: "" }));
      await flush();
    });
    assert.match(document.body.textContent!, /已更新的留言/);
    assert.doesNotMatch(document.body.textContent!, /過期的留言/);
  } finally {
    await f.dispose();
  }
});

test("unconfirmed comment update stays locked across unrelated pages and resolves only when its target is read", async () => {
  let recovering = false,
    reads = 0;
  const f = await fixture(async (request) => {
    if (request.method === "PATCH") {
      recovering = true;
      throw new Error("update outcome unknown");
    }
    if (!new URL(request.url).pathname.endsWith("/comments")) return;
    reads++;
    if (!recovering) return Response.json({ items: [comment()], nextCursor: "page2" });
    const cursor = new URL(request.url).searchParams.get("cursor");
    return cursor === "target"
      ? Response.json({
          items: [comment({ body: "伺服器確認的內容", version: 2 })],
          nextCursor: "",
        })
      : Response.json({
          items: [comment({ id: "d".repeat(32), body: "其他頁的留言" })],
          nextCursor: "target",
        });
  });
  try {
    await f.click("討論");
    await f.click("#1 測試項目");
    await f.click("編輯留言");
    await f.fill("編輯自己的留言", "可能已提交的修改");
    await f.click("儲存留言");
    const tab = () => document.querySelector<HTMLButtonElement>('[aria-label="問題"]')!;
    assert.equal(tab().disabled, true);
    await f.click("載入更多留言");
    assert.equal(tab().disabled, true);
    assert.ok(
      [...document.querySelectorAll("textarea")].some((node) => node.value === "可能已提交的修改"),
    );
    await f.click("重新載入留言");
    assert.equal(tab().disabled, true);
    assert.equal(document.querySelector("[role=alertdialog]"), null);
    await f.click("載入更多留言");
    assert.equal(tab().disabled, false);
    assert.ok(
      [...document.querySelectorAll("textarea")].some((node) => node.value === "伺服器確認的內容"),
    );
    assert.equal(reads, 4);
  } finally {
    await f.dispose();
  }
});

test("Wiki creation retries the same request and conflicting edits preserve the draft", async () => {
  const attempts: string[] = [],
    versions: number[] = [];
  const response = (version: number) =>
    Response.json({
      id,
      title: "巡檢",
      body: "原文",
      version,
      authorId: "A",
      updatedAt: "2026-11-19T00:00:00Z",
    });
  const f = await fixture(async (req) => {
    if (!req.url.includes("/wiki-pages")) return;
    if (req.method === "POST") {
      attempts.push(req.headers.get("Idempotency-Key") + " " + (await req.text()));
      if (attempts.length === 1) throw new Error("network lost");
      return response(1);
    }
    if (req.method === "PATCH") {
      versions.push((await req.json()).expectedVersion);
      if (versions.length === 1)
        return Response.json(
          { error: { code: "conflict", message: "版本已更新" } },
          { status: 409 },
        );
      if (versions.length === 2)
        return Response.json(
          { error: { code: "unavailable", retrySameRequest: true, message: "結果待確認" } },
          { status: 503 },
        );
      return response(3);
    }
    if (req.url.endsWith(id)) return response(versions.length ? 2 : 1);
    return Response.json({ items: [], nextCursor: "" });
  });
  try {
    await f.click("Wiki");
    await f.click("建立頁面");
    await f.fill("標題", "巡檢");
    await f.fill("Markdown 內容", "原文");
    await f.click("儲存");
    await f.click("重試原建立請求");
    assert.equal(attempts.length, 2);
    assert.equal(attempts[0], attempts[1]);
    const permanent = [...document.querySelectorAll("a")].find(
      (a) => a.textContent === "此頁永久連結",
    );
    assert.equal(new URL(permanent!.href).searchParams.get("wiki"), id);
    await f.click("編輯");
    await f.fill("Markdown 內容", "我的修訂");
    const comparison = () =>
      [...document.querySelectorAll<HTMLButtonElement>("button")].find((button) =>
        button.textContent?.startsWith("對照第"),
      )!;
    assert.equal(comparison().type, "button");
    assert.equal(
      comparison().getAttribute("aria-expanded"),
      "false",
      "comparison starts collapsed",
    );
    comparison().focus();
    await f.click("對照第 1 版 · 巡檢");
    assert.equal(comparison().getAttribute("aria-expanded"), "true");
    assert.match(
      document.getElementById(comparison().getAttribute("aria-controls")!)!.textContent!,
      /原文/,
    );
    assert.equal(document.activeElement, comparison(), "toggle retains keyboard focus");
    await f.click("對照第 1 版 · 巡檢");
    assert.equal(comparison().getAttribute("aria-expanded"), "false");
    assert.equal(
      document.querySelector("textarea")?.value,
      "我的修訂",
      "collapse cannot discard the editor draft",
    );
    assert.equal(attempts.length, 2);
    assert.deepEqual(versions, [], "comparison toggles must not submit the form");
    await f.click("對照第 1 版 · 巡檢");
    await f.click("儲存");
    await f.click("讀取目前版本以比對");
    assert.equal(document.querySelector("textarea")?.value, "我的修訂");
    assert.match(comparison().textContent!, /對照第 2 版/);
    assert.equal(
      comparison().getAttribute("aria-expanded"),
      "false",
      "a fresh server version resets comparison state",
    );
    await f.click("儲存");
    assert.deepEqual(versions, [1, 2]);
    assert.equal(document.querySelector("textarea")?.disabled, true);
    await f.click("讀取目前版本以比對");
    assert.equal(document.querySelector("textarea")?.value, "我的修訂");
    assert.equal(document.querySelector("textarea")?.disabled, false);
    await f.click("儲存");
    assert.deepEqual(versions, [1, 2, 2]);
  } finally {
    await f.dispose();
  }
});

test("issue editing follows the response capability even when the displayed creator matches", async () => {
  const f = await fixture(async (request) => {
    if (new URL(request.url).pathname.endsWith("/issues")) {
      return Response.json({ items: [item({ canEdit: false })], nextCursor: "" });
    }
  });
  try {
    await f.click("問題");
    assert.match(document.body.textContent!, /測試項目/);
    assert.equal(document.querySelector('[aria-label="編輯問題：測試項目"]'), null);
  } finally {
    await f.dispose();
  }
});

test("an interrupted label rename keeps its original target through a later denial", async () => {
  const bodies: unknown[] = [];
  const f = await fixture(async (request) => {
    if (request.method === "GET" && request.url.endsWith("/labels"))
      return Response.json({ items: [{ id, name: "原名稱" }], nextCursor: "" });
    if (request.method !== "PATCH") return;
    bodies.push(await request.json());
    if (bodies.length === 1) throw new Error("connection lost");
    if (bodies.length === 2)
      return Response.json(
        { error: { code: "forbidden", message: "權限已變更", retrySameRequest: false } },
        { status: 403 },
      );
    return Response.json({ id, name: "新名稱" });
  });
  try {
    await f.click("標籤");
    await f.click("改名");
    await f.fill("標籤名稱", "新名稱");
    await f.click("儲存新名稱");
    await f.click("重試改名");
    assert.equal(document.querySelector<HTMLButtonElement>('[aria-label="物件"]')!.disabled, true);
    await f.click("重試改名");
    assert.deepEqual(bodies, Array(3).fill({ name: "新名稱", expectedName: "原名稱" }));
  } finally {
    await f.dispose();
  }
});
