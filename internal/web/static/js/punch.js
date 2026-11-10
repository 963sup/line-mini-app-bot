"use strict";

const punchButton = document.querySelector("#punch");
const punchStatus = document.querySelector("#punch-status");
let punchGeneration = 0;
let punchController;
let autoStarted = false;
let exchangedRequest = "";
let exchanging;
let exchangeController;
let exchangeGeneration = 0;
let windowClosed = false;
let punchWatchId;

function stopPunchWatch() {
  if (punchWatchId !== undefined) navigator.geolocation.clearWatch(punchWatchId);
  punchWatchId = undefined;
}

function resetPunch(message) {
  stopPunchWatch();
  punchGeneration += 1;
  exchangeController?.abort();
  exchangeController = undefined;
  exchangeGeneration += 1;
  exchanging = undefined;
  exchangedRequest = "";
  punchController?.abort();
  punchController = undefined;
  clearTimeout(punchTimer);
  punchPhase = "idle";
  punchButton.disabled = false;
  punchButton.removeAttribute("aria-busy");
  if (message) punchStatus.textContent = message;
}

function launchToken() {
  try {
    const params = new URLSearchParams(window.location.search);
    const nested = params.get("liff.state");
    const token = params.get("launch") || (nested ? new URL(nested, window.location.origin).searchParams.get("launch") : "");
    return /^[a-f0-9]{64}$/.test(token || "") ? token : "";
  } catch { return ""; }
}

async function resolveRequest() {
  if (exchangedRequest) return exchangedRequest;
  const launch = launchToken();
  if (!launch) return "";
  if (exchanging) return exchanging;
  const pending = (async () => {
    const generation = exchangeGeneration;
    const token = liff.getIDToken();
    if (!token) throw new Error("無法取得 LINE 身分，請重新登入。");
    const controller = new AbortController();
    exchangeController = controller;
    try {
      const response = await fetch("/api/v1/attendance/launch", {
        method: "POST", credentials: "omit", signal: controller.signal,
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
        body: JSON.stringify({ launch }),
      });
      const data = await response.json();
      if (generation !== exchangeGeneration) throw new Error("操作已停止，請重新開啟打卡頁。");
      if (!response.ok) throw new Error(data?.error?.message || "無法確認打卡連結。");
      if (!/^[a-f0-9]{64}$/.test(data?.requestId || "")) throw new Error("打卡連結回應無效。");
      exchangedRequest = data.requestId;
      return exchangedRequest;
    } finally { if (exchangeController === controller) exchangeController = undefined; }
  })();
  exchanging = pending;
  try { return await pending; } finally { if (exchanging === pending) exchanging = undefined; }
}

function showAttendance() {
  document.querySelector("#attendance").hidden = !launchToken();
}

let punchPhase = "idle";
let punchTimer;
const unknownResult = "打卡結果尚未確認；請先私訊查詢本人紀錄。重試仍使用同一打卡連結。";

function finishPunch(id, message) {
  if (id !== punchGeneration) return;
  stopPunchWatch();
  if (punchPhase === "exchanging") {
    exchangeGeneration += 1;
    exchangeController?.abort();
    exchanging = undefined;
  }
  punchGeneration += 1;
  clearTimeout(punchTimer);
  punchController?.abort();
  punchController = undefined;
  punchPhase = "idle";
  punchButton.disabled = false;
  punchButton.removeAttribute("aria-busy");
  punchStatus.textContent = message;
}

