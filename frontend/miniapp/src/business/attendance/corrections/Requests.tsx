import { FormField } from "@/shared/FormField";
import { PagePanel } from "@/shared/PagePanel";
import { Button } from "@/shared/ui/button";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";
import { Status } from "@/shared/Status";
import React, { useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "../../../identity/Session";
import { useLatestRequest, errorMessage } from "../../../shared/http/lifecycle";
import { displayTime } from "../../../shared/time/time";
type S = components["schemas"];
export function correctionStatus(status: string) {
  return (
    (
      {
        pending: "待審核",
        approved: "已核准",
        rejected: "已駁回",
        cancelled: "已撤回",
        stale: "原紀錄已變動，請重新申請",
        invalid: "核准證據異常，請聯絡管理者",
      } as Record<string, string>
    )[status] || "狀態無法確認"
  );
}
export function Requests({
  review = false,
  disabled = false,
  onSelect,
}: {
  review?: boolean;
  disabled?: boolean;
  onSelect?: (item: S["CorrectionReview"]) => void;
}) {
  const { user, request } = useSession(),
    read = useLatestRequest(),
    groupsRead = useLatestRequest(),
    mutation = useLatestRequest();
  const [items, setItems] = useState<S["CorrectionReview"][]>([]),
    [groups, setGroups] = useState<S["GroupItem"][]>([]),
    [group, setGroup] = useState(""),
    [next, setNext] = useState(""),
    [status, setStatus] = useState(""),
    [busy, setBusy] = useState(false);
  function load(cursor = "") {
    if (review && !group) {
      setStatus("請先載入並選擇審核群組。");
      return;
    }
    setBusy(true);
    setStatus("正在查詢……");
    void read.run(
      (signal) =>
        request(
          (c, s) =>
            c.GET("/api/v1/attendance/corrections", {
              signal: s,
              params: {
                query: {
                  scope: review ? "review" : "own",
                  ...(review ? { group } : {}),
                  cursor,
                  pageSize: 20,
                },
              },
            }),
          { signal },
        ),
      (data) => {
        setItems((previous) =>
          cursor ? [...previous, ...data.items] : data.items,
        );
        setNext(data.nextCursor);
        setBusy(false);
        setStatus(
          data.items.length
            ? "已載入申請；核准前有效紀錄不變。"
            : "本頁沒有申請。",
        );
      },
      (error) => {
        setBusy(false);
        setStatus(errorMessage(error));
      },
    );
  }
  function loadGroups() {
    void groupsRead.run(
      (signal) =>
        request(
          (c, s) =>
            c.GET("/api/v1/attendance/corrections/reviewer-groups", {
              signal: s,
            }),
          { signal },
        ),
      (data) => {
        setGroups(data);
        setStatus(data.length ? "請選擇審核群組。" : "尚未指定審核群組。");
      },
      (error) => setStatus(errorMessage(error)),
    );
  }
  function cancel(id: string) {
    if (busy) return;
    setBusy(true);
    void mutation.run(
      (signal) =>
        request(
          (c, s) =>
            c.POST("/api/v1/attendance/corrections/{id}/cancel", {
              signal: s,
              params: { path: { id } },
            }),
          { signal },
        ),
      () => load(),
      (error) => {
        setBusy(false);
        setStatus(errorMessage(error));
      },
    );
  }
  return (
    <PagePanel heading={review ? "群組申請清單" : "本人申請清單"}>
      <fieldset
        className="grid min-w-0 gap-4 rounded-xl border p-4 disabled:opacity-60"
        disabled={!user || busy || disabled}
      >
        {review && (
          <>
            <Button onClick={loadGroups}>載入審核群組</Button>
            <FormField>
              審核群組
              <NativeSelect
                value={group}
                onChange={(e) => {
                  read.cancel();
                  setGroup(e.target.value);
                  setItems([]);
                  setNext("");
                  setStatus("");
                }}
              >
                <NativeSelectOption value="">請選擇</NativeSelectOption>
                {groups.map((g) => (
                  <NativeSelectOption key={g.id} value={g.id}>
                    {g.name}
                  </NativeSelectOption>
                ))}
              </NativeSelect>
            </FormField>
          </>
        )}
        <Button id="requests-load" onClick={() => load()}>
          查詢{review ? "群組" : "本人"}申請
        </Button>
        <ol id="requests-list">
          {items.map((item) => (
            <li key={item.id}>
              <p>
                {review &&
                  `${item.applicantName || "姓名未留存，請依申請編號核對"}；`}
                {displayTime(item.start)} ～ {displayTime(item.end)}；
                {correctionStatus(item.status || "pending")}
              </p>
              <p>{item.reason}</p>
              {item.decisionReason && <p>審核理由：{item.decisionReason}</p>}
              {review ? (
                <Button onClick={() => onSelect?.(item)}>查看申請與審核</Button>
              ) : (
                (item.status === "pending" || item.status === "stale") && (
                  <Button onClick={() => cancel(item.id)}>撤回申請</Button>
                )
              )}
            </li>
          ))}
        </ol>
        {next && <Button onClick={() => load(next)}>載入更多申請</Button>}
      </fieldset>
      <Status role="status">{status}</Status>
    </PagePanel>
  );
}
