import React from "react";
import { useEffect, useRef, useState, type FormEvent } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "../../../identity/Session";
import { useLatestRequest, errorMessage } from "../../../shared/http/lifecycle";
import { taiwan, displayTime, timestamp } from "../../../shared/time/time";
import { Requests, correctionStatus } from "./Requests";
type S = components["schemas"];
export function View() {
  const { user, request } = useSession(),
    recordRead = useLatestRequest(),
    statusRead = useLatestRequest(),
    submissionRequest = useLatestRequest(),
    groupsRead = useLatestRequest();
  const [selected, setSelected] = useState<S["AttendanceRecord"] | null>(null),
    [group, setGroup] = useState(""),
    [newShift, setNewShift] = useState(false),
    [groups, setGroups] = useState<S["GroupItem"][]>([]),
    [groupCursor, setGroupCursor] = useState(""),
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
    setGroup(group);
    setNewShift(!record);
    if (!user) return;
    if (!record) {
      setStatus("補卡請填寫已結束班次的實際上下班與休息時間；一律須人工審核。");
      return;
    }
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
    if (lock.current || !user || (!selected && !newShift) || !group) return;
    try {
      const body: S["CorrectionSubmission"] = {
        groupId: selected?.GroupID || group,
        ...(selected
          ? {
              recordId: selected.RecordID,
              expectedRevision: selected.Revision,
              expectedEventSequence: selected.EventSequence,
            }
          : {}),
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
              params: { path: { id: correctionID } },
            }),
          { signal },
        ),
      (data) =>
        setApplication(
          correctionStatus(data.status) +
            (data.decisionReason ? `：${data.decisionReason}` : "") +
            (data.status === "approved" ? "，請重新查詢月紀錄。" : ""),
        ),
      (error) => setApplication(errorMessage(error)),
    );
  }
  function loadGroups(append = false) {
    void groupsRead.run(
      (signal) =>
        request(
          (c, s) =>
            c.GET("/api/v1/attendance/groups/page", {
              signal: s,
              params: {
                query: { pageSize: 20, cursor: append ? groupCursor : "" },
              },
            }),
          { signal },
        ),
      (data) => {
        setGroups((previous) =>
          append ? [...previous, ...data.items] : data.items,
        );
        setGroupCursor(data.nextCursor);
      },
      (error) => setStatus(errorMessage(error)),
    );
  }
  return (
    <>
      <section className="panel">
        <h2>補卡／時間更正申請</h2>
        <p>
          已結束班次可申請漏上班、漏下班、整班補卡或時間更正；核准前有效紀錄不變。仍在工作且漏上班者，請於班次結束後申報完整時段。
        </p>
        <p>
          <a id="back-history" href={back}>
            回本人月紀錄選擇班次
          </a>
        </p>
        {newShift && (
          <fieldset disabled={!user || busy}>
            <p>
              沒有可選班次時，可直接申請補卡。首次使用請回原 LINE
              群組輸入「補卡」開啟表單。
            </p>
            <button onClick={() => loadGroups()}>載入本人群組</button>
            <label>
              群組
              <select
                id="correction-group"
                value={group}
                onChange={(e) => setGroup(e.target.value)}
              >
                <option value="">請選擇群組</option>
                {group && !groups.some((g) => g.id === group) && (
                  <option value={group}>目前群組</option>
                )}
                {groups.map((g) => (
                  <option key={g.id} value={g.id}>
                    {g.name}
                  </option>
                ))}
              </select>
            </label>
            {groupCursor && (
              <button onClick={() => loadGroups(true)}>載入更多群組</button>
            )}
          </fieldset>
        )}
        <p id="record-description">
          {selected
            ? `原班次：${displayTime(selected.ClockInAt)} → ${displayTime(selected.ClockOutAt)}。核准前工時不變。`
            : newShift
              ? "申請新增缺失班次；請勿重複申報已有的班次。"
              : "請先從月紀錄選擇要更正的班次。"}
        </p>
        <form id="correction-form" onSubmit={submitCorrection}>
          <fieldset
            id="correction-fields"
            disabled={!user || (!selected && !newShift) || !group || busy}
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
              申請原因
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
              送出申請
            </button>
          </fieldset>
        </form>
        <p id="correction-status" role="status">
          {status}
        </p>
      </section>
      <Requests />
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
