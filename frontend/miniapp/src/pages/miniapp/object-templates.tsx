import { useState } from "react";
import { Button } from "@/shared/ui/button";
import { PagePanel } from "@/shared/PagePanel";
import { Shell } from "@/shell/Shell";
import { Templates } from "@/business/object/Templates";
import { useSession } from "@/identity/Session";
import {
  RepositoryNavigationContext,
  useRepositoryNavigation,
} from "@/business/repository/NavigationGuard";
function Library() {
  const { user, epoch } = useSession();
  return user ? (
    <LibraryContent key={`${user.userId}:${epoch}`} />
  ) : (
    <p>請先登入以使用物件範本。</p>
  );
}
function LibraryContent() {
  const [created, setCreated] = useState<{ repositoryId: string; root: string; count: number }>();
  const guard = useRepositoryNavigation();
  return (
    <RepositoryNavigationContext.Provider value={guard.context}>
      {created ? (
        <PagePanel heading={`已建立 ${created.count} 個物件`}>
          <div className="flex flex-wrap gap-2">
            <Button asChild>
              <a
                href={
                  "/miniapp/repositories?" +
                  new URLSearchParams({ id: created.repositoryId, object: created.root })
                }
              >
                前往儲存庫查看根物件
              </a>
            </Button>
            <Button variant="outline" onClick={() => setCreated(undefined)}>
              返回我的範本
            </Button>
          </div>
        </PagePanel>
      ) : (
        <Templates
          onApplied={(result, repositoryId) =>
            setCreated({ repositoryId, root: result.rootObjectId, count: result.objectIds.length })
          }
        />
      )}
      {guard.dialog}
    </RepositoryNavigationContext.Provider>
  );
}
export default function Page() {
  return (
    <Shell title="物件範本" workflow="object-templates">
      <Library />
    </Shell>
  );
}
