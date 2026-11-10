import { api, onIdentity, latestRead, taiwan, displayTime, errorMessage } from "/assets/app-api.js";

const $ = id => document.getElementById(id);
const monthRead = latestRead(), groupRead = latestRead();
let groupCursor = "", nextCursor = "", cursors = [""], page = 0;
const initial = new URLSearchParams(location.search);
let desiredGroup = initial.get("group") || "";
$("history-month").value = /^\d{4}-\d{2}$/.test(initial.get("month") || "") ? initial.get("month") : taiwan(new Date().toISOString()).slice(0, 7);

function clearMonth() {
  monthRead.cancel();
  nextCursor = ""; cursors = [""]; page = 0;
  $("record-list").replaceChildren(); $("month-summary").textContent = "";
  $("record-page").textContent = "";
  $("previous-records").disabled = true; $("next-records").disabled = true;
}
function context() {
  const group = $("history-group").value, month = $("history-month").value;
  if (!group || !/^\d{4}-\d{2}$/.test(month)) throw new Error("請選擇群組與月份。");
  return { group, month };
}
function renderRecords(data, query, targetPage) {
  const nodes = (data.items || []).map(record => {
    const row = document.createElement("li");
    const title = document.createElement("h3"); title.textContent = displayTime(record.ClockInAt);
    const detail = document.createElement("p"); detail.textContent = `下班：${displayTime(record.ClockOutAt)}${record.RestConfirmed ? "" : "；休息未確認"}`;
    const link = document.createElement("a");
    link.href = `/miniapp/correction?${new URLSearchParams({ ...query, record: record.RecordID })}`;
    link.textContent = "查看班次並提出更正";
    row.append(title, detail, link); return row;
  });
  $("record-list").replaceChildren(...nodes);
  nextCursor = data.nextCursor || ""; page = targetPage;
  $("record-page").textContent = `第 ${page + 1} 頁`;
  $("previous-records").disabled = page === 0; $("next-records").disabled = !nextCursor;
  $("history-status").textContent = nodes.length ? "顯示本人班次；月份摘要包含全月資料。" : "本頁沒有班次。";
}
function loadPage(targetPage = 0, withSummary = false) {
  let query;
  try { query = context(); } catch (error) { $("history-status").textContent = error.message; return; }
  const cursor = cursors[targetPage] || "";
  $("history-status").textContent = "正在讀取紀錄……";
  $("previous-records").disabled = true; $("next-records").disabled = true;
  const params = new URLSearchParams({ ...query, pageSize: "20", cursor });
  window.history.replaceState(null, "", `/miniapp/history?${new URLSearchParams(query)}`);
  void monthRead.run(async signal => {
    const [data, summary] = await Promise.all([
      api(`/api/v1/attendance/records?${params}`, { signal }),
      withSummary ? api(`/api/v1/attendance/monthly/summary?${new URLSearchParams(query)}`, { signal }) : null,
    ]);
    return { data, summary };
  }, ({ data, summary }) => {
    if (summary) $("month-summary").textContent = `全月已確認工作 ${summary.Calculation.Minutes} 分鐘；待確認 ${summary.Calculation.Pending} 筆；共 ${summary.count} 個班次。\n正常／加班分類待認定。`;
    renderRecords(data, query, targetPage);
  }, error => {
    $("history-status").textContent = errorMessage(error);
    $("previous-records").disabled = page === 0; $("next-records").disabled = !nextCursor;
  });
}
function loadGroups(append = false) {
  if (!append) {
    clearMonth(); groupCursor = ""; $("history-group").replaceChildren(new Option("請選擇群組", ""));
  }
  $("load-groups").disabled = true; $("more-groups").disabled = true;
  void groupRead.run(signal => api(`/api/v1/attendance/groups/page?${new URLSearchParams({ pageSize: "20", cursor: groupCursor })}`, { signal }), data => {
    for (const group of data.items || []) {
      const existing = [...$("history-group").options].find(option => option.value === group.id);
      if (existing) existing.textContent = group.name;
      else $("history-group").add(new Option(group.name, group.id));
    }
    if (!append && desiredGroup) {
      // Preserve a direct-link selection outside the first group page. The API still authorizes every read.
      if (![...$("history-group").options].some(option => option.value === desiredGroup)) $("history-group").add(new Option("目前選擇的群組", desiredGroup));
      $("history-group").value = desiredGroup;
    }
    groupCursor = data.nextCursor || "";
    $("more-groups").hidden = !groupCursor;
    $("load-groups").disabled = false; $("more-groups").disabled = false;
    $("history-status").textContent = $("history-group").options.length > 1 ? "請選擇群組與月份後查詢。" : "目前沒有出勤群組。";
    if (!append && desiredGroup) loadPage(0, true);
  }, error => {
    $("history-status").textContent = errorMessage(error);
    $("load-groups").disabled = false; $("more-groups").disabled = false;
  });
}
for (const id of ["history-group", "history-month"]) $(id).addEventListener("change", () => {
  desiredGroup = $("history-group").value;
  clearMonth(); $("history-status").textContent = "條件已變更，請重新查詢。";
});
$("load-groups").addEventListener("click", () => loadGroups());
$("more-groups").addEventListener("click", () => loadGroups(true));
$("load-month").addEventListener("click", () => { clearMonth(); loadPage(0, true); });
$("previous-records").addEventListener("click", () => { if (page > 0) loadPage(page - 1); });
$("next-records").addEventListener("click", () => { if (nextCursor) { cursors[page + 1] = nextCursor; loadPage(page + 1); } });
onIdentity(() => { $("history-controls").disabled = false; loadGroups(); }, () => {
  groupRead.cancel(); clearMonth(); groupCursor = "";
  $("history-group").replaceChildren(); $("more-groups").hidden = true;
  $("history-controls").disabled = true; $("load-groups").disabled = false; $("more-groups").disabled = false;
  $("history-status").textContent = "請先完成身分驗證。";
});
