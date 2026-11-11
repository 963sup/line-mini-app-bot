import React from "react";
import { Shell } from "../../shell/Shell";
import { View } from "../../business/attendance/punch/View";
export default function Page() {
  return (
    <Shell title="群組定位打卡" workflow="punch">
      <View />
    </Shell>
  );
}
