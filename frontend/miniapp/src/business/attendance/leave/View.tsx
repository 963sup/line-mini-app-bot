import { attendanceLaunchToken } from "../launchToken";
import { Item, ItemGroup, ItemContent, ItemTitle, ItemDescription } from "@/shared/ui/item";
import { Badge } from "@/shared/ui/badge";
import { Empty, EmptyHeader, EmptyTitle } from "@/shared/ui/empty";
import { FormField } from "@/shared/FormField";
import { PagePanel } from "@/shared/PagePanel";
import { FieldGroup, FieldSet } from "@/shared/ui/field";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";
import { Input } from "@/shared/ui/input";
import { Textarea } from "@/shared/ui/textarea";
import { Button } from "@/shared/ui/button";
import { Status } from "@/shared/Status";
import { useEffect, useRef, useState, type FormEvent } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "../../../identity/Session";
import { useLatestRequest, errorMessage } from "../../../shared/http/lifecycle";
import { displayTime, timestamp } from "../../../shared/time/time";

type S = components["schemas"];
const kinds = { personal: "事假", sick: "病假", annual: "特休", other: "其他" };
export function View() {
  const { user, request } = useSession();
  const read = useLatestRequest(),
    write = useLatestRequest();
  const [launch, setLaunch] = useState("");
  const [kind, setKind] = useState<S["LeaveSubmission"]["kind"]>("personal");
  const [start, setStart] = useState(""),
    [end, setEnd] = useState("");
  const [reason, setReason] = useState("");
  const [busy, setBusy] = useState(false),
    [sent, setSent] = useState(false);
  const [uncertain, setUncertain] = useState(false);
  const [status, setStatus] = useState("請從原群組輸入「請假」取得表單連結。");
  const [records, setRecords] = useState<S["LeaveRecord"][]>([]);
  const [historyGroup, setHistoryGroup] = useState("");
  const [listStatus, setListStatus] = useState("");
  const [historyFailed, setHistoryFailed] = useState(false);
  const [historyLoaded, setHistoryLoaded] = useState(false);
  const lock = useRef(false),
    attempt = useRef<S["LeaveSubmission"] | null>(null);
  useEffect(() => {
    const token = attendanceLaunchToken(location.search, location.origin, "leave");
    setLaunch(token);
    if(token) setStatus("請填寫原群組的請假資料；送出只代表登記，不代表核准。");
  }, []);
  function loadLeaveHistory() {
    if (!user) return;
    setHistoryFailed(false);
    setHistoryLoaded(false);
    setListStatus("正在讀取本人最近20筆請假登記……");
    void read.run(
      (signal) =>
        request((c, s) => c.GET("/api/v1/attendance/leaves", { signal: s }), {
          signal,
        }),
      (data) => {
        setRecords(data);
        setHistoryLoaded(true);
        setListStatus(
          data.length
            ? "本人最近20筆請假登記（跨群組）。"
            : "目前沒有請假登記。",
        );
      },
      (error) => {setRecords([]);setHistoryFailed(true);setListStatus(errorMessage(error));},
    );
  }
  useEffect(() => {
    loadLeaveHistory();
    return read.cancel;
  }, [user, request]);
  function submitLeave(event: FormEvent) {
    event.preventDefault();
    if (!user || !launch || lock.current || sent) return;
    try {
      if (!uncertain || !attempt.current) {
        const from = timestamp(start),
          to = timestamp(end);
        attempt.current = {
          launch,
          requestKey: crypto.randomUUID(),
          kind,
          start: from,
          end: to,
          reason: reason.trim(),
        };
      }
      const body = attempt.current;
      lock.current = true;
      setBusy(true);
      setUncertain(true);
      setStatus("正在登記請假……");
      void write.run(
        (signal) =>
          request(
            (c, s) => c.POST("/api/v1/attendance/leaves", { signal: s, body }),
            { signal },
          ),
        (data) => {
          lock.current = false;
          setBusy(false);
          setSent(true);
          setUncertain(false);
          setStatus(
            `已登記，尚未核准。申請編號：${data.id}。請另行聯絡主管確認；理由未發送到群組。`,
          );
          loadLeaveHistory();
        },
        (error) => {
          lock.current = false;
          setBusy(false);
          const code = (error as { code?: string })?.code;
          if (code === "expired_launch" || code === "not_member") {
            setLaunch(""); setUncertain(false); attempt.current=null; setStatus(errorMessage(error)); return;
          }
          if (code === "invalid_input") {setUncertain(false);setStatus(errorMessage(error));return;}
          setStatus(
            errorMessage(error) +
              " 請查看本人紀錄確認結果；若仍未確認，可沿用原內容重試。",
          );
          loadLeaveHistory();
        },
      );
    } catch (error) {
      setStatus(errorMessage(error));
    }
  }
  function startAnotherLeave() {
    attempt.current = null;
    setSent(false);
    setUncertain(false);
    setReason("");
    setStatus("請填寫另一筆請假；送出不代表核准。");
  }
  const historyGroups = [...new Set(records.map(record => record.groupId))];
  const selectedGroup = historyGroups.includes(historyGroup) ? historyGroup : "";
  const visibleRecords = selectedGroup ? records.filter(record => record.groupId === selectedGroup) : records;

  return (
    <>
      <PagePanel heading="群組請假登記">
        <p>
          填寫原群組的請假資料，時間均為臺灣時間。第一版僅登記，不會自動核准、扣假或變更打卡紀錄。
        </p>
        {!launch && (
          <p>請回群組輸入「請假」，點選「填寫請假」。下方仍可查詢本人紀錄。</p>
        )}
        {launch && <form onSubmit={submitLeave}>
          <FieldGroup>
            <FieldSet
              className="grid min-w-0 gap-4 rounded-xl border p-4 disabled:opacity-60"
              disabled={!user || !launch || busy || sent || uncertain}
            >
              <FormField>
                假別
                <NativeSelect
                  value={kind}
                  onChange={(e) =>
                    setKind(e.target.value as S["LeaveSubmission"]["kind"])
                  }
                >
                  {Object.entries(kinds).map(([value, label]) => (
                    <NativeSelectOption key={value} value={value}>
                      {label}
                    </NativeSelectOption>
                  ))}
                </NativeSelect>
              </FormField>
              <FormField>
                開始時間
                <Input
                  type="datetime-local"
                  required
                  value={start}
                  onChange={(e) => setStart(e.target.value)}
                />
              </FormField>
              <FormField>
                結束時間
                <Input
                  type="datetime-local"
                  required
                  value={end}
                  onChange={(e) => setEnd(e.target.value)}
                />
              </FormField>
              <FormField>
                請假理由（不發送群組，最多500字）
                <Textarea
                  required
                  maxLength={500}
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                />
              </FormField>
            </FieldSet>
            <Button type="submit" disabled={!user || !launch || busy || sent}>
              {uncertain ? "重試原申請" : "送出請假登記"}
            </Button>
          </FieldGroup>
        </form>}
        {sent && <Button onClick={startAnotherLeave}>填寫另一筆</Button>}
        <Status role="status">
          {user ? status : "請先完成 LINE 身分驗證。"}
        </Status>
      </PagePanel>
      <PagePanel heading="本人請假紀錄">
        <Status role="status">{listStatus}</Status>
        {user && historyFailed && <Button type="button" variant="outline" onClick={loadLeaveHistory}>重試讀取</Button>}
        {user && historyGroups.length > 0 && (
          <FormField>
            群組
            <NativeSelect value={selectedGroup} onChange={event => setHistoryGroup(event.target.value)}>
              <NativeSelectOption value="">全部群組</NativeSelectOption>
              {historyGroups.map(group => <NativeSelectOption key={group} value={group}>{group}</NativeSelectOption>)}
            </NativeSelect>
          </FormField>
        )}
        <ItemGroup className="gap-3">
          {visibleRecords.map((leave) => (
            <Item key={leave.id} role="listitem" variant="outline">
              <ItemContent className="min-w-0">
                <ItemTitle>{kinds[leave.kind]}<Badge variant="outline">已登記、未核准</Badge></ItemTitle>
                <ItemDescription className="line-clamp-none break-words">{displayTime(leave.start)} ～ {displayTime(leave.end)}</ItemDescription>
                <ItemDescription className="line-clamp-none break-words">理由：{leave.reason}</ItemDescription>
                <ItemDescription className="line-clamp-none break-all">群組：{leave.groupId}</ItemDescription>
                <ItemDescription className="line-clamp-none break-all">申請編號：{leave.id}</ItemDescription>
              </ItemContent>
            </Item>
          ))}
        </ItemGroup>
        {historyLoaded && !records.length && <Empty><EmptyHeader><EmptyTitle>目前沒有請假登記</EmptyTitle></EmptyHeader></Empty>}
      </PagePanel>
    </>
  );
}
