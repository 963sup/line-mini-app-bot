import { api, authenticated, onIdentity, latestRead, errorMessage } from "/assets/api.js";
import { taiwan, displayTime } from "/assets/time.js";

const $ = id => document.getElementById(id);
const params = new URLSearchParams(location.search);
const group = params.get("group") || "", recordID = params.get("record") || "";
const recordRead = latestRead(), statusRead = latestRead();
let selected = null, attempt = null, submitting = false, generation = 0;
const back = new URLSearchParams();
if (group) back.set("group", group);
if (/^\d{4}-\d{2}$/.test(params.get("month") || "")) back.set("month", params.get("month"));
$("back-history").href = `/miniapp/history?${back}`;

function timestamp(value) {
  if (!value) throw new Error("請填寫完整的時間。");
  const date = new Date(`${value.length === 16 ? `${value}:00` : value}+08:00`);
  if (!Number.isFinite(date.getTime())) throw new Error("時間格式無效。");
  return date.toISOString();
}
function addBreak(start = "", end = "") {
  const row = document.createElement("div");
  for (const [label, value] of [["休息開始", start], ["休息結束", end]]) {
    const wrap = document.createElement("label"); wrap.textContent = label;
    const input = document.createElement("input"); input.type = "datetime-local"; input.step = "1"; input.required = true; input.value = taiwan(value);
    wrap.append(input); row.append(wrap);
  }
  const remove = document.createElement("button"); remove.type = "button"; remove.className = "secondary"; remove.textContent = "移除";
  remove.addEventListener("click", () => row.remove());
  row.append(remove); $("correction-breaks").append(row);
}
function loadRecord() {
  if (!group || !/^[A-Za-z0-9_-]{1,128}$/.test(recordID)) {
    $("correction-status").textContent = "請從本人月紀錄選擇班次；也可在下方查詢已送出的申請。";
    return;
  }
  $("correction-status").textContent = "正在讀取班次……";
  void recordRead.run(signal => api(`/api/v1/attendance/records/${encodeURIComponent(recordID)}?${new URLSearchParams({ group })}`, { signal }), record => {
    selected = record;
    $("record-description").textContent = `原班次：${displayTime(record.ClockInAt)} → ${displayTime(record.ClockOutAt)}。核准前工時不變。`;
    $("correction-start").value = taiwan(record.ClockInAt); $("correction-end").value = taiwan(record.ClockOutAt);
    $("correction-breaks").replaceChildren();
    for (const rest of record.Breaks || []) addBreak(rest.Start, rest.End);
    $("correction-fields").disabled = false;
    $("correction-status").textContent = "請填寫實際工作與休息時間。";
  }, error => { $("correction-status").textContent = errorMessage(error); });
}
$("add-break").addEventListener("click", () => addBreak());
$("correction-form").addEventListener("submit", async event => {
  event.preventDefault();
  if (submitting || !selected || !authenticated()) return;
  const current = generation;
  try {
    const breaks = [...$("correction-breaks").children].map(row => {
      const inputs = row.querySelectorAll("input");
      return { Start: timestamp(inputs[0].value), End: timestamp(inputs[1].value) };
    });
    const body = { groupId: selected.GroupID, recordId: selected.RecordID,
      expectedRevision: selected.Revision, expectedEventSequence: selected.EventSequence,
      reason: $("correction-reason").value, start: timestamp($("correction-start").value), end: timestamp($("correction-end").value),
      breaks, restConfirmed: $("correction-confirm").checked };
    const serialized = JSON.stringify(body);
    // The key is fixed for this exact draft, including after an ambiguous timeout.
    if (!attempt || attempt.serialized !== serialized) attempt = { serialized, key: crypto.randomUUID() };
    submitting = true; $("correction-fields").disabled = true;
    $("correction-status").textContent = "正在送出申請……";
    const data = await api("/api/v1/attendance/corrections", { body, key: attempt.key });
    if (current !== generation) return;
    $("correction-id").value = data.id;
    $("correction-status").textContent = `申請已受理。編號：${data.id}。請保存編號並提供給授權審核者，可在下方查詢目前結果。`;
  } catch (error) {
    if (current !== generation) return;
    $("correction-status").textContent = `${errorMessage(error)}${error.code === "idempotency_conflict" ? " 請重新讀取班次並確認申請內容。" : " 結果未確認時，請保留本頁並以相同內容重試。"}`;
  } finally {
    if (current === generation) { submitting = false; $("correction-fields").disabled = !selected || !authenticated(); }
  }
});
function requestID() {
  const id = $("correction-id").value.trim();
  if (!/^[A-Za-z0-9_-]{1,128}$/.test(id)) throw new Error("請輸入有效的申請編號。");
  return id;
}
$("correction-id").addEventListener("input", () => { statusRead.cancel(); $("application-status").textContent = ""; });
$("check-correction").addEventListener("click", () => {
  let id;
  try { id = requestID(); } catch (error) { $("application-status").textContent = error.message; return; }
  $("application-status").textContent = "正在查詢申請……";
  void statusRead.run(signal => api(`/api/v1/attendance/corrections/${id}`, { signal }), data => {
    const labels = { pending: "待審核", approved: "已核准，請重新查詢月紀錄。", stale: "原紀錄已變動，請重新載入班次再提出申請。", invalid: "核准證據異常，請聯絡管理者。" };
    $("application-status").textContent = labels[data.status] || "申請狀態無法確認。";
  }, error => { $("application-status").textContent = errorMessage(error); });
});
onIdentity(() => { $("status-controls").disabled = false; loadRecord(); }, () => {
  generation += 1; recordRead.cancel(); statusRead.cancel();
  selected = null; attempt = null; submitting = false;
  $("correction-fields").disabled = true; $("status-controls").disabled = true;
  $("correction-form").reset(); $("correction-breaks").replaceChildren(); $("correction-id").value = "";
  $("record-description").textContent = "請先完成身分驗證，再讀取班次。";
  $("correction-status").textContent = "請先完成身分驗證。"; $("application-status").textContent = "";
});
