import React from "react";
import { createAPIClient, type components } from "@repo/api-client";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { initializeSDK, loadSDK, type Liff } from "../platform/line/sdk";
import { aborted, errorMessage, untilAbort } from "../shared/http/lifecycle";
type Client = ReturnType<typeof createAPIClient>;
type Result<T> = { data?: T; error?: unknown; response: Response };
export type Request = <T>(
  operation: (client: Client, signal: AbortSignal) => Promise<Result<T>>,
  options?: { signal?: AbortSignal; timeoutMs?: number },
) => Promise<T>;
interface Session {
  user: components["schemas"]["CurrentUser"] | null;
  epoch: number;
  status: string;
  busy: boolean;
  ready: boolean;
  inClient: boolean;
  request: Request;
  verify: () => void;
  login: () => void;
  logout: () => void;
  close: () => void;
  autoPunch: React.MutableRefObject<boolean>;
  punchNotice: React.MutableRefObject<string>;
}
const Context = createContext<Session | null>(null);
function unwrap<T>({ data, error, response }: Result<T>): T {
  if (!response.ok) {
    const detail = (error as { error?: { message?: string; code?: string } })
      ?.error;
    throw Object.assign(new Error(detail?.message || "伺服器無法完成請求。"), {
      code: detail?.code,
    });
  }
  if (data === undefined) throw new Error("伺服器回應無效。");
  return data;
}
export function SessionProvider({
  children,
  sdkLoader = loadSDK,
}: {
  children: ReactNode;
  sdkLoader?: () => Promise<Liff>;
}) {
  const [user, setUser] = useState<Session["user"]>(null),
    [epoch, setEpoch] = useState(0),
    [status, setStatus] = useState("正在準備登入……"),
    [busy, setBusy] = useState(true),
    [ready, setReady] = useState(false),
    [inClient, setInClient] = useState(false);
  const sdk = useRef<Liff | null>(null),
    actor = useRef<Session["user"]>(null),
    generation = useRef(0),
    active = useRef(new Set<AbortController>()),
    autoPunch = useRef(false),
    punchNotice = useRef(""),
    closed = useRef(false);
  const invalidate = useCallback(() => {
    generation.current++;
    for (const c of active.current) c.abort();
    active.current.clear();
    actor.current = null;
  }, []);
  const reset = useCallback(() => {
    invalidate();
    setUser(null);
    setEpoch((value) => value + 1);
  }, [invalidate]);
  const verify = useCallback(async () => {
    reset();
    const current = generation.current,
      c = new AbortController();
    active.current.add(c);
    const timer = setTimeout(() => c.abort(), 10000);
    setBusy(true);
    try {
      if (!sdk.current) {
        setStatus("正在初始化 LINE……");
        const [config, loaded] = await untilAbort(
          Promise.all([
            createAPIClient(location.origin)
              .GET("/api/v1/miniapp/config", { signal: c.signal })
              .then(unwrap),
            sdkLoader(),
          ]),
          c.signal,
        );
        await untilAbort(initializeSDK(loaded, config.liffId), c.signal);
        if (current !== generation.current) return;
        sdk.current = loaded;
      }
      const liff = sdk.current;
      setReady(true);
      setInClient(liff.isInClient());
      if (!liff.isLoggedIn()) {
        setStatus("請登入 LINE 後驗證本人身分。");
        return;
      }
      const token = liff.getIDToken();
      if (!token)
        throw new Error("無法取得 LINE 身分，請確認 openid scope 並重新登入。");
      setStatus("正在由伺服器驗證 LINE 身分……");
      const data = await untilAbort(
        createAPIClient(location.origin, token)
          .GET("/api/v1/me", { signal: c.signal })
          .then(unwrap),
        c.signal,
      );
      if (current !== generation.current) return;
      if (!data.userId) throw new Error("身分回應無效。");
      actor.current = data;
      setUser(data);
      setStatus("本人身分驗證成功。");
    } catch (error) {
      if (current === generation.current) setStatus(errorMessage(error));
    } finally {
      clearTimeout(timer);
      active.current.delete(c);
      if (current === generation.current) setBusy(false);
    }
  }, [reset, sdkLoader]);
  const request: Request = useCallback(async (operation, options = {}) => {
    const current = generation.current,
      liff = sdk.current;
    if (!actor.current || !liff?.isLoggedIn())
      throw new Error("請先完成身分驗證。");
    const token = liff.getIDToken();
    if (!token) throw new Error("無法取得 LINE 身分，請重新登入。");
    const c = new AbortController(),
      stop = () => c.abort();
    active.current.add(c);
    options.signal?.addEventListener("abort", stop, { once: true });
    if (options.signal?.aborted) c.abort();
    const timer = setTimeout(stop, options.timeoutMs ?? 15000);
    try {
      const data = unwrap(
        await untilAbort(
          operation(createAPIClient(location.origin, token), c.signal),
          c.signal,
        ),
      );
      if (current !== generation.current) throw aborted();
      return data;
    } finally {
      clearTimeout(timer);
      active.current.delete(c);
      options.signal?.removeEventListener("abort", stop);
    }
  }, []);
  useEffect(() => {
    void verify();
    const hide = () => {
      reset();
      setBusy(false);
    };
    const show = (e: PageTransitionEvent) => {
      if (e.persisted) void verify();
    };
    window.addEventListener("pagehide", hide);
    window.addEventListener("pageshow", show);
    return () => {
      invalidate();
      window.removeEventListener("pagehide", hide);
      window.removeEventListener("pageshow", show);
    };
  }, [verify, reset, invalidate]);
  const logout = useCallback(() => {
    punchNotice.current = "已清除本頁打卡狀態。";
    reset();
    const liff = sdk.current;
    try {
      if (liff && !liff.isInClient()) liff.logout();
    } catch {}
    setBusy(false);
    setStatus("已清除本頁登入狀態；LINE 內建瀏覽器可能在下次開啟時自動登入。");
  }, [reset]);
  const close = useCallback(() => {
    if (!closed.current && sdk.current?.isInClient()) {
      try {
        sdk.current.closeWindow();
        closed.current = true;
      } catch {}
    }
  }, []);
  const value = useMemo(
    () => ({
      user,
      epoch,
      status,
      busy,
      ready,
      inClient,
      request,
      verify: () => {
        void verify();
      },
      login: () => sdk.current?.login({ redirectUri: location.href }),
      logout,
      close,
      autoPunch,
      punchNotice,
    }),
    [
      user,
      epoch,
      status,
      busy,
      ready,
      inClient,
      request,
      verify,
      logout,
      close,
    ],
  );
  return <Context.Provider value={value}>{children}</Context.Provider>;
}
export function useSession() {
  const value = useContext(Context);
  if (!value) throw new Error("SessionProvider is required");
  return value;
}
