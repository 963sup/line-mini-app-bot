import React from "react";
import { Shell } from "@/shell/Shell";
import { Templates } from "@/business/object/Templates";
import { useSession } from "@/identity/Session";
import { RepositoryNavigationContext, useRepositoryNavigation } from "@/business/repository/NavigationGuard";
const ignore = () => {};
function Library() {
  const { user, epoch } = useSession();
  return user ? <LibraryContent key={`${user.userId}:${epoch}`} /> : <p>請先登入以使用物件範本。</p>;
}
function LibraryContent() {
  const guard = useRepositoryNavigation();
  return <RepositoryNavigationContext.Provider value={guard.context}>
    <Templates repositoryId="" standalone initiallyOpen disabled={false} onBlocked={ignore} onApplied={ignore} />
    {guard.dialog}
  </RepositoryNavigationContext.Provider>;
}
export default function Page() {
  return <Shell title="物件範本" workflow="object-templates"><Library /></Shell>;
}
