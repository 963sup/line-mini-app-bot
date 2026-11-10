import { api, authenticated, onIdentity, latestRead, displayTime, errorMessage } from "/assets/app-api.js";

const $ = id => document.getElementById(id);
const reviewRead = latestRead();
let reviewed = "", generation = 0, approving = false;
function clearReview() {
  reviewRead.cancel(); reviewed = ""; $("review-details").textContent = ""; $("approve-correction").disabled = true;
}
function requestID() {
  const id = $("correction-id").value.trim();
  if (!/^[A-Za-z0-9_-]{1,128}$/.test(id)) throw new Error("請輸入有效的申請編號。");
  return id;
}
$("correction-id").addEventListener("input", () => { clearReview(); $("review-status").textContent = "請讀取目前申請的內容。"; });
$("review-correction").addEventListener("click", () => {
  clearReview();
  let id;
  try { id = requestID(); } catch (error) { $("review-status").textContent = error.message; return; }
  $("review-status").textContent = "正在讀取申請內容……";
  void reviewRead.run(signal => api(`/api/v1/attendance/corrections/${id}/review`, { signal }), proposal => {
    reviewed = id;
    $("review-details").textContent = `申請 ${id}\n原因：${proposal.reason}\n上班：${displayTime(proposal.start)}\n下班：${displayTime(proposal.end)}\n休息：${(proposal.breaks || []).map(rest => `${displayTime(rest.Start)} ～ ${displayTime(rest.End)}`).join("；") || "無"}\n休息完整確認：${proposal.restConfirmed ? "是" : "否"}`;
    $("approve-correction").disabled = false;
    $("review-status").textContent = "請確認以上內容後，再核准此申請。";
  }, error => { $("review-status").textContent = errorMessage(error); });
});
$("approve-correction").addEventListener("click", async () => {
  if (approving || !reviewed || reviewed !== $("correction-id").value.trim() || !authenticated()) return;
  const current = generation, id = reviewed;
  approving = true; $("review-controls").disabled = true;
  $("review-status").textContent = "正在核准申請……";
  try {
    await api(`/api/v1/attendance/corrections/${id}/approve`, { body: {} });
    if (current !== generation) return;
    clearReview(); $("review-status").textContent = "已核准。員工可查詢申請結果並重新載入月紀錄。";
  } catch (error) {
    if (current === generation) $("review-status").textContent = errorMessage(error);
  } finally {
    if (current === generation) { approving = false; $("review-controls").disabled = !authenticated(); }
  }
});
onIdentity(() => { $("review-controls").disabled = false; $("review-status").textContent = "請輸入員工提供的申請編號。"; }, () => {
  generation += 1; clearReview(); approving = false;
  $("review-controls").disabled = true; $("correction-id").value = "";
  $("review-status").textContent = "請先完成身分驗證。";
});
