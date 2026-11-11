import React from "react";
import { Shell } from "../../shell/Shell";
import { View } from "../../business/form/journal/View";

export default function Page() {
  return <Shell title="日誌" workflow="journal"><View /></Shell>;
}
