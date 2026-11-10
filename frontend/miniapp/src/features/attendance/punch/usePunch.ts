import { useCallback, useEffect, useRef, useState } from "react";
import { useSession } from "../../../identity/Session";
import { errorMessage } from "../../../shared/http/lifecycle";
import { locate } from "../../../shared/location/position";
export const unknownResult =
  "打卡結果尚未確認；請先私訊查詢本人紀錄。重試仍使用同一打卡連結。";
export function launchToken(search: string, origin: string) {
  try {
    const params = new URLSearchParams(search),
      nested = params.get("liff.state"),
      token =
        params.get("launch") ||
        (nested ? new URL(nested, origin).searchParams.get("launch") : "");
    return /^[a-f0-9]{64}$/.test(token || "") ? token! : "";
  } catch {
    return "";
  }
}
export function usePunch() {
  const { user, request, close, autoPunch, punchNotice } = useSession(),
    [launch, setLaunch] = useState(""),
    [phase, setPhase] = useState("idle"),
    [status, setStatus] = useState(punchNotice.current);
  const active = useRef<AbortController | null>(null),
    generation = useRef(0),
    cached = useRef(""),
    currentPhase = useRef("idle");
  const cancel = useCallback(() => {
    generation.current++;
    active.current?.abort();
    active.current = null;
    currentPhase.current = "idle";
  }, []);
  const start = useCallback(async () => {
    if (active.current || !launch) return;
    if (!user) {
      setStatus("請先登入 LINE。");
      return;
    }
    const c = new AbortController(),
      id = ++generation.current;
    active.current = c;
    const change = (value: string, message: string) => {
      currentPhase.current = value;
      setPhase(value);
      setStatus(message);
    };
    try {
      change("exchanging", "正在確認群組打卡連結……");
      if (!cached.current) {
        const data = await request(
          (client, signal) =>
            client.POST("/api/v1/attendance/launch", {
              signal,
              body: { launch },
            }),
          { signal: c.signal, timeoutMs: 10000 },
        );
        if (!/^[a-f0-9]{64}$/.test(data.requestId))
          throw new Error("打卡連結回應無效。");
        cached.current = data.requestId;
      }
      if (id !== generation.current) return;
      change("locating", "正在取得目前位置；若瀏覽器詢問權限，請允許定位。");
      const position = await locate(c.signal);
      if (id !== generation.current) return;
      change("submitting", "已取得位置，正在保存打卡……");
      const data = await request(
        (client, signal) =>
          client.POST("/api/v1/attendance/punches", {
            signal,
            body: { requestId: cached.current, ...position },
          }),
        { signal: c.signal, timeoutMs: 10000 },
      );
      if (id !== generation.current) return;
      if (
        typeof data.code !== "string" ||
        typeof data.message !== "string" ||
        !data.message
      )
        throw new Error(unknownResult);
      setStatus(data.message);
      if (
        [
          "clocked_in",
          "clocked_out",
          "break_started",
          "break_ended",
          "rest_confirmed",
        ].includes(data.code)
      )
        close();
    } catch (error) {
      if (id === generation.current)
        setStatus(
          currentPhase.current === "submitting" &&
            (error instanceof TypeError ||
              (error instanceof Error && error.name === "AbortError"))
            ? unknownResult
            : errorMessage(error),
        );
    } finally {
      if (id === generation.current) {
        active.current = null;
        currentPhase.current = "idle";
        setPhase("idle");
      }
    }
  }, [user, request, launch, close]);
  useEffect(() => {
    setLaunch(launchToken(location.search, location.origin));
  }, []);
  useEffect(() => {
    const hide = () => {
      const message =
        currentPhase.current === "submitting"
          ? unknownResult
          : currentPhase.current !== "idle"
            ? "頁面切換使定位中斷；返回後請按打卡按鈕重試。"
            : "";
      punchNotice.current = message;
      cancel();
      cached.current = "";
      setPhase("idle");
      if (message) setStatus(message);
    };
    window.addEventListener("pagehide", hide);
    return () => {
      cancel();
      window.removeEventListener("pagehide", hide);
    };
  }, [cancel]);
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
  return { visible: !!launch, busy: phase !== "idle", status, start };
}
