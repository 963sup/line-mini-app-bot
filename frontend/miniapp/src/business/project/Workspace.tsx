import { useOrganizationSelection } from "@/business/organization/OrganizationSelection";
import { useCallback, useEffect, useRef, useState } from "react";
import { useSession } from "@/identity/Session";
import { PagePanel } from "@/shared/PagePanel";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { Item, ItemGroup, ItemContent, ItemTitle } from "@/shared/ui/item";
import { View } from "./View";

type CopySource = { ownerId: string; projectId: string };
type Route = { owner: string; project: string; source: string; copyFrom?: CopySource };
const emptyRoute: Route = { owner: "", project: "", source: "" };

function readRoute(): Route {
  const query = new URLSearchParams(location.search);
  const ownerId = query.get("fromOwner");
  const projectId = query.get("fromProject");
  return {
    owner: query.get("owner") || "",
    project: query.get("project") || "",
    source: query.get("source") || "",
    copyFrom: ownerId && projectId ? { ownerId, projectId } : undefined,
  };
}

export function Workspace() {
  const { user, epoch } = useSession();
  if (!user) return <Status role="status">請先完成身分驗證，以查看專案。</Status>;
  return <AuthenticatedWorkspace key={epoch + ":" + user.userId} />;
}

function AuthenticatedWorkspace() {
  const { user } = useSession();
  const { context, error, refresh } = useOrganizationSelection();
  const organizations = context?.groups.flatMap(group => group.organizations) ?? [];
  const blocked = useRef(false), navigating = useRef(false);
  const [writeBlocked, setWriteBlocked] = useState(false);
  const onBlocked = useCallback((value: boolean) => { blocked.current = value; setWriteBlocked(value); }, []);
  const canWrite = useCallback(() => !navigating.current, []);
  useEffect(() => { navigating.current = false; });
  const [route, setRoute] = useState<Route>(emptyRoute);
  const [ready, setReady] = useState(false);
  useEffect(() => {
    const update = () => {
      const next = readRoute();
      setRoute({ ...next, owner: next.owner || user?.userId || "" });
      setReady(true);
    };
    update();
    window.addEventListener("popstate", update);
    return () => window.removeEventListener("popstate", update);
  }, [user?.userId]);

  const navigate = useCallback((owner: string, project = "", source = "", copyFrom?: CopySource) => {
    if (blocked.current) return;
    navigating.current = true;
    const next = { owner, project, source, copyFrom };
    const query = new URLSearchParams();
    if (owner) query.set("owner", owner);
    if (project) query.set("project", project);
    if (copyFrom) {
      query.set("fromOwner", copyFrom.ownerId);
      query.set("fromProject", copyFrom.projectId);
    }
    if (source) query.set("source", source);
    window.history.pushState(null, "", "/miniapp/projects" + (query.size ? "?" + query : ""));
    setRoute(next);
  }, []);

  const active = organizations.some(org => org.organizationId === route.owner);
  const selectProject = useCallback((id: string) => navigate(route.owner, id, route.source), [navigate, route.owner, route.source]);
  return <div className="flex min-w-0 flex-col gap-4">
    {user && (active || route.owner === user.userId) && <View
      key={route.owner}
      ownerId={route.owner}
      projectId={route.project}
      initialSource={route.source}
      copyFrom={route.copyFrom}
      destinations={[{ id: user.userId, name: "我的個人專案" }, ...organizations.map(org => ({ id: org.organizationId, name: org.name }))]}
      onCopy={destination => navigate(destination, "new", "", { ownerId: route.owner, projectId: route.project })}
      onSelect={selectProject}
      onBlocked={onBlocked}
      canWrite={canWrite}
    />}
    <PagePanel heading="專案擁有者">
      <Button disabled={writeBlocked} onClick={() => user && navigate(user.userId)}>我的個人專案</Button>
      <ItemGroup>{organizations.map(org => <Item key={org.organizationId}><ItemContent><ItemTitle>
        <Button variant="link" disabled={writeBlocked} aria-current={org.organizationId === route.owner ? "page" : undefined} onClick={() => navigate(org.organizationId)}>{org.name}</Button>
      </ItemTitle></ItemContent></Item>)}</ItemGroup>
      {error && <><Status role="status">{error}</Status><Button onClick={refresh}>重試讀取組織</Button></>}
      {ready && context && !active && route.owner !== user?.userId && <Status role="status">無法開啟此組織。</Status>}
    </PagePanel>
  </div>;
}
