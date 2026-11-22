import { useRouter } from "next/router";
import { Shell } from "../../shell/Shell";
import { JournalWorkspace } from "../../business/journal/Workspace";

export default function Page() {
  const router = useRouter();
  const repositoryId = typeof router.query.repositoryId === "string" ? router.query.repositoryId : "";
  const journalId = typeof router.query.journalId === "string" ? router.query.journalId : "";
  const composing = router.query.compose === "1" && !router.query.journalId;

  return (
    <Shell title={composing ? "寫日誌" : "動態"} workflow="feed">
      {router.isReady ? (
        <JournalWorkspace
          dashboard={false}
          repositoryId={repositoryId}
          journalId={journalId}
          onSelectRepository={composing && !repositoryId ? id => {
            void router.replace({ pathname: "/miniapp/feed", query: { repositoryId: id, compose: "1" } });
          } : undefined}
          onCloseComposer={composing ? () => {
            void router.replace({ pathname: "/miniapp/feed", query: repositoryId ? { repositoryId } : {} });
          } : undefined}
        />
      ) : <p role="status">正在開啟日誌……</p>}
    </Shell>
  );
}
