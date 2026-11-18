export interface Liff {
  init(options: { liffId: string }): Promise<void>;
  isLoggedIn(): boolean;
  isInClient(): boolean;
  getIDToken(): string | null;
  getProfile(): Promise<{ userId: string; pictureUrl?: string }>;
  login(options: { redirectUri: string }): void;
  logout(): void;
  closeWindow(): void;
}
declare global {
  interface Window {
    liff?: Liff;
  }
}
let loading: Promise<Liff> | undefined;
const initializing = new WeakMap<Liff, Map<string, Promise<void>>>();
export function loadSDK(): Promise<Liff> {
  if (window.liff) return Promise.resolve(window.liff);
  if (!loading)
    loading = new Promise<Liff>((resolve, reject) => {
      const script = document.createElement("script");
      script.src = "https://static.line-scdn.net/liff/edge/2/sdk.js";
      script.onload = () =>
        window.liff
          ? resolve(window.liff)
          : reject(new Error("LINE SDK 未載入。"));
      script.onerror = () => {
        script.remove();
        reject(new Error("無法載入 LINE SDK，請確認網路後重試。"));
      };
      document.head.append(script);
    }).catch((error) => {
      loading = undefined;
      throw error;
    });
  return loading;
}
export function initializeSDK(sdk: Liff, liffId: string) {
  let entries = initializing.get(sdk);
  if (!entries) {
    entries = new Map();
    initializing.set(sdk, entries);
  }
  let pending = entries.get(liffId);
  if (!pending) {
    pending = sdk.init({ liffId }).catch((error) => {
      entries!.delete(liffId);
      throw error;
    });
    entries.set(liffId, pending);
  }
  return pending;
}
