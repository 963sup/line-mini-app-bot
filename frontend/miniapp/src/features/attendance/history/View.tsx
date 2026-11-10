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
  function loadPage(
    target = 0,
    withSummary = false,
    selected = group,
    selectedMonth = month,
    cursor = cursors[target] || "",
  ) {
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
        setBusy(false);
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
        setGroups((previous) => {
          const all = new Map((append ? previous : []).map((g) => [g.id, g]));
          for (const g of data.items || []) all.set(g.id, g);
          if (desired && !all.has(desired))
            all.set(desired, { id: desired, name: "目前選擇的群組" });
          return [...all.values()];
        });
        setGroupCursor(data.nextCursor || "");
        setGroupsBusy(false);
        setStatus("請選擇群組與月份後查詢。");
        if (!append && desired) loadPage(0, true, desired, selectedMonth, "");
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
    <section className="panel" id="history">
      <h2>本人月紀錄</h2>
      <p className="hint">
        整月摘要與下方分頁明細分開計算。資料可能更新；更正後請重新查詢。
      </p>
      <fieldset id="history-controls" disabled={!user}>
        <div className="actions">
          <button
            id="load-groups"
            disabled={groupsBusy}
            onClick={() => loadGroups()}
          >
            載入本人群組
          </button>
          <button
            id="more-groups"
            hidden={!groupCursor}
            disabled={groupsBusy}
            onClick={() => loadGroups(true)}
          >
            載入更多群組
          </button>
        </div>
        <div className="field-grid">
          <label>
            群組
            <select
              id="history-group"
              value={group}
              onChange={(e) => {
                clear();
                setGroup(e.target.value);
                setStatus("條件已變更，請重新查詢。");
              }}
            >
              <option value="">請選擇群組</option>
              {groups.map((g) => (
                <option key={g.id} value={g.id}>
                  {g.name}
                </option>
              ))}
            </select>
          </label>
          <label>
            月份
            <input
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
          </label>
        </div>
        <button
          id="load-month"
          onClick={() => {
            clear();
            loadPage(0, true, group, month, "");
          }}
        >
          查詢月紀錄
        </button>
      </fieldset>
      <p id="history-status" role="status">
        {status}
      </p>
      <pre id="month-summary" className="summary">
        {summary &&
          `全月暫估合計 ${summary.Calculation.EstimatedMinutes ?? summary.Calculation.Minutes} 分鐘（含已確認 ${summary.Calculation.Minutes} 分鐘）。\n出勤時數 ${summary.Calculation.AttendanceMinutes ?? "—"} 分鐘（未扣休息）；休息待確認 ${summary.Calculation.Unconfirmed ?? "—"} 筆；無法計算 ${summary.Calculation.Uncalculable ?? "—"} 筆；共 ${summary.count} 個班次；其他異常／提醒 ${summary.Calculation.Issues?.filter((issue) => issue.Code !== "rest_unconfirmed").length ?? 0} 項。\n合計裁切本月區間；明細為完整班次。暫估非薪資結算，正常／加班分類待認定。`}
      </pre>
      <ol id="record-list" className="record-list">
        {records.map((record) => (
          <li key={record.RecordID}>
            <h3>{displayTime(record.ClockInAt)}</h3>
            <p>
              下班：{displayTime(record.ClockOutAt)}
              {record.WorktimeStatus === "invalid" ? `；${(record.AttendanceMinutes ?? 0) > 0 ? `出勤 ${record.AttendanceMinutes} 分鐘；` : ""}資料未完整或有異常，暫無法計算工時` :
               record.WorktimeStatus === "estimated" ? `；出勤 ${record.AttendanceMinutes} 分鐘；暫估工時 ${record.EstimatedMinutes} 分鐘（休息待確認）` :
               record.WorktimeStatus === "confirmed" ? `；出勤 ${record.AttendanceMinutes} 分鐘；已確認工時 ${record.EstimatedMinutes} 分鐘` :
               !record.RestConfirmed ? "；休息待確認" : ""}
            </p>
            <a
              href={
                "/miniapp/correction?" +
                new URLSearchParams({ group, month, record: record.RecordID })
              }
            >
              查看班次並提出更正
            </a>
          </li>
        ))}
      </ol>
      <div className="actions">
        <button
          id="previous-records"
          disabled={!user || busy || page === 0}
          onClick={() => loadPage(page - 1)}
        >
          上一頁
        </button>
        <span id="record-page">{loaded && `第 ${page + 1} 頁`}</span>
        <button
          id="next-records"
          disabled={!user || busy || !next}
          onClick={() => {
            setCursors((old) => {
              const a = [...old];
              a[page + 1] = next;
              return a;
            });
            loadPage(page + 1, false, group, month, next);
          }}
        >
          下一頁
        </button>
      </div>
    </section>
  );
}
