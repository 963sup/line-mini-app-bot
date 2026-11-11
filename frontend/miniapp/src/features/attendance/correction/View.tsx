import React from "react";
import { useEffect, useRef, useState, type FormEvent } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "../../../identity/Session";
import { useLatestRequest, errorMessage } from "../../../shared/http/lifecycle";
import { taiwan, displayTime, timestamp } from "../../../shared/time/time";
type S = components["schemas"];
export function View() {
  const { user, request } = useSession(),
    recordRead = useLatestRequest(),
    statusRead = useLatestRequest(),
    submissionRequest = useLatestRequest();
  const [selected, setSelected] = useState<S["AttendanceRecord"] | null>(null),
    [start, setStart] = useState(""),
    [end, setEnd] = useState(""),
    [breaks, setBreaks] = useState<{ Start: string; End: string }[]>([]),
    [reason, setReason] = useState(""),
    [confirmed, setConfirmed] = useState(false),
    [busy, setBusy] = useState(false),
    [id, setID] = useState(""),
    [status, setStatus] = useState("請先完成身分驗證。"),
    [application, setApplication] = useState(""),
    [back, setBack] = useState("/miniapp/history");
  const lock = useRef(false),
    attempt = useRef<{ serialized: string; key: string } | null>(null);
  useEffect(() => {
    const params = new URLSearchParams(location.search),
      group = params.get("group") || "",
      record = params.get("record") || "",
      query = new URLSearchParams();
    if (group) query.set("group", group);
    if (params.get("month")) query.set("month", params.get("month")!);
    setBack("/miniapp/history?" + query);
    if (!user) return;
    if (!group || !validID(record)) {
      setStatus("請從本人月紀錄選擇班次；也可在下方查詢已送出的申請。");
      return;
    }
    setStatus("正在讀取班次……");
    void recordRead.run(
      (signal) =>
        request(
          (c, s) =>
            c.GET("/api/v1/attendance/records/{id}", {
              signal: s,
              params: { path: { id: record }, query: { group } },
            }),
          { signal },
        ),
      (data) => {
        setSelected(data);
        setStart(taiwan(data.ClockInAt));
        setEnd(taiwan(data.ClockOutAt));
        setBreaks(
          (data.Breaks || []).map((b) => ({
            Start: taiwan(b.Start),
            End: taiwan(b.End),
          })),
        );
        setStatus("請填寫實際工作與休息時間。");
      },
      (error) => setStatus(errorMessage(error)),
    );
    return recordRead.cancel;
  }, [user, request]);
  function submitCorrection(event: FormEvent) {
    event.preventDefault();
    if (lock.current || !selected || !user) return;
    try {
      const body: S["CorrectionSubmission"] = {
        groupId: selected.GroupID,
        recordId: selected.RecordID,
        expectedRevision: selected.Revision,
        expectedEventSequence: selected.EventSequence,
        reason,
        start: timestamp(start),
        end: timestamp(end),
        breaks: breaks.map((b) => ({
          Start: timestamp(b.Start),
          End: timestamp(b.End),
        })),
        restConfirmed: confirmed,
      };
      const serialized = JSON.stringify(body);
      if (attempt.current?.serialized !== serialized)
        attempt.current = { serialized, key: crypto.randomUUID() };
      const key = attempt.current.key;
      lock.current = true;
      setBusy(true);
      setStatus("正在送出申請……");
      void submissionRequest.run(
        (signal) =>
          request(
            (c, s) =>
              c.POST("/api/v1/attendance/corrections", {
                signal: s,
                body,
                headers: { "Idempotency-Key": key },
              }),
            { signal },
          ),
        (data) => {
          lock.current = false;
          setBusy(false);
          statusRead.cancel();
          setApplication("");
          setID(data.id);
          setStatus(
            `申請已受理。編號：${data.id}。請保存編號並提供給授權審核者，可在下方查詢目前結果。`,
          );
        },
        (error) => {
          lock.current = false;
          setBusy(false);
          setStatus(
            errorMessage(error) + " 結果未確認時，請保留本頁並以相同內容重試。",
          );
        },
      );
    } catch (error) {
      setStatus(errorMessage(error));
    }
  }
  function checkCorrectionStatus() {
    const correctionID = id.trim();
    if (!validID(correctionID)) {
      setApplication("請輸入有效的申請編號。");
      return;
    }
    setApplication("正在查詢申請……");
    void statusRead.run(
      (signal) =>
        request(
          (c, s) =>
            c.GET("/api/v1/attendance/corrections/{id}", {
              signal: s,
              params: { path: { id } },
            }),
          { signal },
        ),
      (data) =>
        setApplication(
          (
            {
              pending: "待審核",
              approved: "已核准，請重新查詢月紀錄。",
              stale: "原紀錄已變動，請重新載入班次再提出申請。",
              invalid: "核准證據異常，請聯絡管理者。",
            } as Record<string, string>
          )[data.status] || "申請狀態無法確認。",
        ),
      (error) => setApplication(errorMessage(error)),
    );
  }
  return (
    <>
      <section className="panel">
        <h2>提出更正</h2>
        <p>
          <a id="back-history" href={back}>
            回本人月紀錄選擇班次
          </a>
        </p>
        <p id="record-description">
          {selected
            ? `原班次：${displayTime(selected.ClockInAt)} → ${displayTime(selected.ClockOutAt)}。核准前工時不變。`
            : "請先從月紀錄選擇要更正的班次。"}
        </p>
        <form id="correction-form" onSubmit={submitCorrection}>
          <fieldset
            id="correction-fields"
            disabled={!user || !selected || busy}
          >
            <div className="field-grid">
              <label>
                上班時間
                <input
                  id="correction-start"
                  type="datetime-local"
                  step="1"
                  required
                  value={start}
                  onChange={(e) => setStart(e.target.value)}
                />
              </label>
              <label>
                下班時間
                <input
                  id="correction-end"
                  type="datetime-local"
                  step="1"
                  required
                  value={end}
                  onChange={(e) => setEnd(e.target.value)}
                />
              </label>
            </div>
            <fieldset>
              <legend>實際休息區間</legend>
              <div id="correction-breaks">
                {breaks.map((b, index) => (
                  <div key={index}>
                    {(["Start", "End"] as const).map((field) => (
                      <label key={field}>
                        {field === "Start" ? "休息開始" : "休息結束"}
                        <input
                          type="datetime-local"
                          step="1"
                          required
                          value={b[field]}
                          onChange={(e) =>
                            setBreaks((old) =>
                              old.map((v, i) =>
                                i === index
                                  ? { ...v, [field]: e.target.value }
                                  : v,
                              ),
                            )
                          }
                        />
                      </label>
                    ))}
                    <button
                      type="button"
                      onClick={() =>
                        setBreaks((old) => old.filter((_, i) => i !== index))
                      }
                    >
                      移除
                    </button>
                  </div>
                ))}
              </div>
              <button
                id="add-break"
                type="button"
                onClick={() =>
                  setBreaks((old) => [...old, { Start: "", End: "" }])
                }
              >
                新增休息區間
              </button>
            </fieldset>
            <label>
              更正原因
              <textarea
                id="correction-reason"
                required
                maxLength={500}
                value={reason}
                onChange={(e) => setReason(e.target.value)}
              />
            </label>
            <label className="check-row">
              <input
                id="correction-confirm"
                type="checkbox"
                required
                checked={confirmed}
                onChange={(e) => setConfirmed(e.target.checked)}
              />
              已確認實際工作與休息時間；沒有休息區間代表本班無可扣除的休息。
            </label>
            <button id="submit-correction" type="submit">
              送出更正申請
            </button>
          </fieldset>
        </form>
        <p id="correction-status" role="status">
          {status}
        </p>
      </section>
      <section className="panel">
        <h2>查看申請結果</h2>
        <p>輸入送出申請時取得的編號；核准前工時不變。</p>
        <fieldset id="status-controls" disabled={!user}>
          <label>
            申請編號
            <input
              id="correction-id"
              maxLength={128}
              autoComplete="off"
              value={id}
              onChange={(e) => {
                statusRead.cancel();
                setApplication("");
                setID(e.target.value);
              }}
            />
          </label>
          <button id="check-correction" onClick={checkCorrectionStatus}>
            查詢本人申請
          </button>
        </fieldset>
        <p id="application-status" role="status">
          {application}
        </p>
      </section>
    </>
  );
}
export function validID(id: string) {
  return /^[A-Za-z0-9_-]{1,128}$/.test(id);
}
