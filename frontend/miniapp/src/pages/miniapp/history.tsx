import React from "react";
import { Shell } from "../../shell/Shell";
import { View } from "../../business/attendance/records/View";
export default function Page() {
  return (
    <Shell title="本人月紀錄" workflow="history">
      <View />
    </Shell>
  );
}
