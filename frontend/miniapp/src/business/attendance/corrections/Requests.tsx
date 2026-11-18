import { Item, ItemGroup, ItemContent, ItemTitle, ItemDescription, ItemActions } from "@/shared/ui/item";
import { Badge } from "@/shared/ui/badge";
import { Empty, EmptyHeader, EmptyTitle } from "@/shared/ui/empty";
import { Spinner } from "@/shared/ui/spinner";
import { FormField } from "@/shared/FormField";
import { PagePanel } from "@/shared/PagePanel";
import { Button } from "@/shared/ui/button";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";
import { Status } from "@/shared/Status";
import React, { useState, useCallback, useEffect } from "react";
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
  refreshVersion = 0,
  disabled = false,
  onSelect,
}: {
  review?: boolean;
  refreshVersion?: number;
  disabled?: boolean;
  onSelect?: (item: S["CorrectionReview"]) => void;
}) {
  const { user, request } = useSession(),
    { run: runRead, cancel: cancelRead } = useLatestRequest(),
    groupsRead = useLatestRequest(),
    mutation = useLatestRequest();
  const [items, setItems] = useState<S["CorrectionReview"][]>([]),
    [groups, setGroups] = useState<S["GroupItem"][]>([]),
    [group, setGroup] = useState(""),
    [next, setNext] = useState(""),
    [status, setStatus] = useState(""),
    [busy, setBusy] = useState(false),
    [loaded, setLoaded] = useState(false);
  const load = useCallback((cursor = "") => {
    if (review && !group) {
      setStatus("請先載入並選擇審核群組。");
      return;
    }
    setBusy(true);
    setLoaded(false);
    setStatus("正在查詢……");
    void runRead(
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
        setLoaded(true);
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
  }, [review, group, request, runRead]);
  useEffect(() => {
    if (!refreshVersion) return;
    setItems([]);
    setLoaded(false);
    setNext("");
    if (!review || group) load();
  }, [refreshVersion, review, group, load]);
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
                  cancelRead();
                  setGroup(e.target.value);
                  setLoaded(false);
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
        {busy && <p role="status" className="flex items-center gap-2"><Spinner aria-hidden="true" />正在處理申請清單</p>}
        <ItemGroup id="requests-list" className="gap-3" aria-busy={busy}>
          {items.map((item) => (
            <Item key={item.id} role="listitem" variant="outline">
              <ItemContent className="min-w-0">
                <ItemTitle>{review ? item.applicantName || "姓名未留存，請依申請編號核對" : "補卡／時間更正"}</ItemTitle>
                <ItemDescription className="line-clamp-none break-words">{displayTime(item.start)} ～ {displayTime(item.end)}</ItemDescription>
                <Badge variant={item.status === "approved" ? "secondary" : "outline"}>{correctionStatus(item.status || "pending")}</Badge>
                <ItemDescription className="line-clamp-none break-words">{item.reason}</ItemDescription>
                {item.decisionReason && <ItemDescription className="line-clamp-none break-words">審核理由：{item.decisionReason}</ItemDescription>}
              </ItemContent>
              <ItemActions>
                {review ? (
                  <Button onClick={() => onSelect?.(item)}>查看申請與審核</Button>
                ) : (
                  (item.status === "pending" || item.status === "stale") && (
                    <Button variant="outline" onClick={() => cancel(item.id)}>撤回申請</Button>
                  )
                )}
              </ItemActions>
            </Item>
          ))}
        </ItemGroup>
        {loaded && !busy && !items.length && <Empty><EmptyHeader><EmptyTitle>本頁沒有申請</EmptyTitle></EmptyHeader></Empty>}
        {next && <Button onClick={() => load(next)}>載入更多申請</Button>}
      </fieldset>
      <Status role="status">{status}</Status>
    </PagePanel>
  );
}
