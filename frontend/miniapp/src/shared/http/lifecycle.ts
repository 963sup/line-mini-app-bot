import { useCallback, useEffect, useRef } from "react";
export function aborted() {
  return new DOMException("操作已中止或逾時，請重試。", "AbortError");
}
export function untilAbort<T>(
  promise: Promise<T>,
  signal: AbortSignal,
): Promise<T> {
  return new Promise((resolve, reject) => {
    const stop = () => reject(aborted());
    if (signal.aborted) {
      stop();
      return;
    }
    signal.addEventListener("abort", stop, { once: true });
    promise.then(
      (value) => {
        signal.removeEventListener("abort", stop);
        if (signal.aborted) stop();
        else resolve(value);
      },
      (error) => {
        signal.removeEventListener("abort", stop);
        reject(error);
      },
    );
  });
}
export function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : "操作失敗，請重試。";
}
// A generation check also rejects responses from transports that ignore cancellation.
export function useLatestRequest() {
  const version = useRef(0),
    controller = useRef<AbortController | null>(null);
  const cancel = useCallback(() => {
    version.current++;
    controller.current?.abort();
    controller.current = null;
  }, []);
  useEffect(() => cancel, [cancel]);
  const run = useCallback(
    async <T>(
      work: (signal: AbortSignal) => Promise<T>,
      success: (value: T) => void,
      failure: (error: unknown) => void,
    ) => {
      cancel();
      const current = version.current,
        c = new AbortController();
      controller.current = c;
      try {
        const data = await untilAbort(work(c.signal), c.signal);
        if (current === version.current) success(data);
      } catch (error) {
        if (current === version.current) failure(error);
      } finally {
        if (controller.current === c) controller.current = null;
      }
    },
    [cancel],
  );
  return { run, cancel };
}
