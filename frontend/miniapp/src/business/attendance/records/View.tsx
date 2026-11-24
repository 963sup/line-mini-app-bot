import { FieldSet } from "@/shared/ui/field";
import { ButtonGroup, ButtonGroupText } from "@/shared/ui/button_group";
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
  TableCaption,
} from "@/shared/ui/table";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from "@/shared/ui/empty";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { FormField } from "@/shared/FormField";
import { PagePanel } from "@/shared/PagePanel";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { useEffect, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "../../../identity/Session";
import { useLatestRequest, errorMessage } from "../../../shared/http/lifecycle";
import { taiwan, displayTime } from "../../../shared/time/time";
type S = components["schemas"];
export function View() {
  const { user, request } = useSession();
  const monthRead = useLatestRequest();
  const [month, setMonth] = useState("");
  const [records, setRecords] = useState<S["AttendanceRecord"][]>([]);
  const [page, setPage] = useState(0);
  const [cursors, setCursors] = useState([""]);
  const [next, setNext] = useState("");
  const [busy, setBusy] = useState(false);
  const [loaded, setLoaded] = useState(false);
  const [failure, setFailure] = useState<{ message: string; target: number; cursor: string }>();

  function clear() {
    monthRead.cancel();
    setRecords([]);
    setCursors([""]);
    setPage(0);
    setNext("");
    setBusy(false);
    setLoaded(false);
    setFailure(undefined);
  }

  function loadRecords(selectedMonth: string, target = 0, cursor = cursors[target] || "") {
    if (!user || !/^\d{4}-(0[1-9]|1[0-2])$/.test(selectedMonth)) return;
    setBusy(true);
    setFailure(undefined);
    const url = new URL(location.href);
    url.searchParams.delete("group");
    url.searchParams.set("month", selectedMonth);
    window.history.replaceState(window.history.state, "", url);
    void monthRead.run(
      (signal) =>
        request(
          (c, s) =>
            c.GET("/api/v1/attendance/records", {
              signal: s,
              params: { query: { month: selectedMonth, pageSize: 20, cursor } },
            }),
          { signal },
        ),
      (data) => {
        setRecords(data.items || []);
        setNext(data.nextCursor || "");
        setPage(target);
        setLoaded(true);
        setBusy(false);
      },
      (error) => {
        const httpStatus = (error as { httpStatus?: number }).httpStatus;
        if (httpStatus === 401 || httpStatus === 403 || httpStatus === 404) clear();
        setBusy(false);
        setFailure({ message: errorMessage(error), target, cursor });
      },
    );
  }

  useEffect(() => {
    const requestedMonth = new URLSearchParams(location.search).get("month") || "";
    const initialMonth = /^\d{4}-(0[1-9]|1[0-2])$/.test(requestedMonth)
      ? requestedMonth
      : taiwan(new Date().toISOString()).slice(0, 7);
    clear();
    setMonth(initialMonth);
    loadRecords(initialMonth, 0, "");
    return () => monthRead.cancel();
  }, [user, request]);

  return (
    <PagePanel id="history" heading="我的出勤紀錄">
      <p>
        <a
          className="break-words text-primary underline underline-offset-4 focus-visible:outline-2 focus-visible:outline-ring"
          href={
            "/miniapp/attendance?" +
            new URLSearchParams({ tab: "correction", ...(month ? { month } : {}) })
          }
        >
          申請補卡
        </a>
      </p>
      <FieldSet id="history-controls" disabled={!user}>
        <FormField>
          月份
          <Input
            id="history-month"
            type="month"
            required
            value={month}
            onChange={(e) => {
              const selectedMonth = e.target.value;
              clear();
              setMonth(selectedMonth);
              loadRecords(selectedMonth, 0, "");
            }}
          />
        </FormField>
      </FieldSet>
      {!user && <p role="status">請先完成身分驗證。</p>}
      {user && !month && <p role="status">請選擇月份。</p>}
      {user && (
        <ReadFeedback
          busy={busy}
          label="正在讀取出勤紀錄"
          message={failure && "無法載入出勤紀錄。" + failure.message}
          onRetry={failure && (() => loadRecords(month, failure.target, failure.cursor))}
        />
      )}
      {user && records.length > 0 && (
        <Table aria-busy={busy}>
          <TableCaption>本人班次明細；工時單位為分鐘</TableCaption>
          <TableHeader>
            <TableRow>
              <TableHead>群組</TableHead>
              <TableHead>上班</TableHead>
              <TableHead>下班</TableHead>
              <TableHead>工時與狀態</TableHead>
              <TableHead>操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody id="record-list">
            {records.map((record) => (
              <TableRow key={record.RecordID}>
                <TableCell>{record.GroupID}</TableCell>
                <TableCell>{displayTime(record.ClockInAt)}</TableCell>
                <TableCell>{displayTime(record.ClockOutAt)}</TableCell>
                <TableCell className="whitespace-normal min-w-48">
                  {record.WorktimeStatus === "invalid"
                    ? `；${(record.AttendanceMinutes ?? 0) > 0 ? `出勤 ${record.AttendanceMinutes} 分鐘；` : ""}資料未完整或有異常，暫無法計算工時`
                    : record.WorktimeStatus === "estimated"
                      ? `；出勤 ${record.AttendanceMinutes} 分鐘；暫估工時 ${record.EstimatedMinutes} 分鐘（休息待確認）`
                      : record.WorktimeStatus === "confirmed"
                        ? `；出勤 ${record.AttendanceMinutes} 分鐘；已確認工時 ${record.EstimatedMinutes} 分鐘`
                        : !record.RestConfirmed
                          ? "；休息待確認"
                          : ""}
                </TableCell>
                <TableCell>
                  <Button asChild variant="link">
                    <a
                      href={
                        "/miniapp/attendance?tab=correction&" +
                        new URLSearchParams({
                          group: record.GroupID,
                          month,
                          record: record.RecordID,
                        })
                      }
                    >
                      查看／補卡
                    </a>
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
      {user && loaded && !busy && !failure && !records.length && (
        <Empty>
          <EmptyHeader>
            <EmptyTitle>{page === 0 ? "本月沒有出勤紀錄" : "本頁沒有出勤紀錄"}</EmptyTitle>
            <EmptyDescription>可選擇其他月份，或在漏打卡時申請補卡。</EmptyDescription>
          </EmptyHeader>
        </Empty>
      )}
      {user && (page > 0 || next) && (
        <ButtonGroup aria-label="出勤紀錄分頁" className="mt-4">
          <Button
            id="previous-records"
            disabled={!user || busy || page === 0}
            onClick={() => loadRecords(month, page - 1)}
          >
            上一頁
          </Button>
          <ButtonGroupText id="record-page" aria-live="polite">
            {loaded && `第 ${page + 1} 頁`}
          </ButtonGroupText>
          <Button
            id="next-records"
            disabled={!user || busy || !next}
            onClick={() => {
              setCursors((old) => {
                const a = [...old];
                a[page + 1] = next;
                return a;
              });
              loadRecords(month, page + 1, next);
            }}
          >
            下一頁
          </Button>
        </ButtonGroup>
      )}
    </PagePanel>
  );
}