async function startPunch() {
  if (punchPhase !== "idle") return;
  if (!window.liff?.isLoggedIn()) { punchStatus.textContent = "請先登入 LINE。"; return; }
  if (!window.isSecureContext || !navigator.geolocation) { punchStatus.textContent = "此瀏覽器無法定位，請使用 HTTPS 並確認定位權限。"; return; }
  const id = ++punchGeneration;
  punchPhase = "exchanging";
  punchButton.disabled = true;
  punchButton.setAttribute("aria-busy", "true");
  punchStatus.textContent = "正在確認群組打卡連結……";
  punchTimer = setTimeout(() => finishPunch(id, "確認連結逾時，請重試。"), 10000);
  let requestId;
  try { requestId = await resolveRequest(); }
  catch (error) { finishPunch(id, error.name === "AbortError" ? "確認連結逾時，請重試。" : error.message || "無法確認打卡連結。"); return; }
  if (id !== punchGeneration) return;
  clearTimeout(punchTimer);
  if (!requestId) { finishPunch(id, "打卡連結無效，請回群組重新開啟。"); return; }
  punchPhase = "locating";
  punchButton.disabled = true;
  punchButton.setAttribute("aria-busy", "true");
  punchStatus.textContent = "正在取得目前位置；若瀏覽器詢問權限，請允許定位。";
  // The browser timeout can exclude permission and visibility waits.
  punchTimer = setTimeout(() => finishPunch(id, "定位等待逾時，請檢查權限後重試。"), 15000);
  try {
    punchWatchId = navigator.geolocation.watchPosition(async (position) => {
      if (id !== punchGeneration || punchPhase !== "locating") return;
      stopPunchWatch();
      clearTimeout(punchTimer);
      try {
        const { latitude, longitude, accuracy } = position.coords;
        const capturedAt = new Date(position.timestamp);
        if (!Number.isFinite(latitude) || Math.abs(latitude) > 90 || !Number.isFinite(longitude) || Math.abs(longitude) > 180 ||
            !Number.isFinite(accuracy) || accuracy < 0 || !Number.isFinite(capturedAt.getTime())) {
          throw new Error("裝置傳回的位置無效，請重新定位。");
        }
        const token = liff.getIDToken();
        if (!token) throw new Error("無法取得 LINE 身分，請重新登入。");
        punchPhase = "submitting";
        punchController = new AbortController();
        punchTimer = setTimeout(() => finishPunch(id, unknownResult), 10000);
        punchStatus.textContent = "已取得位置，正在保存打卡……";
        const response = await fetch("/api/v1/attendance/punches", {
          method: "POST", credentials: "omit", signal: punchController.signal,
          headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
          body: JSON.stringify({ requestId, latitude, longitude, accuracyMeters: accuracy, capturedAt: capturedAt.toISOString() }),
        });
        const data = await response.json().catch(() => null);
        if (id !== punchGeneration || punchPhase !== "submitting") return;
        if (!response.ok) throw new Error(data?.error?.message || "打卡失敗，請重試。");
        if (typeof data?.code !== "string" || typeof data?.message !== "string" || !data.message) {
          throw new Error(unknownResult);
        }
        finishPunch(id, data.message);
        if (!windowClosed && ["clocked_in", "clocked_out", "break_started", "break_ended", "rest_confirmed"].includes(data.code) && liff.isInClient()) {
          windowClosed = true;
          try { liff.closeWindow(); } catch { windowClosed = false; }
        }
      } catch (error) {
        if (id !== punchGeneration) return;
        finishPunch(id, punchPhase === "submitting" && (error.name === "AbortError" || error instanceof TypeError) ? unknownResult : error.message || "打卡失敗，請重試。");
      }
    }, (error) => {
      stopPunchWatch();
      const messages = { 1: "未取得定位權限，請調整 LINE、瀏覽器與系統設定後重試。", 2: "目前無法取得位置，請稍後重試。", 3: "定位逾時，請重試。" };
      finishPunch(id, messages[error.code] || "定位失敗，請重試。");
    }, { enableHighAccuracy: false, timeout: 10000, maximumAge: 0 });
  } catch { finishPunch(id, "瀏覽器無法啟動定位，請確認權限後重試。"); }
}

punchButton.addEventListener("click", startPunch);
window.addEventListener("miniapp:identity-ready", () => {
  if (autoStarted || !launchToken()) return;
  autoStarted = true;
  startPunch();
});

window.addEventListener("pagehide", () => {
  const message = punchPhase === "submitting" ? unknownResult :
    punchPhase !== "idle" ? "頁面切換使定位中斷；返回後請按打卡按鈕重試。" : "";
  resetPunch(message);
});
document.querySelector("#logout").addEventListener("click", () => resetPunch("已清除本頁打卡狀態。"));
showAttendance();
