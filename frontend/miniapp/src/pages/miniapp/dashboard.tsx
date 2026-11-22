import { Shell } from "../../shell/Shell";
import { JournalWorkspace } from "../../business/journal/Workspace";

export default function Page() {
  return <Shell title="儀表板" workflow="dashboard"><JournalWorkspace dashboard={true} /></Shell>;
}
