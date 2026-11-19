import React from "react";
import { Shell } from "../../shell/Shell";
import { JournalWorkspace } from "../../business/journal/Workspace";

export default function Page() {
  return <Shell title="動態" workflow="feed"><JournalWorkspace dashboard={false} /></Shell>;
}
