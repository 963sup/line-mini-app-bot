import { listMemberGroups } from "@/business/group/memberGroups";
import { useCallback, useEffect, useRef, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";
import { PagePanel } from "@/shared/PagePanel";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { Item, ItemGroup, ItemContent, ItemTitle } from "@/shared/ui/item";
import { View } from "./View";

type Group = components["schemas"]["Group"];
type Route = { owner: string; project: string; source: string };
const emptyRoute: Route = { owner: "", project: "", source: "" };

function readRoute(): Route {
  const query = new URLSearchParams(location.search);
  return { owner: query.get("owner") || query.get("group") || "", project: query.get("project") || "", source: query.get("source") || "" };
}

export function Workspace() {
  const { user, epoch } = useSession();
  if (!user) return <Status role="status">請先完成身分驗證，以查看專案。</Status>;
  return <AuthenticatedWorkspace key={epoch + ":" + user.userId} />;
}

function AuthenticatedWorkspace() {
  const { request, user } = useSession();
  const list = useLatestRequest(), detail = useLatestRequest();
  const blocked = useRef(false), navigating = useRef(false);
  const [writeBlocked, setWriteBlocked] = useState(false);
  const onBlocked = useCallback((value: boolean) => { blocked.current = value; setWriteBlocked(value); }, []);
  const canWrite = useCallback(() => !navigating.current, []);
  useEffect(() => { navigating.current = false; });
  const [route, setRoute] = useState<Route>(emptyRoute);
  const [ready, setReady] = useState(false);
  const [groups, setGroups] = useState<Group[]>([]), [cursor, setCursor] = useState("");
  const [selected, setSelected] = useState<{ id: string; item: Group | null }>({ id: "", item: null });
  const [busy, setBusy] = useState(false), [status, setStatus] = useState("");
  const [listFailed, setListFailed] = useState(false);
  const [detailStatus, setDetailStatus] = useState("");

  useEffect(() => {
    const update = () => { setRoute(readRoute()); setReady(true); };
    update();
    window.addEventListener("popstate", update);
    return () => window.removeEventListener("popstate", update);
  }, []);

  const navigate = useCallback((owner: string, project = "", source = "") => {
    if (blocked.current) return;
    navigating.current = true;
    const next = { owner, project, source };
    const query = new URLSearchParams();
    if (owner) query.set("owner", owner);
    if (project) query.set("project", project);
    if (source) query.set("source", source);
    window.history.pushState(null, "", "/miniapp/projects" + (query.size ? "?" + query : ""));
    setRoute(next);
  }, []);

  const load = useCallback((next = "") => {
    if (blocked.current) return;
    setListFailed(false); setBusy(true); setCursor(""); setStatus("正在核驗可開啟的群組……");
    if (!next) setGroups([]);
    void list.run(signal => listMemberGroups(request, signal, next || undefined), page => {
      setGroups(previous => [...new Map([...(next ? previous : []), ...page.items].map(item => [item.id, item])).values()]);
      setCursor(page.nextCursor); setBusy(false);
      setStatus(!page.nextCursor && !next && !page.items.length ? "目前沒有可開啟的群組。" : "");
    }, error => { setGroups([]); setListFailed(true); setBusy(false); setStatus(errorMessage(error)); });
  }, [request, list.run]);
  useEffect(() => { load(); return list.cancel; }, [load, list.cancel]);

  useEffect(() => {
    detail.cancel(); setSelected({ id: route.owner, item: null });
    if (!ready || !route.owner) { setDetailStatus(""); return; }
    if (route.owner === user?.userId) { setDetailStatus(""); return; }
    if (!/^C[0-9a-f]{32}$/.test(route.owner)) { setDetailStatus("群組連結無效。"); return; }
    setDetailStatus("正在核驗群組資格……");
    void detail.run(signal => request((client, s) => client.GET("/api/v1/groups/{id}", {
      signal: s, params: { path: { id: route.owner } },
    }), { signal }), item => {
      if (item.id !== route.owner || item.groupId !== route.owner) { setDetailStatus("無法開啟此群組。"); return; }
      setSelected({ id: route.owner, item }); setDetailStatus("");
    }, error => setDetailStatus(errorMessage(error)));
    return detail.cancel;
  }, [ready, user?.userId, route.owner, request, detail.run, detail.cancel]);

  const active = selected.id === route.owner ? selected.item : null;
  const selectProject = useCallback((id: string) => navigate(route.owner, id, route.source), [navigate, route.owner, route.source]);
  return <>
    {(active || (user && route.owner === user.userId)) && <View key={route.owner} ownerId={route.owner} projectId={route.project} initialSource={route.source} onSelect={selectProject} onBlocked={onBlocked} canWrite={canWrite} />}
    <PagePanel heading="專案群組">
      <Status role="status">{status}</Status>
      {listFailed && !writeBlocked && <Button type="button" variant="outline" onClick={() => load()}>重試讀取</Button>}
      <Button disabled={writeBlocked} onClick={() => user && navigate(user.userId)}>我的個人專案</Button>
      <ItemGroup>
        {groups.map(group => <Item key={group.id} variant="outline"><ItemContent><ItemTitle>
          <Button variant="link" disabled={writeBlocked} aria-current={group.id === route.owner ? "page" : undefined} onClick={() => navigate(group.id)}>{group.name || "工程群組"}</Button>
        </ItemTitle></ItemContent></Item>)}
      </ItemGroup>
      {cursor && <Button variant="outline" disabled={busy || writeBlocked} onClick={() => load(cursor)}>載入更多專案群組</Button>}
      {ready && !route.owner && <p>請選擇個人或群組專案。</p>}
      <Status role="status">{detailStatus}</Status>
    </PagePanel>
  </>;
}
