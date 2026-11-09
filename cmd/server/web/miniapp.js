"use strict";

const miniStatus = document.querySelector("#mini-status");
const loginButton = document.querySelector("#login");
const verifyButton = document.querySelector("#verify");
const logoutButton = document.querySelector("#logout");
const identity = document.querySelector("#identity");
let verification;
let generation = 0;

async function initialize() {
  try {
    if (!window.liff) throw new Error("LINE SDK 未載入，請確認網路後重新整理。");
    const response = await fetch("/api/v1/miniapp/config", { cache: "no-store" });
    const config = await response.json();
    if (!response.ok) throw new Error(config.error?.message || "MINI App 尚未設定。");
    await liff.init({ liffId: config.liffId });
    if (!liff.isLoggedIn()) {
      loginButton.hidden = false;
      miniStatus.textContent = "請登入 LINE 後驗證本人身分。";
      return;
    }
    verifyButton.hidden = false;
    logoutButton.hidden = false;
    miniStatus.textContent = "LINE 已登入，正在確認身分……";
    await verifyIdentity();
  } catch (error) {
    miniStatus.textContent = error.message || "無法初始化 LINE，請稍後重試。";
  }
}

async function verifyIdentity() {
  verification?.abort();
  const id = ++generation;
  const controller = new AbortController();
  verification = controller;
  const timer = setTimeout(() => controller.abort(), 10000);
  verifyButton.disabled = true;
  identity.hidden = true;
  try {
    const token = liff.getIDToken();
    if (!token) throw new Error("無法取得 ID token，請確認 Channel 已啟用 openid scope，並重新登入。");
    miniStatus.textContent = "正在由伺服器驗證 LINE 身分……";
    const response = await fetch("/api/v1/me", {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
      credentials: "omit",
      signal: controller.signal,
    });
    const data = await response.json();
    if (id !== generation) return;
    if (!response.ok) throw new Error(data.error?.message || "身分驗證失敗。");
    document.querySelector("#display-name").textContent = data.displayName || "未提供名稱";
    document.querySelector("#user-id").textContent = data.userId;
    identity.hidden = false;
    logoutButton.hidden = false;
    miniStatus.textContent = "本人身分驗證成功，資料未保存。";
    window.dispatchEvent(new Event("miniapp:identity-ready"));
  } catch (error) {
    if (id !== generation) return;
    miniStatus.textContent = error.name === "AbortError" ? "驗證已中止或逾時，請重試。" :
      (error instanceof TypeError ? "無法連線至伺服器，請重試。" : error.message);
  } finally {
    clearTimeout(timer);
    if (id === generation) verifyButton.disabled = false;
  }
}

loginButton.addEventListener("click", () => {
  liff.login({ redirectUri: window.location.href });
});
verifyButton.addEventListener("click", () => { void verifyIdentity(); });
logoutButton.addEventListener("click", () => {
  generation += 1;
  verification?.abort();
  const inClient = liff.isInClient();
  try {
    if (!inClient) liff.logout();
  } catch {
    // The local display is still cleared if the SDK cannot log out.
  }
  identity.hidden = true;
  document.querySelector("#display-name").textContent = "";
  document.querySelector("#user-id").textContent = "";
  verifyButton.hidden = !inClient;
  verifyButton.disabled = false;
  logoutButton.hidden = true;
  loginButton.hidden = inClient;
  miniStatus.textContent = "已清除本頁登入狀態；LINE 內建瀏覽器可能在下次開啟時自動登入。";
});
window.addEventListener("pagehide", () => {
  generation += 1;
  verification?.abort();
  identity.hidden = true;
  verifyButton.disabled = false;
});
void initialize();
