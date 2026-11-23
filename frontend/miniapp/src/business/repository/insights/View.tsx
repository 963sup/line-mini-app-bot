import { ReadFeedback } from "@/shared/ReadFeedback";
import React, { useCallback, useEffect, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { useLatestRequest, errorMessage } from "@/shared/http/lifecycle";
import { PagePanel } from "@/shared/PagePanel";
import { displayTime } from "@/shared/time/time";
import { Card, CardHeader, CardDescription, CardContent } from "@/shared/ui/card";
import { Skeleton } from "@/shared/ui/skeleton";

const metrics = [
  ["openIssues", "待處理問題"],
  ["inProgressIssues", "進行中問題"],
  ["completedIssues", "已完成問題"],
  ["discussions", "討論"],
  ["milestones", "里程碑"],
  ["projects", "以本庫為初始來源的專案"],
] as const;
export function View({ repositoryId }: { repositoryId: string }) {
  const { user, request } = useSession(),
    { run, cancel } = useLatestRequest();
  const [data, setData] = useState<components["schemas"]["RepositoryStatistics"] | null>(null),
    [busy, setBusy] = useState(false),
    [status, setStatus] = useState("");
  const load = useCallback(() => {
    if (!user) return;
    setData(null);
    setBusy(true);
    setStatus("");
    void run(
      (signal) =>
        request(
          (c, s) =>
            c.GET("/api/v1/repositories/{repositoryId}/statistics", {
              signal: s,
              params: { path: { repositoryId } },
            }),
          { signal },
        ),
      (v) => {
        setData(v);
        setBusy(false);
        setStatus("");
      },
      (e) => {
        setData(null);
        setBusy(false);
        setStatus(errorMessage(e));
      },
    );
  }, [user, request, repositoryId, run]);
  useEffect(() => {
    load();
    return cancel;
  }, [load, cancel]);
  return (
    <PagePanel heading="現況總覽" description="查看這個儲存庫的問題狀態與協作項目數量。">
      <ReadFeedback
        busy={busy}
        message={status}
        label="正在讀取現況……"
        onRetry={user && status ? load : undefined}
      />
      {busy && (
        <div aria-hidden="true" className="grid min-w-0 grid-cols-2 gap-4 xl:grid-cols-3">
          {metrics.map(([key]) => (
            <Skeleton key={key} className="h-32 w-full" />
          ))}
        </div>
      )}
      {data && (
        <>
          <div
            role="group"
            aria-label="儲存庫統計"
            className="grid min-w-0 grid-cols-2 gap-4 xl:grid-cols-3"
          >
            {metrics.map(([key, label]) => (
              <Card key={key} className="min-w-0">
                <CardHeader>
                  <CardDescription>{label}</CardDescription>
                </CardHeader>
                <CardContent>
                  <p
                    aria-label={`${label}：${data[key]}`}
                    className="break-all text-2xl font-semibold tabular-nums sm:text-3xl"
                  >
                    {data[key].toLocaleString("zh-TW")}
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>
          <p className="text-sm text-muted-foreground">
            統計時間：<time dateTime={data.observedAt}>{displayTime(data.observedAt)}</time>
          </p>
        </>
      )}
    </PagePanel>
  );
}
