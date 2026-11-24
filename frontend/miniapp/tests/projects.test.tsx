import React, { act } from "react";
import { createRoot } from "react-dom/client";
import test from "node:test";
import assert from "node:assert/strict";
import "./sidebar_dom";
import { Shell } from "./shell_fixture";
import { Workspace } from "../src/business/project/Workspace";

const sa = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
  sb = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
  rid = "c".repeat(32),
  pid = "d".repeat(32);
const organizationContext = {
  groups: [
    {
      label: "Organizations",
      organizations: [
        { organizationId: sa, name: "Group A" },
        { organizationId: sb, name: "Group B" },
      ],
    },
  ],
  selected: null,
  enterprises: [],
};
const project = {
  id: pid,
  ownerId: sa,
  defaultRepositoryId: rid,
  title: "Project A",
  body: "Project body",
  canEdit: true,
  creatorId: "A",
  version: 3,
  issueCount: 0,
  completedIssueCount: 0,
  createdAt: "2026-11-01T00:00:00Z",
  updatedAt: "2026-11-01T00:00:00Z",
};
const repository = {
  id: rid,
  ownerId: sa,
  name: "Source A",
  status: "ACTIVE",
  version: 1,
  createdAt: "2026-11-01T00:00:00Z",
  updatedAt: "2026-11-01T00:00:00Z",
};
const flush = async () => {
  for (let i = 0; i < 10; i++) await new Promise<void>((resolve) => setImmediate(resolve));
};

