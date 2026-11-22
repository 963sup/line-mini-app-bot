import { useRouter } from "next/router";
import { Shell } from "../../shell/Shell";
import { JournalWorkspace } from "../../business/journal/Workspace";

export default function Page() {
  const router = useRouter();
  const repositoryId = typeof router.query.repositoryId === "string" ? router.query.repositoryId : "";
  return <Shell title="儀表板" workflow="dashboard">{router.isReady && <JournalWorkspace dashboard repositoryId={repositoryId} />}</Shell>;
}
