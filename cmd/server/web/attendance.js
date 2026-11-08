"use strict";

const punchButton = document.querySelector("#punch");
const punchStatus = document.querySelector("#punch-status");
let punchGeneration = 0;
let punchController;

function resetPunch(message) {
  punchGeneration += 1;
  punchController?.abort();
  punchController = undefined;
  punchButton.disabled = false;
  punchButton.removeAttribute("aria-busy");
  if (message) punchStatus.textContent = message;
}

function attendanceRequest() {
  try {
  const params = new URLSearchParams(window.location.search);
  const nested = params.get("liff.state");
  const request = params.get("request") || (nested ? new URL(nested, window.location.origin).searchParams.get("request") : "");
  return /^[a-f0-9]{64}$/.test(request || "") ? request : "";
  } catch { return ""; }
}

function showAttendance() {
  document.querySelector("#attendance").hidden = !attendanceRequest();
}

punchButton.addEventListener("click", () => {
  const requestId = attendanceRequest();
  if (!requestId) return;
  if (!window.liff?.isLoggedIn()) { punchStatus.textContent = "請先登入 LINE。"; return; }
  if (!window.isSecureContext || !navigator.geolocation) { punchStatus.textContent = "此瀏覽器無法定位，請使用 HTTPS 並確認定位權限。"; return; }
  const generation = ++punchGeneration;
  punchButton.disabled = true;
  punchButton.setAttribute("aria-busy", "true");
  punchStatus.textContent = "請允許定位；取得位置後會自動送出打卡。";
  punchController = new AbortController();
  function finish(message) {
    if (generation !== punchGeneration) return;
    punchStatus.textContent = message;
    punchButton.disabled = false;
    punchButton.removeAttribute("aria-busy");
  }
  try { navigator.geolocation.getCurrentPosition(async (position) => {
    if (generation !== punchGeneration) return;
    try {
      const { latitude, longitude, accuracy } = position.coords;
      const capturedAt = position.timestamp;
      if (!Number.isFinite(capturedAt)) throw new Error("裝置傳回的位置時間無效，請重新定位。");
      const token = liff.getIDToken();
      if (!token) throw new Error("無法取得 LINE 身分，請重新登入。");
      punchStatus.textContent = "已取得位置，正在保存打卡……";
      const response = await fetch("/api/v1/attendance/punches", {
        method: "POST", credentials: "omit", signal: punchController.signal,
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
        body: JSON.stringify({ requestId, latitude, longitude,
          accuracyMeters: accuracy, capturedAt: new Date(capturedAt).toISOString() }),
      });
      const data = await response.json().catch(() => null);
      if (!response.ok) throw new Error(data?.error?.message || "打卡失敗，請重試。");
      if (!data?.message) throw new Error("無法確認打卡結果，請私訊查詢紀錄。");
      finish(data.message);
    } catch (error) {
      finish(error.name === "AbortError" ? "操作已取消；若資料已送出，請私訊查詢紀錄。" : error.message);
    }
  }, (error) => {
    const messages = { 1: "未取得定位權限，請調整 LINE、瀏覽器與系統設定後重試。", 2: "目前無法取得位置，請稍後重試。", 3: "定位逾時，請重試。" };
    finish(messages[error.code] || "定位失敗，請重試。");
  });
  } catch { finish("瀏覽器無法啟動定位，請確認權限後重試。"); }
});

window.addEventListener("pagehide", () => resetPunch());
document.querySelector("#logout").addEventListener("click", () => resetPunch("已清除本頁打卡狀態。"));
showAttendance();
