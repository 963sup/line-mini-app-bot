import { FormField } from "@/shared/FormField";
import { PagePanel } from "@/shared/PagePanel";
import { Input } from "@/shared/ui/input";
import { Button } from "@/shared/ui/button";
import { Textarea } from "@/shared/ui/textarea";
import { Status } from "@/shared/Status";
import React from "react";
import { useRef, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "../../../identity/Session";
import { useLatestRequest, errorMessage } from "../../../shared/http/lifecycle";
import { displayTime } from "../../../shared/time/time";
import { Requests, correctionStatus } from "../corrections/Requests";
export function View() {
  const { user, request } = useSession(),
    read = useLatestRequest(),
    approvalRequest = useLatestRequest(),
    lock = useRef(false);
  const [id, setID] = useState(""),
    [proposal, setProposal] = useState<
      components["schemas"]["CorrectionReview"] | null
    >(null),
    [reviewed, setReviewed] = useState(""),
    [reason, setReason] = useState(""),
    [busy, setBusy] = useState(false),
    [status, setStatus] = useState("請輸入員工提供的申請編號。");
  function clearReview() {
    read.cancel();
    setProposal(null);
    setReviewed("");
    setReason("");
  }
  function loadCorrection() {
    clearReview();
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
              params: { path: { id: correctionID } },
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
  function approveCorrection(reject = false) {
    if (lock.current || !user || !reviewed || reviewed !== id.trim()) return;
    if (reject && !reason.trim()) {
      setStatus("請填寫駁回理由。");
      return;
    }
    lock.current = true;
    setBusy(true);
    setStatus(reject ? "正在駁回申請……" : "正在核准申請……");
    void approvalRequest.run(
      (signal) =>
        request(
          (c, s) =>
            reject
              ? c.POST("/api/v1/attendance/corrections/{id}/reject", {
                  signal: s,
                  params: { path: { id: reviewed } },
                  body: { reason },
                })
              : c.POST("/api/v1/attendance/corrections/{id}/approve", {
                  signal: s,
                  params: { path: { id: reviewed } },
                }),
          { signal },
        ),
      () => {
        clearReview();
        lock.current = false;
        setBusy(false);
        setStatus(
          reject
            ? "已駁回，原紀錄不變。"
            : "已核准。員工可查詢申請結果並重新載入月紀錄。",
        );
      },
      (error) => {
        lock.current = false;
        setBusy(false);
        setStatus(errorMessage(error));
      },
    );
  }
  return (
    <>
      <Requests
        review
        disabled={busy}
        onSelect={(item) => {
          clearReview();
          setID(item.id);
          setProposal(item);
          setReviewed(item.id);
          setStatus("請確認原資料與申報內容。");
        }}
      />
      <PagePanel heading="授權審核者">
        <p>
          取得員工提供的申請編號後，先讀取內容再核准。權限由伺服器確認，不能審核自己的申請。
        </p>
        <fieldset
          className="grid min-w-0 gap-4 rounded-xl border p-4 disabled:opacity-60"
          id="review-controls"
          disabled={!user || busy}
        >
          <FormField>
            申請編號
            <Input
              id="correction-id"
              maxLength={128}
              autoComplete="off"
              value={id}
              onChange={(e) => {
                clearReview();
                setID(e.target.value);
                setStatus("請讀取目前申請的內容。");
              }}
            />
          </FormField>
          <Button id="review-correction" onClick={loadCorrection}>
            讀取審核內容
          </Button>
          <pre
            id="review-details"
            className="my-4 max-w-full overflow-x-auto whitespace-pre-wrap break-words rounded-xl border bg-muted p-4 font-sans text-sm empty:hidden"
          >
            {proposal &&
              `申請 ${reviewed}\n申請人：${proposal.applicantName || "姓名未留存，請依申請編號核對"}\n狀態：${correctionStatus(proposal.status || "pending")}\n原紀錄：${proposal.before ? `${displayTime(proposal.before.ClockInAt)} ～ ${displayTime(proposal.before.ClockOutAt)}` : proposal.newRecord ? "無，申請新增補卡班次" : "請核對原班次"}\n審核理由：${proposal.decisionReason || "—"}\n原因：${proposal.reason}\n上班：${displayTime(proposal.start)}\n下班：${displayTime(proposal.end)}\n休息：${(proposal.breaks || []).map((b) => `${displayTime(b.Start)} ～ ${displayTime(b.End)}`).join("；") || "無"}\n休息完整確認：${proposal.restConfirmed ? "是" : "否"}`}
          </pre>
          <Button
            id="approve-correction"
            disabled={
              !reviewed ||
              reviewed !== id.trim() ||
              (!!proposal?.status && proposal.status !== "pending")
            }
            onClick={() => approveCorrection()}
          >
            核准已讀取的申請
          </Button>
          <FormField>
            駁回理由
            <Textarea
              id="rejection-reason"
              maxLength={500}
              value={reason}
              onChange={(e) => setReason(e.target.value)}
            />
          </FormField>
          <Button
            id="reject-correction"
            disabled={
              !reviewed ||
              reviewed !== id.trim() ||
              !reason.trim() ||
              (!!proposal?.status &&
                !["pending", "stale"].includes(proposal.status))
            }
            onClick={() => approveCorrection(true)}
          >
            駁回已讀取的申請
          </Button>
        </fieldset>
        <Status id="review-status" role="status">
          {user ? status : "請先完成身分驗證。"}
        </Status>
      </PagePanel>
    </>
  );
}
