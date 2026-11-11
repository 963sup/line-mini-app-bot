import React from "react";
import { useEffect, useRef, useState, type FormEvent } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "../../../identity/Session";
import { useLatestRequest, errorMessage } from "../../../shared/http/lifecycle";
import { displayTime, timestamp } from "../../../shared/time/time";

type S = components["schemas"];
const kinds = { personal: "事假", sick: "病假", annual: "特休", other: "其他" };
export function View() {
  const { user, request } = useSession();
  const read = useLatestRequest(), write = useLatestRequest();
  const [launch, setLaunch] = useState("");
  const [kind, setKind] = useState<S["LeaveSubmission"]["kind"]>("personal");
  const [start, setStart] = useState(""), [end, setEnd] = useState("");
  const [reason, setReason] = useState("");
  const [busy, setBusy] = useState(false), [sent, setSent] = useState(false);
  const [uncertain, setUncertain] = useState(false);
  const [status, setStatus] = useState("請從原群組輸入「請假」取得表單連結。");
  const [records, setRecords] = useState<S["LeaveRecord"][]>([]);
  const [listStatus, setListStatus] = useState("");
  const lock = useRef(false), attempt = useRef<S["LeaveSubmission"] | null>(null);
  useEffect(() => {
    try {
      const params = new URLSearchParams(location.search);
      const nested = params.get("liff.state");
      const token = params.get("leave") || (nested ? new URL(nested, location.origin).searchParams.get("leave") : "");
      if (/^[a-f0-9]{64}$/.test(token || "")) {
        setLaunch(token!);
        setStatus("請填寫原群組的請假資料；送出只代表登記，不代表核准。");
      }
    } catch {}
  }, []);
  function loadLeaveHistory() {
    if (!user) return;
    setListStatus("正在讀取本人最近20筆請假登記……");
    void read.run(
      (signal) => request((c, s) => c.GET("/api/v1/attendance/leaves", { signal: s }), { signal }),
      (data) => { setRecords(data); setListStatus(data.length ? "本人最近20筆請假登記（跨群組）。" : "目前沒有請假登記。"); },
      (error) => setListStatus(errorMessage(error)),
    );
  }
  useEffect(() => { loadLeaveHistory(); return read.cancel; }, [user, request]);
  function submitLeave(event: FormEvent) {
    event.preventDefault();
    if (!user || !launch || lock.current || sent) return;
    try {
      if (!uncertain || !attempt.current) {
        const from = timestamp(start), to = timestamp(end);
        if (Date.parse(to) <= Date.parse(from) || Date.parse(to) - Date.parse(from) > 366 * 86400000)
          throw new Error("結束須晚於開始，請假期間最多366天。");
        if (!reason.trim()) throw new Error("請填寫請假理由。");
        attempt.current = { launch, requestKey: crypto.randomUUID(), kind, start: from, end: to, reason: reason.trim() };
      }
      const body = attempt.current;
      lock.current = true;
      setBusy(true);
      setUncertain(true);
      setStatus("正在登記請假……");
      void write.run(
        (signal) => request((c, s) => c.POST("/api/v1/attendance/leaves", { signal: s, body }), { signal }),
        (data) => {
          lock.current = false; setBusy(false); setSent(true); setUncertain(false);
          setStatus(`已登記，尚未核准。申請編號：${data.id}。請另行聯絡主管確認；理由未發送到群組。`);
          loadLeaveHistory();
        },
        (error) => {
          lock.current = false; setBusy(false);
          const code = (error as { code?: string })?.code;
          if (code === "invalid_input" || code === "not_member") setUncertain(false);
          setStatus(errorMessage(error) + " 請先重新載入本人紀錄確認結果；未確認時，重試會沿用原內容。重新開頁前也請先確認紀錄。");
        },
      );
    } catch (error) { setStatus(errorMessage(error)); }
  }
  function startAnotherLeave() {
    attempt.current = null; setSent(false); setUncertain(false); setReason("");
    setStatus("請填寫另一筆請假；送出不代表核准。");
  }
  return <>
    <section className="panel">
      <h2>群組請假登記</h2>
      <p>填寫原群組的請假資料，時間均為臺灣時間。第一版僅登記，不會自動核准、扣假或變更打卡紀錄。</p>
      {!launch && <p>請回群組輸入「請假」，點選「填寫請假」。下方仍可查詢本人紀錄。</p>}
      <form onSubmit={submitLeave}>
        <fieldset disabled={!user || !launch || busy || sent || uncertain}>
          <label>假別<select value={kind} onChange={(e) => setKind(e.target.value as S["LeaveSubmission"]["kind"])}>
            {Object.entries(kinds).map(([value, label]) => <option key={value} value={value}>{label}</option>)}
          </select></label>
          <label>開始時間<input type="datetime-local" required value={start} onChange={(e) => setStart(e.target.value)} /></label>
          <label>結束時間<input type="datetime-local" required value={end} onChange={(e) => setEnd(e.target.value)} /></label>
          <label>請假理由（不發送群組，最多500字）<textarea required maxLength={500} value={reason} onChange={(e) => setReason(e.target.value)} /></label>
        </fieldset>
        <button type="submit" disabled={!user || !launch || busy || sent}>{uncertain ? "重試原申請" : "送出請假登記"}</button>
      </form>
      {sent && <button onClick={startAnotherLeave}>填寫另一筆</button>}
      <p role="status">{user ? status : "請先完成 LINE 身分驗證。"}</p>
    </section>
    <section className="panel">
      <h2>本人請假紀錄</h2>
      <button disabled={!user} onClick={loadLeaveHistory}>重新載入本人紀錄</button>
      <p role="status">{listStatus}</p>
      <ul>{records.map((l) => <li key={l.id}>
        <strong>{kinds[l.kind]}：已登記、未核准</strong><br />
        {displayTime(l.start)} ～ {displayTime(l.end)}<br />
        理由：{l.reason}<br />群組：{l.groupId}<br />申請編號：{l.id}
      </li>)}</ul>
    </section>
  </>;
}
