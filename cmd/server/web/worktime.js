"use strict";
(() => {
  const $ = id => document.getElementById(id);
  let records = [], selected = null, reviewed = "";
  const controllers = new Set();
  const taiwan = value => value && !value.startsWith("0001-") ? new Date(new Date(value).getTime() + 28800000).toISOString().slice(0, 19) : "";
  const timestamp = value => {
    if (!value) throw new Error("請填寫完整的時間。");
    return new Date(`${value.length === 16 ? `${value}:00` : value}+08:00`).toISOString();
  };
  $("history-month").value = taiwan(new Date().toISOString()).slice(0, 7);
  async function api(path, body) {
    const token = window.liff?.getIDToken();
    if (!token || !liff.isLoggedIn()) throw new Error("請先登入 LINE。");
    const controller = new AbortController(); controllers.add(controller);
    const timer = setTimeout(() => controller.abort(), 15000);
    try {
      const response = await fetch(path, {method: body === undefined ? "GET" : "POST", credentials: "omit", cache: "no-store",
        headers: {Authorization: `Bearer ${token}`, ...(body === undefined ? {} : {"Content-Type": "application/json"})},
        body: body === undefined ? undefined : JSON.stringify(body), signal: controller.signal});
      const data = await response.json();
      if (!response.ok) throw new Error(data.error?.message || "操作失敗，請重試。");
      return data;
    } finally { clearTimeout(timer); controllers.delete(controller); }
  }
  async function run(button, action) {
    button.disabled = true;
    try { await action(); }
    catch (error) { $("work-status").textContent = error.name === "AbortError" ? "操作已中止或逾時；如已送出，請查詢申請結果。" : error.message; }
    finally { button.disabled = button.id === "approve-correction" && !reviewed; }
  }
  function addBreak(start = "", end = "") {
    const row = document.createElement("div");
    for (const [label, value] of [["休息開始", start], ["休息結束", end]]) {
      const wrap = document.createElement("label"); wrap.textContent = label;
      const input = document.createElement("input"); input.type = "datetime-local"; input.step = "1"; input.required = true; input.value = taiwan(value);
      wrap.append(input); row.append(wrap);
    }
    const remove = document.createElement("button"); remove.type = "button"; remove.textContent = "移除";
    remove.addEventListener("click", () => row.remove()); row.append(remove); $("correction-breaks").append(row);
  }
  function selectRecord() {
    selected = records.find(r => r.RecordID === $("correction-record").value) || null;
    $("correction-start").value = taiwan(selected?.ClockInAt);
    $("correction-end").value = taiwan(selected?.ClockOutAt);
    $("correction-breaks").replaceChildren();
    for (const b of selected?.Breaks || []) addBreak(b.Start, b.End);
    $("correction-confirm").checked = false; $("correction-reason").value = "";
  }
  async function loadMonth() {
    const group = $("history-group").value, month = $("history-month").value;
    if (!group || !month) throw new Error("請先選擇群組與月份。");
    const report = await api(`/api/v1/attendance/monthly?${new URLSearchParams({group, month})}`);
    records = report.Records || [];
    $("correction-record").replaceChildren(...records.map(r => new Option(`${taiwan(r.ClockInAt).replace("T", " ")} → ${taiwan(r.ClockOutAt).replace("T", " ") || "尚未下班"}`, r.RecordID)));
    $("month-summary").textContent = `已確認工作 ${report.Calculation.Minutes} 分鐘；待確認 ${report.Calculation.Pending} 筆。正常／加班分類待認定。`;
    selectRecord();
  }
  $("load-groups").addEventListener("click", e => run(e.currentTarget, async () => {
    const groups = await api("/api/v1/attendance/groups");
    $("history-group").replaceChildren(...groups.map(g => new Option(g.name, g.id)));
    $("work-status").textContent = groups.length ? "請選擇月份後查詢。" : "目前沒有出勤群組。";
    records = []; $("correction-record").replaceChildren(); selectRecord();
  }));
  $("load-month").addEventListener("click", e => run(e.currentTarget, loadMonth));
  for (const id of ["history-group", "history-month"]) $(id).addEventListener("change", () => { records = []; $("correction-record").replaceChildren(); $("month-summary").textContent = ""; selectRecord(); });
  $("correction-record").addEventListener("change", selectRecord);
  $("add-break").addEventListener("click", () => addBreak());
  $("correction-form").addEventListener("submit", e => {
    e.preventDefault(); run(e.submitter, async () => {
      if (!selected) throw new Error("請先查詢並選擇班次。");
      const breaks = [...$("correction-breaks").children].map(row => {const inputs = row.querySelectorAll("input"); return {Start: timestamp(inputs[0].value), End: timestamp(inputs[1].value)};});
      const data = await api("/api/v1/attendance/corrections", {groupId: selected.GroupID, recordId: selected.RecordID,
        expectedRevision: selected.Revision, expectedEventSequence: selected.EventSequence,
        reason: $("correction-reason").value, start: timestamp($("correction-start").value), end: timestamp($("correction-end").value),
        breaks, restConfirmed: $("correction-confirm").checked});
      $("correction-id").value = data.id; clearReview();
      $("work-status").textContent = `已送出。申請編號：${data.id}。請提供給授權審核者；核准前工時不變。`;
    });
  });
  function requestID() {
    const id = $("correction-id").value.trim();
    if (!/^[A-Za-z0-9_-]{1,128}$/.test(id)) throw new Error("請輸入申請編號。");
    return id;
  }
  function clearReview() { reviewed = ""; $("review-details").textContent = ""; $("approve-correction").disabled = true; }
  $("correction-id").addEventListener("input", clearReview);
  $("check-correction").addEventListener("click", e => run(e.currentTarget, async () => {
    const data = await api(`/api/v1/attendance/corrections/${requestID()}`);
    const labels = {pending: "待審核", approved: "已核准，可重新查詢月報", stale: "原紀錄已變動，請重新載入月報並申請", invalid: "核准證據異常，請聯絡管理者"};
    $("work-status").textContent = labels[data.status] || "申請狀態無法確認。";
  }));
  $("review-correction").addEventListener("click", e => run(e.currentTarget, async () => {
    clearReview(); const id = requestID(); const c = await api(`/api/v1/attendance/corrections/${id}/review`);
    if (requestID() !== id) return;
    $("review-details").textContent = `申請 ${id}\n原因：${c.reason}\n上班：${taiwan(c.start)}\n下班：${taiwan(c.end)}\n休息：${(c.breaks || []).map(b => `${taiwan(b.Start)} ～ ${taiwan(b.End)}`).join("；") || "無"}\n休息完整確認：${c.restConfirmed ? "是" : "否"}`;
    reviewed = id; $("approve-correction").disabled = false;
  }));
  $("approve-correction").addEventListener("click", e => run(e.currentTarget, async () => {
    if (!reviewed || reviewed !== requestID()) throw new Error("請先讀取申請內容。");
    await api(`/api/v1/attendance/corrections/${reviewed}/approve`, {});
    clearReview(); $("work-status").textContent = "已核准；員工可查詢申請結果並重新載入月報。";
  }));
  function reset() {
    for (const c of controllers) c.abort();
    records = []; selected = null; clearReview();
    for (const id of ["history-group", "correction-record", "correction-breaks"]) $(id).replaceChildren();
    $("correction-form").reset(); $("correction-id").value = ""; $("month-summary").textContent = ""; $("work-status").textContent = "";
  }
  $("logout").addEventListener("click", reset); window.addEventListener("pagehide", reset);
})();
