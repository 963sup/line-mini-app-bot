import { attendanceLaunchToken } from "../launchToken";
import { PagePanel } from "@/shared/PagePanel";
import { Spinner } from "@/shared/ui/spinner";
import { Button } from "@/shared/ui/button";
import { Status } from "@/shared/Status";
import React, { useCallback, useEffect, useRef, useState } from "react";
import { useSession } from "../../../identity/Session";
import { errorMessage } from "../../../shared/http/lifecycle";
import { locate } from "../../../shared/location/position";

const unknownResult =
  "打卡結果尚未確認；請先私訊查詢本人紀錄。重試仍使用同一打卡連結。";

export function View() {
  const { user, request, close, autoPunch, punchNotice } = useSession(),
    [launch, setLaunch] = useState(""),
    [busy, setBusy] = useState(false),
    [status, setStatus] = useState(punchNotice.current);
  const active = useRef<{ controller: AbortController; submitting: boolean } | null>(null);
  const cached = useRef("");
  const start = useCallback(async () => {
    if (active.current || !launch) return;
    if (!user) {
      setStatus("請先登入 LINE。");
      return;
    }
    const attempt = { controller: new AbortController(), submitting: false };
    const { signal } = attempt.controller;
    active.current = attempt;
    setBusy(true);
    try {
      setStatus("正在確認群組打卡連結……");
      if (!cached.current) {
        const data = await request(
          (client, signal) =>
            client.POST("/api/v1/attendance/launch", {
              signal,
              body: { launch },
            }),
          { signal, timeoutMs: 10000 },
        );
        if (signal.aborted) return;
        if (!/^[a-f0-9]{64}$/.test(data.requestId))
          throw new Error("打卡連結回應無效。");
        cached.current = data.requestId;
      }
      setStatus("正在確認這次操作是否需要打卡……");
      const preflight = await request(
        (client, signal) =>
          client.POST("/api/v1/attendance/punches/preflight", {
            signal,
            body: { requestId: cached.current },
          }),
        { signal, timeoutMs: 10000 },
      );
      if (signal.aborted) return;
      if (
        typeof preflight.code !== "string" ||
        !preflight.code ||
        typeof preflight.message !== "string" ||
        !preflight.message
      )
        throw new Error("打卡狀態回應無效，請重試。");
      if (preflight.code !== "needs_position") {
        setStatus(preflight.message);
        if (preflight.accepted === true) close();
        return;
      }
      setStatus("正在取得目前位置；若瀏覽器詢問權限，請允許定位。");
      const position = await locate(signal);
      if (signal.aborted) return;
      attempt.submitting = true;
      setStatus("已取得位置，正在保存打卡……");
      const data = await request(
        (client, signal) =>
          client.POST("/api/v1/attendance/punches", {
            signal,
            body: { requestId: cached.current, ...position },
          }),
        { signal, timeoutMs: 10000 },
      );
      if (signal.aborted) return;
      if (
        typeof data.code !== "string" ||
        typeof data.message !== "string" ||
        !data.message
      )
        throw new Error(unknownResult);
      setStatus(data.message);
      if (data.accepted === true) close();
    } catch (error) {
      if (!signal.aborted)
        setStatus(
          attempt.submitting &&
            (error instanceof TypeError ||
              (error instanceof Error && error.name === "AbortError"))
            ? unknownResult
            : errorMessage(error),
        );
    } finally {
      if (!signal.aborted) {
        active.current = null;
        setBusy(false);
      }
    }
  }, [user, request, launch, close]);
  useEffect(() => {
    setLaunch(attendanceLaunchToken(location.search, location.origin, "launch"));
  }, []);
  useEffect(() => {
    const cancel = () => {
      active.current?.controller.abort();
      active.current = null;
    };
    const hide = () => {
      const message =
        active.current?.submitting
          ? unknownResult
          : active.current
            ? "頁面切換使操作中斷；返回後請按打卡按鈕重新確認。"
            : "";
      punchNotice.current = message;
      cancel();
      cached.current = "";
      setBusy(false);
      if (message) setStatus(message);
    };
    window.addEventListener("pagehide", hide);
    return () => {
      cancel();
      window.removeEventListener("pagehide", hide);
    };
  }, [punchNotice]);
  // Defer intent until the effect survives Strict Mode's setup/cleanup probe.
  useEffect(() => {
    let stopped = false;
    queueMicrotask(() => {
      if (!stopped && user && launch && !autoPunch.current) {
        autoPunch.current = true;
        void start();
      }
    });
    return () => {
      stopped = true;
    };
  }, [user, launch, start, autoPunch]);
  return (
    <PagePanel
      id="attendance"
      className="bg-secondary"
      hidden={!launch}
      heading="確認並完成打卡"
    >
      <p>
        先確認打卡狀態；需要新增紀錄時才取得位置，保存本次位置、精度與採樣時間。
      </p>
      <Button
        id="punch"
        disabled={busy}
        aria-busy={busy}
        onClick={() => void start()}
      >
        {busy && <Spinner aria-label="處理中" />}
        確認並完成打卡
      </Button>
      <Status id="punch-status" role="status" aria-live="polite">
        {status}
      </Status>
    </PagePanel>
  );
}
