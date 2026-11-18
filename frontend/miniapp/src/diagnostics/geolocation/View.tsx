import { Spinner } from "@/shared/ui/spinner";
import { Button } from "@/shared/ui/button";
import { Status } from "@/shared/Status";
import React from "react";
import { useEffect, useRef, useState } from "react";
import { locate, type Position } from "../../shared/location/position";
import { errorMessage, untilAbort } from "../../shared/http/lifecycle";
export function View() {
  const [busy, setBusy] = useState(false),
    [status, setStatus] = useState("尚未要求定位。"),
    [position, setPosition] = useState<Position | null>(null),
    [valid, setValid] = useState(false);
  const active = useRef<AbortController | null>(null);
  useEffect(() => {
    const stop = () => {
      active.current?.abort();
      active.current = null;
      setBusy(false);
      setValid(false);
      setStatus("操作已中止，請重試。");
    };
    window.addEventListener("pagehide", stop);
    return () => {
      active.current?.abort();
      active.current = null;
      window.removeEventListener("pagehide", stop);
    };
  }, []);
  async function start() {
    if (active.current) return;
    const c = new AbortController();
    active.current = c;
    setBusy(true);
    setValid(false);
    setStatus("正在取得目前位置……");
    let timer: ReturnType<typeof setTimeout> | undefined;
    try {
      const value = await locate(c.signal);
      if (active.current !== c) return;
      setStatus("正在驗證位置……");
      timer = setTimeout(() => c.abort(), 10000);
      const response = await untilAbort(
        fetch("/api/location/validate", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "omit",
          cache: "no-store",
          signal: c.signal,
          body: JSON.stringify(value),
        }),
        c.signal,
      );
      const data = await untilAbort(response.json(), c.signal);
      if (active.current !== c) return;
      if (!response.ok || data.valid !== true)
        throw new Error(
          typeof data.error === "string" ? data.error : "位置驗證失敗。",
        );
      setPosition(value);
      setValid(true);
      setStatus("位置驗證成功。");
    } catch (error) {
      if (active.current === c) setStatus(errorMessage(error));
    } finally {
      clearTimeout(timer);
      if (active.current === c) {
        active.current = null;
        setBusy(false);
      }
    }
  }
  return (
    <>
      <div className="mt-4 flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center">
        <Button id="locate" disabled={busy} onClick={() => void start()}>
          {busy && <Spinner aria-label="處理中" />}
          取得位置並驗證
        </Button>
      </div>
      <Status id="status" role="status">
        {status}
      </Status>
      <dl className="grid gap-2 text-sm [&_dd]:break-all [&_dd]:text-muted-foreground" id="result" hidden={!position}>
        <div>
          <dt>緯度</dt>
          <dd id="latitude">{position?.latitude}</dd>
        </div>
        <div>
          <dt>經度</dt>
          <dd id="longitude">{position?.longitude}</dd>
        </div>
        <div>
          <dt>估計精度</dt>
          <dd id="accuracy">{position?.accuracyMeters} 公尺</dd>
        </div>
        <div>
          <dt>採樣時間</dt>
          <dd id="captured">
            {position?.capturedAt}
            {!valid && position && "（上次取得）"}
          </dd>
        </div>
        <div>
          <dt>伺服器驗證</dt>
          <dd id="validation">{valid ? "通過" : "尚未驗證"}</dd>
        </div>
      </dl>
    </>
  );
}
