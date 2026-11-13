import React from "react";
import { Shell } from "../../shell/Shell";
import { Workspace } from "../../business/project/Workspace";

export default function Page() {
  return <Shell title="專案" workflow="projects"><Workspace /></Shell>;
}
