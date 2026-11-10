import React from "react";
import { Shell } from "../../shell/Shell";
import { View } from "../../features/attendance/history/View";
export default function Page() {
  return (
    <Shell title="本人月紀錄" workflow="history">
      <View />
    </Shell>
  );
}
