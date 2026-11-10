import React from "react";
import { Shell } from "../../shell/Shell";
import { View } from "../../features/attendance/correction/View";
export default function Page() {
  return (
    <Shell title="更正與申請結果" workflow="correction">
      <View />
    </Shell>
  );
}
