import { memberGroups } from "../memberGroups";
import { FieldSet } from "@/shared/ui/field";
import { ButtonGroup, ButtonGroupText } from "@/shared/ui/button-group";
import { Badge } from "@/shared/ui/badge";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell, TableCaption } from "@/shared/ui/table";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from "@/shared/ui/empty";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { FormField } from "@/shared/FormField";
import { PagePanel } from "@/shared/PagePanel";
import { Button } from "@/shared/ui/button";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";
import { Input } from "@/shared/ui/input";
import { Status } from "@/shared/Status";
import React from "react";
import { useEffect, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "../../../identity/Session";
import { useLatestRequest, errorMessage } from "../../../shared/http/lifecycle";
import { taiwan, displayTime } from "../../../shared/time/time";
type S = components["schemas"];
export function View() {
  const { user, request } = useSession(),
    groupsRead = useLatestRequest(),
    monthRead = useLatestRequest();
  const [groups, setGroups] = useState<S["GroupItem"][]>([]),
    [group, setGroup] = useState(""),
    [month, setMonth] = useState(""),
    [groupCursor, setGroupCursor] = useState(""),
    [membershipGroups, setMembershipGroups] = useState(false),
    [groupsBusy, setGroupsBusy] = useState(false);
  const [records, setRecords] = useState<S["AttendanceRecord"][]>([]),
    [summary, setSummary] = useState<S["MonthlySummary"] | null>(null),
    [page, setPage] = useState(0),
    [cursors, setCursors] = useState([""]),
    [next, setNext] = useState(""),
    [busy, setBusy] = useState(false),
    [loaded, setLoaded] = useState(false),
    [status, setStatus] = useState("請先完成身分驗證。");
  function clear() {
    monthRead.cancel();
    setRecords([]);
    setSummary(null);
    setCursors([""]);
    setPage(0);
    setNext("");
    setBusy(false);
    setLoaded(false);
  }
  function loadRecords({ target = 0, selected = group, selectedMonth = month, cursor = cursors[target] || "", withSummary = false }:
    {target?:number; selected?:string; selectedMonth?:string; cursor?:string; withSummary?:boolean}) {
    if (!selected || !/^\d{4}-\d{2}$/.test(selectedMonth)) {
      setStatus("請選擇群組與月份。");
      return;
    }
    setBusy(true);
    setStatus("正在讀取紀錄……");
    const query = { group: selected, month: selectedMonth };
    window.history.replaceState(
      null,
      "",
      "/miniapp/history?" + new URLSearchParams(query),
    );
    void monthRead.run(
      async (signal) => {
        const [data, total] = await Promise.all([
          request(
            (c, s) =>
              c.GET("/api/v1/attendance/records", {
                signal: s,
                params: { query: { ...query, pageSize: 20, cursor } },
              }),
            { signal },
          ),
          withSummary
            ? request(
                (c, s) =>
                  c.GET("/api/v1/attendance/monthly/summary", {
                    signal: s,
                    params: { query },
                  }),
                { signal },
              )
            : null,
        ]);
        return { data, total };
      },
      ({ data, total }) => {
        setRecords(data.items || []);
        if (total) setSummary(total);
        setNext(data.nextCursor || "");
        setPage(target);
        setLoaded(true);
        setBusy(false);
        setStatus(
          data.items?.length
            ? "顯示本人班次；月份摘要包含全月資料。"
            : "本頁沒有班次。",
        );
      },
      (error) => {
        const httpStatus=(error as {httpStatus?:number}).httpStatus;
        if (withSummary || httpStatus === 401 || httpStatus === 403 || httpStatus === 404) clear();
        else setBusy(false);
        setStatus(errorMessage(error));
      },
    );
  }
  function loadGroups(append = false, desired = group, selectedMonth = month) {
    if (!append) {
      clear();
      setGroups([]);
      setGroupCursor("");
    }
    setGroupsBusy(true);
    void groupsRead.run(
      async (signal) => {
        const fromMembership = append && membershipGroups;
        const history = fromMembership ? null : await request(
          (c, s) =>
            c.GET("/api/v1/attendance/groups/page", {
              signal: s,
              params: {
                query: { pageSize: 20, cursor: append ? groupCursor : "" },
              },
            }),
          { signal },
        );
        if (history && (desired || append || history.items.length || history.nextCursor)) {
          return { ...history, fromMembership: false };
        }
        const available = await memberGroups(request, signal, fromMembership ? groupCursor : undefined);
        return {
          ...available,
          fromMembership: true,
        };
      },
      (data) => {
        setGroups((previous) => {
          const all = new Map((append ? previous : []).map((g) => [g.id, g]));
          for (const g of data.items || []) all.set(g.id, g);
          if (desired && !all.has(desired))
            all.set(desired, { id: desired, name: "目前選擇的群組" });
          return [...all.values()];
        });
        setGroupCursor(data.nextCursor || "");
        setMembershipGroups(data.fromMembership);
        setGroupsBusy(false);
        setStatus(!append && !desired && !data.items?.length
          ? "目前沒有出勤群組或可存取的 LINE 群組。請先在原 LINE 群組與機器人互動後重新載入。"
          : "請選擇群組與月份後查詢。");
        if (!append && desired) loadRecords({selected:desired,selectedMonth,withSummary:true,cursor:""});
      },
      (error) => {
        setGroupsBusy(false);
        setStatus(errorMessage(error));
      },
    );
  }
  useEffect(() => {
    const params = new URLSearchParams(location.search),
      g = params.get("group") || "",
      m = /^\d{4}-\d{2}$/.test(params.get("month") || "")
        ? params.get("month")!
        : taiwan(new Date().toISOString()).slice(0, 7);
    setGroup(g);
    setMonth(m);
    if (user) loadGroups(false, g, m);
    return () => {
      groupsRead.cancel();
      monthRead.cancel();
    };
  }, [user, request]);
  return (
    <PagePanel id="history" heading="本人月紀錄">
      <p>
        <a
          className="break-words text-primary underline underline-offset-4 focus-visible:outline-2 focus-visible:outline-ring"
          href={
            "/miniapp/correction?" + new URLSearchParams(group ? { group } : {})
          }
        >
          補卡申請（漏上班、漏下班或整班皆漏）
        </a>
      </p>
      <p className="text-sm leading-relaxed text-muted-foreground">
        整月摘要與下方分頁明細分開計算。資料可能更新；更正後請重新查詢。
      </p>
      <FieldSet
        className="grid min-w-0 gap-4 rounded-xl border p-4 disabled:opacity-60"
        id="history-controls"
        disabled={!user}
      >
        <div className="mt-4 flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center">
          <Button
            id="load-groups"
            disabled={groupsBusy}
            onClick={() => loadGroups()}
          >
            載入本人群組
          </Button>
          <Button
            id="more-groups"
            hidden={!groupCursor}
            disabled={groupsBusy}
            onClick={() => loadGroups(true)}
          >
            載入更多群組
          </Button>
        </div>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <FormField>
            群組
            <NativeSelect
              id="history-group"
              disabled={groupsBusy || !groups.length}
              aria-busy={groupsBusy}
              value={group}
              onChange={(e) => {
                clear();
                setGroup(e.target.value);
                setStatus("條件已變更，請重新查詢。");
              }}
            >
              <NativeSelectOption value="">{groupsBusy ? "正在載入群組……" : groups.length ? "請選擇群組" : "目前沒有可選群組"}</NativeSelectOption>
              {groups.map((g) => (
                <NativeSelectOption key={g.id} value={g.id}>
                  {g.name}
                </NativeSelectOption>
              ))}
            </NativeSelect>
          </FormField>
          <FormField>
            月份
            <Input
              id="history-month"
              type="month"
              required
              value={month}
              onChange={(e) => {
                clear();
                setMonth(e.target.value);
                setStatus("條件已變更，請重新查詢。");
              }}
            />
          </FormField>
        </div>
        <Button
          id="load-month"
          onClick={() => {
            clear();
            loadRecords({withSummary:true,cursor:""});
          }}
        >
          查詢月紀錄
        </Button>
      </FieldSet>
      <Status id="history-status" role="status">
        {status}
      </Status>
      <section id="month-summary" aria-label="全月工時摘要" hidden={!summary} className="my-4 space-y-3">
        {summary && <>
          <dl className="grid grid-cols-2 gap-3">
            <div><dt>全月暫估合計</dt><dd className="text-xl font-semibold">{summary.Calculation.EstimatedMinutes ?? summary.Calculation.Minutes} 分鐘</dd></div>
            <div><dt>其中已確認</dt><dd className="text-xl font-semibold">{summary.Calculation.Minutes} 分鐘</dd></div>
            <div><dt>出勤時數（未扣休息）</dt><dd>{summary.Calculation.AttendanceMinutes ?? "—"} 分鐘</dd></div>
            <div><dt>班次</dt><dd>{summary.count} 班</dd></div>
          </dl>
          <dl className="flex flex-wrap gap-3">
            {([["休息待確認", summary.Calculation.Unconfirmed], ["無法計算", summary.Calculation.Uncalculable],
              ["其他異常／提醒", summary.Calculation.OtherNoticeCount]] as const)
              .map(([label, count]) => <div key={label} className="flex items-center gap-1">
                <dt>{label}</dt><dd>{count ? <Badge variant="secondary">{count} 筆</Badge> : count ?? "—"}</dd>
              </div>)}
          </dl>
          <p className="text-sm text-muted-foreground">合計裁切本月區間；明細為完整班次。暫估非薪資結算，正常／加班分類待認定。</p>
        </>}
      </section>
      {busy && <ReadFeedback busy label="正在讀取出勤明細" />}
      <Table aria-busy={busy}>
        <TableCaption>本人班次明細；工時單位為分鐘</TableCaption>
        <TableHeader>
          <TableRow>
            <TableHead>上班</TableHead>
            <TableHead>下班</TableHead>
            <TableHead>工時與狀態</TableHead>
            <TableHead>操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody id="record-list">
          {records.map((record) => (
            <TableRow key={record.RecordID}>
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
                  <a href={"/miniapp/correction?" + new URLSearchParams({ group, month, record: record.RecordID })}>
                    查看班次並提出更正
                  </a>
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      {loaded && !busy && !records.length && (
        <Empty>
          <EmptyHeader>
            <EmptyTitle>本頁沒有班次</EmptyTitle>
            <EmptyDescription>可調整群組或月份後重新查詢。</EmptyDescription>
          </EmptyHeader>
        </Empty>
      )}
      <ButtonGroup aria-label="出勤紀錄分頁" className="mt-4">
        <Button
          id="previous-records"
          disabled={!user || busy || page === 0}
          onClick={() => loadRecords({target:page-1})}
        >
          上一頁
        </Button>
        <ButtonGroupText id="record-page" aria-live="polite">{loaded && `第 ${page + 1} 頁`}</ButtonGroupText>
        <Button
          id="next-records"
          disabled={!user || busy || !next}
          onClick={() => {
            setCursors((old) => {
              const a = [...old];
              a[page + 1] = next;
              return a;
            });
            loadRecords({target:page+1,cursor:next});
          }}
        >
          下一頁
        </Button>
      </ButtonGroup>
    </PagePanel>
  );
}