async function fixture(
  path: string,
  handle: (request: Request, actor: string) => Promise<Response>,
) {
  window.history.replaceState(null, "", path);
  document.body.innerHTML = '<div id="root"></div>';
  let actor = "A";
  const oldFetch = globalThis.fetch,
    oldSDK = window.liff,
    requests: Request[] = [];
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
    requests.push(request.clone());
    const path = new URL(request.url).pathname;
    if (path.endsWith("/config")) return Response.json({ liffId: "projects-test" });
    if (path.endsWith("/session"))
      return Response.json({
        user: { userId: actor, displayName: actor },
        csrfToken: "c".repeat(43),
        expiresAt: new Date(Date.now() + 1800000).toISOString(),
      });
    return handle(request, actor);
  };
  const root = createRoot(document.getElementById("root")!);
  await act(async () => {
    root.render(
      <Shell title="專案" workflow="projects">
        <Workspace />
      </Shell>,
    );
    await flush();
  });
  return {
    requests,
    async click(text: string) {
      await act(async () => {
        const button = [...document.querySelectorAll("button")].find(
          (node) => node.textContent === text || node.getAttribute("aria-label") === text,
        );
        assert.ok(button, `missing ${text}`);
        button.click();
        await flush();
      });
    },
    async chooseKind(value: string) {
      await act(async () => {
        document
          .querySelector<HTMLButtonElement>(
            `[role="radiogroup"][aria-label="來源種類"] [value="${value}"]`,
          )!
          .click();
        await flush();
      });
    },
    async select(label: string, value: string) {
      await act(async () => {
        const select = [...document.querySelectorAll("label")]
          .find((node) => node.textContent?.startsWith(label))
          ?.querySelector<HTMLSelectElement>("select");
        assert.ok(select, `missing ${label}`);
        select.value = value;
        select.dispatchEvent(new window.Event("change", { bubbles: true }));
        await flush();
      });
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
    async route(value: string) {
      await act(async () => {
        window.history.replaceState(null, "", value);
        window.dispatchEvent(new window.PopStateEvent("popstate"));
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

async function normal(request: Request): Promise<Response> {
  const path = new URL(request.url).pathname;
  if (path === "/api/v1/me/organization-context") return Response.json(organizationContext);
  if (path.endsWith("/projects"))
    return Response.json({ items: path.includes(sa) ? [project] : [], nextCursor: "" });
  if (path.endsWith("/projects/" + pid)) return Response.json(project);
  if (path.endsWith("/references") || path.endsWith("/items") || path.endsWith("/issues"))
    return Response.json({ items: [], nextCursor: "" });
  if (path === "/api/v1/repositories")
    return Response.json({ items: [repository], nextCursor: "" });
  if (path === "/api/v1/repositories/" + rid) return Response.json(repository);
  throw new Error("unexpected " + path);
}

test("projects are a sidebar peer and open independently through Group APIs", async () => {
  const f = await fixture("/miniapp/projects", normal);
  try {
    const projectLink = document.querySelector<HTMLAnchorElement>('a[href="/miniapp/projects"]')!;
    const repoLink = Array.from(
      document.querySelectorAll<HTMLButtonElement>("button[aria-expanded]"),
    ).find((button) => button.textContent === "儲存庫")!;
    assert.ok(projectLink);
    assert.ok(repoLink);
    assert.equal(
      projectLink.closest('[data-sidebar="menu-item"]')?.parentElement,
      repoLink.closest('[data-sidebar="menu-item"]')?.parentElement,
    );
    assert.equal(projectLink.getAttribute("aria-current"), "page");
    assert.equal(
      f.requests.filter((request) =>
        new URL(request.url).pathname.startsWith("/api/v1/repositories"),
      ).length,
      0,
      "Project landing must not require repository listing",
    );
    await f.click("Group A");
    assert.ok(
      f.requests.some(
        (request) => new URL(request.url).pathname === `/api/v1/owners/${sa}/projects`,
      ),
    );
    assert.equal(
      f.requests.filter((request) =>
        new URL(request.url).pathname.startsWith("/api/v1/repositories"),
      ).length,
      0,
    );
  } finally {
    await f.dispose();
  }
});

test("Project deep links and popstate reauthorize Group and reset previous Project detail", async () => {
  const f = await fixture(`/miniapp/projects?owner=${sa}&project=${pid}`, normal);
  try {
    assert.ok(document.body.textContent!.includes("Project body"));
    await f.click("Group B");
    assert.equal(new URLSearchParams(location.search).get("owner"), sb);
    assert.ok(!document.body.textContent!.includes("Project body"));
    await f.route(`/miniapp/projects?owner=${sa}&project=${pid}`);
    assert.ok(document.body.textContent!.includes("Project body"));
    assert.equal(
      f.requests.filter(
        (request) => new URL(request.url).pathname === `/api/v1/owners/${sa}/projects/${pid}`,
      ).length,
      2,
    );
  } finally {
    await f.dispose();
  }
});

test("identity changes reject late Group detail even when fetch ignores cancellation", async () => {
  let finish!: (response: Response) => void;
  const f = await fixture(
    `/miniapp/projects?owner=${sa}&project=${pid}`,
    async (request, actor) => {
      if (new URL(request.url).pathname === "/api/v1/me/organization-context") {
        if (actor === "A")
          return new Promise((resolve) => {
            finish = resolve;
          });
        return Response.json(
          { error: { code: "not_found", message: "群組不可讀取" } },
          { status: 404 },
        );
      }
      return normal(request);
    },
  );
  try {
    await f.verifyAs("B");
    await act(async () => {
      finish(Response.json(organizationContext));
      await flush();
    });
    assert.ok(document.body.textContent!.includes("群組不可讀取"));
    assert.ok(!document.body.textContent!.includes("Project body"));
    assert.ok(
      !f.requests.some(
        (request) => new URL(request.url).pathname === `/api/v1/owners/${sa}/projects/${pid}`,
      ),
    );
  } finally {
    await f.dispose();
  }
});

test("canonical create retries retain source, payload and key after an unknown result", async () => {
  const posts: { key: string | null; body: unknown }[] = [];
  const f = await fixture(`/miniapp/projects?owner=${sa}&source=${rid}`, async (request) => {
    if (request.method === "POST" && new URL(request.url).pathname.endsWith("/projects")) {
      posts.push({ key: request.headers.get("Idempotency-Key"), body: await request.json() });
      return posts.length === 1
        ? Response.json({ error: { code: "unavailable", message: "結果不明" } }, { status: 503 })
        : Response.json(project);
    }
    return normal(request);
  });
  try {
    await f.click("新增專案");
    const input = document.querySelector<HTMLInputElement>('input[maxlength="120"]')!;
    await act(async () => {
      Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!.call(
        input,
        "Same project",
      );
      input.dispatchEvent(new Event("input", { bubbles: true }));
      await flush();
    });
    await f.click("繼續填寫專案");
    assert.equal(input.value, "Same project");
    await f.click("建立專案");
    await f.click("重試原建立請求");
    assert.equal(posts.length, 2);
    assert.deepEqual(posts[0], posts[1]);
    assert.deepEqual(posts[0].body, { title: "Same project", body: "", defaultRepositoryId: rid });
    assert.equal(new URLSearchParams(location.search).get("project"), pid);
  } finally {
    await f.dispose();
  }
});

test("a source deep link from another Group cannot become a creation source", async () => {
  const f = await fixture(
    `/miniapp/projects?owner=${sa}&source=${rid}&project=new`,
    async (request) => {
      const path = new URL(request.url).pathname;
      if (path === "/api/v1/repositories") {
        assert.equal(new URL(request.url).searchParams.get("ownerId"), sa);
        return Response.json({ items: [], nextCursor: "" });
      }
      if (path === "/api/v1/repositories/" + rid)
        return Response.json({ ...repository, ownerId: sb });
      return normal(request);
    },
  );
  try {
    assert.ok(document.body.textContent!.includes("來源擁有者不符"));
    assert.equal(document.querySelector<HTMLSelectElement>("select")!.value, "");
    assert.ok(
      !f.requests.some(
        (request) =>
          request.method === "POST" && new URL(request.url).pathname.endsWith("/projects"),
      ),
    );
  } finally {
    await f.dispose();
  }
});

test("unreadable source preserves Project metadata without inventing progress or fetching source items", async () => {
  const f = await fixture(`/miniapp/projects?owner=${sa}&project=${pid}`, async (request) => {
    if (new URL(request.url).pathname.endsWith("/projects/" + pid))
      return Response.json({ ...project, sourceAvailable: false });
    return normal(request);
  });
  try {
    assert.ok(document.body.textContent!.includes("來源儲存庫目前不可讀取"));
    assert.ok(document.body.textContent!.includes("Project body"));
    assert.ok(!document.body.textContent!.includes("尚無問題"));
    assert.equal(
      document.querySelector<HTMLInputElement>('input[maxlength="120"]')!.disabled,
      false,
    );
    assert.ok(
      !f.requests.some((request) => /\/(items|issues)$/.test(new URL(request.url).pathname)),
    );
  } finally {
    await f.dispose();
  }
});

const sourceId = "e".repeat(32),
  parentId = "f".repeat(32),
  refId = "1".repeat(32);
const reference = (overrides = {}) => ({
  id: refId,
  sourceType: "issue",
  sourceId,
  sourceRepositoryId: rid,
  sourceUnavailable: false,
  sourceTitle: "最新問題",
  sourceState: "open",
  ...overrides,
});
const objectSource = {
  id: parentId,
  name: "父物件",
  quantity: "10",
  unit: "SET",
  completedQuantity: "3",
  type: "work",
  status: "in_progress",
};
const sourceCatalog = async (request: Request) => {
  const path = new URL(request.url).pathname;
  if (path.endsWith("/objects"))
    return Response.json({
      items: new URL(request.url).searchParams.has("parentId")
        ? [{ ...objectSource, id: sourceId, parentId, name: "子工序", quantity: "20", unit: "M2" }]
        : [objectSource],
      nextCursor: "",
    });
  if (path.endsWith("/issues"))
    return Response.json({
      items: [{ id: sourceId, title: "可選問題", number: 12 }],
      nextCursor: "",
    });
  if (path.endsWith("/discussions"))
    return Response.json({ items: [{ id: sourceId, title: "可選討論" }], nextCursor: "" });
  return normal(request);
};

test("Project reference list shows current four typed sources without mixed progress aggregation", async () => {
  const refs = [
    reference({
      sourceType: "object",
      sourceTitle: "最新物件",
      quantity: "10",
      unit: "SET",
      completedQuantity: "3",
      type: "work",
      sourceState: "in_progress",
    }),
    reference({
      id: "2".repeat(32),
      sourceType: "object",
      sourceTitle: "最新子工序",
      quantity: "20",
      unit: "M2",
      type: "non_work",
    }),
    reference({ id: "3".repeat(32) }),
    reference({
      id: "4".repeat(32),
      sourceType: "discussion",
      sourceTitle: "最新討論",
      sourceState: "",
    }),
  ];
  const f = await fixture(`/miniapp/projects?owner=${sa}&project=${pid}`, async (request) =>
    new URL(request.url).pathname.endsWith("/references")
      ? Response.json({ items: refs, nextCursor: "" })
      : sourceCatalog(request),
  );
  try {
    for (const title of ["最新物件", "最新子工序", "最新問題", "最新討論"])
      assert.ok(document.body.textContent!.includes(title));
    assert.match(document.body.textContent!, /進度：3\/10 SET/);
    assert.match(document.body.textContent!, /進度：不適用/);
    assert.match(document.body.textContent!, /不會修改來源/);
    assert.ok(![...document.querySelectorAll("p")].some((node) => node.textContent === "狀態："));
    assert.equal(
      f.requests.filter((request) => request.method !== "GET").length,
      1,
      "only session creation writes",
    );
  } finally {
    await f.dispose();
  }
});

for (const kind of ["object", "subobject", "issue", "discussion"])
  test(`Project adds ${kind} by source catalog with exact typed identity`, async () => {
    const bodies: Record<string, unknown>[] = [];
    const f = await fixture(`/miniapp/projects?owner=${sa}&project=${pid}`, async (request) => {
      if (request.method === "POST" && new URL(request.url).pathname.endsWith("/references")) {
        bodies.push(await request.json());
        return Response.json({ ...project, version: 4 });
      }
      return sourceCatalog(request);
    });
    try {
      await f.chooseKind(kind);
      if (kind === "subobject") {
        assert.ok(
          !f.requests.some((request) => new URL(request.url).pathname.endsWith("/subobjects")),
        );
        await f.select("父物件", parentId);
      }
      await f.select("選擇既有", kind === "object" ? parentId : sourceId);
      await f.click("加入專案");
      assert.deepEqual(bodies, [
        {
          sourceRepositoryId: rid,
          sourceType: kind === "subobject" ? "object" : kind,
          sourceId: kind === "object" ? parentId : sourceId,
          expectedVersion: 3,
        },
      ]);
    } finally {
      await f.dispose();
    }
  });

test("unknown reference add retains body after forbidden retry and blocks metadata/Project/Group navigation", async () => {
  const attempts: unknown[] = [];
  let metadataWrites = 0;
  const f = await fixture(`/miniapp/projects?owner=${sa}&project=${pid}`, async (request) => {
    const path = new URL(request.url).pathname;
    if (request.method === "POST" && path.endsWith("/references")) {
      attempts.push(await request.json());
      assert.equal(request.headers.get("Idempotency-Key"), null);
      if (attempts.length === 1) throw new Error("unknown");
      if (attempts.length === 2)
        return Response.json(
          { error: { code: "forbidden", message: "forbidden" } },
          { status: 403 },
        );
      return Response.json({ ...project, version: 4 });
    }
    if (request.method === "PATCH") metadataWrites++;
    return sourceCatalog(request);
  });
  try {
    await f.select("選擇既有", sourceId);
    await f.click("加入專案");
    assert.equal(
      [...document.querySelectorAll("button")].some((node) => node.textContent === "重新載入專案"),
      false,
    );
    for (const label of ["Group B", "新增專案"])
      assert.equal(
        [...document.querySelectorAll<HTMLButtonElement>("button")].find(
          (node) => node.textContent === label,
        )!.disabled,
        true,
      );
    await f.click("Group B");
    assert.equal(new URLSearchParams(location.search).get("owner"), sa);
    await act(async () => {
      document
        .querySelector("form")!
        .dispatchEvent(new window.Event("submit", { bubbles: true, cancelable: true }));
      await flush();
    });
    assert.equal(metadataWrites, 0);
    await f.click("重試原引用操作");
    assert.equal(
      [...document.querySelectorAll<HTMLButtonElement>("button")].find(
        (node) => node.textContent === "Group B",
      )!.disabled,
      true,
    );
    await f.click("重試原引用操作");
    assert.equal(attempts.length, 3);
    assert.deepEqual(attempts[0], attempts[1]);
    assert.deepEqual(attempts[0], attempts[2]);
  } finally {
    await f.dispose();
  }
});

test("unavailable source strips all old display payload but owner can remove reference with frozen retry", async () => {
  const deletes: string[] = [];
  const f = await fixture(`/miniapp/projects?owner=${sa}&project=${pid}`, async (request) => {
    const path = new URL(request.url).pathname;
    if (path.endsWith("/projects/" + pid))
      return Response.json({ ...project, sourceAvailable: false });
    if (request.method === "DELETE") {
      deletes.push(request.url);
      if (deletes.length === 1) throw new Error("unknown");
      return Response.json({ ...project, sourceAvailable: false, version: 4 });
    }
    if (path.endsWith("/references"))
      return Response.json({
        items: [
          reference({
            sourceType: "object",
            sourceUnavailable: true,
            sourceTitle: "不得顯示的舊秘密",
            quantity: "999999",
            unit: "秘密單位",
            completedQuantity: "999",
          }),
        ],
        nextCursor: "",
      });
    return normal(request);
  });
  try {
    assert.ok(!document.body.textContent!.includes("不得顯示的舊秘密"));
    assert.ok(!document.body.textContent!.includes("秘密單位"));
    assert.ok(
      ![...document.querySelectorAll("label")].some((node) =>
        node.textContent?.startsWith("選擇既有"),
      ),
    );
    await f.click("移出專案：不可讀取的物件");
    await f.click("重試原引用操作");
    assert.equal(deletes.length, 2);
    assert.equal(deletes[0], deletes[1]);
    assert.equal(new URL(deletes[0]).searchParams.get("expectedVersion"), "3");
  } finally {
    await f.dispose();
  }
});

test("reference source-kind change synchronously invalidates old selected source", async () => {
  let posts = 0;
  const f = await fixture(`/miniapp/projects?owner=${sa}&project=${pid}`, async (request) => {
    if (request.method === "POST" && new URL(request.url).pathname.endsWith("/references")) posts++;
    return sourceCatalog(request);
  });
  try {
    await f.select("選擇既有", sourceId);
    const add = [...document.querySelectorAll<HTMLButtonElement>("button")].find(
      (node) => node.textContent === "加入專案",
    )!;
    await act(async () => {
      document.querySelector<HTMLButtonElement>('[role="radio"][value="discussion"]')!.click();
      add.click();
      await flush();
    });
    assert.equal(posts, 0);
    assert.equal(
      [...document.querySelectorAll<HTMLButtonElement>("button")].find(
        (node) => node.textContent === "加入專案",
      )!.disabled,
      true,
    );
  } finally {
    await f.dispose();
  }
});

test("member reads references while only Project creator receives source selection and removal controls", async () => {
  const f = await fixture(`/miniapp/projects?owner=${sa}&project=${pid}`, async (request) => {
    const path = new URL(request.url).pathname;
    if (path.endsWith("/projects/" + pid))
      return Response.json({ ...project, canEdit: false, creatorId: "B" });
    if (path.endsWith("/references"))
      return Response.json({ items: [reference()], nextCursor: "" });
    return normal(request);
  });
  try {
    assert.match(document.body.textContent!, /最新問題/);
    assert.ok(
      ![...document.querySelectorAll("label")].some((node) =>
        node.textContent?.startsWith("選擇既有"),
      ),
    );
    assert.ok(
      ![...document.querySelectorAll("button")].some((node) =>
        node.textContent?.startsWith("移出專案"),
      ),
    );
  } finally {
    await f.dispose();
  }
});

for (const first of ["reference", "metadata"])
  test(`Project metadata and reference write handshake is synchronous (${first} first)`, async () => {
    let references = 0,
      metadata = 0;
    const f = await fixture(`/miniapp/projects?owner=${sa}&project=${pid}`, async (request) => {
      if (request.method === "POST" && new URL(request.url).pathname.endsWith("/references")) {
        references++;
        throw new Error("unknown reference");
      }
      if (request.method === "PATCH") {
        metadata++;
        throw new Error("unknown metadata");
      }
      return sourceCatalog(request);
    });
    try {
      await f.select("選擇既有", sourceId);
      const add = [...document.querySelectorAll<HTMLButtonElement>("button")].find(
        (node) => node.textContent === "加入專案",
      )!;
      const form = document.querySelector("form")!;
      await act(async () => {
        if (first === "reference") {
          add.click();
          form.dispatchEvent(new window.Event("submit", { bubbles: true, cancelable: true }));
        } else {
          form.dispatchEvent(new window.Event("submit", { bubbles: true, cancelable: true }));
          add.click();
        }
        await flush();
      });
      assert.equal(references, first === "reference" ? 1 : 0);
      assert.equal(metadata, first === "metadata" ? 1 : 0);
    } finally {
      await f.dispose();
    }
  });

for (const scope of ["group", "identity"])
  test(`Project reference private payload clears after ${scope} switch and late response`, async () => {
    let release: ((value: Response) => void) | undefined;
    let aborted = false;
    const f = await fixture(
      `/miniapp/projects?owner=${sa}&project=${pid}`,
      async (request, actor) => {
        if (new URL(request.url).pathname.endsWith("/references")) {
          if (actor !== "A") return Response.json({ items: [], nextCursor: "" });
          request.signal.addEventListener("abort", () => {
            aborted = true;
          });
          return new Promise<Response>((resolve) => {
            release = resolve;
          });
        }
        return sourceCatalog(request);
      },
    );
    try {
      if (scope === "group") await f.click("Group B");
      else await f.verifyAs("B");
      assert.equal(aborted, true);
      await act(async () => {
        release?.(
          Response.json({ items: [reference({ sourceTitle: "舊引用秘密" })], nextCursor: "" }),
        );
        await flush();
      });
      assert.ok(!document.body.textContent!.includes("舊引用秘密"));
    } finally {
      await f.dispose();
    }
  });

test("Project reference parent change rejects stale selected child in the same batch", async () => {
  let posts = 0;
  const f = await fixture(`/miniapp/projects?owner=${sa}&project=${pid}`, async (request) => {
    const path = new URL(request.url).pathname;
    if (request.method === "POST" && path.endsWith("/references")) posts++;
    if (path.endsWith("/objects") && !new URL(request.url).searchParams.has("parentId"))
      return Response.json({
        items: [objectSource, { ...objectSource, id: "2".repeat(32), name: "另一父物件" }],
        nextCursor: "",
      });
    return sourceCatalog(request);
  });
  try {
    await f.chooseKind("subobject");
    await f.select("父物件", parentId);
    await f.select("選擇既有", sourceId);
    const add = [...document.querySelectorAll<HTMLButtonElement>("button")].find(
      (node) => node.textContent === "加入專案",
    )!;
    await act(async () => {
      const select = [...document.querySelectorAll("label")]
        .find((node) => node.textContent?.startsWith("父物件"))!
        .querySelector("select")!;
      select.value = "2".repeat(32);
      select.dispatchEvent(new window.Event("change", { bubbles: true }));
      add.click();
      await flush();
    });
    assert.equal(posts, 0);
  } finally {
    await f.dispose();
  }
});

test("Project unknown update retains reload control after a failed detail read", async () => {
  let reads = 0;
  const f = await fixture(`/miniapp/projects?owner=${sa}&project=${pid}`, async (request) => {
    if (request.method === "PATCH") throw new Error("unknown update");
    if (new URL(request.url).pathname.endsWith("/projects/" + pid)) {
      reads++;
      if (reads === 2)
        return Response.json(
          { error: { code: "unavailable", message: "暫時讀取失敗" } },
          { status: 503 },
        );
      return Response.json(project);
    }
    return normal(request);
  });
  try {
    await f.click("儲存內容");
    await f.click("重新讀取專案");
    assert.match(document.body.textContent!, /暫時讀取失敗/);
    await f.click("重新讀取專案");
    assert.equal(reads, 3);
    assert.ok(
      [...document.querySelectorAll<HTMLButtonElement>("button")].find(
        (button) => button.textContent === "儲存內容",
      )?.disabled === false,
    );
  } finally {
    await f.dispose();
  }
});

test("unknown committed reference can be reconciled after source permission loss and then removed", async () => {
  let present = false,
    posts = 0,
    deletes = 0;
  const f = await fixture(`/miniapp/projects?owner=${sa}&project=${pid}`, async (request) => {
    const path = new URL(request.url).pathname;
    if (path.endsWith("/references") && request.method === "POST") {
      posts++;
      if (posts === 1) {
        present = true;
        throw new Error("committed reply lost");
      }
      return Response.json(
        { error: { code: "forbidden", message: "來源無法存取" } },
        { status: 403 },
      );
    }
    if (path.endsWith("/reference-state")) {
      const query = new URL(request.url).searchParams;
      assert.equal(query.get("sourceId"), sourceId);
      assert.equal(query.get("desiredPresent"), "true");
      return Response.json({ present, matches: present });
    }
    if (request.method === "DELETE") {
      deletes++;
      present = false;
      return Response.json({ ...project, version: 5, sourceAvailable: false });
    }
    if (path.endsWith("/projects/" + pid) && posts)
      return Response.json({ ...project, version: 4, sourceAvailable: false });
    if (path.endsWith("/references"))
      return Response.json({
        items: present
          ? [reference({ sourceUnavailable: true, sourceTitle: "不應顯示的來源" })]
          : [],
        nextCursor: "",
      });
    return sourceCatalog(request);
  });
  try {
    await f.select("選擇既有", sourceId);
    await f.click("加入專案");
    await f.click("重試原引用操作");
    await f.click("核對引用結果");
    assert.equal(posts, 2);
    assert.ok(!document.body.textContent!.includes("不應顯示的來源"));
    await f.click("移出專案：不可讀取的問題");
    assert.equal(deletes, 1);
    assert.equal(present, false);
    assert.equal(
      [...document.querySelectorAll<HTMLButtonElement>("button")].find(
        (button) => button.textContent === "Group B",
      )!.disabled,
      false,
    );
  } finally {
    await f.dispose();
  }
});

test("creator links a repository from a second group before adding its issue", async () => {
  const second = "2".repeat(32);
  const writes: Record<string, unknown>[] = [];
  const f = await fixture(`/miniapp/projects?owner=${sa}&project=${pid}`, async (request) => {
    const path = new URL(request.url).pathname;
    if (path === "/api/v1/repositories")
      return Response.json({
        items: [repository, { ...repository, id: second, ownerId: sb, name: "Source B" }],
        nextCursor: "",
      });
    if (request.method === "PATCH" || (request.method === "POST" && path.endsWith("/references"))) {
      writes.push(await request.json());
      return Response.json({
        ...project,
        linkedRepositoryIds: [rid, second],
        version: 3 + writes.length,
      });
    }
    return sourceCatalog(request);
  });
  try {
    await f.select("來源儲存庫", second);
    await f.click("連結此儲存庫");
    assert.equal(writes[0]?.linkRepositoryId, second);
    await f.select("來源儲存庫", second);
    await f.select("選擇既有", sourceId);
    await f.click("加入專案");
    assert.equal(writes[1]?.sourceRepositoryId, second);
    assert.equal(writes[1]?.expectedVersion, 4);
  } finally {
    await f.dispose();
  }
});

test("inspection on a SubObject submits once and reads the saved pending result", async () => {
  const saved = {
    id: "submission",
    note: "配線完成",
    status: "pending",
    submittedBy: "A",
    submittedAt: "2026-11-01T00:00:00Z",
    sourceId,
    sourceRepositoryId: rid,
    sourceVersion: 7,
    sourceType: "object",
  };
  let writes = 0;
  const f = await fixture(`/miniapp/projects?owner=${sa}&project=${pid}`, async (req) => {
    const path = new URL(req.url).pathname;
    if (path.endsWith("/references"))
      return Response.json({
        items: [reference({ sourceType: "object", inspectionMode: "submit_if_empty" })],
        nextCursor: "",
      });
    if (path.endsWith("/inspection-submissions")) {
      if (req.method === "POST") {
        assert.deepEqual(await req.json(), { note: saved.note });
        writes++;
        return Response.json(saved);
      }
      return Response.json(writes ? [saved] : []);
    }
    return sourceCatalog(req);
  });
  try {
    await f.click("查看驗收紀錄");
    const field = [...document.querySelectorAll("label")]
      .find((node) => node.textContent?.startsWith("檢查說明"))!
      .querySelector("textarea")!;
    await act(async () => {
      Object.getOwnPropertyDescriptor(window.HTMLTextAreaElement.prototype, "value")!.set!.call(
        field,
        saved.note,
      );
      field.dispatchEvent(new window.Event("input", { bubbles: true }));
      await flush();
    });
    await f.click("查看驗收紀錄");
    assert.equal(field.value, saved.note);
    await f.click("提交驗收");
    assert.equal(writes, 1);
    await f.click("查看驗收紀錄");
    assert.match(document.body.textContent!, /配線完成.*待審/);
  } finally {
    await f.dispose();
  }
});

test("personal projects use the signed-in owner without looking up a group", async () => {
  const owner = "U" + "1".repeat(32);
  const f = await fixture("/miniapp/projects", normal);
  try {
    await f.verifyAs(owner);
    await f.click("我的個人專案");
    const paths = f.requests.map((request) => new URL(request.url).pathname);
    assert.ok(paths.includes(`/api/v1/owners/${owner}/projects`));
    assert.ok(!paths.includes(`/api/v1/groups/${owner}`));
  } finally {
    await f.dispose();
  }
});

test("temporary pagination failure preserves references and retries the same cursor", async () => {
  const cursors: (string | null)[] = [];
  const f = await fixture(`/miniapp/projects?owner=${sa}&project=${pid}`, async (request) => {
    const url = new URL(request.url);
    if (!url.pathname.endsWith("/references")) return sourceCatalog(request);
    cursors.push(url.searchParams.get("cursor"));
    if (cursors.length === 2) throw new TypeError("network offline");
    if (!url.searchParams.has("cursor"))
      return Response.json({ items: [reference()], nextCursor: "next" });
    return Response.json({
      items: [reference({ id: "2".repeat(32), sourceTitle: "第二頁問題" })],
      nextCursor: "",
    });
  });
  try {
    await f.click("載入更多專案引用");
    assert.match(document.body.textContent!, /最新問題/);
    await f.click("載入更多專案引用");
    assert.deepEqual(cursors, [null, "next", "next"]);
    assert.match(document.body.textContent!, /第二頁問題/);
  } finally {
    await f.dispose();
  }
});

test("copy selects a destination and submits the source project without linking repositories", async () => {
  const posts: unknown[] = [];
  const f = await fixture(`/miniapp/projects?owner=${sa}&project=${pid}`, async (request) => {
    if (request.method === "POST") {
      assert.equal(new URL(request.url).pathname, `/api/v1/owners/${sb}/projects`);
      posts.push(await request.json());
      return Response.json({
        ...project,
        ownerId: sb,
        id: "e".repeat(32),
        defaultRepositoryId: undefined,
      });
    }
    return normal(request);
  });
  try {
    await f.select("複製到另一個擁有者", sb);
    const input = document.querySelector<HTMLInputElement>('input[maxlength="120"]')!;
    await act(async () => {
      Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!.call(
        input,
        "Audit copy",
      );
      input.dispatchEvent(new Event("input", { bubbles: true }));
      await flush();
    });
    await f.click("建立專案");
    assert.deepEqual(posts, [
      { title: "Audit copy", body: "", copyFrom: { ownerId: sa, projectId: pid } },
    ]);
    assert.equal(new URLSearchParams(location.search).get("project"), "e".repeat(32));
  } finally {
    await f.dispose();
  }
});

test("save field options and a view, then edit the copied settings independently", async () => {
  type Settings = import("@repo/api-client").components["schemas"]["ProjectSettings"];
  const copyId = "e".repeat(32);
  const projects = new Map<string, typeof project & { settings?: Settings }>([[pid, project]]);
  const f = await fixture(`/miniapp/projects?owner=${sa}&project=${pid}`, async (request) => {
    const path = new URL(request.url).pathname;
    const id = path.split("/").at(-1)!;
    if (request.method === "PATCH" && projects.has(id)) {
      const input = await request.json();
      const current = projects.get(id)!;
      assert.equal(input.expectedVersion, current.version);
      const updated = { ...current, settings: input.settings, version: current.version + 1 };
      projects.set(id, updated);
      return Response.json(updated);
    }
    if (request.method === "POST" && path.endsWith("/projects")) {
      const input = await request.json();
      assert.deepEqual(input.copyFrom, { ownerId: sa, projectId: pid });
      const copy = {
        ...structuredClone(projects.get(pid)!),
        ownerId: sb,
        id: copyId,
        title: input.title,
        version: 1,
      };
      projects.set(copyId, copy);
      return Response.json(copy);
    }
    if (projects.has(id)) return Response.json(projects.get(id));
    return normal(request);
  });
  async function fill(label: string, value: string) {
    const element = [...document.querySelectorAll("label")]
      .find((node) => node.textContent?.startsWith(label))
      ?.querySelector("input,textarea");
    assert.ok(element, label);
    await act(async () => {
      Object.getOwnPropertyDescriptor(Object.getPrototypeOf(element), "value")!.set!.call(
        element,
        value,
      );
      element.dispatchEvent(new Event("input", { bubbles: true }));
      await flush();
    });
  }
  try {
    await f.click("新增欄位");
    await fill("欄位名稱", "Priority");
    await fill("選項", "High\nLow");
    await f.click("新增檢視");
    await fill("檢視名稱", "Triage");
    await f.click("儲存內容");
    const settings = projects.get(pid)!.settings!;
    assert.deepEqual(settings.fields[0].options, ["High", "Low"]);
    assert.equal(settings.views[0].fieldId, settings.fields[0].id);
    await f.select("複製到另一個擁有者", sb);
    await fill("標題", "Audit");
    await f.click("建立專案");
    await fill("欄位名稱", "Severity");
    await f.click("儲存內容");
    assert.equal(projects.get(copyId)!.settings!.fields[0].name, "Severity");
    assert.equal(projects.get(pid)!.settings!.fields[0].name, "Priority");
  } finally {
    await f.dispose();
  }
});

test("project list keeps the first page and retries the failed cursor", async () => {
  const cursors: (string | null)[] = [];
  const f = await fixture(`/miniapp/projects?owner=${sa}`, async (request) => {
    const url = new URL(request.url);
    if (!url.pathname.endsWith("/projects")) return normal(request);
    cursors.push(url.searchParams.get("cursor"));
    if (cursors.length === 2) throw new TypeError("network offline");
    return Response.json(
      cursors.length === 1
        ? { items: [project], nextCursor: "next" }
        : { items: [{ ...project, id: "2".repeat(32), title: "第二頁專案" }], nextCursor: "" },
    );
  });
  try {
    await f.click("載入更多專案");
    assert.ok(document.body.textContent!.includes(project.title));
    await f.click("重試讀取");
    assert.deepEqual(cursors, [null, "next", "next"]);
    assert.match(document.body.textContent!, /第二頁專案/);
  } finally {
    await f.dispose();
  }
});

test("owner navigation preserves a project draft until discard is confirmed", async () => {
  const f = await fixture(`/miniapp/projects?owner=${sa}&project=new`, normal);
  try {
    await f.select("來源儲存庫", rid);
    await f.click("Group B");
    assert.equal(new URLSearchParams(location.search).get("owner"), sa);
    await f.click("繼續編輯");
    assert.equal(document.querySelector<HTMLSelectElement>("form select")!.value, rid);
    await f.click("Group B");
    await f.click("放棄修改");
    assert.equal(new URLSearchParams(location.search).get("owner"), sb);
  } finally {
    await f.dispose();
  }
});

test("clearing a prefilled project source is an edit and restoring it makes the form clean", async () => {
  const f = await fixture(`/miniapp/projects?owner=${sa}&project=new&source=${rid}`, normal);
  try {
    assert.equal(document.querySelector<HTMLSelectElement>("form select")!.value, rid);
    await f.select("來源儲存庫", "");
    await f.click("Group B");
    await f.click("繼續編輯");
    assert.equal(document.querySelector<HTMLSelectElement>("form select")!.value, "");
    await f.select("來源儲存庫", rid);
    await f.click("Group B");
    assert.equal(new URLSearchParams(location.search).get("owner"), sb);
  } finally {
    await f.dispose();
  }
});

test("a rejected project update stays editable without rereading the project", async () => {
  let writes = 0;
  const f = await fixture(`/miniapp/projects?owner=${sa}&project=${pid}`, async (request) => {
    if (request.method !== "PATCH") return normal(request);
    writes++;
    return writes === 1
      ? Response.json(
          { error: { code: "invalid_input", message: "欄位無效", retrySameRequest: false } },
          { status: 400 },
        )
      : Response.json(project);
  });
  try {
    await f.click("儲存內容");
    assert.match(document.body.textContent!, /欄位無效/);
    assert.equal(
      document.querySelector<HTMLInputElement>('input[maxlength="120"]')!.matches(":disabled"),
      false,
    );
    await f.click("儲存內容");
    assert.equal(writes, 2);
  } finally {
    await f.dispose();
  }
});
