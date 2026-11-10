import React from "react";
import { useRef, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "../../../identity/Session";
import { useLatestRequest, errorMessage } from "../../../shared/http/lifecycle";
import { displayTime } from "../../../shared/time/time";
export function View() {
  const { user, request } = useSession(),
    read = useLatestRequest(),
    approveRead = useLatestRequest(),
    lock = useRef(false);
  const [id, setID] = useState(""),
    [proposal, setProposal] = useState<
      components["schemas"]["CorrectionReview"] | null
    >(null),
    [reviewed, setReviewed] = useState(""),
    [busy, setBusy] = useState(false),
    [status, setStatus] = useState("請輸入員工提供的申請編號。");
  function clear() {
    read.cancel();
    setProposal(null);
    setReviewed("");
  }
  function load() {
    clear();
    const correctionID = id.trim();
    if (!/^[A-Za-z0-9_-]{1,128}$/.test(correctionID)) {
      setStatus("請輸入有效的申請編號。");
      return;
    }
    setStatus("正在讀取申請內容……");
    void read.run(
      (signal) =>
        request(
          (c, s) =>
            c.GET("/api/v1/attendance/corrections/{id}/review", {
              signal: s,
              params: { path: { id } },
            }),
          { signal },
        ),
      (data) => {
        setProposal(data);
        setReviewed(correctionID);
        setStatus("請確認以上內容後，再核准此申請。");
      },
      (error) => setStatus(errorMessage(error)),
    );
  }
  function approve() {
    if (lock.current || !user || !reviewed || reviewed !== id.trim()) return;
    lock.current = true;
    setBusy(true);
    setStatus("正在核准申請……");
    void approveRead.run(
      (signal) =>
        request(
          (c, s) =>
            c.POST("/api/v1/attendance/corrections/{id}/approve", {
              signal: s,
              params: { path: { id: reviewed } },
            }),
          { signal },
        ),
      () => {
        clear();
        lock.current = false;
        setBusy(false);
        setStatus("已核准。員工可查詢申請結果並重新載入月紀錄。");
      },
      (error) => {
        lock.current = false;
        setBusy(false);
        setStatus(errorMessage(error));
      },
    );
  }
  return (
    <section className="panel">
      <h2>授權審核者</h2>
      <p>
        取得員工提供的申請編號後，先讀取內容再核准。權限由伺服器確認，不能審核自己的申請。
      </p>
      <fieldset id="review-controls" disabled={!user || busy}>
        <label>
          申請編號
          <input
            id="correction-id"
            maxLength={128}
            autoComplete="off"
            value={id}
            onChange={(e) => {
              clear();
              setID(e.target.value);
              setStatus("請讀取目前申請的內容。");
            }}
          />
        </label>
        <button id="review-correction" onClick={load}>
          讀取審核內容
        </button>
        <pre id="review-details" className="summary">
          {proposal &&
            `申請 ${reviewed}\n原因：${proposal.reason}\n上班：${displayTime(proposal.start)}\n下班：${displayTime(proposal.end)}\n休息：${(proposal.breaks || []).map((b) => `${displayTime(b.Start)} ～ ${displayTime(b.End)}`).join("；") || "無"}\n休息完整確認：${proposal.restConfirmed ? "是" : "否"}`}
        </pre>
        <button
          id="approve-correction"
          disabled={!reviewed || reviewed !== id.trim()}
          onClick={approve}
        >
          核准已讀取的申請
        </button>
      </fieldset>
      <p id="review-status" role="status">
        {user ? status : "請先完成身分驗證。"}
      </p>
    </section>
  );
}
