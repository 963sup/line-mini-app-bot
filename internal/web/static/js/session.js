"use strict";

const miniStatus = document.querySelector("#mini-status");
const loginButton = document.querySelector("#login");
const verifyButton = document.querySelector("#verify");
const logoutButton = document.querySelector("#logout");
const identity = document.querySelector("#identity");
let verification, initialization;
let generation = 0, initialized = false;

function resetIdentity() {
  generation += 1;
  verification?.abort(); initialization?.abort();
  document.documentElement.dataset.authenticated = "false";
  identity.hidden = true;
  document.querySelector("#display-name").textContent = "";
  document.querySelector("#user-id").textContent = "";
  window.dispatchEvent(new Event("miniapp:identity-reset"));
}
async function initialize() {
  resetIdentity();
  const current = generation;
  const controller = new AbortController(); initialization = controller;
  const timer = setTimeout(() => controller.abort(), 10000);
  try {
    if (!window.liff) throw new Error("LINE SDK 未載入，請確認網路後重新整理。");
    const response = await fetch("/api/v1/miniapp/config", { cache: "no-store", signal: controller.signal });
    const config = await response.json();
    if (current !== generation) return;
    if (!response.ok) throw new Error(config.error?.message || "MINI App 尚未設定。");
    await liff.init({ liffId: config.liffId });
    if (current !== generation) return;
    initialized = true;
    if (!liff.isLoggedIn()) {
      loginButton.hidden = false;
      miniStatus.textContent = "請登入 LINE 後驗證本人身分。";
      return;
    }
    clearTimeout(timer); initialization = undefined;
    await verifyIdentity();
  } catch (error) {
    if (current === generation) miniStatus.textContent = error.name === "AbortError" ? "初始化逾時，請重新整理。" : error.message || "無法初始化 LINE，請稍後重試。";
  } finally { clearTimeout(timer); if (initialization === controller) initialization = undefined; }
}
async function verifyIdentity() {
  resetIdentity();
  const current = generation;
  const controller = new AbortController(); verification = controller;
  const timer = setTimeout(() => controller.abort(), 10000);
  verifyButton.hidden = false; verifyButton.disabled = true; logoutButton.hidden = false; loginButton.hidden = true;
  try {
    const token = liff.getIDToken();
    if (!token) throw new Error("無法取得 LINE 身分，請確認 openid scope 並重新登入。");
    miniStatus.textContent = "正在由伺服器驗證 LINE 身分……";
    const response = await fetch("/api/v1/me", {
      headers: { Authorization: `Bearer ${token}` }, cache: "no-store", credentials: "omit", signal: controller.signal,
    });
    const data = await response.json();
    if (current !== generation) return;
    if (!response.ok) throw new Error(data.error?.message || "身分驗證失敗。");
    document.querySelector("#display-name").textContent = data.displayName || "未提供名稱";
    document.querySelector("#user-id").textContent = data.userId;
    identity.hidden = false;
    document.documentElement.dataset.authenticated = "true";
    miniStatus.textContent = "本人身分驗證成功。";
    window.dispatchEvent(new Event("miniapp:identity-ready"));
  } catch (error) {
    if (current === generation) miniStatus.textContent = error.name === "AbortError" ? "驗證已中止或逾時，請重試。" : error instanceof TypeError ? "無法連線至伺服器，請重試。" : error.message;
  } finally { clearTimeout(timer); if (current === generation) verifyButton.disabled = false; }
}
loginButton.addEventListener("click", () => { liff.login({ redirectUri: window.location.href }); });
verifyButton.addEventListener("click", () => { void (initialized ? verifyIdentity() : initialize()); });
logoutButton.addEventListener("click", () => {
  resetIdentity();
  const inClient = liff.isInClient();
  try { if (!inClient) liff.logout(); } catch { /* Local identity is already cleared. */ }
  verifyButton.hidden = !inClient; verifyButton.disabled = false;
  logoutButton.hidden = true; loginButton.hidden = inClient;
  miniStatus.textContent = "已清除本頁登入狀態；LINE 內建瀏覽器可能在下次開啟時自動登入。";
});
window.addEventListener("pagehide", () => { resetIdentity(); verifyButton.disabled = false; });
window.addEventListener("pageshow", event => { if (event.persisted) void (initialized ? verifyIdentity() : initialize()); });
void initialize();
