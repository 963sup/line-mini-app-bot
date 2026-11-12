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
  pictureUrl: string | undefined;
  canLogout: boolean;
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
  const [pictureUrl, setPictureUrl] = useState<string>();
  const [user, setUser] = useState<Session["user"]>(null),
    [epoch, setEpoch] = useState(0),
    [status, setStatus] = useState("正在準備登入……"),
    [busy, setBusy] = useState(true),
    [ready, setReady] = useState(false),
    [inClient, setInClient] = useState(false);
  const sdk = useRef<Liff | null>(null),
    actor = useRef<Session["user"]>(null),
    csrf = useRef(""),
    loggingOut = useRef(false),
    expiresAt = useRef(0),
    lastActivity = useRef(0),
    lastRefreshAttempt = useRef(0),
    refreshPending = useRef<Promise<void> | null>(null),
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
    expiresAt.current = 0;
    lastRefreshAttempt.current = 0;
    refreshPending.current = null;
  }, []);
  const reset = useCallback(() => {
    invalidate();
    setUser(null);
    setPictureUrl(undefined);
    setEpoch((value) => value + 1);
  }, [invalidate]);
  const verify = useCallback(async () => {
    if (loggingOut.current) return;
    reset();
    csrf.current = "";
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
      const session = await untilAbort(
        createAPIClient(location.origin, token)
          .POST("/api/v1/session", { body: {}, params: { header: { Origin: location.origin } }, signal: c.signal })
          .then(unwrap),
        c.signal,
      );
      if (current !== generation.current) return;
      const data = session.user;
      if (!data.userId || !session.csrfToken || !Number.isFinite(Date.parse(session.expiresAt))) throw new Error("身分回應無效。");
      csrf.current = session.csrfToken;
      expiresAt.current = Date.parse(session.expiresAt);
      lastActivity.current = Date.now();
      actor.current = data;
      setUser(data);
      setStatus("本人身分驗證成功。");
      // Profile is display-only; authorization continues to use the server session.
      void untilAbort(liff.getProfile(), AbortSignal.timeout(5000)).then((profile) => {
        if (current !== generation.current || profile.userId !== data.userId || !profile.pictureUrl) return;
        const picture = new URL(profile.pictureUrl);
        if (picture.protocol === "https:" && picture.hostname === "profile.line-scdn.net") {
          setPictureUrl(picture.href);
        }
      }).catch(() => { /* Missing profile permission must not block login. */ });
    } catch (error) {
      if (current === generation.current) setStatus(errorMessage(error));
    } finally {
      clearTimeout(timer);
      active.current.delete(c);
      if (current === generation.current) setBusy(false);
    }
  }, [reset, sdkLoader]);
  const refreshIfDue = useCallback(async () => {
    const now = Date.now();
    if (!actor.current || !csrf.current || loggingOut.current ||
        document.visibilityState === "hidden" || now - lastActivity.current > 120000 ||
        expiresAt.current - now > 600000) return;
    if (refreshPending.current) return refreshPending.current;
    if (now - lastRefreshAttempt.current < 60000) return;
    lastRefreshAttempt.current = now;
    const current = generation.current, c = new AbortController();
    active.current.add(c);
    const timer = setTimeout(() => c.abort(), 10000);
    const pending = (async () => {
      try {
        const result = await untilAbort(
          createAPIClient(location.origin, undefined, csrf.current)
            .POST("/api/v1/session/refresh", {
              body: {}, params: { header: { Origin: location.origin, "X-CSRF-Token": csrf.current } }, signal: c.signal,
            }), c.signal);
        if (current !== generation.current) throw aborted();
        if (result.response.status === 401 || result.response.status === 403) {
          reset();
          csrf.current = "";
          setStatus("登入已失效，請重新驗證身分。");
        }
        const value = unwrap(result);
        if (current !== generation.current) throw aborted();
        if (value.user.userId !== actor.current?.userId || value.csrfToken !== csrf.current ||
            !Number.isFinite(Date.parse(value.expiresAt))) throw new Error("登入續期回應無效，請重新驗證。");
        expiresAt.current = Date.parse(value.expiresAt);
      } finally {
        clearTimeout(timer);
        active.current.delete(c);
      }
    })();
    refreshPending.current = pending;
    try { await pending; }
    finally { if (refreshPending.current === pending) refreshPending.current = null; }
  }, [reset]);
  const request: Request = useCallback(async (operation, options = {}) => {
    const current = generation.current,
      liff = sdk.current;
    if (!actor.current || !csrf.current || !liff?.isLoggedIn())
      throw new Error("請先完成身分驗證。");
    await refreshIfDue();
    if (current !== generation.current) throw aborted();
    const c = new AbortController(),
      stop = () => c.abort();
    active.current.add(c);
    options.signal?.addEventListener("abort", stop, { once: true });
    if (options.signal?.aborted) c.abort();
    const timer = setTimeout(stop, options.timeoutMs ?? 15000);
    try {
      const result = await untilAbort(
        operation(createAPIClient(location.origin, undefined, csrf.current), c.signal),
        c.signal,
      );
      if (current === generation.current && result.response.status === 401) {
        reset();
        csrf.current = "";
        setStatus("登入已失效，請重新驗證身分。");
      }
      const data = unwrap(result);
      if (current !== generation.current) throw aborted();
      return data;
    } finally {
      clearTimeout(timer);
      active.current.delete(c);
      options.signal?.removeEventListener("abort", stop);
    }
  }, [reset, refreshIfDue]);
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
  useEffect(() => {
    const renew = () => { void refreshIfDue().catch(() => {}); };
    const activity = (event: Event) => {
      if (!event.isTrusted) return;
      lastActivity.current = Date.now();
      renew();
    };
    for (const name of ["pointerdown", "keydown", "touchstart", "wheel"]) {
      window.addEventListener(name, activity, { passive: true });
    }
    document.addEventListener("visibilitychange", renew);
    const timer = setInterval(renew, 60000);
    return () => {
      clearInterval(timer);
      for (const name of ["pointerdown", "keydown", "touchstart", "wheel"]) window.removeEventListener(name, activity);
      document.removeEventListener("visibilitychange", renew);
    };
  }, [refreshIfDue]);
  const logout = useCallback(async () => {
    if (loggingOut.current) return;
    loggingOut.current = true;
    punchNotice.current = "已清除本頁打卡狀態。";
    reset();
    setBusy(true);
    setStatus("正在撤銷登入……");
    try {
      const result = await createAPIClient(location.origin, undefined, csrf.current)
        .DELETE("/api/v1/session", {
          params: { header: { Origin: location.origin } },
          signal: AbortSignal.timeout(10000),
        });
      if (!result.response.ok) throw new Error("伺服器登出尚未確認，請重試登出。");
      csrf.current = "";
      const liff = sdk.current;
      if (liff && !liff.isInClient()) liff.logout();
      setStatus("已撤銷登入並清除本頁狀態；LINE 內建瀏覽器下次開啟時可能自動登入。");
    } catch (error) {
      // Keep the token only to allow revocation to be retried; business requests stay disabled.
      setStatus(errorMessage(error) + " 請重試登出，或重新驗證後再登出。");
    } finally {
      loggingOut.current = false;
      setBusy(false);
    }
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
      pictureUrl,
      canLogout: !!user || !!csrf.current,
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
      pictureUrl,
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
