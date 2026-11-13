import React, { useCallback, useEffect, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";
import { PagePanel } from "@/shared/PagePanel";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { Item, ItemGroup, ItemContent, ItemTitle } from "@/shared/ui/item";
import { View } from "./View";

type Space = components["schemas"]["Space"];
type Route = { space: string; project: string; source: string };
const emptyRoute: Route = { space: "", project: "", source: "" };

function readRoute(): Route {
  const query = new URLSearchParams(location.search);
  return { space: query.get("space") || "", project: query.get("project") || "", source: query.get("source") || "" };
}

export function Workspace() {
  const { user, epoch } = useSession();
  if (!user) return <Status role="status">請先完成身分驗證，以查看專案。</Status>;
  return <AuthenticatedWorkspace key={epoch + ":" + user.userId} />;
}

function AuthenticatedWorkspace() {
  const { request } = useSession();
  const list = useLatestRequest(), detail = useLatestRequest();
  const [route, setRoute] = useState<Route>(emptyRoute);
  const [ready, setReady] = useState(false);
  const [spaces, setSpaces] = useState<Space[]>([]), [cursor, setCursor] = useState("");
  const [selected, setSelected] = useState<{ id: string; item: Space | null }>({ id: "", item: null });
  const [busy, setBusy] = useState(false), [status, setStatus] = useState("");
  const [detailStatus, setDetailStatus] = useState("");

  useEffect(() => {
    const update = () => { setRoute(readRoute()); setReady(true); };
    update();
    window.addEventListener("popstate", update);
    return () => window.removeEventListener("popstate", update);
  }, []);

  const navigate = useCallback((space: string, project = "", source = "") => {
    const next = { space, project, source };
    const query = new URLSearchParams();
    if (space) query.set("space", space);
    if (project) query.set("project", project);
    if (source) query.set("source", source);
    window.history.pushState(null, "", "/miniapp/projects" + (query.size ? "?" + query : ""));
    setRoute(next);
  }, []);

  const load = useCallback((next = "") => {
    setBusy(true); setCursor(""); setStatus("正在核驗可開啟的空間……");
    if (!next) setSpaces([]);
    void list.run(signal => request((client, s) => client.GET("/api/v1/spaces", {
      signal: s, params: { query: { cursor: next || undefined } },
    }), { signal }), page => {
      setSpaces(previous => [...new Map([...(next ? previous : []), ...page.items].map(item => [item.id, item])).values()]);
      setCursor(page.nextCursor); setBusy(false);
      setStatus(page.nextCursor ? "可繼續載入空間。" : "已讀取可開啟的空間。");
    }, error => { setSpaces([]); setBusy(false); setStatus(errorMessage(error)); });
  }, [request, list.run]);
  useEffect(() => { load(); return list.cancel; }, [load, list.cancel]);

  useEffect(() => {
    detail.cancel(); setSelected({ id: route.space, item: null });
    if (!ready || !route.space) { setDetailStatus(""); return; }
    if (!/^C[0-9a-f]{32}$/.test(route.space)) { setDetailStatus("空間連結無效。"); return; }
    setDetailStatus("正在核驗空間資格……");
    void detail.run(signal => request((client, s) => client.GET("/api/v1/spaces/{id}", {
      signal: s, params: { path: { id: route.space } },
    }), { signal }), item => {
      if (item.id !== route.space || item.groupId !== route.space || item.status !== "ACTIVE") { setDetailStatus("無法開啟此空間。"); return; }
      setSelected({ id: route.space, item }); setDetailStatus("");
    }, error => setDetailStatus(errorMessage(error)));
    return detail.cancel;
  }, [ready, route.space, request, detail.run, detail.cancel]);

  const active = selected.id === route.space ? selected.item : null;
  const selectProject = useCallback((id: string) => navigate(route.space, id, route.source), [navigate, route.space, route.source]);
  return <>
    <PagePanel heading="專案空間">
      <p>專案由空間管理，來源儲存庫只提供可讀取的事項。</p>
      <Button variant="outline" disabled={busy} onClick={() => load()}>重新載入專案空間</Button>
      <Status role="status">{status}</Status>
      <ItemGroup>
        {spaces.map(space => <Item key={space.id} variant="outline"><ItemContent><ItemTitle>
          <Button variant="link" aria-current={space.id === route.space ? "page" : undefined} onClick={() => navigate(space.id)}>{space.name || "群組空間"}</Button>
        </ItemTitle></ItemContent></Item>)}
      </ItemGroup>
      {cursor && <Button variant="outline" disabled={busy} onClick={() => load(cursor)}>載入更多專案空間</Button>}
      {ready && !route.space && <p>請選擇專案所屬空間。</p>}
      <Status role="status">{detailStatus}</Status>
    </PagePanel>
    {active && <View key={active.id} spaceId={active.id} projectId={route.project} initialSource={route.source} onSelect={selectProject} />}
  </>;
}
