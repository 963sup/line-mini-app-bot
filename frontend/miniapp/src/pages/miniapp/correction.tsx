import React from "react";
import { Shell } from "../../shell/Shell";
import { View } from "../../business/attendance/corrections/View";
export default function Page() {
  return (
    <Shell title="補卡／時間更正" workflow="correction">
      <View />
    </Shell>
  );
}
