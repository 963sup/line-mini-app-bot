import { useState } from "react";
import { Shell } from "@/shell/Shell";
import { PagePanel } from "@/shared/PagePanel";
import { useSession } from "@/identity/Session";
import { View as Objects } from "@/business/object/View";
import { RepositoryTarget } from "@/business/repository/RepositoryTarget";
import { RepositoryNavigationContext, useRepositoryNavigation } from "@/business/repository/NavigationGuard";

function Workspace() {
  const [target, setTarget] = useState("");
  const guard = useRepositoryNavigation();
  return <RepositoryNavigationContext.Provider value={guard.context}>
    <RepositoryTarget value={target} disabled={guard.blocked} onChange={id => guard.navigate(() => { setTarget(id); guard.settled(); })} />
    {target && <Objects repositoryId={target} documentTool />}
    {guard.dialog}
  </RepositoryNavigationContext.Provider>;
}

function Tool() {
  const { user, epoch } = useSession();
  return <PagePanel heading="從文件辨識物件">{user ? <Workspace key={user.userId + ":" + epoch} /> : <p>請先登入以辨識文件。</p>}</PagePanel>;
}

export default function Page() {
  return <Shell title="從文件辨識物件" workflow="object-recognition"><Tool /></Shell>;
}
