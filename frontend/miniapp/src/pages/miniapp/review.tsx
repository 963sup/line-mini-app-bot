import React from "react";
import { Shell } from "../../shell/Shell";
import { View } from "../../features/attendance/review/View";
export default function Page() {
  return (
    <Shell title="授權審核" workflow="review">
      <View />
    </Shell>
  );
}
