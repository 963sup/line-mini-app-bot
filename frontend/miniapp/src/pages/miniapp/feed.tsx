import React from "react";
import { useRouter } from "next/router";
import { Shell } from "../../shell/Shell";
import { JournalWorkspace } from "../../business/journal/Workspace";

export default function Page() {
  const router = useRouter();
  const composing = router.query.compose === "1" && !router.query.journalId;

  return (
    <Shell title={composing ? "寫日誌" : "動態"} workflow="feed">
      {router.isReady && <JournalWorkspace dashboard={false} journalId={typeof router.query.journalId === "string" ? router.query.journalId : ""} onCloseComposer={composing ? () => { void router.replace("/miniapp/feed"); } : undefined} />}
    </Shell>
  );
}
