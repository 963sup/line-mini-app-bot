import { useRouter } from "next/router";
import { Shell } from "../../shell/Shell";
import { JournalWorkspace } from "../../business/journal/Workspace";

export default function Page() {
  const router = useRouter();
  const repositoryId = typeof router.query.repositoryId === "string" ? router.query.repositoryId : "";
  const composing = router.query.compose === "1" && !router.query.journalId;

  return (
    <Shell title={composing ? "寫日誌" : "動態"} workflow="feed">
      {router.isReady ? <JournalWorkspace dashboard={false} repositoryId={repositoryId} onSelectRepository={composing && !repositoryId ? (id) => { void router.replace("/miniapp/feed?" + new URLSearchParams({repositoryId:id,compose:"1"})); } : undefined} journalId={typeof router.query.journalId === "string" ? router.query.journalId : ""} onCloseComposer={composing ? () => { void router.replace(repositoryId ? "/miniapp/feed?" + new URLSearchParams({repositoryId}) : "/miniapp/feed"); } : undefined} /> : <p role="status">請先登入以查閱群組日誌。</p>}
    </Shell>
  );
}
