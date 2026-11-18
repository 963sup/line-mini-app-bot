import { aborted } from "../http/lifecycle";
export interface Position {
  latitude: number;
  longitude: number;
  accuracyMeters: number;
  capturedAt: string;
}
export function locate(signal: AbortSignal): Promise<Position> {
  return new Promise((resolve, reject) => {
    if (!window.isSecureContext || !navigator.geolocation) {
      reject(new Error("此瀏覽器無法定位，請使用 HTTPS 並確認定位權限。"));
      return;
    }
    let watch: number | undefined,
      done = false;
    const cleanup = () => {
      clearTimeout(timer);
      signal.removeEventListener("abort", stop);
      if (watch !== undefined) navigator.geolocation.clearWatch(watch);
    };
    const fail = (error: Error) => {
      if (done) return;
      done = true;
      cleanup();
      reject(error);
    };
    const stop = () => fail(aborted());
    const timer = setTimeout(
      () => fail(new Error("定位等待逾時，請檢查權限後重試。")),
      15000,
    );
    if (signal.aborted) {
      stop();
      return;
    }
    signal.addEventListener("abort", stop, { once: true });
    try {
      watch = navigator.geolocation.watchPosition(
        (position) => {
          if (done) return;
          const { latitude, longitude, accuracy } = position.coords,
            date = new Date(position.timestamp);
          if (
            !Number.isFinite(latitude) ||
            Math.abs(latitude) > 90 ||
            !Number.isFinite(longitude) ||
            Math.abs(longitude) > 180 ||
            !Number.isFinite(accuracy) ||
            accuracy < 0 ||
            !Number.isFinite(date.getTime())
          ) {
            fail(new Error("裝置傳回的位置無效，請重新定位。"));
            return;
          }
          done = true;
          cleanup();
          resolve({
            latitude,
            longitude,
            accuracyMeters: accuracy,
            capturedAt: date.toISOString(),
          });
        },
        (error) =>
          fail(
            new Error(
              (
                {
                  1: "未取得定位權限，請調整 LINE、瀏覽器與系統設定後重試。",
                  2: "目前無法取得位置，請稍後重試。",
                  3: "定位逾時，請重試。",
                } as Record<number, string>
              )[error.code] || "定位失敗，請重試。",
            ),
          ),
        { enableHighAccuracy: false, timeout: 10000, maximumAge: 0 },
      );
      if (done) navigator.geolocation.clearWatch(watch);
    } catch {
      fail(new Error("瀏覽器無法啟動定位，請確認權限後重試。"));
    }
  });
}
