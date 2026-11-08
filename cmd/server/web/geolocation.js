"use strict";
(() => {

const locate = document.querySelector("#locate");
const status = document.querySelector("#status");
const result = document.querySelector("#result");
let requestID = 0;
let controller;
let deadline;

function stopRequest(message) {
  requestID += 1;
  clearTimeout(deadline);
  controller?.abort();
  controller = undefined;
  locate.disabled = false;
  result.hidden = true;
  status.textContent = message;
}

function finishRequest(id, message) {
  if (id !== requestID) return;
  clearTimeout(deadline);
  locate.disabled = false;
  status.textContent = message;
}

function requestCurrentPosition() {
  if (locate.disabled) return;
  stopRequest("正在取得位置，請回應瀏覽器的權限提示……");
  const id = requestID;
  if (!window.isSecureContext) {
    status.textContent = "定位需要 HTTPS；本機請使用 localhost 開啟。";
    return;
  }
  if (!navigator.geolocation) {
    status.textContent = "此瀏覽器不支援定位，請改用支援定位的瀏覽器。";
    return;
  }
  locate.disabled = true;
  // Geolocation's own timeout may not cover a pending permission prompt.
  deadline = setTimeout(() => stopRequest("定位等待逾時，請檢查權限後重試。"), 15000);
  try {
    navigator.geolocation.getCurrentPosition(
      (position) => { void submitPosition(id, position); },
      (error) => {
        const messages = {
          1: "未取得定位權限。請檢查瀏覽器、LINE 與系統的定位設定。",
          2: "目前無法取得位置，請確認裝置定位與網路狀態後再試。",
          3: "取得位置逾時，請稍後再按一次按鈕。",
        };
        finishRequest(id, messages[error.code] || "定位失敗，請稍後重試。");
      },
      { enableHighAccuracy: false, timeout: 10000, maximumAge: 0 },
    );
  } catch {
    finishRequest(id, "瀏覽器無法啟動定位，請檢查權限或改用其他瀏覽器。");
  }
}

async function submitPosition(id, position) {
  if (id !== requestID) return;
  clearTimeout(deadline);
  controller = new AbortController();
  deadline = setTimeout(() => stopRequest("伺服器回應逾時，請重新取得位置。"), 10000);
  try {
    const { latitude, longitude, accuracy } = position.coords;
    const capturedAt = new Date(position.timestamp).toISOString();
    const validLatitude = Number.isFinite(latitude) && Math.abs(latitude) <= 90;
    const validLongitude = Number.isFinite(longitude) && Math.abs(longitude) <= 180;
    if (!validLatitude || !validLongitude || !Number.isFinite(accuracy) || accuracy < 0) {
      throw new Error("裝置傳回的位置資料無效，請重新定位。");
    }
    document.querySelector("#latitude").textContent = latitude.toFixed(6);
    document.querySelector("#longitude").textContent = longitude.toFixed(6);
    document.querySelector("#accuracy").textContent = `約 ${Math.round(accuracy)} 公尺`;
    document.querySelector("#captured").textContent = new Date(capturedAt).toLocaleString();
    document.querySelector("#validation").textContent = "驗證中……";
    result.hidden = false;
    status.textContent = "已取得位置，正在傳送至本站驗證……";
    const response = await fetch("/api/location/validate", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "omit",
      signal: controller.signal,
      body: JSON.stringify({ latitude, longitude, accuracyMeters: accuracy, capturedAt }),
    });
    const data = await response.json();
    if (id !== requestID) return;
    if (!response.ok || data.valid !== true) {
      throw new Error(data.error || "伺服器未通過位置驗證。");
    }
    document.querySelector("#validation").textContent = "格式與時間驗證通過";
    finishRequest(id, "位置驗證完成，資料未保存。");
  } catch (error) {
    if (id !== requestID) return;
    document.querySelector("#validation").textContent = "未通過驗證";
    const message = error instanceof TypeError ? "無法連線至伺服器，請檢查網路後重試。" : error.message;
    finishRequest(id, message || "位置驗證失敗，請重試。");
  }
}

locate.addEventListener("click", requestCurrentPosition);
window.addEventListener("pagehide", () => stopRequest("已停止定位。"));
})();
