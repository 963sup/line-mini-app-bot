"use strict";
(() => {
  const $ = id => document.getElementById(id);
  const locate = $("locate");
  if (!locate) return;
  const status = $("status");
  const result = $("result");
  let generation = 0;
  let phase = "idle";
  let timer;
  let controller;
  let lastCapturedAt;

  function clearPending() {
    clearTimeout(timer);
    timer = undefined;
    controller?.abort();
    controller = undefined;
  }

  function showResultAge() {
    if (result.hidden || !lastCapturedAt) return;
    $("captured").textContent = `${lastCapturedAt.toLocaleString()}（上次取得）`;
  }

  function finish(id, message, success = false) {
    if (id !== generation) return;
    generation += 1;
    clearPending();
    phase = success ? "success" : "error";
    locate.disabled = false;
    status.textContent = message;
    if (!success) showResultAge();
  }

  function requestCurrentPosition() {
    if (phase === "locating" || phase === "validating") return;
    clearPending();
    const id = ++generation;
    if (!window.isSecureContext) {
      finish(id, "定位需要 HTTPS；本機請使用 localhost 開啟。");
      return;
    }
    if (!navigator.geolocation) {
      finish(id, "此瀏覽器不支援定位，請改用支援定位的瀏覽器。");
      return;
    }
    phase = "locating";
    locate.disabled = true;
    showResultAge();
    status.textContent = result.hidden ? "正在取得目前位置……" : "正在更新位置；下方仍顯示上次結果……";
    // The browser's timeout may exclude a pending permission prompt.
    timer = setTimeout(() => finish(id, "定位等待逾時，請檢查權限後重試。"), 15000);
    try {
      navigator.geolocation.getCurrentPosition(
        position => { void submitPosition(id, position); },
        error => {
          const messages = {
            1: "未取得定位權限。請檢查瀏覽器、LINE 與系統的定位設定。",
            2: "目前無法取得位置，請確認裝置定位與網路狀態後再試。",
            3: "取得位置逾時，請稍後再按一次按鈕。",
          };
          finish(id, messages[error.code] || "定位失敗，請稍後重試。");
        },
        { enableHighAccuracy: false, timeout: 10000, maximumAge: 0 },
      );
    } catch {
      finish(id, "瀏覽器無法啟動定位，請檢查權限或改用其他瀏覽器。");
    }
  }

  async function submitPosition(id, position) {
    if (id !== generation || phase !== "locating") return;
    clearTimeout(timer);
    try {
      const { latitude, longitude, accuracy } = position.coords;
      const capturedAt = new Date(position.timestamp);
      if (!Number.isFinite(latitude) || Math.abs(latitude) > 90 ||
          !Number.isFinite(longitude) || Math.abs(longitude) > 180 ||
          !Number.isFinite(accuracy) || accuracy < 0 || !Number.isFinite(capturedAt.getTime())) {
        throw new Error("裝置傳回的位置資料無效，請重新定位。");
      }
      phase = "validating";
      controller = new AbortController();
      timer = setTimeout(() => finish(id, "伺服器回應逾時，請重新取得位置。"), 10000);
      status.textContent = "已取得新位置，正在驗證……";
      const response = await fetch("/api/location/validate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "omit",
        signal: controller.signal,
        body: JSON.stringify({ latitude, longitude, accuracyMeters: accuracy, capturedAt: capturedAt.toISOString() }),
      });
      const data = await response.json();
      if (id !== generation || phase !== "validating") return;
      if (!response.ok || data.valid !== true) {
        throw new Error(typeof data.error === "string" ? data.error : "伺服器未通過位置驗證。");
      }
      $("latitude").textContent = latitude.toFixed(6);
      $("longitude").textContent = longitude.toFixed(6);
      $("accuracy").textContent = `約 ${Math.round(accuracy)} 公尺`;
      lastCapturedAt = capturedAt;
      $("captured").textContent = capturedAt.toLocaleString();
      $("validation").textContent = "格式與時間驗證通過";
      result.hidden = false;
      finish(id, "位置驗證完成，資料未保存。", true);
    } catch (error) {
      if (id !== generation || phase === "error") return;
      finish(id, error instanceof TypeError ? "無法連線至伺服器，請檢查網路後重試。" : error.message || "位置驗證失敗，請重試。");
    }
  }

  locate.addEventListener("click", requestCurrentPosition);
  window.addEventListener("pagehide", () => {
    generation += 1;
    clearPending();
    phase = "idle";
    locate.disabled = false;
    showResultAge();
    status.textContent = "已停止定位。";
  });
})();
