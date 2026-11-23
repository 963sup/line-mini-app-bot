import { FieldSet } from "@/shared/ui/field";
import {
  Item,
  ItemGroup,
  ItemContent,
  ItemTitle,
  ItemDescription,
  ItemActions,
} from "@/shared/ui/item";
import { Badge } from "@/shared/ui/badge";
import { Empty, EmptyHeader, EmptyTitle } from "@/shared/ui/empty";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { PagePanel } from "@/shared/PagePanel";
import { Button } from "@/shared/ui/button";
import { Status } from "@/shared/Status";
import { useState, useCallback, useEffect } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "../../../identity/Session";
import { useLatestRequest, errorMessage } from "../../../shared/http/lifecycle";
import { displayTime } from "../../../shared/time/time";
type S = components["schemas"];
const statusLabels = {
  pending: "待審核",
  approved: "已核准",
  rejected: "已駁回",
  cancelled: "已撤回",
  stale: "原紀錄已變動，請重新申請",
  invalid: "核准證據異常，請聯絡管理者",
} satisfies Record<S["CorrectionLifecycleState"], string>;
export function correctionStatus(status: S["CorrectionReview"]["status"]) {
  return (status && statusLabels[status]) || "狀態無法確認";
}
export function Requests({
  review = false,
  refreshVersion = 0,
  disabled = false,
  onSelect,
  onClearReview,
}: {
  review?: boolean;
  refreshVersion?: number;
  disabled?: boolean;
  onClearReview?: () => void;
  onSelect?: (item: S["CorrectionReview"]) => void;
}) {
  const { user, request } = useSession(),
    { run: runRead, cancel: cancelRead } = useLatestRequest(),
    mutation = useLatestRequest();
  const [items, setItems] = useState<S["CorrectionReview"][]>([]),
    [next, setNext] = useState(""),
    [status, setStatus] = useState(""),
    [busy, setBusy] = useState(false),
    [loaded, setLoaded] = useState(false),
    [showHistory, setShowHistory] = useState(false),
    [failed, setFailed] = useState(false);
  const load = useCallback(
    (cursor = "") => {
      setBusy(true);
      setFailed(false);
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
                    cursor,
                    pageSize: 20,
                  },
                },
              }),
            { signal },
          ),
        (data) => {
          if (!Array.isArray(data.items)) throw new Error("申請清單回應無效，請重新查詢。");
          setItems((previous) => (cursor ? [...previous, ...data.items] : data.items));
          setLoaded(true);
          setNext(data.nextCursor);
          setBusy(false);
          setStatus("");
        },
        (error) => {
          setBusy(false);
          setItems([]);
          setFailed(true);
          setNext("");
          onClearReview?.();
          setStatus(
            review && (error as { httpStatus?: number }).httpStatus === 403
              ? "目前沒有補卡審核權限。請聯絡管理員確認審核群組設定。"
              : errorMessage(error),
          );
        },
      );
    },
    [review, request, runRead, onClearReview],
  );
  useEffect(() => {
    setItems([]);
    setFailed(false);
    setShowHistory(false);
    setLoaded(false);
    setNext("");
    if (user) load();
    else {
      setBusy(false);
      setStatus("");
    }
    return cancelRead;
  }, [user, refreshVersion, load, cancelRead]);
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
  const cutoff = Date.now() - 30 * 24 * 60 * 60 * 1000;
  const visible =
    review || showHistory
      ? items
      : items.filter(
          (item) =>
            item.status === "pending" ||
            !(Date.parse(item.reviewedAt || item.createdAt || "") < cutoff),
        );
  return (
    <PagePanel heading={review ? "待我審核" : "補卡申請進度"}>
      <FieldSet
        className="grid min-w-0 gap-4 rounded-xl border p-4 disabled:opacity-60"
        disabled={!user || busy || disabled}
      >
        {(review || failed) && (
          <Button id="requests-load" onClick={() => load()}>
            {review ? "重新整理待辦" : "重試讀取申請"}
          </Button>
        )}
        {!review && loaded && (next || visible.length !== items.length || showHistory) && (
          <Button variant="outline" onClick={() => setShowHistory((value) => !value)}>
            {showHistory ? "只看近期與待審核申請" : "查看較早的申請"}
          </Button>
        )}
        {busy && <ReadFeedback busy label="正在處理申請清單" />}
        <ItemGroup id="requests-list" className="gap-3" aria-busy={busy}>
          {visible.map((item) => (
            <Item key={item.id} role="listitem" variant="outline">
              <ItemContent className="min-w-0">
                <ItemTitle>
                  {review ? item.applicantName || "姓名未留存，請依申請編號核對" : "補卡／時間更正"}
                </ItemTitle>
                <ItemDescription className="line-clamp-none break-words">
                  {displayTime(item.start)} ～ {displayTime(item.end)}
                </ItemDescription>
                {item.groupId && (
                  <ItemDescription className="break-all">群組：{item.groupId}</ItemDescription>
                )}
                <Badge variant={item.status === "approved" ? "secondary" : "outline"}>
                  {correctionStatus(item.status)}
                </Badge>
                <ItemDescription className="line-clamp-none break-words">
                  {item.reason}
                </ItemDescription>
                {item.decisionReason && (
                  <ItemDescription className="line-clamp-none break-words">
                    審核理由：{item.decisionReason}
                  </ItemDescription>
                )}
              </ItemContent>
              <ItemActions>
                {review ? (
                  <Button onClick={() => onSelect?.(item)}>查看申請與審核</Button>
                ) : (
                  item.canCancel && (
                    <Button variant="outline" onClick={() => cancel(item.id)}>
                      撤回申請
                    </Button>
                  )
                )}
              </ItemActions>
            </Item>
          ))}
        </ItemGroup>
        {loaded && !busy && !visible.length && (
          <Empty>
            <EmptyHeader>
              <EmptyTitle>
                {next ? "本頁沒有符合的申請，可繼續查看" : "目前沒有符合的申請"}
              </EmptyTitle>
            </EmptyHeader>
          </Empty>
        )}
        {next && <Button onClick={() => load(next)}>載入更多申請</Button>}
      </FieldSet>
      <Status role="status">{status}</Status>
    </PagePanel>
  );
}
