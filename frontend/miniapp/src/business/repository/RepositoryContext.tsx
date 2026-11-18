import React, { createContext, useCallback, useContext, useEffect, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "../../identity/Session";
import { errorMessage, useLatestRequest } from "../../shared/http/lifecycle";

type Repository = components["schemas"]["Repository"];
type Route = { ready: boolean; scope: string; selected: string };
type Catalog = { owner: string; items: Repository[]; cursor: string; busy: boolean; loaded: boolean; status: string };
type Detail = { owner: string; id: string; item: Repository | null; busy: boolean; status: string; failed: boolean };
const emptyCatalog: Catalog = { owner: "", items: [], cursor: "", busy: false, loaded: false, status: "" };
const emptyDetail: Detail = { owner: "", id: "", item: null, busy: false, status: "", failed: false };

function repositoryRoute(): Route {
  const params = new URLSearchParams(location.search);
  let scope = params.get("id") || params.get("repository") || "";
  try {
    const nested = params.get("liff.state");
    if (nested) {
      const query = new URL(nested, location.origin).searchParams;
      scope ||= query.get("id") || query.get("repository") || "";
    }
  } catch {}
  return { ready: true, scope, selected: scope || params.get("selected") || "" };
}

function useRepositoryState(workflow: string) {
  const { user, epoch, request } = useSession();
  const [route, setRoute] = useState<Route>({ ready: false, scope: "", selected: "" });
  const owner = `${epoch}:${user?.userId || ""}:${route.scope}`;
  const [catalog, setCatalog] = useState<Catalog>(emptyCatalog);
  const [detail, setDetail] = useState<Detail>(emptyDetail);
  const { run: listRun, cancel: listCancel } = useLatestRequest();
  const { run: detailRun, cancel: detailCancel } = useLatestRequest();
  useEffect(() => setRoute(workflow === "repositories" ? repositoryRoute() : { ready: true, scope: "", selected: "" }), [workflow]);

  const loadPage = useCallback((cursor = "") => {
    if (!user || !route.ready || route.scope) return;
    setCatalog((previous) => ({
      ...emptyCatalog, owner, busy: true, status: "正在讀取空間……",
      items: cursor && previous.owner === owner ? previous.items : [],
    }));
    void listRun(
      (signal) => request((client, s) => client.GET("/api/v1/repositories", {
        signal: s, params: { query: { cursor } },
      }), { signal }),
      (page) => setCatalog((previous) => {
        const values = [...(cursor && previous.owner === owner ? previous.items : []), ...page.items];
        const items = [...new Map(values.filter((item) => item.status === "ACTIVE").map((item) => [item.id, item])).values()];
        return { owner, items, cursor: page.nextCursor, busy: false, loaded: true,
          status: items.length ? "選擇空間查看資料。" : page.nextCursor ? "本頁沒有可開啟的空間，可繼續載入。" : "目前沒有可開啟的空間。" };
      }),
      (error) => setCatalog({ ...emptyCatalog, owner, loaded: true, status: errorMessage(error) }),
    );
  }, [user, owner, request, route, listRun]);

  const loadDetail = useCallback(() => {
    if (!user || !route.ready || !route.selected) return;
    setDetail({ ...emptyDetail, owner, id: route.selected, busy: true, status: "正在核驗空間存取資格……" });
    void detailRun(
      (signal) => request((client, s) => client.GET("/api/v1/repositories/{id}", {
        signal: s, params: { path: { id: route.selected } },
      }), { signal }),
      (item) => {
        if (item.status !== "ACTIVE" || item.id !== route.selected) {
          setDetail({ ...emptyDetail, owner, id: route.selected, failed: true, status: "無法開啟此空間。" });
          return;
        }
        setDetail({ owner, id: route.selected, item, busy: false, status: "已載入空間。", failed: false });
      },
      (error) => setDetail({ ...emptyDetail, owner, id: route.selected, failed: true, status: errorMessage(error) }),
    );
  }, [user, owner, request, route, detailRun]);

  useEffect(() => {
    listCancel();
    detailCancel();
    setCatalog(emptyCatalog);
    setDetail(emptyDetail);
    if (user && route.ready) {
      if (route.selected) loadDetail();
      if (!route.scope && workflow === "repositories") loadPage();
    }
    return () => { listCancel(); detailCancel(); };
  }, [owner, user, route, workflow, loadPage, loadDetail, listCancel, detailCancel]);

  // Guard during render as well as effect cleanup: a previous identity's data
  // must never appear in the first render after logout or re-verification.
  const currentCatalog = user && catalog.owner === owner ? catalog : emptyCatalog;
  const currentDetail = user && detail.owner === owner ? detail : emptyDetail;
  const items = route.scope
    ? currentDetail.item ? [currentDetail.item] : []
    : currentCatalog.items.filter((item) => !(currentDetail.failed && item.id === currentDetail.id));
  const busy = route.scope ? currentDetail.busy : currentCatalog.busy;
  const status = !user ? "請先完成身分驗證。" : route.scope
    ? currentDetail.status : currentCatalog.status || "展開後載入可開啟的空間。";
  const ensure = () => {
    if (!currentCatalog.loaded && !currentCatalog.busy && !route.scope) loadPage();
  };
  return { items, busy, status, scope: route.scope, selectedID: route.selected,
    cursor: route.scope ? "" : currentCatalog.cursor,
    selected: currentDetail.item, detailBusy: currentDetail.busy,
    detailStatus: !user ? "請先完成身分驗證。" : currentDetail.status,
    canLoad: !!user && route.ready,
    ensure, loadMore: () => loadPage(currentCatalog.cursor),
    reload: () => route.scope ? loadDetail() : loadPage(), reloadDetail: loadDetail,
    href: (id: string) => "/miniapp/repositories?" + new URLSearchParams({ [route.scope ? "id" : "selected"]: id }),
  };
}

type RepositoryState = ReturnType<typeof useRepositoryState>;
const Context = createContext<RepositoryState | null>(null);
export function RepositoryProvider({ workflow, children }: { workflow: string; children: React.ReactNode }) {
  const state = useRepositoryState(workflow);
  return <Context.Provider value={state}>{children}</Context.Provider>;
}
export function useRepositories() {
  const value = useContext(Context);
  if (!value) throw new Error("RepositoryProvider is required");
  return value;
}
