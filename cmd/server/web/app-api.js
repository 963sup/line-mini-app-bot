// Shared request and identity lifecycle; tokens stay in the LIFF SDK.
const controllers = new Set();
const resets = new Set();
let identityGeneration = 0;

export const authenticated = () => document.documentElement.dataset.authenticated === "true";
export const taiwan = value => value && !value.startsWith("0001-")
  ? new Date(new Date(value).getTime() + 28800000).toISOString().slice(0, 19) : "";
export const displayTime = value => taiwan(value).replace("T", " ") || "尚未下班";
export const errorMessage = error => error.name === "AbortError"
  ? "操作已中止或逾時；已送出的申請可使用相同內容重試。"
  : error instanceof TypeError ? "無法連線至伺服器，請稍後重試。" : error.message;

export function onIdentity(ready, reset) {
  resets.add(reset);
  window.addEventListener("miniapp:identity-ready", ready);
  if (authenticated()) ready();
}

window.addEventListener("miniapp:identity-reset", () => {
  identityGeneration += 1;
  for (const controller of controllers) controller.abort();
  for (const reset of resets) reset();
});

export async function api(path, { body, signal, key } = {}) {
  if (!authenticated() || !window.liff?.isLoggedIn()) throw new Error("請先完成 LINE 身分驗證。");
  const token = window.liff.getIDToken();
  if (!token) throw new Error("LINE 登入已失效，請重新登入。");
  const generation = identityGeneration;
  const controller = new AbortController();
  const cancel = () => controller.abort();
  signal?.addEventListener("abort", cancel, { once: true });
  if (signal?.aborted) controller.abort();
  controllers.add(controller);
  const timer = setTimeout(cancel, 15000);
  try {
    const response = await fetch(path, {
      method: body === undefined ? "GET" : "POST", credentials: "omit", cache: "no-store", signal: controller.signal,
      headers: { Authorization: `Bearer ${token}`, ...(body === undefined ? {} : { "Content-Type": "application/json" }), ...(key ? { "Idempotency-Key": key } : {}) },
      body: body === undefined ? undefined : JSON.stringify(body),
    });
    const data = await response.json();
    if (generation !== identityGeneration || controller.signal.aborted) throw new DOMException("Obsolete request", "AbortError");
    if (!response.ok) {
      const error = new Error(data.error?.message || "操作失敗，請重試。");
      error.code = data.error?.code;
      throw error;
    }
    return data;
  } finally {
    clearTimeout(timer);
    controllers.delete(controller);
    signal?.removeEventListener("abort", cancel);
  }
}

// Cancellation saves work; the generation check also rejects responses already resolved.
export function latestRead() {
  let generation = 0, controller;
  return {
    cancel() { generation += 1; controller?.abort(); },
    async run(load, render, fail) {
      controller?.abort();
      controller = new AbortController();
      const current = ++generation;
      try {
        const value = await load(controller.signal);
        if (current === generation && authenticated()) render(value);
      } catch (error) {
        if (current === generation && authenticated()) fail(error);
      }
    },
  };
}
