import { createLocationClient } from "@repo/api-client";
import { Spinner } from "@/shared/ui/spinner";
import { Button } from "@/shared/ui/button";
import { Status } from "@/shared/Status";
import { PagePanel } from "@/shared/PagePanel";
import { Badge } from "@/shared/ui/badge";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from "@/shared/ui/empty";
import { useEffect, useRef, useState } from "react";
import { locate, type Position } from "../../shared/location/position";
import { errorMessage, untilAbort } from "../../shared/http/lifecycle";
export function View() {
  const [busy, setBusy] = useState(false),
    [status, setStatus] = useState(""),
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
      const {data,error} = await untilAbort(
        createLocationClient(location.origin).POST("/api/location/validate", {
          body: value, signal: c.signal,
        }), c.signal,
      );
      if (active.current !== c) return;
      if (data?.valid !== true) throw new Error(error?.error || "位置驗證失敗。");
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
    <PagePanel heading="定位測試" description="取得裝置目前位置，再交由伺服器驗證，不會儲存。結果包含座標、估計精度與採樣時間。">
      <div className="flex flex-wrap items-center gap-2">
        <Button id="locate" disabled={busy} onClick={start}>
          {busy && <Spinner data-icon="inline-start" aria-label="處理中" />}
          取得位置並驗證
        </Button>
      </div>
      <Status id="status" role="status">
        {status}
      </Status>
      {!position && !busy && !status && <Empty><EmptyHeader>
        <EmptyTitle>尚未取得位置</EmptyTitle>
        <EmptyDescription>按「取得位置並驗證」，並允許瀏覽器使用位置資訊。</EmptyDescription>
      </EmptyHeader></Empty>}
      <dl className="grid gap-4 text-sm sm:grid-cols-2 [&_dt]:font-medium [&_dd]:break-all [&_dd]:text-muted-foreground" id="result" hidden={!position}>
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
          <dd id="validation"><Badge variant={valid ? "secondary" : "outline"}>{valid ? "通過" : "尚未驗證"}</Badge></dd>
        </div>
      </dl>
    </PagePanel>
  );
}
