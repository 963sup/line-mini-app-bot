import { Check } from "lucide-react";
import { useOrganizationSelection } from "@/business/organization/OrganizationSelection";
import { useCallback, useEffect, useRef, useState } from "react";
import { useSession } from "@/identity/Session";
import { PagePanel } from "@/shared/PagePanel";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { View, type EditorNavigation } from "./View";

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
  const editor = useRef<EditorNavigation>(null);
  const historyIndex = useRef(0);
  useEffect(() => {
    const update = () => {
      const next = readRoute();
      setRoute({ ...next, owner: next.owner || user?.userId || "" });
      setReady(true);
    };
    const history = window.history;
    historyIndex.current = history.state?.projectIndex ?? 0;
    history.replaceState({ ...history.state, projectIndex: historyIndex.current }, "");
    let restoring = false;
    const traverse = (event: PopStateEvent) => {
      const target: number | undefined = event.state?.projectIndex;
      if (restoring && target === historyIndex.current) {
        restoring = false;
        event.stopImmediatePropagation();
        return;
      }
      if (target !== undefined && target !== historyIndex.current &&
          (blocked.current || editor.current?.allowHistoryChange() === false)) {
        // popstate fires after the URL moves. Return to the original entry without truncating forward history.
        event.stopImmediatePropagation();
        restoring = true;
        history.go(historyIndex.current - target);
        return;
      }
      historyIndex.current = target ?? 0;
      update();
    };
    update();
    window.addEventListener("popstate", traverse, true);
    return () => window.removeEventListener("popstate", traverse, true);
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
    window.history.pushState({ projectIndex: ++historyIndex.current }, "", "/miniapp/projects" + (query.size ? "?" + query : ""));
    setRoute(next);
  }, []);

  const owners = user ? [{ id: user.userId, name: "我的個人專案" }, ...organizations.map(org => ({ id: org.organizationId, name: org.name }))] : [];
  const active = organizations.some(org => org.organizationId === route.owner);
  const selectProject = useCallback((id: string) => navigate(route.owner, id, route.source), [navigate, route.owner, route.source]);
  return <div className="flex min-w-0 flex-col gap-4">
    <PagePanel heading="專案擁有者">
      <nav aria-label="專案擁有者" className="flex flex-wrap gap-2">
        {owners.map(owner => {
          const current = owner.id === route.owner;
          return <Button key={owner.id} variant={current ? "default" : "outline"}
            className="h-auto min-h-11 max-w-full whitespace-normal text-left [overflow-wrap:anywhere]"
            disabled={writeBlocked} aria-current={current ? "page" : undefined}
            onClick={() => {
              const select = () => navigate(owner.id);
              if (editor.current) editor.current.navigate(select);
              else select();
            }}>
            {current && <Check aria-hidden="true" />}
            {owner.name}
          </Button>;
        })}
      </nav>
      {error && <><Status role="status">{error}</Status><Button onClick={refresh}>重試讀取組織</Button></>}
      {ready && context && !active && route.owner !== user?.userId && <Status role="status">無法開啟此組織。</Status>}
    </PagePanel>
    {user && (active || route.owner === user.userId) && <View
      key={route.owner}
      ref={editor}
      ownerId={route.owner}
      projectId={route.project}
      initialSource={route.source}
      copyFrom={route.copyFrom}
      destinations={owners}
      onCopy={destination => navigate(destination, "new", "", { ownerId: route.owner, projectId: route.project })}
      onSelect={selectProject}
      onBlocked={onBlocked}
      canWrite={canWrite}
    />}
  </div>;
}
