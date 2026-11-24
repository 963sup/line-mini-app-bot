import { memberGroups } from "../memberGroups";
import { FormField } from "@/shared/FormField";
import { PagePanel } from "@/shared/PagePanel";
import { Button } from "@/shared/ui/button";
import { Checkbox } from "@/shared/ui/checkbox";
import { Field, FieldLabel, FieldGroup, FieldSet } from "@/shared/ui/field";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native_select";
import { Input } from "@/shared/ui/input";
import { Textarea } from "@/shared/ui/textarea";
import { Status } from "@/shared/Status";
import { useEffect, useRef, useState, type FormEvent } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "../../../identity/Session";
import { rejectedAttempt, useLatestRequest, errorMessage } from "../../../shared/http/lifecycle";
import { taiwan, displayTime, timestamp } from "../../../shared/time/time";
import { Requests } from "./Requests";
type S = components["schemas"];
export function View() {
  const { user, request } = useSession(),
    recordRead = useLatestRequest(),
    submissionRequest = useLatestRequest(),
    groupsRead = useLatestRequest();
  const [selected, setSelected] = useState<S["AttendanceRecord"] | null>(null),
    [group, setGroup] = useState(""),
    [newShift, setNewShift] = useState(false),
    [groups, setGroups] = useState<S["GroupItem"][]>([]),
    [groupCursor, setGroupCursor] = useState(""),
    [groupLoading, setGroupLoading] = useState(false),
    [groupError, setGroupError] = useState(""),
    [start, setStart] = useState(""),
    [end, setEnd] = useState(""),
    [editBothTimes, setEditBothTimes] = useState(false),
    [breaks, setBreaks] = useState<{ Start: string; End: string }[]>([]),
    [reason, setReason] = useState(""),
    [confirmed, setConfirmed] = useState(false),
    [uncertain, setUncertain] = useState(false),
    [busy, setBusy] = useState(false),
    [refreshVersion, setRefreshVersion] = useState(0),
    [status, setStatus] = useState("請先完成身分驗證。"),
    [back, setBack] = useState("/miniapp/attendance");
  const lock = useRef(false),
    attempt = useRef<{ body: S["CorrectionSubmission"]; key: string } | null>(null);
  useEffect(() => {
    const params = new URLSearchParams(location.search),
      group = params.get("group") || "",
      record = params.get("record") || "",
      query = new URLSearchParams();
    if (params.get("month")) query.set("month", params.get("month")!);
    setBack("/miniapp/attendance" + (query.size ? "?" + query : ""));
    setGroup(group);
    setGroups([]);
    setGroupCursor("");
    setNewShift(!record);
    setSelected(null);
    setStart("");
    setEnd("");
    setEditBothTimes(false);
    setBreaks([]);
    setReason("");
    setConfirmed(false);
    setGroupLoading(false);
    setGroupError("");
    setStatus(user ? "" : "請先完成身分驗證。");
    if (!user) return;
    if (!record) {
      if (!group) loadGroups();
      return groupsRead.cancel;
    }
    if (!group || !validID(record)) {
      setStatus("請從出勤紀錄選擇要更正的班次。");
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
        setStatus("");
      },
      (error) => setStatus(errorMessage(error)),
    );
    return recordRead.cancel;
  }, [user, request]);
  function submitCorrection(event: FormEvent) {
    event.preventDefault();
    if (lock.current || !user || (!selected && !newShift) || !group) return;
    try {
      const body: S["CorrectionSubmission"] =
        uncertain && attempt.current
          ? attempt.current.body
          : {
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
      const retry = uncertain;
      if (
        !attempt.current ||
        (!retry && JSON.stringify(attempt.current.body) !== JSON.stringify(body))
      )
        attempt.current = { body, key: crypto.randomUUID() };
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
          setUncertain(false);
          setRefreshVersion((value) => value + 1);
          setStatus(`已送出，等待審核。申請編號：${data.id}。`);
        },
        (error) => {
          lock.current = false;
          setBusy(false);
          const unknown =
            retry ||
            !rejectedAttempt(error, [
              "invalid_correction",
              "invalid_idempotency_key",
              "not_member",
            ]);
          if (!unknown) attempt.current = null;
          setUncertain(unknown);
          setStatus(errorMessage(error) + " 結果未確認時，請保留本頁並以相同內容重試。");
        },
      );
    } catch (error) {
      setStatus(errorMessage(error));
    }
  }
  function loadGroups(append = false) {
    setGroupLoading(true);
    setGroupError("");
    void groupsRead.run(
      (signal) => memberGroups(request, signal, append ? groupCursor : undefined),
      (data) => {
        setGroupLoading(false);
        const options = data.items;
        setGroups((previous) => (append ? [...previous, ...options] : options));
        if (!append && !options.length && !data.nextCursor && !group) {
          setStatus("目前沒有可選群組。請在原 LINE 群組輸入「補卡」開啟表單。");
        }
        if (!append && options.length === 1 && !data.nextCursor) setGroup(options[0].id);
        setGroupCursor(data.nextCursor);
      },
      (error) => {
        setGroupLoading(false);
        setGroupError(errorMessage(error));
      },
    );
  }
  const originalStart = selected ? taiwan(selected.ClockInAt) : "";
  const originalEnd = selected ? taiwan(selected.ClockOutAt) : "";
  const oneTimeMissing = !!selected && !!originalStart !== !!originalEnd;
  const needsGroupPicker = newShift && (!group || groups.length > 1 || !!groupCursor);
  return (
    <>
      <PagePanel heading={newShift ? "申請補卡" : "更正出勤紀錄"}>
        <p>班次結束後申請，核准後生效。</p>
        <p>
          <a
            className="break-words text-primary underline underline-offset-4 focus-visible:outline-2 focus-visible:outline-ring"
            id="back-history"
            href={back}
          >
            查看出勤紀錄
          </a>
        </p>
        {needsGroupPicker && (
          <FieldSet
            className="grid min-w-0 gap-4 rounded-xl border p-4 disabled:opacity-60"
            disabled={!user || busy || uncertain}
          >
            {groupLoading && <p role="status">正在讀取群組……</p>}
            {groupError && (
              <div role="alert">
                <p>{groupError}</p>
                <Button onClick={() => loadGroups()}>重試讀取群組</Button>
              </div>
            )}
            {(groups.length > 0 || !!groupCursor) && (
              <FormField>
                群組
                <NativeSelect
                  id="correction-group"
                  value={group}
                  onChange={(e) => setGroup(e.target.value)}
                >
                  <NativeSelectOption value="">請選擇群組</NativeSelectOption>
                  {groups.map((g) => (
                    <NativeSelectOption key={g.id} value={g.id}>
                      {g.name}
                    </NativeSelectOption>
                  ))}
                </NativeSelect>
              </FormField>
            )}
            {groupCursor && <Button onClick={() => loadGroups(true)}>載入更多群組</Button>}
          </FieldSet>
        )}
        {newShift && group && !needsGroupPicker && <p>群組：{groups[0]?.name || group}</p>}
        {selected && (
          <p id="record-description">
            原班次：{originalStart ? displayTime(selected.ClockInAt) : "缺上班時間"} →{" "}
            {originalEnd ? displayTime(selected.ClockOutAt) : "缺下班時間"}
          </p>
        )}
        <form id="correction-form" onSubmit={submitCorrection}>
          <FieldGroup>
            <FieldSet
              className="grid min-w-0 gap-4 rounded-xl border p-4 disabled:opacity-60"
              id="correction-fields"
              disabled={!user || (!selected && !newShift) || !group || busy || uncertain}
            >
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                {(!oneTimeMissing || editBothTimes || !originalStart) && (
                  <FormField>
                    上班時間
                    <Input
                      id="correction-start"
                      type="datetime-local"
                      step="1"
                      required
                      value={start}
                      onChange={(e) => setStart(e.target.value)}
                    />
                  </FormField>
                )}
                {(!oneTimeMissing || editBothTimes || !originalEnd) && (
                  <FormField>
                    下班時間
                    <Input
                      id="correction-end"
                      type="datetime-local"
                      step="1"
                      required
                      value={end}
                      onChange={(e) => setEnd(e.target.value)}
                    />
                  </FormField>
                )}
              </div>
              {oneTimeMissing && !editBothTimes && (
                <Button type="button" variant="outline" onClick={() => setEditBothTimes(true)}>
                  {originalStart ? "也要更正上班時間" : "也要更正下班時間"}
                </Button>
              )}
              <details
                key={selected?.RecordID || "new-shift"}
                className="space-y-3"
                onInvalidCapture={(event) => {
                  event.currentTarget.open = true;
                }}
              >
                <summary className="cursor-pointer">
                  休息時間（{breaks.length ? `${breaks.length} 段` : "未填寫"}）
                </summary>
                <div id="correction-breaks">
                  {breaks.map((b, index) => (
                    <div key={index}>
                      {(["Start", "End"] as const).map((field) => (
                        <FormField key={field}>
                          {field === "Start" ? "休息開始" : "休息結束"}
                          <Input
                            type="datetime-local"
                            step="1"
                            required
                            value={b[field]}
                            onChange={(e) =>
                              setBreaks((old) =>
                                old.map((v, i) =>
                                  i === index ? { ...v, [field]: e.target.value } : v,
                                ),
                              )
                            }
                          />
                        </FormField>
                      ))}
                      <Button
                        type="button"
                        onClick={() => setBreaks((old) => old.filter((_, i) => i !== index))}
                      >
                        移除
                      </Button>
                    </div>
                  ))}
                </div>
                <Button
                  id="add-break"
                  type="button"
                  onClick={() => setBreaks((old) => [...old, { Start: "", End: "" }])}
                >
                  新增休息區間
                </Button>
              </details>
              <FormField>
                申請原因
                <Textarea
                  id="correction-reason"
                  required
                  maxLength={500}
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                />
              </FormField>
              <Field orientation="horizontal">
                <Checkbox
                  id="correction-confirm"
                  required
                  checked={confirmed}
                  onCheckedChange={(checked) => setConfirmed(checked === true)}
                />
                <FieldLabel htmlFor="correction-confirm">
                  我已確認上下班及休息時間。未填休息時間表示沒有休息。
                </FieldLabel>
              </Field>
              <Button id="submit-correction" type="submit">
                送出申請
              </Button>
            </FieldSet>
            {uncertain && (
              <Button type="submit" disabled={busy}>
                重試原申請
              </Button>
            )}
          </FieldGroup>
        </form>
        <Status id="correction-status" role="status">
          {status}
        </Status>
      </PagePanel>
      <Requests refreshVersion={refreshVersion} />
    </>
  );
}
export function validID(id: string) {
  return /^[A-Za-z0-9_-]{1,128}$/.test(id);
}
