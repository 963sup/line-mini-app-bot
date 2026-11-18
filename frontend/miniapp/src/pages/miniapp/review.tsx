import React from "react";
import { Shell } from "../../shell/Shell";
import { View } from "../../business/attendance/approvals/View";
export default function Page() {
  return (
    <Shell title="授權審核" workflow="review">
      <View />
    </Shell>
  );
}
