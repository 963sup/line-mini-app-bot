"use strict";

const punchButton = document.querySelector("#punch");
const punchStatus = document.querySelector("#punch-status");
let punchGeneration = 0;
let punchController;
let punchTimer;

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
  punchStatus.textContent = "正在取得位置，請允許定位……";
  clearTimeout(punchTimer);
  punchController = new AbortController();
  function finish(message) {
    if (generation !== punchGeneration) return;
    clearTimeout(punchTimer);
    punchStatus.textContent = message;
    punchButton.disabled = false;
  }
  punchTimer = setTimeout(() => {
    finish("等待定位或伺服器逾時，請重試；若已送出，可私訊查詢紀錄。");
    punchGeneration += 1;
    punchController.abort();
  }, 30000);
  try { navigator.geolocation.getCurrentPosition(async (position) => {
    if (generation !== punchGeneration) return;
    try {
      const token = liff.getIDToken();
      if (!token) throw new Error("無法取得 LINE 身分，請重新登入。");
      punchStatus.textContent = "正在保存打卡，請稍候……";
      const response = await fetch("/api/v1/attendance/punches", {
        method: "POST", credentials: "omit", signal: punchController.signal,
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
        body: JSON.stringify({ requestId, latitude: position.coords.latitude, longitude: position.coords.longitude,
          accuracyMeters: position.coords.accuracy, capturedAt: new Date(position.timestamp).toISOString() }),
      });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error?.message || "打卡失敗，請重試。");
      finish(data.message);
    } catch (error) {
      finish(error.name === "AbortError" ? "操作逾時，請重試或私訊查詢紀錄。" : error.message);
    }
  }, (error) => {
    const messages = { 1: "未取得定位權限，請調整 LINE、瀏覽器與系統設定後重試。", 2: "目前無法取得位置，請稍後重試。", 3: "定位逾時，請重試。" };
    finish(messages[error.code] || "定位失敗，請重試。");
  }, { enableHighAccuracy: false, timeout: 10000, maximumAge: 0 });
  } catch { finish("瀏覽器無法啟動定位，請確認權限後重試。"); }
});

window.addEventListener("pagehide", () => {
  punchGeneration += 1;
  clearTimeout(punchTimer);
  punchController?.abort();
  punchButton.disabled = false;
});
document.querySelector("#logout").addEventListener("click", () => {
  punchGeneration += 1;
  clearTimeout(punchTimer);
  punchController?.abort();
  punchButton.disabled = false;
  punchStatus.textContent = "已清除本頁打卡狀態。";
});
showAttendance();
